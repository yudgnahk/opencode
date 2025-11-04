package provider

import (
	"context"
)

// ModelsDevRegistry interface for accessing models.dev data
type ModelsDevRegistry interface {
	GetModels(providerID string) []string
}

// Provider is the interface all AI providers must implement
type Provider interface {
	// Complete generates a non-streaming completion
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)

	// Stream generates a streaming completion
	Stream(ctx context.Context, req CompletionRequest) (Stream, error)

	// Name returns provider name
	Name() string

	// Models returns available models
	Models() []string

	// SetModelsDevRegistry sets the models.dev registry for dynamic model lookup
	SetModelsDevRegistry(registry ModelsDevRegistry)
}

// CompletionRequest is the unified request format
type CompletionRequest struct {
	Messages      []Message `json:"messages"`
	Model         string    `json:"model"`
	MaxTokens     int       `json:"maxTokens"`
	Temperature   float64   `json:"temperature"`
	Tools         []Tool    `json:"tools,omitempty"`
	SystemPrompt  string    `json:"systemPrompt,omitempty"`
	StopSequences []string  `json:"stopSequences,omitempty"`
}

// CompletionResponse is the unified response format
type CompletionResponse struct {
	ID         string         `json:"id"`
	Content    []ContentBlock `json:"content"`
	StopReason string         `json:"stopReason"`
	Usage      Usage          `json:"usage"`
	ToolCalls  []ToolCall     `json:"toolCalls,omitempty"`
}

// Stream represents a streaming response
type Stream interface {
	Next() bool
	Chunk() StreamChunk
	Error() error
	Close() error
}

// StreamChunk is a single chunk in a stream
type StreamChunk struct {
	Type       ChunkType `json:"type"`
	Delta      string    `json:"delta,omitempty"`
	ToolCall   *ToolCall `json:"toolCall,omitempty"`
	StopReason string    `json:"stopReason,omitempty"`
}

type ChunkType string

const (
	ChunkTypeText     ChunkType = "text"
	ChunkTypeToolCall ChunkType = "tool_call"
	ChunkTypeStop     ChunkType = "stop"
)

type Message struct {
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

type ContentBlock struct {
	Type       string      `json:"type"` // "text", "tool_use", "tool_result"
	Text       string      `json:"text,omitempty"`
	ToolUse    *ToolUse    `json:"toolUse,omitempty"`
	ToolResult *ToolResult `json:"toolResult,omitempty"`
}

type ToolUse struct {
	ID    string      `json:"id"`
	Name  string      `json:"name"`
	Input interface{} `json:"input"`
}

type ToolResult struct {
	ToolUseID string      `json:"toolUseId"`
	Content   interface{} `json:"content"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"` // JSON Schema
}

type ToolCall struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Arguments interface{} `json:"arguments"`
}

type Usage struct {
	InputTokens  int `json:"inputTokens"`
	OutputTokens int `json:"outputTokens"`
}
