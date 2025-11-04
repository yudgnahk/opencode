package session

import (
	"context"
	"errors"
	"testing"

	"github.com/opencode/opencode-go/internal/provider"
)

func setupMockProvider(t *testing.T, registry *provider.Registry, name string, response *provider.CompletionResponse, err error) {
	t.Helper()

	mockProv := &mockSummarizeProvider{
		name:     name,
		response: response,
		err:      err,
	}

	if regErr := registry.Register(mockProv); regErr != nil {
		t.Fatalf("failed to register provider: %v", regErr)
	}
}

type mockSummarizeProvider struct {
	name     string
	response *provider.CompletionResponse
	err      error
}

func (m *mockSummarizeProvider) Name() string {
	return m.name
}

func (m *mockSummarizeProvider) Models() []string {
	return []string{"test-model"}
}

func (m *mockSummarizeProvider) Complete(ctx context.Context, req provider.CompletionRequest) (*provider.CompletionResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

func (m *mockSummarizeProvider) Stream(ctx context.Context, req provider.CompletionRequest) (provider.Stream, error) {
	return nil, errors.New("streaming not implemented")
}

func (m *mockSummarizeProvider) SetModelsDevRegistry(registry provider.ModelsDevRegistry) {
	// Mock provider doesn't need modelsdev registry
}

func TestManager_Summarize_Success(t *testing.T) {
	store := setupTestStorage(t)
	defer store.Close()

	registry := provider.NewRegistry()

	// Register mock provider
	response := &provider.CompletionResponse{
		Content: []provider.ContentBlock{
			{Type: "text", Text: "Implementing session summarization feature with AI-generated summaries"},
		},
	}
	setupMockProvider(t, registry, "mock", response, nil)

	manager := NewManager(store, registry)
	ctx := context.Background()

	// Create session with mock provider
	session, err := manager.Create(ctx, "test-project", "mock", "test-model")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Add messages
	_, err = manager.AddMessage(ctx, session.ID, RoleUser, []ContentBlock{
		{Type: "text", Text: "Please implement session summarization"},
	})
	if err != nil {
		t.Fatalf("failed to add user message: %v", err)
	}

	_, err = manager.AddMessage(ctx, session.ID, RoleAssistant, []ContentBlock{
		{Type: "text", Text: "I'll help implement session summarization"},
	})
	if err != nil {
		t.Fatalf("failed to add assistant message: %v", err)
	}

	// Test summarization
	summary, err := manager.Summarize(ctx, session.ID)
	if err != nil {
		t.Fatalf("failed to summarize: %v", err)
	}

	expectedSummary := "Implementing session summarization feature with AI-generated summaries"
	if summary != expectedSummary {
		t.Errorf("expected summary %q, got %q", expectedSummary, summary)
	}

	// Verify session title was updated
	updated, err := manager.Get(ctx, session.ID)
	if err != nil {
		t.Fatalf("failed to get updated session: %v", err)
	}

	if updated.Title != expectedSummary {
		t.Errorf("expected title %q, got %q", expectedSummary, updated.Title)
	}
}

func TestManager_Summarize_NoMessages(t *testing.T) {
	store := setupTestStorage(t)
	defer store.Close()

	registry := provider.NewRegistry()
	manager := NewManager(store, registry)
	ctx := context.Background()

	// Create session without messages
	session, err := manager.Create(ctx, "test-project", "mock", "test-model")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Test summarization with no messages
	_, err = manager.Summarize(ctx, session.ID)
	if err == nil {
		t.Error("expected error for empty message list, got nil")
	}

	expectedErr := "no messages to summarize"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestManager_Summarize_SessionNotFound(t *testing.T) {
	store := setupTestStorage(t)
	defer store.Close()

	registry := provider.NewRegistry()
	manager := NewManager(store, registry)
	ctx := context.Background()

	// Test with non-existent session
	_, err := manager.Summarize(ctx, "non-existent-id")
	if err == nil {
		t.Error("expected error for non-existent session, got nil")
	}
}

func TestManager_Summarize_ProviderNotFound(t *testing.T) {
	store := setupTestStorage(t)
	defer store.Close()

	registry := provider.NewRegistry()
	manager := NewManager(store, registry)
	ctx := context.Background()

	// Create session with unregistered provider
	session, err := manager.Create(ctx, "test-project", "unregistered-provider", "test-model")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Add a message
	_, err = manager.AddMessage(ctx, session.ID, RoleUser, []ContentBlock{
		{Type: "text", Text: "Test message"},
	})
	if err != nil {
		t.Fatalf("failed to add message: %v", err)
	}

	// Test summarization with unregistered provider
	_, err = manager.Summarize(ctx, session.ID)
	if err == nil {
		t.Error("expected error for unregistered provider, got nil")
	}
}

func TestManager_Summarize_ProviderError(t *testing.T) {
	store := setupTestStorage(t)
	defer store.Close()

	registry := provider.NewRegistry()

	// Register mock provider that returns error
	providerErr := errors.New("provider failure")
	setupMockProvider(t, registry, "mock", nil, providerErr)

	manager := NewManager(store, registry)
	ctx := context.Background()

	// Create session
	session, err := manager.Create(ctx, "test-project", "mock", "test-model")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Add a message
	_, err = manager.AddMessage(ctx, session.ID, RoleUser, []ContentBlock{
		{Type: "text", Text: "Test message"},
	})
	if err != nil {
		t.Fatalf("failed to add message: %v", err)
	}

	// Test summarization with provider error
	_, err = manager.Summarize(ctx, session.ID)
	if err == nil {
		t.Error("expected error from provider, got nil")
	}
}

func TestManager_Summarize_LongSummaryTruncation(t *testing.T) {
	store := setupTestStorage(t)
	defer store.Close()

	registry := provider.NewRegistry()

	// Create long summary (>100 chars)
	longSummary := "This is a very long summary that exceeds one hundred characters and should be truncated when used as the session title to keep things manageable"

	response := &provider.CompletionResponse{
		Content: []provider.ContentBlock{
			{Type: "text", Text: longSummary},
		},
	}
	setupMockProvider(t, registry, "mock", response, nil)

	manager := NewManager(store, registry)
	ctx := context.Background()

	// Create session
	session, err := manager.Create(ctx, "test-project", "mock", "test-model")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Add a message
	_, err = manager.AddMessage(ctx, session.ID, RoleUser, []ContentBlock{
		{Type: "text", Text: "Test message"},
	})
	if err != nil {
		t.Fatalf("failed to add message: %v", err)
	}

	// Test summarization
	summary, err := manager.Summarize(ctx, session.ID)
	if err != nil {
		t.Fatalf("failed to summarize: %v", err)
	}

	// Full summary should be returned
	if summary != longSummary {
		t.Errorf("expected full summary, got truncated")
	}

	// But title should be truncated to 100 chars + "..."
	updated, err := manager.Get(ctx, session.ID)
	if err != nil {
		t.Fatalf("failed to get updated session: %v", err)
	}

	expectedTitle := longSummary[:100] + "..."
	if updated.Title != expectedTitle {
		t.Errorf("expected truncated title of length %d, got length %d", len(expectedTitle), len(updated.Title))
	}

	if len(updated.Title) != 103 { // 100 + "..."
		t.Errorf("expected title length 103, got %d", len(updated.Title))
	}
}

func TestManager_Summarize_MultipleContentBlocks(t *testing.T) {
	store := setupTestStorage(t)
	defer store.Close()

	registry := provider.NewRegistry()

	response := &provider.CompletionResponse{
		Content: []provider.ContentBlock{
			{Type: "text", Text: "Summary from multiple blocks"},
		},
	}
	setupMockProvider(t, registry, "mock", response, nil)

	manager := NewManager(store, registry)
	ctx := context.Background()

	// Create session
	session, err := manager.Create(ctx, "test-project", "mock", "test-model")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Add message with multiple text content blocks
	_, err = manager.AddMessage(ctx, session.ID, RoleUser, []ContentBlock{
		{Type: "text", Text: "First part of message."},
		{Type: "text", Text: "Second part of message."},
		{Type: "tool_use", Text: "Should be ignored"},
	})
	if err != nil {
		t.Fatalf("failed to add message: %v", err)
	}

	// Test summarization
	summary, err := manager.Summarize(ctx, session.ID)
	if err != nil {
		t.Fatalf("failed to summarize: %v", err)
	}

	expectedSummary := "Summary from multiple blocks"
	if summary != expectedSummary {
		t.Errorf("expected summary %q, got %q", expectedSummary, summary)
	}
}

func TestBuildConversationText(t *testing.T) {
	messages := []*Message{
		{
			Role: RoleUser,
			Content: []ContentBlock{
				{Type: "text", Text: "Hello"},
			},
		},
		{
			Role: RoleAssistant,
			Content: []ContentBlock{
				{Type: "text", Text: "Hi there"},
			},
		},
		{
			Role: RoleUser,
			Content: []ContentBlock{
				{Type: "text", Text: "How are you?"},
			},
		},
	}

	result := buildConversationText(messages)

	expected := "user: Hello\n\nassistant: Hi there\n\nuser: How are you?\n\n"
	if result != expected {
		t.Errorf("expected:\n%q\ngot:\n%q", expected, result)
	}
}

func TestExtractTextFromMessage(t *testing.T) {
	tests := []struct {
		name     string
		message  *Message
		expected string
	}{
		{
			name: "single text block",
			message: &Message{
				Content: []ContentBlock{
					{Type: "text", Text: "Hello"},
				},
			},
			expected: "Hello",
		},
		{
			name: "multiple text blocks",
			message: &Message{
				Content: []ContentBlock{
					{Type: "text", Text: "Hello"},
					{Type: "text", Text: "World"},
				},
			},
			expected: "Hello World",
		},
		{
			name: "mixed content types",
			message: &Message{
				Content: []ContentBlock{
					{Type: "text", Text: "Hello"},
					{Type: "tool_use", Text: "ignored"},
					{Type: "text", Text: "World"},
				},
			},
			expected: "Hello World",
		},
		{
			name: "no text blocks",
			message: &Message{
				Content: []ContentBlock{
					{Type: "tool_use", Text: "ignored"},
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTextFromMessage(tt.message)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "short string",
			input:    "Hello",
			maxLen:   10,
			expected: "Hello",
		},
		{
			name:     "exact length",
			input:    "Hello",
			maxLen:   5,
			expected: "Hello",
		},
		{
			name:     "needs truncation",
			input:    "Hello World",
			maxLen:   5,
			expected: "Hello...",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   5,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExtractTextFromContent(t *testing.T) {
	tests := []struct {
		name     string
		content  []provider.ContentBlock
		expected string
	}{
		{
			name: "single text block",
			content: []provider.ContentBlock{
				{Type: "text", Text: "Summary text"},
			},
			expected: "Summary text",
		},
		{
			name: "multiple blocks, first is text",
			content: []provider.ContentBlock{
				{Type: "text", Text: "Summary text"},
				{Type: "tool_use", Text: "ignored"},
			},
			expected: "Summary text",
		},
		{
			name: "no text blocks",
			content: []provider.ContentBlock{
				{Type: "tool_use", Text: "ignored"},
			},
			expected: "",
		},
		{
			name:     "empty content",
			content:  []provider.ContentBlock{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTextFromContent(tt.content)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
