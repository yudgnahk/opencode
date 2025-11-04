package session

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/opencode/opencode-go/internal/provider"
	"github.com/opencode/opencode-go/internal/storage"
)

// TestCompletionServiceIntegration tests the full flow from session to provider to response
func TestCompletionServiceIntegration(t *testing.T) {
	t.Run("Complete flow with mock provider", func(t *testing.T) {
		ctx := context.Background()

		// Setup temp storage
		tempDir := t.TempDir()
		storePath := filepath.Join(tempDir, "test.db")
		store, err := storage.New(storePath)
		if err != nil {
			t.Fatalf("Failed to create storage: %v", err)
		}
		defer store.Close()

		// Create session manager
		registry := provider.NewRegistry()
		sessionMgr := NewManager(store, registry)

		// Create provider registry

		// Register mock provider
		mockProvider := &mockProvider{
			name:   "mock",
			models: []string{"mock-model"},
			response: &provider.CompletionResponse{
				ID: "test-response-1",
				Content: []provider.ContentBlock{{
					Type: "text",
					Text: "This is a test response from the mock provider",
				}},
				StopReason: "end_turn",
				Usage: provider.Usage{
					InputTokens:  10,
					OutputTokens: 15,
				},
			},
		}

		if err := registry.Register(mockProvider); err != nil {
			t.Fatalf("Failed to register mock provider: %v", err)
		}

		// Create completion service
		completionSvc := NewCompletionService(sessionMgr, registry)

		// Create a session
		sess, err := sessionMgr.Create(ctx, "test-project", "mock", "mock-model")
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Send a message
		err = completionSvc.SendMessage(ctx, sess.ID, "Hello, can you help me?")
		if err != nil {
			t.Fatalf("SendMessage failed: %v", err)
		}

		// Verify messages were stored
		messages, err := sessionMgr.GetMessages(ctx, sess.ID)
		if err != nil {
			t.Fatalf("GetMessages failed: %v", err)
		}

		// Should have user message + assistant response
		if len(messages) != 2 {
			t.Fatalf("Expected 2 messages, got %d", len(messages))
		}

		// Check user message
		if messages[0].Role != RoleUser {
			t.Errorf("First message should be user, got %s", messages[0].Role)
		}
		if messages[0].Content[0].Text != "Hello, can you help me?" {
			t.Errorf("User message content mismatch")
		}

		// Check assistant message
		if messages[1].Role != RoleAssistant {
			t.Errorf("Second message should be assistant, got %s", messages[1].Role)
		}
		if messages[1].Content[0].Text != "This is a test response from the mock provider" {
			t.Errorf("Assistant message content mismatch")
		}

		// Verify session state updated to idle
		updatedSess, err := sessionMgr.Get(ctx, sess.ID)
		if err != nil {
			t.Fatalf("Failed to get updated session: %v", err)
		}
		if updatedSess.State != StateIdle {
			t.Errorf("Expected session state to be idle, got %s", updatedSess.State)
		}
	})

	t.Run("Streaming flow with mock provider", func(t *testing.T) {
		ctx := context.Background()

		// Setup temp storage
		tempDir := t.TempDir()
		storePath := filepath.Join(tempDir, "test.db")
		store, err := storage.New(storePath)
		if err != nil {
			t.Fatalf("Failed to create storage: %v", err)
		}
		defer store.Close()

		// Create session manager
		registry := provider.NewRegistry()
		sessionMgr := NewManager(store, registry)

		// Create provider registry

		// Create mock stream
		mockStream := &mockStream{
			chunks: []provider.StreamChunk{
				{Type: provider.ChunkTypeText, Delta: "Hello "},
				{Type: provider.ChunkTypeText, Delta: "from "},
				{Type: provider.ChunkTypeText, Delta: "streaming!"},
				{Type: provider.ChunkTypeStop, StopReason: "end_turn"},
			},
		}

		// Register mock provider
		mockProvider := &mockProvider{
			name:       "mock",
			models:     []string{"mock-model"},
			streamResp: mockStream,
		}

		if err := registry.Register(mockProvider); err != nil {
			t.Fatalf("Failed to register mock provider: %v", err)
		}

		// Create completion service
		completionSvc := NewCompletionService(sessionMgr, registry)

		// Create a session
		sess, err := sessionMgr.Create(ctx, "test-project", "mock", "mock-model")
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Start streaming
		stream, err := completionSvc.StreamMessage(ctx, sess.ID, "Tell me something")
		if err != nil {
			t.Fatalf("StreamMessage failed: %v", err)
		}
		defer stream.Close()

		// Collect chunks
		var chunks []provider.StreamChunk
		for stream.Next() {
			chunks = append(chunks, stream.Chunk())
		}

		// Check for errors
		if err := stream.Error(); err != nil {
			t.Fatalf("Stream error: %v", err)
		}

		// Verify chunks
		if len(chunks) != 4 {
			t.Fatalf("Expected 4 chunks, got %d", len(chunks))
		}

		// Check text chunks
		expectedDeltas := []string{"Hello ", "from ", "streaming!"}
		for i, delta := range expectedDeltas {
			if chunks[i].Delta != delta {
				t.Errorf("Chunk %d delta mismatch: got %s, expected %s", i, chunks[i].Delta, delta)
			}
		}

		// Check stop chunk
		if chunks[3].Type != provider.ChunkTypeStop {
			t.Errorf("Last chunk should be stop, got %s", chunks[3].Type)
		}
	})

	t.Run("Retry on provider failure", func(t *testing.T) {
		ctx := context.Background()

		// Setup temp storage
		tempDir := t.TempDir()
		storePath := filepath.Join(tempDir, "test.db")
		store, err := storage.New(storePath)
		if err != nil {
			t.Fatalf("Failed to create storage: %v", err)
		}
		defer store.Close()

		// Create session manager
		registry := provider.NewRegistry()
		sessionMgr := NewManager(store, registry)

		// Create provider registry

		// Create counting provider that fails once then succeeds
		attempts := 0
		countingProvider := &countingProvider{
			name: "counting",
			completeFunc: func(ctx context.Context, req provider.CompletionRequest) (*provider.CompletionResponse, error) {
				attempts++
				if attempts == 1 {
					// Return a retryable error with 503 in message
					return nil, &apiError{
						statusCode: 503,
						message:    "Service unavailable: 503 - temporary outage",
					}
				}
				return &provider.CompletionResponse{
					ID: "success-after-retry",
					Content: []provider.ContentBlock{{
						Type: "text",
						Text: "Success after retry",
					}},
				}, nil
			},
		}

		// Wrap with retry
		retryConfig := provider.RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 10 * time.Millisecond,
			MaxBackoff:     100 * time.Millisecond,
			Multiplier:     2.0,
		}
		retryable := provider.NewRetryableProvider(countingProvider, retryConfig)

		if err := registry.Register(retryable); err != nil {
			t.Fatalf("Failed to register provider: %v", err)
		}

		// Create completion service
		completionSvc := NewCompletionService(sessionMgr, registry)

		// Create a session
		sess, err := sessionMgr.Create(ctx, "test-project", "counting", "test-model")
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Send message - should retry and succeed
		err = completionSvc.SendMessage(ctx, sess.ID, "Test retry")
		if err != nil {
			t.Fatalf("SendMessage failed: %v", err)
		}

		// Verify retry occurred
		if attempts != 2 {
			t.Errorf("Expected 2 attempts (1 failure + 1 success), got %d", attempts)
		}

		// Verify response received
		messages, err := sessionMgr.GetMessages(ctx, sess.ID)
		if err != nil {
			t.Fatalf("GetMessages failed: %v", err)
		}

		if len(messages) != 2 {
			t.Fatalf("Expected 2 messages, got %d", len(messages))
		}

		if messages[1].Content[0].Text != "Success after retry" {
			t.Errorf("Unexpected response content: %s", messages[1].Content[0].Text)
		}
	})

	t.Run("Provider not found error", func(t *testing.T) {
		ctx := context.Background()

		// Setup temp storage
		tempDir := t.TempDir()
		storePath := filepath.Join(tempDir, "test.db")
		store, err := storage.New(storePath)
		if err != nil {
			t.Fatalf("Failed to create storage: %v", err)
		}
		defer store.Close()

		// Create session manager
		registry := provider.NewRegistry()
		sessionMgr := NewManager(store, registry)

		// Create provider registry (empty)

		// Create completion service
		completionSvc := NewCompletionService(sessionMgr, registry)

		// Create a session with non-existent provider
		sess, err := sessionMgr.Create(ctx, "test-project", "nonexistent", "model")
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Try to send message - should fail
		err = completionSvc.SendMessage(ctx, sess.ID, "This should fail")
		if err == nil {
			t.Error("Expected error for non-existent provider")
		}

		// Verify session state is error
		updatedSess, err := sessionMgr.Get(ctx, sess.ID)
		if err != nil {
			t.Fatalf("Failed to get updated session: %v", err)
		}
		if updatedSess.State != StateError {
			t.Errorf("Expected session state to be error, got %s", updatedSess.State)
		}
	})
}

