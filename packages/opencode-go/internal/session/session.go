package session

import (
	"time"
)

// Session matches the TypeScript Session.Info schema
type Session struct {
	ID        string                 `json:"id"`
	ProjectID string                 `json:"projectId"`
	Directory string                 `json:"directory"`
	Title     string                 `json:"title"`
	Version   string                 `json:"version"` // OpenCode version, not model version
	ParentID  string                 `json:"parentId,omitempty"`
	Time      SessionTime            `json:"time"`
	Share     *SessionShare          `json:"share,omitempty"`
	Summary   *SessionSummary        `json:"summary,omitempty"`
	Revert    *SessionRevert         `json:"revert,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`

	// Additional fields for Go implementation (not in TypeScript schema)
	// These are stored separately in metadata or inferred from messages
	Provider   string    `json:"-"` // Not persisted, inferred from Version or metadata
	Model      string    `json:"-"` // Not persisted, inferred from messages or metadata
	MessageIDs []string  `json:"-"` // Not persisted in session, loaded from message storage
	UpdatedAt  time.Time `json:"-"` // Helper for sorting, derived from Time.Updated
}

// SessionTime stores Unix millisecond timestamps
type SessionTime struct {
	Created    int64 `json:"created"`
	Updated    int64 `json:"updated"`
	Compacting int64 `json:"compacting,omitempty"`
}

type SessionShare struct {
	URL string `json:"url"`
}

type SessionSummary struct {
	Diffs []interface{} `json:"diffs"`
}

type SessionRevert struct {
	MessageID string `json:"messageID"`
	PartID    string `json:"partID,omitempty"`
	Snapshot  string `json:"snapshot,omitempty"`
	Diff      string `json:"diff,omitempty"`
}

// SessionListItem is a lightweight representation for listing sessions
// This avoids loading full nested objects
type SessionListItem struct {
	ID           string `json:"id"`
	ProjectID    string `json:"projectId"`
	Title        string `json:"title"`
	Version      string `json:"version"`
	ParentID     string `json:"parentId,omitempty"`
	Created      int64  `json:"created"`
	Updated      int64  `json:"updated"`
	MessageCount int    `json:"messageCount"`
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
