package session

import (
	"context"
	"fmt"
	"time"

	"github.com/opencode/opencode-go/internal/provider"
)

// CompletionService handles AI completions for sessions
type CompletionService struct {
	manager  *Manager
	registry *provider.Registry
}

// NewCompletionService creates a new completion service
func NewCompletionService(manager *Manager, registry *provider.Registry) *CompletionService {
	return &CompletionService{
		manager:  manager,
		registry: registry,
	}
}

// SendMessage adds a user message and generates an AI response
func (s *CompletionService) SendMessage(ctx context.Context, sessionID, content string) error {
	// Get session
	sess, err := s.manager.Get(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Update state
	s.manager.Update(ctx, sessionID, map[string]interface{}{
		"state": StateProcessing,
	})

	// Create user message
	userContent := []ContentBlock{{
		Type: "text",
		Text: content,
	}}

	// Save user message
	if _, err := s.manager.AddMessage(ctx, sessionID, RoleUser, userContent); err != nil {
		return err
	}

	// Get provider
	prov, err := s.registry.Get(sess.Provider)
	if err != nil {
		s.manager.Update(ctx, sessionID, map[string]interface{}{
			"state": StateError,
		})
		return fmt.Errorf("provider not found: %w", err)
	}

	// Build completion request
	messages, err := s.manager.GetMessages(ctx, sessionID)
	if err != nil {
		return err
	}

	req := s.buildCompletionRequest(sess, messages)

	// Log the request details
	fmt.Printf("DEBUG [CompletionService]: Sending completion request - Provider: %s, Model: %s, Messages: %d\n",
		sess.Provider, sess.Model, len(req.Messages))

	// Call provider
	startAPI := time.Now()
	resp, err := prov.Complete(ctx, req)
	apiDuration := time.Since(startAPI)
	fmt.Printf("DEBUG [Completion]: API call took %v\n", apiDuration)
	if err != nil {
		s.manager.Update(ctx, sessionID, map[string]interface{}{
			"state": StateError,
		})
		return fmt.Errorf("completion failed: %w", err)
	}

	// Create assistant message
	assistantContent := s.convertContent(resp.Content)

	// Save assistant message
	assistantMsg, err := s.manager.AddMessage(ctx, sessionID, RoleAssistant, assistantContent)
	if err != nil {
		return err
	}

	// Update message metadata with usage info
	assistantMsg.ToolCalls = s.convertToolCalls(resp.ToolCalls)
	assistantMsg.Metadata = map[string]interface{}{
		"usage": map[string]int{
			"inputTokens":  resp.Usage.InputTokens,
			"outputTokens": resp.Usage.OutputTokens,
		},
		"stopReason": resp.StopReason,
	}

	// Save the updated message with metadata
	if err := s.manager.storage.WriteJSON([]string{"message", sessionID, assistantMsg.ID}, assistantMsg); err != nil {
		return fmt.Errorf("failed to update message metadata: %w", err)
	}

	// Update state
	s.manager.Update(ctx, sessionID, map[string]interface{}{
		"state": StateIdle,
	})

	totalDuration := time.Since(startAPI)
	fmt.Printf("DEBUG [Completion]: Total completion time: %v (API: %v, Post-processing: %v)\n",
		totalDuration, apiDuration, totalDuration-apiDuration)

	return nil
}

// StreamMessage sends a message and streams the response
func (s *CompletionService) StreamMessage(ctx context.Context, sessionID, content string) (provider.Stream, error) {
	// Get session
	sess, err := s.manager.Get(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Update state
	s.manager.Update(ctx, sessionID, map[string]interface{}{
		"state": StateProcessing,
	})

	// Create user message
	userContent := []ContentBlock{{
		Type: "text",
		Text: content,
	}}

	// Save user message
	if _, err := s.manager.AddMessage(ctx, sessionID, RoleUser, userContent); err != nil {
		return nil, err
	}

	// Get provider
	prov, err := s.registry.Get(sess.Provider)
	if err != nil {
		s.manager.Update(ctx, sessionID, map[string]interface{}{
			"state": StateError,
		})
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	// Build completion request
	messages, err := s.manager.GetMessages(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	req := s.buildCompletionRequest(sess, messages)

	// Start streaming
	stream, err := prov.Stream(ctx, req)
	if err != nil {
		s.manager.Update(ctx, sessionID, map[string]interface{}{
			"state": StateError,
		})
		return nil, fmt.Errorf("streaming failed: %w", err)
	}

	return stream, nil
}

// buildCompletionRequest converts session messages to provider format
func (s *CompletionService) buildCompletionRequest(sess *Session, messages []*Message) provider.CompletionRequest {
	providerMessages := make([]provider.Message, 0, len(messages))

	for _, msg := range messages {
		content := make([]provider.ContentBlock, len(msg.Content))
		for i, block := range msg.Content {
			content[i] = provider.ContentBlock{
				Type: block.Type,
				Text: block.Text,
			}
		}

		providerMessages = append(providerMessages, provider.Message{
			Role:    string(msg.Role),
			Content: content,
		})
	}

	return provider.CompletionRequest{
		Messages:    providerMessages,
		Model:       sess.Model,
		MaxTokens:   4096,
		Temperature: 0.7,
	}
}

// convertContent converts provider content to session content
func (s *CompletionService) convertContent(providerContent []provider.ContentBlock) []ContentBlock {
	content := make([]ContentBlock, len(providerContent))
	for i, block := range providerContent {
		content[i] = ContentBlock{
			Type: block.Type,
			Text: block.Text,
		}
	}
	return content
}

// convertToolCalls converts provider tool calls to session tool calls
func (s *CompletionService) convertToolCalls(providerCalls []provider.ToolCall) []ToolCall {
	calls := make([]ToolCall, len(providerCalls))
	for i, call := range providerCalls {
		calls[i] = ToolCall{
			ID:        call.ID,
			Name:      call.Name,
			Arguments: call.Arguments,
		}
	}
	return calls
}
