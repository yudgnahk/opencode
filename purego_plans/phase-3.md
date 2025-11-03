# Phase 3: Session Management

**Duration**: Weeks 11-16 (6 weeks)  
**Team**: 2 engineers  
**Complexity**: 🟠 High  
**Dependencies**: Phases 1-2 complete

---

## Objectives

1. Implement session CRUD operations
2. Create message data structures and storage
3. Build event streaming system (SSE)
4. Add session state management
5. Implement concurrent session handling
6. Add session history and branching

**Success Criterion**: Can create/manage sessions and store messages (AI integration comes in Phase 4)

---

## Architecture Overview

### Session Lifecycle

```
Create Session → Add Messages → Stream Events → Update State → Fork/Delete
```

### Key Components

- **Session**: Container for conversation
- **Message**: Individual chat messages (user/assistant/system)
- **Event Bus**: Real-time updates via Server-Sent Events
- **State Manager**: Track session state and history

---

## Detailed Tasks

### Week 11-12: Core Session Management

#### Task 11.1: Session Data Model

Create `internal/session/session.go`:

```go
package session

import (
    "time"
)

type Session struct {
    ID          string                 `json:"id"`
    ProjectID   string                 `json:"projectId"`
    Title       string                 `json:"title"`
    Model       string                 `json:"model"`
    Provider    string                 `json:"provider"`
    State       SessionState           `json:"state"`
    Metadata    map[string]interface{} `json:"metadata"`
    CreatedAt   time.Time              `json:"createdAt"`
    UpdatedAt   time.Time              `json:"updatedAt"`
    MessageIDs  []string               `json:"messageIds"`  // Ordered list
    ParentID    string                 `json:"parentId,omitempty"`  // For forks
}

type SessionState string

const (
    StateIdle       SessionState = "idle"
    StateProcessing SessionState = "processing"
    StateError      SessionState = "error"
    StateCompleted  SessionState = "completed"
)

type Message struct {
    ID        string         `json:"id"`
    SessionID string         `json:"sessionId"`
    Role      Role           `json:"role"`
    Content   []ContentBlock `json:"content"`
    ToolCalls []ToolCall     `json:"toolCalls,omitempty"`
    CreatedAt time.Time      `json:"createdAt"`
    Metadata  map[string]interface{} `json:"metadata"`
}

type Role string

const (
    RoleUser      Role = "user"
    RoleAssistant Role = "assistant"
    RoleSystem    Role = "system"
    RoleTool      Role = "tool"
)

type ContentBlock struct {
    Type string      `json:"type"`  // "text", "image", "tool_use", "tool_result"
    Text string      `json:"text,omitempty"`
    Data interface{} `json:"data,omitempty"`
}

type ToolCall struct {
    ID        string      `json:"id"`
    Name      string      `json:"name"`
    Arguments interface{} `json:"arguments"`
    Result    interface{} `json:"result,omitempty"`
}
```

#### Task 11.2: Session Manager

Create `internal/session/manager.go`:

```go
package session

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "sync"
    "time"

    "github.com/opencode/opencode-go/internal/storage"
)

type Manager struct {
    storage *storage.Storage

    // In-memory cache for active sessions
    sessions map[string]*Session
    mu       sync.RWMutex

    // Event bus for real-time updates
    eventBus *EventBus
}

func NewManager(storage *storage.Storage) *Manager {
    return &Manager{
        storage:  storage,
        sessions: make(map[string]*Session),
        eventBus: NewEventBus(),
    }
}

// Create creates a new session
func (m *Manager) Create(ctx context.Context, projectID, provider, model string) (*Session, error) {
    session := &Session{
        ID:        generateID(),
        ProjectID: projectID,
        Title:     "New Session",
        Model:     model,
        Provider:  provider,
        State:     StateIdle,
        Metadata:  make(map[string]interface{}),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
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
        ID:        generateID(),
        ProjectID: original.ProjectID,
        Title:     original.Title + " (forked)",
        Model:     original.Model,
        Provider:  original.Provider,
        State:     StateIdle,
        Metadata:  original.Metadata,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        ParentID:  original.ID,
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

func generateID() string {
    b := make([]byte, 16)
    rand.Read(b)
    return hex.EncodeToString(b)
}
```

