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
		// Load from storage
		var loaded Session
		if err := m.storage.GetJSON("sessions", sessionID, &loaded); err != nil {
			return nil, err
		}
		if loaded.ID == "" {
			return nil, fmt.Errorf("session not found: %s", sessionID)
		}
		session = &loaded
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

	// Save message
	if err := m.storage.SetJSON("messages", message.ID, message); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	// Update session
	session.MessageIDs = append(session.MessageIDs, message.ID)
	session.UpdatedAt = time.Now()

	if err := m.storage.SetJSON("sessions", session.ID, session); err != nil {
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

// GetMessages retrieves all messages for a session
func (m *Manager) GetMessages(ctx context.Context, sessionID string) ([]*Message, error) {
	session, err := m.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	messages := make([]*Message, 0, len(session.MessageIDs))

	for _, msgID := range session.MessageIDs {
		var msg Message
		if err := m.storage.GetJSON("messages", msgID, &msg); err != nil {
			continue
		}
		messages = append(messages, &msg)
	}

	return messages, nil
}

// UpdateMessage updates a message (for streaming)
func (m *Manager) UpdateMessage(ctx context.Context, messageID string, content []ContentBlock) error {
	var message Message
	if err := m.storage.GetJSON("messages", messageID, &message); err != nil {
		return err
	}

	message.Content = content

	if err := m.storage.SetJSON("messages", messageID, &message); err != nil {
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
	var message Message
	if err := m.storage.GetJSON("messages", messageID, &message); err != nil {
		return err
	}

	// Lock to prevent race conditions when modifying session
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get session from cache
	session, ok := m.sessions[message.SessionID]
	if !ok {
		// Load from storage
		var loaded Session
		if err := m.storage.GetJSON("sessions", message.SessionID, &loaded); err != nil {
			return err
		}
		if loaded.ID == "" {
			return fmt.Errorf("session not found: %s", message.SessionID)
		}
		session = &loaded
		m.sessions[message.SessionID] = session
	}

	// Remove from session
	newMessageIDs := make([]string, 0)
	for _, id := range session.MessageIDs {
		if id != messageID {
			newMessageIDs = append(newMessageIDs, id)
		}
	}
	session.MessageIDs = newMessageIDs

	if err := m.storage.SetJSON("sessions", session.ID, session); err != nil {
		return err
	}

	// Delete message
	if err := m.storage.Delete("messages", messageID); err != nil {
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
		var loaded Session
		if err := m.storage.GetJSON("sessions", sessionID, &loaded); err != nil {
			return err
		}
		if loaded.ID == "" {
			return fmt.Errorf("session not found: %s", sessionID)
		}
		session = &loaded
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

	// Delete messages from storage
	for _, id := range messagesToDelete {
		if err := m.storage.Delete("messages", id); err != nil {
			// Continue even if delete fails
			continue
		}
	}

	// Update session with remaining messages
	session.MessageIDs = session.MessageIDs[:cutoffIndex]
	session.UpdatedAt = time.Now()

	if err := m.storage.SetJSON("sessions", session.ID, session); err != nil {
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
