package types

import "time"

// Session represents a chat session
type Session struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Message represents a conversation message
type Message struct {
	ID        string         `json:"id"`
	Role      string         `json:"role"`
	Content   []ContentBlock `json:"content"`
	CreatedAt time.Time      `json:"createdAt"`
}

// ContentBlock represents message content
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// Project represents workspace project info
type Project struct {
	ID   string `json:"id"`
	Path string `json:"path"`
	Name string `json:"name"`
}

// Agent represents an AI agent configuration
type Agent struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Tool represents an available tool
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

// Config represents OpenCode configuration
type Config struct {
	Providers map[string]ProviderConfig `json:"providers"`
	Agents    []Agent                   `json:"agents"`
}

type ProviderConfig struct {
	APIKey string `json:"apiKey"`
	Model  string `json:"model"`
}