---

### Week 13: Message Management

#### Task 13.1: Message Operations

Create `internal/session/messages.go`:

```go
package session

import (
    "context"
    "fmt"
    "time"
)

// AddMessage adds a message to a session
func (m *Manager) AddMessage(ctx context.Context, sessionID string, role Role, content []ContentBlock) (*Message, error) {
    session, err := m.Get(ctx, sessionID)
    if err != nil {
        return nil, err
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

    // Remove from session
    session, err := m.Get(context.Background(), message.SessionID)
    if err != nil {
        return err
    }

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
```

---

### Week 14-15: Event Streaming System

#### Task 14.1: Event Bus

Create `internal/session/eventbus.go`:

```go
package session

import (
    "sync"
    "time"
)

type EventType string

const (
    EventSessionCreated EventType = "session.created"
    EventSessionUpdated EventType = "session.updated"
    EventSessionDeleted EventType = "session.deleted"
    EventMessageAdded   EventType = "message.added"
    EventMessageUpdated EventType = "message.updated"
    EventMessageDeleted EventType = "message.deleted"
    EventToolCall       EventType = "tool.call"
    EventToolResult     EventType = "tool.result"
)

type Event struct {
    Type      EventType   `json:"type"`
    SessionID string      `json:"sessionId"`
    Data      interface{} `json:"data"`
    Timestamp time.Time   `json:"timestamp"`
}

type EventBus struct {
    subscribers map[string][]chan Event  // sessionID -> channels
    mu          sync.RWMutex
}

func NewEventBus() *EventBus {
    return &EventBus{
        subscribers: make(map[string][]chan Event),
    }
}

// Subscribe creates a subscription for session events
func (eb *EventBus) Subscribe(sessionID string) chan Event {
    ch := make(chan Event, 100)  // Buffered channel

    eb.mu.Lock()
    eb.subscribers[sessionID] = append(eb.subscribers[sessionID], ch)
    eb.mu.Unlock()

    return ch
}

// Unsubscribe removes a subscription
func (eb *EventBus) Unsubscribe(sessionID string, ch chan Event) {
    eb.mu.Lock()
    defer eb.mu.Unlock()

    subs := eb.subscribers[sessionID]
    for i, sub := range subs {
        if sub == ch {
            // Remove from slice
            eb.subscribers[sessionID] = append(subs[:i], subs[i+1:]...)
            close(ch)
            break
        }
    }
}

// Publish sends an event to all subscribers
func (eb *EventBus) Publish(event Event) {
    event.Timestamp = time.Now()

    eb.mu.RLock()
    subscribers := eb.subscribers[event.SessionID]
    eb.mu.RUnlock()

    for _, ch := range subscribers {
        select {
        case ch <- event:
            // Sent successfully
        default:
            // Channel full, skip (client too slow)
        }
    }
}

// SubscribeAll subscribes to all events (for monitoring)
func (eb *EventBus) SubscribeAll() chan Event {
    ch := make(chan Event, 100)

    eb.mu.Lock()
    eb.subscribers["*"] = append(eb.subscribers["*"], ch)
    eb.mu.Unlock()

    return ch
}
```

#### Task 14.2: SSE Streaming Handler

Update `internal/server/handlers.go`:

