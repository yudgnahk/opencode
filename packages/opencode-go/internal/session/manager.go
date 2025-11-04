package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
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

	// Session index for fast listing (projectID -> sessionID -> metadata)
	sessionIndex map[string]map[string]*SessionListItem
	indexMu      sync.RWMutex
}

func NewManager(storage *storage.Storage, providerRegistry *provider.Registry) *Manager {
	m := &Manager{
		storage:          storage,
		providerRegistry: providerRegistry,
		sessions:         make(map[string]*Session),
		eventBus:         NewEventBus(),
		sessionIndex:     make(map[string]map[string]*SessionListItem),
	}

	// Load index from storage on initialization
	m.loadIndex()

	return m
}

// loadIndex loads the session index from storage
func (m *Manager) loadIndex() {
	m.indexMu.Lock()
	defer m.indexMu.Unlock()

	// Try to load existing index
	var index map[string]map[string]*SessionListItem
	if err := m.storage.ReadJSON([]string{"_index", "sessions"}, &index); err == nil && index != nil {
		m.sessionIndex = index
		return
	}

	// If no index exists or loading failed, build it from scratch
	m.rebuildIndex()
}

// copyIndex creates a deep copy of the session index
// Note: caller should hold indexMu lock
func (m *Manager) copyIndex() map[string]map[string]*SessionListItem {
	indexCopy := make(map[string]map[string]*SessionListItem, len(m.sessionIndex))
	for projectID, projectSessions := range m.sessionIndex {
		indexCopy[projectID] = make(map[string]*SessionListItem, len(projectSessions))
		for sessionID, item := range projectSessions {
			itemCopy := *item
			indexCopy[projectID][sessionID] = &itemCopy
		}
	}
	return indexCopy
}

// rebuildIndex scans all session files and rebuilds the index
func (m *Manager) rebuildIndex() {
	// Note: caller must hold indexMu lock
	m.sessionIndex = make(map[string]map[string]*SessionListItem)

	// List all session files
	fileInfos, err := m.storage.ListInfo([]string{"session"})
	if err != nil {
		return
	}

	for _, info := range fileInfos {
		var session Session
		if err := m.storage.ReadJSON(info.Key, &session); err != nil {
			continue
		}

		if m.sessionIndex[session.ProjectID] == nil {
			m.sessionIndex[session.ProjectID] = make(map[string]*SessionListItem)
		}

		// Count messages by listing message files
		messageCount := 0
		if session.ID != "" {
			msgFiles, err := m.storage.ListInfo([]string{"message", session.ID})
			if err == nil {
				messageCount = len(msgFiles)
			}
		}

		m.sessionIndex[session.ProjectID][session.ID] = &SessionListItem{
			ID:           session.ID,
			ProjectID:    session.ProjectID,
			Title:        session.Title,
			Version:      session.Version,
			ParentID:     session.ParentID,
			Created:      session.Time.Created,
			Updated:      session.Time.Updated,
			MessageCount: messageCount,
		}
	}

	// Save index to storage
	m.storage.WriteJSON([]string{"_index", "sessions"}, m.sessionIndex)
}

// updateIndex updates the index for a single session
func (m *Manager) updateIndex(session *Session) {
	m.indexMu.Lock()

	if m.sessionIndex[session.ProjectID] == nil {
		m.sessionIndex[session.ProjectID] = make(map[string]*SessionListItem)
	}

	// Count messages by listing message files
	messageCount := 0
	if session.ID != "" {
		msgFiles, _ := m.storage.ListInfo([]string{"message", session.ID})
		messageCount = len(msgFiles)
	}

	m.sessionIndex[session.ProjectID][session.ID] = &SessionListItem{
		ID:           session.ID,
		ProjectID:    session.ProjectID,
		Title:        session.Title,
		Version:      session.Version,
		ParentID:     session.ParentID,
		Created:      session.Time.Created,
		Updated:      session.Time.Updated,
		MessageCount: messageCount,
	}

	// Make a copy of the index for async persistence
	indexCopy := m.copyIndex()
	m.indexMu.Unlock()

	// Persist index to storage (async to avoid blocking)
	go m.storage.WriteJSON([]string{"_index", "sessions"}, indexCopy)
}

// removeFromIndex removes a session from the index
func (m *Manager) removeFromIndex(projectID, sessionID string) {
	m.indexMu.Lock()

	if m.sessionIndex[projectID] != nil {
		delete(m.sessionIndex[projectID], sessionID)

		// Clean up empty project map
		if len(m.sessionIndex[projectID]) == 0 {
			delete(m.sessionIndex, projectID)
		}
	}

	// Make a copy of the index for async persistence
	indexCopy := m.copyIndex()
	m.indexMu.Unlock()

	// Persist index to storage (async to avoid blocking)
	go m.storage.WriteJSON([]string{"_index", "sessions"}, indexCopy)
}

