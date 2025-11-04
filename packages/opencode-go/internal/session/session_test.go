package session

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/opencode/opencode-go/internal/provider"
	"github.com/opencode/opencode-go/internal/storage"
)

func setupTestStorage(t *testing.T) *storage.Storage {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := storage.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	return store
}

func TestManager_Create(t *testing.T) {
	store := setupTestStorage(t)
	defer store.Close()

	registry := provider.NewRegistry()
	manager := NewManager(store, registry)
	ctx := context.Background()

	session, err := manager.Create(ctx, "test-project", "/test/dir", "1.0.0")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	if session.ID == "" {
		t.Error("session ID should not be empty")
	}

	if session.ProjectID != "test-project" {
		t.Errorf("expected ProjectID 'test-project', got %s", session.ProjectID)
	}

	if session.Directory != "/test/dir" {
		t.Errorf("expected Directory '/test/dir', got %s", session.Directory)
	}

	if session.Version != "1.0.0" {
		t.Errorf("expected Version '1.0.0', got %s", session.Version)
	}
}

func TestManager_Get(t *testing.T) {
	store := setupTestStorage(t)
	defer store.Close()

	registry := provider.NewRegistry()
	manager := NewManager(store, registry)
	ctx := context.Background()

	// Create a session
	created, err := manager.Create(ctx, "test-project", "/test/dir", "1.0.0")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Retrieve the session
	retrieved, err := manager.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to get session: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("expected ID %s, got %s", created.ID, retrieved.ID)
	}

	if retrieved.ProjectID != created.ProjectID {
		t.Errorf("expected ProjectID %s, got %s", created.ProjectID, retrieved.ProjectID)
	}
}

func TestManager_List(t *testing.T) {
	store := setupTestStorage(t)
	defer store.Close()

	registry := provider.NewRegistry()
	manager := NewManager(store, registry)
	ctx := context.Background()

	// Create multiple sessions
	_, err := manager.Create(ctx, "project-1", "/test/dir1", "1.0.0")
	if err != nil {
		t.Fatalf("failed to create session 1: %v", err)
	}

	_, err = manager.Create(ctx, "project-1", "/test/dir2", "1.0.0")
	if err != nil {
		t.Fatalf("failed to create session 2: %v", err)
	}

	_, err = manager.Create(ctx, "project-2", "/test/dir3", "1.0.0")
	if err != nil {
		t.Fatalf("failed to create session 3: %v", err)
	}

	// List all sessions
	allSessions, err := manager.List(ctx, "")
	if err != nil {
		t.Fatalf("failed to list all sessions: %v", err)
	}

	if len(allSessions) != 3 {
		t.Errorf("expected 3 sessions, got %d", len(allSessions))
	}

	// List sessions for project-1
	project1Sessions, err := manager.List(ctx, "project-1")
	if err != nil {
		t.Fatalf("failed to list project-1 sessions: %v", err)
	}

	if len(project1Sessions) != 2 {
		t.Errorf("expected 2 sessions for project-1, got %d", len(project1Sessions))
	}
}

func TestManager_Update(t *testing.T) {
	store := setupTestStorage(t)
	defer store.Close()

	registry := provider.NewRegistry()
	manager := NewManager(store, registry)
	ctx := context.Background()

	// Create a session
	session, err := manager.Create(ctx, "test-project", "/test/dir", "1.0.0")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Update the session
	updates := map[string]interface{}{
		"title": "Updated Title",
		"state": StateProcessing,
	}

	err = manager.Update(ctx, session.ID, updates)
	if err != nil {
		t.Fatalf("failed to update session: %v", err)
	}

	// Retrieve the updated session
	updated, err := manager.Get(ctx, session.ID)
	if err != nil {
		t.Fatalf("failed to get updated session: %v", err)
	}

	if updated.Title != "Updated Title" {
		t.Errorf("expected Title 'Updated Title', got %s", updated.Title)
	}
}

