package session

import (
	"context"
	"fmt"
)

type TodoList struct {
	SessionID string `json:"sessionId"`
	Todos     []Todo `json:"todos"`
}

type Todo struct {
	ID       string     `json:"id"`
	Content  string     `json:"content"`
	Status   TodoStatus `json:"status"`
	Priority string     `json:"priority"` // high, medium, low
}

type TodoStatus string

const (
	TodoPending    TodoStatus = "pending"
	TodoInProgress TodoStatus = "in_progress"
	TodoCompleted  TodoStatus = "completed"
	TodoCancelled  TodoStatus = "cancelled"
)

// GetTodos retrieves todos for a session
func (m *Manager) GetTodos(ctx context.Context, sessionID string) (*TodoList, error) {
	var todos TodoList
	if err := m.storage.GetJSON("todos", sessionID, &todos); err != nil {
		// Return empty list if not found
		return &TodoList{
			SessionID: sessionID,
			Todos:     []Todo{},
		}, nil
	}

	// Ensure SessionID is set even if data exists
	if todos.SessionID == "" {
		todos.SessionID = sessionID
	}

	// Ensure Todos slice is initialized
	if todos.Todos == nil {
		todos.Todos = []Todo{}
	}

	return &todos, nil
}

// UpdateTodos updates the todo list for a session
func (m *Manager) UpdateTodos(ctx context.Context, sessionID string, todos []Todo) error {
	todoList := &TodoList{
		SessionID: sessionID,
		Todos:     todos,
	}

	if err := m.storage.SetJSON("todos", sessionID, todoList); err != nil {
		return err
	}

	// Emit event
	m.eventBus.Publish(Event{
		Type:      "todos.updated",
		SessionID: sessionID,
		Data:      todoList,
	})

	return nil
}

// AddTodo adds a new todo
func (m *Manager) AddTodo(ctx context.Context, sessionID, content, priority string) error {
	todos, err := m.GetTodos(ctx, sessionID)
	if err != nil {
		return err
	}

	newTodo := Todo{
		ID:       generateID(),
		Content:  content,
		Status:   TodoPending,
		Priority: priority,
	}

	todos.Todos = append(todos.Todos, newTodo)

	return m.UpdateTodos(ctx, sessionID, todos.Todos)
}

// UpdateTodoStatus updates a single todo's status
func (m *Manager) UpdateTodoStatus(ctx context.Context, sessionID, todoID string, status TodoStatus) error {
	todos, err := m.GetTodos(ctx, sessionID)
	if err != nil {
		return err
	}

	for i, todo := range todos.Todos {
		if todo.ID == todoID {
			todos.Todos[i].Status = status
			return m.UpdateTodos(ctx, sessionID, todos.Todos)
		}
	}

	return fmt.Errorf("todo not found: %s", todoID)
}