// EventBus returns the event bus for this manager
func (m *Manager) EventBus() *EventBus {
	return m.eventBus
}

// Create creates a new session
// For backward compatibility, if directory looks like a provider name and version looks like a model,
// they will be stored in metadata
func (m *Manager) Create(ctx context.Context, projectID, directory, version string) (*Session, error) {
	now := time.Now().UnixMilli()
	session := &Session{
		ID:        generateID(),
		ProjectID: projectID,
		Directory: directory,
		Title:     "New Session - " + time.Now().Format(time.RFC3339),
		Version:   version,
		Time: SessionTime{
			Created: now,
			Updated: now,
		},
		Metadata: make(map[string]interface{}),
	}

	// For backward compatibility with tests: if directory/version look like provider/model,
	// store them in metadata and internal fields
	// This handles legacy calls like Create(ctx, "project", "anthropic", "claude-3-opus")
	if !strings.HasPrefix(directory, "/") && version != "" && !strings.Contains(version, ".") {
		// Looks like (provider, model) instead of (directory, version)
		session.Provider = directory
		session.Model = version
		session.Metadata["provider"] = directory
		session.Metadata["model"] = version
		session.Directory = "" // Clear directory since it was actually provider
		session.Version = ""   // Clear version since it was actually model
	}

	// Save to storage with hierarchical path: session/{projectID}/{sessionID}
	if err := m.storage.WriteJSON([]string{"session", projectID, session.ID}, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	// Add to cache
	m.mu.Lock()
	m.sessions[session.ID] = session
	m.mu.Unlock()

	// Update index
	m.updateIndex(session)

	// Emit event
	m.eventBus.Publish(Event{
		Type:      EventSessionCreated,
		SessionID: session.ID,
		Data:      session,
	})

	return session, nil
}

// Get retrieves a session by ID
// Note: This requires knowing the projectID. For now, we search all projects.
func (m *Manager) Get(ctx context.Context, id string) (*Session, error) {
	// Check cache first
	m.mu.RLock()
	session, ok := m.sessions[id]
	m.mu.RUnlock()

	if ok {
		// Restore internal fields if not already done
		m.restoreInternalFields(session)
		return session, nil
	}

	// We need to search across projects since we don't know the projectID
	// List all sessions and find the matching one
	allSessions, err := m.List(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, s := range allSessions {
		if s.ID == id {
			// Add to cache
			m.restoreInternalFields(s)
			m.mu.Lock()
			m.sessions[id] = s
			m.mu.Unlock()
			return s, nil
		}
	}

	return nil, fmt.Errorf("session not found: %s", id)
}

// List returns all sessions for a project using in-memory index for fast O(1) lookup
func (m *Manager) List(ctx context.Context, projectID string) ([]*Session, error) {
	m.indexMu.RLock()
	defer m.indexMu.RUnlock()

	var items []*SessionListItem

	if projectID != "" {
		// Get sessions for specific project
		if projectSessions, ok := m.sessionIndex[projectID]; ok {
			items = make([]*SessionListItem, 0, len(projectSessions))
			for _, item := range projectSessions {
				items = append(items, item)
			}
		}
	} else {
		// Get all sessions across all projects
		totalCount := 0
		for _, projectSessions := range m.sessionIndex {
			totalCount += len(projectSessions)
		}
		items = make([]*SessionListItem, 0, totalCount)
		for _, projectSessions := range m.sessionIndex {
			for _, item := range projectSessions {
				items = append(items, item)
			}
		}
	}

	// Sort by updated time (newest first)
	sort.Slice(items, func(i, j int) bool {
		return items[i].Updated > items[j].Updated
	})

	// Convert SessionListItem to full Session objects
	// We need to return full Session for backward compatibility
	sessions := make([]*Session, 0, len(items))
	for _, item := range items {
		var session Session
		if err := m.storage.ReadJSON([]string{"session", item.ProjectID, item.ID}, &session); err != nil {
			continue
		}
		// Restore internal fields from metadata
		m.restoreInternalFields(&session)
		sessions = append(sessions, &session)
	}

	return sessions, nil
}

// ListLight returns lightweight session metadata from in-memory index (fastest)
// Use this when you don't need the full MessageIDs array
func (m *Manager) ListLight(ctx context.Context, projectID string) ([]*SessionListItem, error) {
	m.indexMu.RLock()
	defer m.indexMu.RUnlock()

	var items []*SessionListItem

	if projectID != "" {
		// Get sessions for specific project
		if projectSessions, ok := m.sessionIndex[projectID]; ok {
			items = make([]*SessionListItem, 0, len(projectSessions))
			for _, item := range projectSessions {
				// Return a copy to avoid race conditions
				itemCopy := *item
				items = append(items, &itemCopy)
			}
		}
	} else {
		// Get all sessions across all projects
		totalCount := 0
		for _, projectSessions := range m.sessionIndex {
			totalCount += len(projectSessions)
		}
		items = make([]*SessionListItem, 0, totalCount)
		for _, projectSessions := range m.sessionIndex {
			for _, item := range projectSessions {
				// Return a copy to avoid race conditions
				itemCopy := *item
				items = append(items, &itemCopy)
			}
		}
	}

	// Sort by updated time (newest first)
	sort.Slice(items, func(i, j int) bool {
		return items[i].Updated > items[j].Updated
	})

	return items, nil
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

	session.Time.Updated = time.Now().UnixMilli()

	// Save with hierarchical path: session/{projectID}/{sessionID}
	if err := m.storage.WriteJSON([]string{"session", session.ProjectID, session.ID}, session); err != nil {
		return err
	}

	// Update index
	m.updateIndex(session)

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
	// Get session to find its projectID
	session, err := m.Get(ctx, id)
	if err != nil {
		return err
	}

	// Remove from cache
	m.mu.Lock()
	delete(m.sessions, id)
	m.mu.Unlock()

	// Remove from index
	m.removeFromIndex(session.ProjectID, id)

	// Remove from storage: session/{projectID}/{sessionID}
	if err := m.storage.Delete([]string{"session", session.ProjectID, id}); err != nil {
		return err
	}

	// Delete all messages: message/{sessionID}/*
	messages, _ := m.GetMessages(ctx, id)
	for _, msg := range messages {
		m.storage.Delete([]string{"message", id, msg.ID})
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

	// Find the index of fromMessageID to determine which messages to copy
	messageIndex := -1
	for i, msgID := range original.MessageIDs {
		if msgID == fromMessageID {
			messageIndex = i
			break
		}
	}

	if messageIndex == -1 {
		return nil, fmt.Errorf("message not found: %s", fromMessageID)
	}

	// Create new session
	now := time.Now().UnixMilli()
	forked := &Session{
		ID:        generateID(),
		ProjectID: original.ProjectID,
		Directory: original.Directory,
		Title:     "Child session - " + time.Now().Format(time.RFC3339),
		Version:   original.Version,
		ParentID:  original.ID,
		Time: SessionTime{
			Created: now,
			Updated: now,
		},
		Metadata:   original.Metadata,
		MessageIDs: []string{}, // Will be populated as we copy messages
	}

	// Copy messages up to and including fromMessageID
	messagesToCopy := original.MessageIDs[:messageIndex+1]
	for _, msgID := range messagesToCopy {
		// Read original message
		var msg Message
		if err := m.storage.ReadJSON([]string{"message", original.ID, msgID}, &msg); err != nil {
			return nil, fmt.Errorf("failed to read message %s: %w", msgID, err)
		}

		// Update message to point to new session
		msg.SessionID = forked.ID

		// Save message under new session: message/{forkedSessionID}/{messageID}
		if err := m.storage.WriteJSON([]string{"message", forked.ID, msgID}, &msg); err != nil {
			return nil, fmt.Errorf("failed to save message %s: %w", msgID, err)
		}

		// Add to forked session's message list
		forked.MessageIDs = append(forked.MessageIDs, msgID)
	}

	// Save with hierarchical path: session/{projectID}/{sessionID}
	if err := m.storage.WriteJSON([]string{"session", forked.ProjectID, forked.ID}, forked); err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.sessions[forked.ID] = forked
	m.mu.Unlock()

	// Update index
	m.updateIndex(forked)

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

// restoreInternalFields populates internal-only fields from metadata
func (m *Manager) restoreInternalFields(session *Session) {
	if session.Metadata != nil {
		if provider, ok := session.Metadata["provider"].(string); ok {
			session.Provider = provider
		}
		if model, ok := session.Metadata["model"].(string); ok {
			session.Model = model
		}
	}
	// Also populate UpdatedAt helper field
	session.UpdatedAt = time.UnixMilli(session.Time.Updated)
}
