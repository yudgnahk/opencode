package tool

import (
	"context"
	"fmt"

	"github.com/opencode/opencode-go/internal/provider"
	"github.com/opencode/opencode-go/internal/session"
	"github.com/opencode/opencode-go/internal/storage"
)

// TaskRequest represents a request to spawn a sub-agent task
type TaskRequest struct {
	Description  string `json:"description" validate:"required"`
	Prompt       string `json:"prompt" validate:"required"`
	SubagentType string `json:"subagentType"` // "general", "codex", etc.
}

// TaskResponse represents the result of a sub-agent task
type TaskResponse struct {
	Result    string `json:"result"`
	SessionID string `json:"sessionId"`
}

// TaskExecutor handles sub-agent task spawning
type TaskExecutor struct {
	sessionManager    *session.Manager
	completionService *session.CompletionService
	providerRegistry  *provider.Registry
	projectID         string
}

// NewTaskExecutor creates a new task executor
func NewTaskExecutor(
	storage *storage.Storage,
	providerRegistry *provider.Registry,
	projectID string,
) *TaskExecutor {
	sessionMgr := session.NewManager(storage, providerRegistry)
	completionSvc := session.NewCompletionService(sessionMgr, providerRegistry)

	return &TaskExecutor{
		sessionManager:    sessionMgr,
		completionService: completionSvc,
		providerRegistry:  providerRegistry,
		projectID:         projectID,
	}
}

// Execute spawns a sub-agent and executes the task
func (t *TaskExecutor) Execute(ctx context.Context, req TaskRequest) (*TaskResponse, error) {
	// Default to general sub-agent type
	if req.SubagentType == "" {
		req.SubagentType = "general"
	}

	// Determine provider and model based on subagent type
	provider, model := t.getProviderForSubagent(req.SubagentType)

	// Create a new session for the sub-agent
	subSession, err := t.sessionManager.Create(ctx, t.projectID, provider, model)
	if err != nil {
		return nil, fmt.Errorf("failed to create sub-session: %w", err)
	}

	// Set session title to the task description
	if err := t.sessionManager.Update(ctx, subSession.ID, map[string]interface{}{
		"title": truncateTitle(req.Description),
		"metadata": map[string]interface{}{
			"isSubAgent": true,
			"type":       req.SubagentType,
		},
	}); err != nil {
		return nil, fmt.Errorf("failed to update sub-session: %w", err)
	}

	// Send the prompt to the sub-agent and get response
	if err := t.completionService.SendMessage(ctx, subSession.ID, req.Prompt); err != nil {
		return nil, fmt.Errorf("sub-agent failed: %w", err)
	}

	// Get the response
	messages, err := t.sessionManager.GetMessages(ctx, subSession.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve messages: %w", err)
	}

	if len(messages) < 2 {
		return nil, fmt.Errorf("sub-agent produced no response")
	}

	// The last message should be the assistant's response
	lastMessage := messages[len(messages)-1]
	if lastMessage.Role != session.RoleAssistant {
		return nil, fmt.Errorf("expected assistant response, got %s", lastMessage.Role)
	}

	result := extractTextFromMessage(lastMessage)

	return &TaskResponse{
		Result:    result,
		SessionID: subSession.ID,
	}, nil
}

// getProviderForSubagent returns the appropriate provider and model for a subagent type
func (t *TaskExecutor) getProviderForSubagent(subagentType string) (string, string) {
	switch subagentType {
	case "codex":
		// Use advanced reasoning model for complex tasks
		return "anthropic", "claude-3-5-sonnet-20241022"
	case "general":
		// Use default model for general tasks
		return "anthropic", "claude-3-5-sonnet-20241022"
	case "gemini":
		// Use Gemini for large context tasks
		return "gemini", "gemini-2.0-flash-exp"
	default:
		// Default to Anthropic
		return "anthropic", "claude-3-5-sonnet-20241022"
	}
}

// extractTextFromMessage extracts text content from a message
func extractTextFromMessage(msg *session.Message) string {
	for _, block := range msg.Content {
		if block.Type == "text" {
			return block.Text
		}
	}
	return ""
}

// truncateTitle truncates a title to a maximum length
func truncateTitle(title string) string {
	maxLen := 100
	if len(title) <= maxLen {
		return title
	}
	return title[:maxLen] + "..."
}
