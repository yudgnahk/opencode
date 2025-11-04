package session

import (
	"context"
	"fmt"
	"strings"

	"github.com/opencode/opencode-go/internal/provider"
)

// Summarize creates a summary of session messages
func (m *Manager) Summarize(ctx context.Context, sessionID string) (string, error) {
	messages, err := m.GetMessages(ctx, sessionID)
	if err != nil {
		return "", err
	}

	if len(messages) == 0 {
		return "", fmt.Errorf("no messages to summarize")
	}

	// Build conversation text
	conversationText := buildConversationText(messages)

	// Get session to determine provider
	session, err := m.Get(ctx, sessionID)
	if err != nil {
		return "", err
	}

	// Get provider
	prov, err := m.providerRegistry.Get(session.Provider)
	if err != nil {
		return "", fmt.Errorf("provider not found: %w", err)
	}

	// Create summarization request
	req := provider.CompletionRequest{
		Messages: []provider.Message{
			{
				Role: "user",
				Content: []provider.ContentBlock{
					{
						Type: "text",
						Text: fmt.Sprintf("Please provide a concise 1-2 sentence summary of the following conversation. Focus on the main task or topic:\n\n%s", conversationText),
					},
				},
			},
		},
		Model:       session.Model,
		MaxTokens:   500,
		Temperature: 0.3,
	}

	resp, err := prov.Complete(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}

	summary := extractTextFromContent(resp.Content)
	summary = strings.TrimSpace(summary)

	// Update session title with summary (truncated to 100 chars)
	if err := m.Update(ctx, sessionID, map[string]interface{}{
		"title": truncate(summary, 100),
	}); err != nil {
		return summary, err // Return summary even if update fails
	}

	return summary, nil
}

// buildConversationText extracts text from all messages
func buildConversationText(messages []*Message) string {
	var builder strings.Builder

	for _, msg := range messages {
		text := extractTextFromMessage(msg)
		if text != "" {
			builder.WriteString(string(msg.Role))
			builder.WriteString(": ")
			builder.WriteString(text)
			builder.WriteString("\n\n")
		}
	}

	return builder.String()
}

// extractTextFromMessage extracts text content from a message
func extractTextFromMessage(msg *Message) string {
	var parts []string

	for _, block := range msg.Content {
		if block.Type == "text" && block.Text != "" {
			parts = append(parts, block.Text)
		}
	}

	return strings.Join(parts, " ")
}

// extractTextFromContent extracts text from provider content blocks
func extractTextFromContent(content []provider.ContentBlock) string {
	for _, block := range content {
		if block.Type == "text" && block.Text != "" {
			return block.Text
		}
	}
	return ""
}

// truncate truncates a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
