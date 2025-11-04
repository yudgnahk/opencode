package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/opencode/opencode-go/internal/provider"
	"github.com/opencode/opencode-go/internal/storage"
)

type Manager struct {
	storage          *storage.Storage
	providerRegistry *provider.Registry

	// In-memory cache for active sessions
	sessions map[string]*Session
	mu       sync.RWMutex

	// Event bus for real-time updates
	eventBus *EventBus
}

func NewManager(storage *storage.Storage, providerRegistry *provider.Registry) *Manager {
	return &Manager{
		storage:          storage,
		providerRegistry: providerRegistry,
		sessions:         make(map[string]*Session),
		eventBus:         NewEventBus(),
	}
}

// EventBus returns the event bus for this manager
func (m *Manager) EventBus() *EventBus {
	return m.eventBus
}

// Create creates a new session
func (m *Manager) Create(ctx context.Context, projectID, provider, model string) (*Session, error) {
	session := &Session{
		ID:         generateID(),
		ProjectID:  projectID,
		Title:      "New Session",
		Model:      model,
		Provider:   provider,
		State:      StateIdle,
		Metadata:   make(map[string]interface{}),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		MessageIDs: []string{},
	}

	// Save to storage
	if err := m.storage.SetJSON("sessions", session.ID, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	// Add to cache
	m.mu.Lock()
	m.sessions[session.ID] = session
	m.mu.Unlock()

	// Emit event
	m.eventBus.Publish(Event{
		Type:      EventSessionCreated,
		SessionID: session.ID,
		Data:      session,
	})

	return session, nil
}

// Get retrieves a session by ID
func (m *Manager) Get(ctx context.Context, id string) (*Session, error) {
	// Check cache first
	m.mu.RLock()
	session, ok := m.sessions[id]
	m.mu.RUnlock()

	if ok {
		return session, nil
	}

	// Load from storage
	var loaded Session
	if err := m.storage.GetJSON("sessions", id, &loaded); err != nil {
		return nil, err
	}

	if loaded.ID == "" {
		return nil, fmt.Errorf("session not found: %s", id)
	}

	// Add to cache
	m.mu.Lock()
	m.sessions[id] = &loaded
	m.mu.Unlock()

	return &loaded, nil
}

// List returns all sessions for a project
func (m *Manager) List(ctx context.Context, projectID string) ([]*Session, error) {
	keys, err := m.storage.List("sessions")
	if err != nil {
		return nil, err
	}

	sessions := make([]*Session, 0)
	for _, key := range keys {
		var session Session
		if err := m.storage.GetJSON("sessions", key, &session); err != nil {
			continue
		}

		if projectID == "" || session.ProjectID == projectID {
			sessions = append(sessions, &session)
		}
	}

	// Sort by updated time (newest first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt)
	})

	return sessions, nil
}

// Update updates session fields
func (m *Manager) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	session, err := m.Get(ctx, id)
	if err != nil {
		return err
	}

	// Apply updates
	if title, ok := updates["title"].(string); ok {
		session.Title = title
	}
	if state, ok := updates["state"].(SessionState); ok {
		session.State = state
	}

	session.UpdatedAt = time.Now()

	// Save
	if err := m.storage.SetJSON("sessions", session.ID, session); err != nil {
		return err
	}

	// Emit event
	m.eventBus.Publish(Event{
		Type:      EventSessionUpdated,
		SessionID: session.ID,
		Data:      session,
	})

	return nil
}

// Delete deletes a session
func (m *Manager) Delete(ctx context.Context, id string) error {
	// Remove from cache
	m.mu.Lock()
	delete(m.sessions, id)
	m.mu.Unlock()

	// Remove from storage
	if err := m.storage.Delete("sessions", id); err != nil {
		return err
	}

	// Delete all messages
	messages, _ := m.GetMessages(ctx, id)
	for _, msg := range messages {
		m.storage.Delete("messages", msg.ID)
	}

	// Emit event
	m.eventBus.Publish(Event{
		Type:      EventSessionDeleted,
		SessionID: id,
	})

	return nil
}

// Fork creates a copy of a session
func (m *Manager) Fork(ctx context.Context, id string, fromMessageID string) (*Session, error) {
	original, err := m.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Create new session
	forked := &Session{
		ID:         generateID(),
		ProjectID:  original.ProjectID,
		Title:      original.Title + " (forked)",
		Model:      original.Model,
		Provider:   original.Provider,
		State:      StateIdle,
		Metadata:   original.Metadata,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		ParentID:   original.ID,
		MessageIDs: []string{},
	}

	// Copy messages up to fromMessageID
	messages, err := m.GetMessages(ctx, id)
	if err != nil {
		return nil, err
	}

	for _, msg := range messages {
		forked.MessageIDs = append(forked.MessageIDs, msg.ID)
		if msg.ID == fromMessageID {
			break
		}
	}

	// Save
	if err := m.storage.SetJSON("sessions", forked.ID, forked); err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.sessions[forked.ID] = forked
	m.mu.Unlock()

	return forked, nil
}

// Branch is an alias for Fork to match TypeScript backend API
func (m *Manager) Branch(ctx context.Context, id string, fromMessageID string) (*Session, error) {
	return m.Fork(ctx, id, fromMessageID)
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
