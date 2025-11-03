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
	subscribers map[string][]chan Event // sessionID -> channels
	mu          sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan Event),
	}
}

// Subscribe creates a subscription for session events
func (eb *EventBus) Subscribe(sessionID string) chan Event {
	ch := make(chan Event, 100) // Buffered channel

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
