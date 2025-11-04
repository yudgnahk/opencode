package session

import (
	"context"
	"fmt"
	"time"
)

// AddMessage adds a message to a session
func (m *Manager) AddMessage(ctx context.Context, sessionID string, role Role, content []ContentBlock) (*Message, error) {
	// Lock to prevent race conditions when modifying session
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get session (must check cache first while holding lock)
	session, ok := m.sessions[sessionID]
	if !ok {
		// Temporarily unlock to call Get (which handles locking internally)
		m.mu.Unlock()
		loaded, err := m.Get(ctx, sessionID)
		m.mu.Lock()
		if err != nil {
			return nil, err
		}
		session = loaded
		m.sessions[sessionID] = session
	}

	message := &Message{
		ID:        generateID(),
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Save message: message/{sessionID}/{messageID}
	if err := m.storage.WriteJSON([]string{"message", sessionID, message.ID}, message); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	// Update session
	session.MessageIDs = append(session.MessageIDs, message.ID)
	session.UpdatedAt = time.Now()

	// Save session: session/{projectID}/{sessionID}
	if err := m.storage.WriteJSON([]string{"session", session.ProjectID, session.ID}, session); err != nil {
		return nil, err
	}

	// Emit event
	m.eventBus.Publish(Event{
		Type:      EventMessageAdded,
		SessionID: sessionID,
		Data:      message,
	})

	return message, nil
}

// GetMessage retrieves a single message by ID
func (m *Manager) GetMessage(ctx context.Context, sessionID string, messageID string) (*Message, error) {
	var msg Message
	// Read message: message/{sessionID}/{messageID}
	if err := m.storage.ReadJSON([]string{"message", sessionID, messageID}, &msg); err != nil {
		return nil, fmt.Errorf("message not found: %w", err)
	}
	return &msg, nil
}

// GetMessages retrieves all messages for a session
func (m *Manager) GetMessages(ctx context.Context, sessionID string) ([]*Message, error) {
	session, err := m.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	messages := make([]*Message, 0, len(session.MessageIDs))

	for _, msgID := range session.MessageIDs {
		var msg Message
		// Read message: message/{sessionID}/{messageID}
		if err := m.storage.ReadJSON([]string{"message", sessionID, msgID}, &msg); err != nil {
			continue
		}
		messages = append(messages, &msg)
	}

	return messages, nil
}

// UpdateMessage updates a message (for streaming)
func (m *Manager) UpdateMessage(ctx context.Context, messageID string, content []ContentBlock) error {
	// We need to find the message by searching all sessions
	// This is inefficient but necessary without knowing the sessionID
	allSessions, err := m.List(ctx, "")
	if err != nil {
		return err
	}

	var message Message
	var sessionID string
	found := false

	for _, session := range allSessions {
		for _, msgID := range session.MessageIDs {
			if msgID == messageID {
				if err := m.storage.ReadJSON([]string{"message", session.ID, messageID}, &message); err != nil {
					return err
				}
				sessionID = session.ID
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return fmt.Errorf("message not found: %s", messageID)
	}

	message.Content = content

	// Save message: message/{sessionID}/{messageID}
	if err := m.storage.WriteJSON([]string{"message", sessionID, messageID}, &message); err != nil {
		return err
	}

	// Emit event
	m.eventBus.Publish(Event{
		Type:      EventMessageUpdated,
		SessionID: message.SessionID,
		Data:      &message,
	})

	return nil
}

// DeleteMessage removes a message from a session
func (m *Manager) DeleteMessage(ctx context.Context, messageID string) error {
	// Find the session containing this message
	allSessions, err := m.List(ctx, "")
	if err != nil {
		return err
	}

	var sessionID string
	var session *Session
	found := false

	for _, s := range allSessions {
		for _, msgID := range s.MessageIDs {
			if msgID == messageID {
				session = s
				sessionID = s.ID
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return fmt.Errorf("message not found: %s", messageID)
	}

	// Lock to prevent race conditions when modifying session
	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove from session
	newMessageIDs := make([]string, 0)
	for _, id := range session.MessageIDs {
		if id != messageID {
			newMessageIDs = append(newMessageIDs, id)
		}
	}
	session.MessageIDs = newMessageIDs

	// Save session: session/{projectID}/{sessionID}
	if err := m.storage.WriteJSON([]string{"session", session.ProjectID, session.ID}, session); err != nil {
		return err
	}

	// Delete message: message/{sessionID}/{messageID}
	if err := m.storage.Delete([]string{"message", sessionID, messageID}); err != nil {
		return err
	}

	return nil
}

// Revert removes all messages after a given message ID (inclusive or exclusive based on inclusive flag)
func (m *Manager) Revert(ctx context.Context, sessionID string, messageID string, inclusive bool) error {
	// Lock to prevent race conditions when modifying session
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get session from cache
	session, ok := m.sessions[sessionID]
	if !ok {
		// Load from storage
		m.mu.Unlock()
		loaded, err := m.Get(ctx, sessionID)
		m.mu.Lock()
		if err != nil {
			return err
		}
		session = loaded
		m.sessions[sessionID] = session
	}

	// Find the message index
	messageIndex := -1
	for i, id := range session.MessageIDs {
		if id == messageID {
			messageIndex = i
			break
		}
	}

	if messageIndex == -1 {
		return fmt.Errorf("message not found: %s", messageID)
	}

	// Determine cutoff index
	cutoffIndex := messageIndex
	if !inclusive {
		cutoffIndex = messageIndex + 1
	}

	// Get messages to delete
	messagesToDelete := session.MessageIDs[cutoffIndex:]

	// Delete messages from storage: message/{sessionID}/{messageID}
	for _, id := range messagesToDelete {
		if err := m.storage.Delete([]string{"message", sessionID, id}); err != nil {
			// Continue even if delete fails
			continue
		}
	}

	// Update session with remaining messages
	session.MessageIDs = session.MessageIDs[:cutoffIndex]
	session.UpdatedAt = time.Now()

	// Save session: session/{projectID}/{sessionID}
	if err := m.storage.WriteJSON([]string{"session", session.ProjectID, session.ID}, session); err != nil {
		return err
	}

	// Emit event
	m.eventBus.Publish(Event{
		Type:      EventSessionUpdated,
		SessionID: sessionID,
		Data:      session,
	})

	return nil
}
