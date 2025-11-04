package session

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/opencode/opencode-go/internal/provider"
	"github.com/opencode/opencode-go/internal/storage"
)

func setupTestStorageTodo(t *testing.T) *storage.Storage {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test-todo.db")

	store, err := storage.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	return store
}

func TestManager_GetTodos_EmptyList(t *testing.T) {
	store := setupTestStorageTodo(t)
	defer store.Close()

	registry := provider.NewRegistry()
	manager := NewManager(store, registry)

	ctx := context.Background()
	sessionID := "test-session"

	todos, err := manager.GetTodos(ctx, sessionID)
	if err != nil {
		t.Fatalf("GetTodos failed: %v", err)
	}

	if todos.SessionID != sessionID {
		t.Errorf("Expected SessionID %s, got %s", sessionID, todos.SessionID)
	}

	if len(todos.Todos) != 0 {
		t.Errorf("Expected empty todo list, got %d todos", len(todos.Todos))
	}
}

func TestManager_AddTodo(t *testing.T) {
	store := setupTestStorageTodo(t)
	defer store.Close()

	registry := provider.NewRegistry()
	manager := NewManager(store, registry)

	ctx := context.Background()
	sessionID := "test-session"

	// Add first todo
	err := manager.AddTodo(ctx, sessionID, "Implement feature X", "high")
	if err != nil {
		t.Fatalf("AddTodo failed: %v", err)
	}

	// Verify todo was added
	todos, err := manager.GetTodos(ctx, sessionID)
	if err != nil {
		t.Fatalf("GetTodos failed: %v", err)
	}

	if len(todos.Todos) != 1 {
		t.Fatalf("Expected 1 todo, got %d", len(todos.Todos))
	}

	todo := todos.Todos[0]
	if todo.Content != "Implement feature X" {
		t.Errorf("Expected content 'Implement feature X', got '%s'", todo.Content)
	}

	if todo.Priority != "high" {
		t.Errorf("Expected priority 'high', got '%s'", todo.Priority)
	}

	if todo.Status != TodoPending {
		t.Errorf("Expected status 'pending', got '%s'", todo.Status)
	}

	if todo.ID == "" {
		t.Error("Expected non-empty todo ID")
	}
}

func TestManager_AddTodo_Multiple(t *testing.T) {
	store := setupTestStorageTodo(t)
	defer store.Close()

	registry := provider.NewRegistry()
	manager := NewManager(store, registry)

	ctx := context.Background()
	sessionID := "test-session"

	// Add multiple todos
	manager.AddTodo(ctx, sessionID, "Task 1", "high")
	manager.AddTodo(ctx, sessionID, "Task 2", "medium")
	manager.AddTodo(ctx, sessionID, "Task 3", "low")

	todos, err := manager.GetTodos(ctx, sessionID)
	if err != nil {
		t.Fatalf("GetTodos failed: %v", err)
	}

	if len(todos.Todos) != 3 {
		t.Fatalf("Expected 3 todos, got %d", len(todos.Todos))
	}

	// Verify each todo
	expected := []struct {
		content  string
		priority string
	}{
		{"Task 1", "high"},
		{"Task 2", "medium"},
		{"Task 3", "low"},
	}

	for i, exp := range expected {
		todo := todos.Todos[i]
		if todo.Content != exp.content {
			t.Errorf("Todo %d: expected content '%s', got '%s'", i, exp.content, todo.Content)
		}
		if todo.Priority != exp.priority {
			t.Errorf("Todo %d: expected priority '%s', got '%s'", i, exp.priority, todo.Priority)
		}
		if todo.Status != TodoPending {
			t.Errorf("Todo %d: expected status 'pending', got '%s'", i, todo.Status)
		}
	}
}