func TestManager_Delete(t *testing.T) {
	store := setupTestStorage(t)
	defer store.Close()

	registry := provider.NewRegistry()
	manager := NewManager(store, registry)
	ctx := context.Background()

	// Create a session
	session, err := manager.Create(ctx, "test-project", "/test/dir", "1.0.0")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Delete the session
	err = manager.Delete(ctx, session.ID)
	if err != nil {
		t.Fatalf("failed to delete session: %v", err)
	}

	// Try to retrieve the deleted session
	_, err = manager.Get(ctx, session.ID)
	if err == nil {
		t.Error("expected error when getting deleted session, got nil")
	}
}

func TestManager_AddMessage(t *testing.T) {
	store := setupTestStorage(t)
	defer store.Close()

	registry := provider.NewRegistry()
	manager := NewManager(store, registry)
	ctx := context.Background()

	// Create a session
	session, err := manager.Create(ctx, "test-project", "/test/dir", "1.0.0")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Add a message
	content := []ContentBlock{
		{Type: "text", Text: "Hello, world!"},
	}

	message, err := manager.AddMessage(ctx, session.ID, RoleUser, content)
	if err != nil {
		t.Fatalf("failed to add message: %v", err)
	}

	if message.ID == "" {
		t.Error("message ID should not be empty")
	}

	if message.SessionID != session.ID {
		t.Errorf("expected SessionID %s, got %s", session.ID, message.SessionID)
	}

	if message.Role != RoleUser {
		t.Errorf("expected Role 'user', got %s", message.Role)
	}

	if len(message.Content) != 1 {
		t.Errorf("expected 1 content block, got %d", len(message.Content))
	}
}

func TestManager_GetMessages(t *testing.T) {
	store := setupTestStorage(t)
	defer store.Close()

	registry := provider.NewRegistry()
	manager := NewManager(store, registry)
	ctx := context.Background()

	// Create a session
	session, err := manager.Create(ctx, "test-project", "/test/dir", "1.0.0")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Add multiple messages
	content1 := []ContentBlock{{Type: "text", Text: "Message 1"}}
	_, err = manager.AddMessage(ctx, session.ID, RoleUser, content1)
	if err != nil {
		t.Fatalf("failed to add message 1: %v", err)
	}

	content2 := []ContentBlock{{Type: "text", Text: "Message 2"}}
	_, err = manager.AddMessage(ctx, session.ID, RoleAssistant, content2)
	if err != nil {
		t.Fatalf("failed to add message 2: %v", err)
	}

	// Retrieve messages
	messages, err := manager.GetMessages(ctx, session.ID)
	if err != nil {
		t.Fatalf("failed to get messages: %v", err)
	}

	if len(messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(messages))
	}

	if messages[0].Role != RoleUser {
		t.Errorf("expected first message role 'user', got %s", messages[0].Role)
	}

	if messages[1].Role != RoleAssistant {
		t.Errorf("expected second message role 'assistant', got %s", messages[1].Role)
	}
}

func TestManager_Fork(t *testing.T) {
	store := setupTestStorage(t)
	defer store.Close()

	registry := provider.NewRegistry()
	manager := NewManager(store, registry)
	ctx := context.Background()

	// Create a session with messages
	session, err := manager.Create(ctx, "test-project", "/test/dir", "1.0.0")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Add messages
	content1 := []ContentBlock{{Type: "text", Text: "Message 1"}}
	msg1, err := manager.AddMessage(ctx, session.ID, RoleUser, content1)
	if err != nil {
		t.Fatalf("failed to add message 1: %v", err)
	}

	content2 := []ContentBlock{{Type: "text", Text: "Message 2"}}
	_, err = manager.AddMessage(ctx, session.ID, RoleAssistant, content2)
	if err != nil {
		t.Fatalf("failed to add message 2: %v", err)
	}

	// Fork at message 1
	forked, err := manager.Fork(ctx, session.ID, msg1.ID)
	if err != nil {
		t.Fatalf("failed to fork session: %v", err)
	}

	if forked.ID == session.ID {
		t.Error("forked session should have different ID")
	}

	if forked.ParentID != session.ID {
		t.Errorf("expected ParentID %s, got %s", session.ID, forked.ParentID)
	}

	// Verify forked session has correct number of messages
	forkedMessages, err := manager.GetMessages(ctx, forked.ID)
	if err != nil {
		t.Fatalf("failed to get forked messages: %v", err)
	}
	if len(forkedMessages) != 1 {
		t.Errorf("expected forked session to have 1 message, got %d", len(forkedMessages))
	}
}

