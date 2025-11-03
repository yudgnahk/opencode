package session

import (
	"time"
)

type Session struct {
	ID         string                 `json:"id"`
	ProjectID  string                 `json:"projectId"`
	Title      string                 `json:"title"`
	Model      string                 `json:"model"`
	Provider   string                 `json:"provider"`
	State      SessionState           `json:"state"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time              `json:"createdAt"`
	UpdatedAt  time.Time              `json:"updatedAt"`
	MessageIDs []string               `json:"messageIds"`         // Ordered list
	ParentID   string                 `json:"parentId,omitempty"` // For forks
}

type SessionState string

const (
	StateIdle       SessionState = "idle"
	StateProcessing SessionState = "processing"
	StateError      SessionState = "error"
	StateCompleted  SessionState = "completed"
)

type Message struct {
	ID        string                 `json:"id"`
	SessionID string                 `json:"sessionId"`
	Role      Role                   `json:"role"`
	Content   []ContentBlock         `json:"content"`
	ToolCalls []ToolCall             `json:"toolCalls,omitempty"`
	CreatedAt time.Time              `json:"createdAt"`
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
	Type string      `json:"type"` // "text", "image", "tool_use", "tool_result"
	Text string      `json:"text,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

type ToolCall struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Arguments interface{} `json:"arguments"`
	Result    interface{} `json:"result,omitempty"`
}