func TestManager_UpdateTodoStatus(t *testing.T) {
	store := setupTestStorageTodo(t)
	defer store.Close()

	registry := provider.NewRegistry()
	manager := NewManager(store, registry)

	ctx := context.Background()
	sessionID := "test-session"

	// Add a todo
	manager.AddTodo(ctx, sessionID, "Implement feature", "high")

	// Get the todo ID
	todos, _ := manager.GetTodos(ctx, sessionID)
	todoID := todos.Todos[0].ID

	// Update status to in_progress
	err := manager.UpdateTodoStatus(ctx, sessionID, todoID, TodoInProgress)
	if err != nil {
		t.Fatalf("UpdateTodoStatus failed: %v", err)
	}

	// Verify status updated
	todos, _ = manager.GetTodos(ctx, sessionID)
	if todos.Todos[0].Status != TodoInProgress {
		t.Errorf("Expected status 'in_progress', got '%s'", todos.Todos[0].Status)
	}

	// Update to completed
	err = manager.UpdateTodoStatus(ctx, sessionID, todoID, TodoCompleted)
	if err != nil {
		t.Fatalf("UpdateTodoStatus failed: %v", err)
	}

	todos, _ = manager.GetTodos(ctx, sessionID)
	if todos.Todos[0].Status != TodoCompleted {
		t.Errorf("Expected status 'completed', got '%s'", todos.Todos[0].Status)
	}
}