```go
func (s *Server) handleMessagesStream(w http.ResponseWriter, r *http.Request) {
    sessionID := chi.URLParam(r, "id")

    // Set SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "streaming not supported", http.StatusInternalServerError)
        return
    }

    // Subscribe to events
    eventCh := s.sessionManager.eventBus.Subscribe(sessionID)
    defer s.sessionManager.eventBus.Unsubscribe(sessionID, eventCh)

    // Send keepalive every 30 seconds
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case event := <-eventCh:
            // Marshal event to JSON
            data, err := json.Marshal(event)
            if err != nil {
                continue
            }

            // Write SSE format
            fmt.Fprintf(w, "event: %s\n", event.Type)
            fmt.Fprintf(w, "data: %s\n\n", data)
            flusher.Flush()

        case <-ticker.C:
            // Send keepalive
            fmt.Fprintf(w, ": keepalive\n\n")
            flusher.Flush()

        case <-r.Context().Done():
            // Client disconnected
            return
        }
    }
}
```

---

### Week 16: Session History & State

#### Task 16.1: Session History

Create `internal/session/history.go`:

```go
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
        return nil  // Nothing to compact
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
```

---

## Testing

### Unit Tests

- [ ] Test session CRUD operations
- [ ] Test message adding/retrieval
- [ ] Test session forking
- [ ] Test event bus pub/sub
- [ ] Test concurrent session access
- [ ] Test history pagination
- [ ] Test session compaction

### Integration Tests

- [ ] Test full session lifecycle via API
- [ ] Test SSE streaming with multiple clients
- [ ] Test event delivery under load
- [ ] Test session persistence across server restarts

### Load Testing

- [ ] 10+ concurrent sessions
- [ ] 100+ messages per session
- [ ] Multiple SSE connections
- [ ] Memory leak testing (24-hour soak)

---

## Deliverables

- [x] Session data model and manager
- [x] Message storage and retrieval
- [x] Event bus for real-time updates
- [x] SSE streaming endpoint
- [x] Session forking functionality
- [x] History and pagination
- [x] Compaction support
- [x] Comprehensive tests

---

## Success Metrics

- [ ] Can create/update/delete sessions
- [ ] Messages persist correctly
- [ ] SSE streaming works with multiple clients
- [ ] No race conditions (go test -race passes)
- [ ] Handles 10+ concurrent sessions
- [ ] Session forking preserves messages
- [ ] Events delivered in real-time (<100ms)

---

## Phase 3 Checklist

### Data Models

- [ ] Define Session struct
- [ ] Define Message struct
- [ ] Define ContentBlock struct
- [ ] Define SessionState enum
- [ ] Define Role enum

### Session Manager

- [ ] Implement Create
- [ ] Implement Get (with caching)
- [ ] Implement List
- [ ] Implement Update
- [ ] Implement Delete
- [ ] Implement Fork
- [ ] Add mutex for concurrent access

### Message Management

- [ ] Implement AddMessage
- [ ] Implement GetMessages
- [ ] Implement UpdateMessage (for streaming)
- [ ] Implement DeleteMessage
- [ ] Link messages to sessions

### Event System

- [ ] Create EventBus
- [ ] Implement Subscribe/Unsubscribe
- [ ] Implement Publish
- [ ] Add buffered channels
- [ ] Handle slow consumers

### SSE Streaming

- [ ] Implement SSE handler
- [ ] Add keepalive mechanism
- [ ] Handle client disconnects
- [ ] Test with multiple clients

### Storage

- [ ] Store sessions in BoltDB
- [ ] Store messages in BoltDB
- [ ] Implement session-message linking
- [ ] Add indexes for queries

### Advanced Features

- [ ] Session history with pagination
- [ ] Session compaction
- [ ] Session state tracking
- [ ] Metadata support

### Testing

- [ ] Unit tests for manager
- [ ] Unit tests for event bus
- [ ] Integration tests for API
- [ ] Race condition testing
- [ ] Load testing (10+ sessions)
- [ ] Memory leak testing

---

## Next Phase

**Phase 4**: AI Provider Integration (Weeks 17-28)

- Anthropic Claude integration
- OpenAI integration
- Streaming responses
- Tool calling
- Multi-provider support