// Mock implementations for testing

type mockProvider struct {
	name       string
	models     []string
	response   *provider.CompletionResponse
	streamResp provider.Stream
	err        error
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) Models() []string {
	return m.models
}

func (m *mockProvider) Complete(ctx context.Context, req provider.CompletionRequest) (*provider.CompletionResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

func (m *mockProvider) Stream(ctx context.Context, req provider.CompletionRequest) (provider.Stream, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.streamResp, nil
}

type mockStream struct {
	chunks []provider.StreamChunk
	index  int
	err    error
}

func (m *mockStream) Next() bool {
	if m.index >= len(m.chunks) {
		return false
	}
	m.index++
	return true
}

func (m *mockStream) Chunk() provider.StreamChunk {
	if m.index == 0 || m.index > len(m.chunks) {
		return provider.StreamChunk{}
	}
	return m.chunks[m.index-1]
}

func (m *mockStream) Error() error {
	return m.err
}

func (m *mockStream) Close() error {
	return nil
}

type countingProvider struct {
	name         string
	completeFunc func(context.Context, provider.CompletionRequest) (*provider.CompletionResponse, error)
}

func (c *countingProvider) Name() string {
	return c.name
}

func (c *countingProvider) Models() []string {
	return []string{"test-model"}
}

func (c *countingProvider) Complete(ctx context.Context, req provider.CompletionRequest) (*provider.CompletionResponse, error) {
	return c.completeFunc(ctx, req)
}

func (c *countingProvider) Stream(ctx context.Context, req provider.CompletionRequest) (provider.Stream, error) {
	return nil, os.ErrNotExist
}

type apiError struct {
	statusCode int
	message    string
}

func (e *apiError) Error() string {
	return e.message
}