func TestManager_UpdateTodoStatus_NotFound(t *testing.T) {
	store := setupTestStorageTodo(t)
	defer store.Close()

	registry := provider.NewRegistry()
	manager := NewManager(store, registry)

	ctx := context.Background()
	sessionID := "test-session"

	// Try to update non-existent todo
	err := manager.UpdateTodoStatus(ctx, sessionID, "non-existent-id", TodoCompleted)
	if err == nil {
		t.Fatal("Expected error for non-existent todo, got nil")
	}

	expectedErr := "todo not found: non-existent-id"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestManager_UpdateTodos(t *testing.T) {
	store := setupTestStorageTodo(t)
	defer store.Close()

	registry := provider.NewRegistry()
	manager := NewManager(store, registry)

	ctx := context.Background()
	sessionID := "test-session"

	// Create custom todo list
	todos := []Todo{
		{
			ID:       "todo-1",
			Content:  "Task 1",
			Status:   TodoPending,
			Priority: "high",
		},
		{
			ID:       "todo-2",
			Content:  "Task 2",
			Status:   TodoInProgress,
			Priority: "medium",
		},
	}

	err := manager.UpdateTodos(ctx, sessionID, todos)
	if err != nil {
		t.Fatalf("UpdateTodos failed: %v", err)
	}

	// Verify todos were saved
	retrieved, err := manager.GetTodos(ctx, sessionID)
	if err != nil {
		t.Fatalf("GetTodos failed: %v", err)
	}

	if len(retrieved.Todos) != 2 {
		t.Fatalf("Expected 2 todos, got %d", len(retrieved.Todos))
	}

	// Verify first todo
	todo1 := retrieved.Todos[0]
	if todo1.ID != "todo-1" || todo1.Content != "Task 1" || todo1.Status != TodoPending {
		t.Error("First todo does not match expected values")
	}

	// Verify second todo
	todo2 := retrieved.Todos[1]
	if todo2.ID != "todo-2" || todo2.Content != "Task 2" || todo2.Status != TodoInProgress {
		t.Error("Second todo does not match expected values")
	}
}

func TestManager_UpdateTodos_EventEmission(t *testing.T) {
	store := setupTestStorageTodo(t)
	defer store.Close()

	registry := provider.NewRegistry()
	manager := NewManager(store, registry)

	ctx := context.Background()
	sessionID := "test-session"

	// Subscribe to events
	eventChan := manager.EventBus().Subscribe(sessionID)

	// Update todos
	todos := []Todo{
		{ID: "1", Content: "Test", Status: TodoPending, Priority: "high"},
	}
	manager.UpdateTodos(ctx, sessionID, todos)

	// Wait for event
	select {
	case event := <-eventChan:
		if event.Type != "todos.updated" {
			t.Errorf("Expected event type 'todos.updated', got '%s'", event.Type)
		}
		if event.SessionID != sessionID {
			t.Errorf("Expected session ID '%s', got '%s'", sessionID, event.SessionID)
		}

		// Verify event data
		todoList, ok := event.Data.(*TodoList)
		if !ok {
			t.Fatal("Event data is not TodoList")
		}
		if len(todoList.Todos) != 1 {
			t.Errorf("Expected 1 todo in event data, got %d", len(todoList.Todos))
		}
	default:
		t.Fatal("No event received")
	}
}

func TestManager_UpdateTodoStatus_AllStatuses(t *testing.T) {
	store := setupTestStorageTodo(t)
	defer store.Close()

	registry := provider.NewRegistry()
	manager := NewManager(store, registry)

	ctx := context.Background()
	sessionID := "test-session"

	// Add a todo
	manager.AddTodo(ctx, sessionID, "Test all statuses", "medium")

	todos, _ := manager.GetTodos(ctx, sessionID)
	todoID := todos.Todos[0].ID

	// Test all status transitions
	statuses := []TodoStatus{
		TodoInProgress,
		TodoCompleted,
		TodoCancelled,
		TodoPending,
	}

	for _, status := range statuses {
		err := manager.UpdateTodoStatus(ctx, sessionID, todoID, status)
		if err != nil {
			t.Fatalf("Failed to update to status '%s': %v", status, err)
		}

		todos, _ = manager.GetTodos(ctx, sessionID)
		if todos.Todos[0].Status != status {
			t.Errorf("Expected status '%s', got '%s'", status, todos.Todos[0].Status)
		}
	}
}

func TestManager_UpdateTodos_PreservesOrder(t *testing.T) {
	store := setupTestStorageTodo(t)
	defer store.Close()

	registry := provider.NewRegistry()
	manager := NewManager(store, registry)

	ctx := context.Background()
	sessionID := "test-session"

	// Add todos
	manager.AddTodo(ctx, sessionID, "First", "high")
	manager.AddTodo(ctx, sessionID, "Second", "medium")
	manager.AddTodo(ctx, sessionID, "Third", "low")

	// Retrieve and verify order
	todos, _ := manager.GetTodos(ctx, sessionID)

	if len(todos.Todos) != 3 {
		t.Fatalf("Expected 3 todos, got %d", len(todos.Todos))
	}

	if todos.Todos[0].Content != "First" {
		t.Errorf("Expected first todo 'First', got '%s'", todos.Todos[0].Content)
	}
	if todos.Todos[1].Content != "Second" {
		t.Errorf("Expected second todo 'Second', got '%s'", todos.Todos[1].Content)
	}
	if todos.Todos[2].Content != "Third" {
		t.Errorf("Expected third todo 'Third', got '%s'", todos.Todos[2].Content)
	}
}

func TestManager_GetTodos_MultipleSession(t *testing.T) {
	store := setupTestStorageTodo(t)
	defer store.Close()

	registry := provider.NewRegistry()
	manager := NewManager(store, registry)

	ctx := context.Background()
	session1 := "session-1"
	session2 := "session-2"

	// Add todos to different sessions
	manager.AddTodo(ctx, session1, "Session 1 Task", "high")
	manager.AddTodo(ctx, session2, "Session 2 Task", "medium")

	// Verify session 1 todos
	todos1, _ := manager.GetTodos(ctx, session1)
	if len(todos1.Todos) != 1 {
		t.Fatalf("Session 1: Expected 1 todo, got %d", len(todos1.Todos))
	}
	if todos1.Todos[0].Content != "Session 1 Task" {
		t.Errorf("Session 1: Wrong todo content")
	}

	// Verify session 2 todos
	todos2, _ := manager.GetTodos(ctx, session2)
	if len(todos2.Todos) != 1 {
		t.Fatalf("Session 2: Expected 1 todo, got %d", len(todos2.Todos))
	}
	if todos2.Todos[0].Content != "Session 2 Task" {
		t.Errorf("Session 2: Wrong todo content")
	}
}