func TestEventBus_PubSub(t *testing.T) {
	bus := NewEventBus()

	sessionID := "test-session"
	ch := bus.Subscribe(sessionID)

	event := Event{
		Type:      EventSessionCreated,
		SessionID: sessionID,
		Data:      "test data",
	}

	bus.Publish(event)

	select {
	case received := <-ch:
		if received.Type != EventSessionCreated {
			t.Errorf("expected event type %s, got %s", EventSessionCreated, received.Type)
		}
		if received.SessionID != sessionID {
			t.Errorf("expected session ID %s, got %s", sessionID, received.SessionID)
		}
	default:
		t.Error("expected to receive event, got none")
	}

	bus.Unsubscribe(sessionID, ch)
}

func TestManager_GetHistory(t *testing.T) {
	store := setupTestStorage(t)
	defer store.Close()

	registry := provider.NewRegistry()
	manager := NewManager(store, registry)
	ctx := context.Background()

	// Create a session with messages
	session, err := manager.Create(ctx, "test-project", "/test/dir", "1.0.0")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Add 10 messages
	for i := 0; i < 10; i++ {
		content := []ContentBlock{{Type: "text", Text: "Message"}}
		_, err = manager.AddMessage(ctx, session.ID, RoleUser, content)
		if err != nil {
			t.Fatalf("failed to add message: %v", err)
		}
	}

	// Get history with pagination
	history, err := manager.GetHistory(ctx, session.ID, 5, 0)
	if err != nil {
		t.Fatalf("failed to get history: %v", err)
	}

	if history.Total != 10 {
		t.Errorf("expected total 10, got %d", history.Total)
	}

	if len(history.Messages) != 5 {
		t.Errorf("expected 5 messages, got %d", len(history.Messages))
	}

	if history.Limit != 5 {
		t.Errorf("expected limit 5, got %d", history.Limit)
	}

	if history.Offset != 0 {
		t.Errorf("expected offset 0, got %d", history.Offset)
	}
}

func TestManager_Revert(t *testing.T) {
	store := setupTestStorage(t)
	defer store.Close()

	registry := provider.NewRegistry()
	manager := NewManager(store, registry)
	ctx := context.Background()

	// Create a session with messages
	session, err := manager.Create(ctx, "test-project", "/test/dir", "1.0.0")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Add 5 messages
	messages := make([]*Message, 5)
	for i := 0; i < 5; i++ {
		content := []ContentBlock{{Type: "text", Text: fmt.Sprintf("Message %d", i+1)}}
		role := RoleUser
		if i%2 == 1 {
			role = RoleAssistant
		}
		msg, err := manager.AddMessage(ctx, session.ID, role, content)
		if err != nil {
			t.Fatalf("failed to add message %d: %v", i+1, err)
		}
		messages[i] = msg
	}

	t.Run("revert exclusive (keep message at index)", func(t *testing.T) {
		// Revert to message 2 (index 2), exclusive - keeps messages 0, 1, 2
		err := manager.Revert(ctx, session.ID, messages[2].ID, false)
		if err != nil {
			t.Fatalf("failed to revert: %v", err)
		}

		// Verify session has only 3 messages
		sessionMessages, err := manager.GetMessages(ctx, session.ID)
		if err != nil {
			t.Fatalf("failed to get messages: %v", err)
		}

		if len(sessionMessages) != 3 {
			t.Errorf("expected 3 messages after revert, got %d", len(sessionMessages))
		}

		// Verify we kept the right messages
		if sessionMessages[0].ID != messages[0].ID {
			t.Error("expected message 0 to remain")
		}
		if sessionMessages[1].ID != messages[1].ID {
			t.Error("expected message 1 to remain")
		}
		if sessionMessages[2].ID != messages[2].ID {
			t.Error("expected message 2 to remain")
		}
	})
}

