package session

import (
	"context"
)

// GetHistory returns session history with pagination
func (m *Manager) GetHistory(ctx context.Context, sessionID string, limit, offset int) (*History, error) {
	messages, err := m.GetMessages(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	total := len(messages)

	// Apply pagination
	start := offset
	end := offset + limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	return &History{
		SessionID: sessionID,
		Messages:  messages[start:end],
		Total:     total,
		Offset:    offset,
		Limit:     limit,
	}, nil
}

type History struct {
	SessionID string     `json:"sessionId"`
	Messages  []*Message `json:"messages"`
	Total     int        `json:"total"`
	Offset    int        `json:"offset"`
	Limit     int        `json:"limit"`
}

// Compact removes old messages to save space
func (m *Manager) Compact(ctx context.Context, sessionID string, keepLast int) error {
	messages, err := m.GetMessages(ctx, sessionID)
	if err != nil {
		return err
	}

	if len(messages) <= keepLast {
		return nil // Nothing to compact
	}

	// Delete old messages
	toDelete := messages[:len(messages)-keepLast]
	for _, msg := range toDelete {
		m.storage.Delete("messages", msg.ID)
	}

	// Update session
	session, err := m.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	session.MessageIDs = session.MessageIDs[len(toDelete):]

	return m.storage.SetJSON("sessions", session.ID, session)
}