func TestManager_RevertInclusive(t *testing.T) {
	store := setupTestStorage(t)
	defer store.Close()

	registry := provider.NewRegistry()
	manager := NewManager(store, registry)
	ctx := context.Background()

	// Create a session with messages
	session, err := manager.Create(ctx, "test-project", "/test/dir", "1.0.0")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Add 5 messages
	messages := make([]*Message, 5)
	for i := 0; i < 5; i++ {
		content := []ContentBlock{{Type: "text", Text: fmt.Sprintf("Message %d", i+1)}}
		role := RoleUser
		if i%2 == 1 {
			role = RoleAssistant
		}
		msg, err := manager.AddMessage(ctx, session.ID, role, content)
		if err != nil {
			t.Fatalf("failed to add message %d: %v", i+1, err)
		}
		messages[i] = msg
	}

	// Revert to message 2 (index 2), inclusive - keeps only messages 0, 1
	err = manager.Revert(ctx, session.ID, messages[2].ID, true)
	if err != nil {
		t.Fatalf("failed to revert: %v", err)
	}

	// Verify session has only 2 messages
	sessionMessages, err := manager.GetMessages(ctx, session.ID)
	if err != nil {
		t.Fatalf("failed to get messages: %v", err)
	}

	if len(sessionMessages) != 2 {
		t.Errorf("expected 2 messages after inclusive revert, got %d", len(sessionMessages))
	}

	// Verify we kept the right messages
	if sessionMessages[0].ID != messages[0].ID {
		t.Error("expected message 0 to remain")
	}
	if sessionMessages[1].ID != messages[1].ID {
		t.Error("expected message 1 to remain")
	}
}

func TestManager_RevertNotFound(t *testing.T) {
	store := setupTestStorage(t)
	defer store.Close()

	registry := provider.NewRegistry()
	manager := NewManager(store, registry)
	ctx := context.Background()

	// Create a session
	session, err := manager.Create(ctx, "test-project", "/test/dir", "1.0.0")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Try to revert to non-existent message
	err = manager.Revert(ctx, session.ID, "nonexistent-message-id", false)
	if err == nil {
		t.Error("expected error when reverting to non-existent message")
	}
}

func TestManager_Branch(t *testing.T) {
	store := setupTestStorage(t)
	defer store.Close()

	registry := provider.NewRegistry()
	manager := NewManager(store, registry)
	ctx := context.Background()

	// Create a session with messages
	session, err := manager.Create(ctx, "test-project", "/test/dir", "1.0.0")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Add messages
	content1 := []ContentBlock{{Type: "text", Text: "Message 1"}}
	msg1, err := manager.AddMessage(ctx, session.ID, RoleUser, content1)
	if err != nil {
		t.Fatalf("failed to add message 1: %v", err)
	}

	content2 := []ContentBlock{{Type: "text", Text: "Message 2"}}
	_, err = manager.AddMessage(ctx, session.ID, RoleAssistant, content2)
	if err != nil {
		t.Fatalf("failed to add message 2: %v", err)
	}

	// Branch at message 1
	branched, err := manager.Branch(ctx, session.ID, msg1.ID)
	if err != nil {
		t.Fatalf("failed to branch session: %v", err)
	}

	// Verify branch has same behavior as Fork
	if branched.ID == session.ID {
		t.Error("branched session should have different ID")
	}

	if branched.ParentID != session.ID {
		t.Errorf("expected ParentID %s, got %s", session.ID, branched.ParentID)
	}

	// Verify branched session has correct number of messages
	branchedMessages, err := manager.GetMessages(ctx, branched.ID)
	if err != nil {
		t.Fatalf("failed to get branched messages: %v", err)
	}
	if len(branchedMessages) != 1 {
		t.Errorf("expected branched session to have 1 message, got %d", len(branchedMessages))
	}
}
