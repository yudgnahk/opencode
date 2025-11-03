package provider

import (
	"context"
	"errors"
	"testing"
	"time"
)

// MockProvider for testing
type MockProvider struct {
	name       string
	models     []string
	response   *CompletionResponse
	streamResp *MockStream
	err        error
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) Models() []string {
	return m.models
}

func (m *MockProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

func (m *MockProvider) Stream(ctx context.Context, req CompletionRequest) (Stream, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.streamResp, nil
}

// CountingProvider for testing retries
type CountingProvider struct {
	name         string
	completeFunc func(context.Context, CompletionRequest) (*CompletionResponse, error)
}

func (c *CountingProvider) Name() string {
	return c.name
}

func (c *CountingProvider) Models() []string {
	return []string{"test-model"}
}

func (c *CountingProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	return c.completeFunc(ctx, req)
}

func (c *CountingProvider) Stream(ctx context.Context, req CompletionRequest) (Stream, error) {
	return nil, errors.New("not implemented")
}

// MockStream for testing
type MockStream struct {
	chunks []StreamChunk
	index  int
	err    error
}

func (m *MockStream) Next() bool {
	if m.index >= len(m.chunks) {
		return false
	}
	m.index++
	return true
}

func (m *MockStream) Chunk() StreamChunk {
	if m.index == 0 || m.index > len(m.chunks) {
		return StreamChunk{}
	}
	return m.chunks[m.index-1]
}

func (m *MockStream) Error() error {
	return m.err
}

func (m *MockStream) Close() error {
	return nil
}

func TestRegistry(t *testing.T) {
	t.Run("Register and Get", func(t *testing.T) {
		registry := NewRegistry()
		provider := &MockProvider{
			name:   "test",
			models: []string{"model-1", "model-2"},
		}

		err := registry.Register(provider)
		if err != nil {
			t.Fatalf("Register failed: %v", err)
		}

		retrieved, err := registry.Get("test")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if retrieved.Name() != "test" {
			t.Errorf("Expected provider name 'test', got '%s'", retrieved.Name())
		}
	})

	t.Run("Duplicate registration", func(t *testing.T) {
		registry := NewRegistry()
		provider := &MockProvider{name: "test"}

		registry.Register(provider)
		err := registry.Register(provider)

		if err == nil {
			t.Error("Expected error for duplicate registration")
		}
	})

	t.Run("Get non-existent provider", func(t *testing.T) {
		registry := NewRegistry()

		_, err := registry.Get("nonexistent")
		if err == nil {
			t.Error("Expected error for non-existent provider")
		}
	})

	t.Run("List providers", func(t *testing.T) {
		registry := NewRegistry()
		registry.Register(&MockProvider{name: "provider1"})
		registry.Register(&MockProvider{name: "provider2"})

		names := registry.List()
		if len(names) != 2 {
			t.Errorf("Expected 2 providers, got %d", len(names))
		}
	})

	t.Run("List models", func(t *testing.T) {
		registry := NewRegistry()
		registry.Register(&MockProvider{
			name:   "provider1",
			models: []string{"model-a", "model-b"},
		})
		registry.Register(&MockProvider{
			name:   "provider2",
			models: []string{"model-c"},
		})

		models := registry.ListModels()
		if len(models) != 2 {
			t.Errorf("Expected 2 providers in models map, got %d", len(models))
		}
		if len(models["provider1"]) != 2 {
			t.Errorf("Expected 2 models for provider1, got %d", len(models["provider1"]))
		}
	})

	t.Run("Unregister provider", func(t *testing.T) {
		registry := NewRegistry()
		registry.Register(&MockProvider{name: "test"})
		registry.Unregister("test")

		_, err := registry.Get("test")
		if err == nil {
			t.Error("Expected error after unregistering provider")
		}
	})
}

func TestRetryableProvider(t *testing.T) {
	t.Run("Success on first try", func(t *testing.T) {
		mockProvider := &MockProvider{
			name: "test",
			response: &CompletionResponse{
				ID: "test-id",
				Content: []ContentBlock{{
					Type: "text",
					Text: "Hello",
				}},
			},
		}

		config := RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 10 * time.Millisecond,
			MaxBackoff:     100 * time.Millisecond,
			Multiplier:     2.0,
		}

		retryable := NewRetryableProvider(mockProvider, config)

		ctx := context.Background()
		req := CompletionRequest{
			Messages: []Message{{
				Role: "user",
				Content: []ContentBlock{{
					Type: "text",
					Text: "Hi",
				}},
			}},
			Model:     "test-model",
			MaxTokens: 100,
		}

		resp, err := retryable.Complete(ctx, req)
		if err != nil {
			t.Fatalf("Expected success, got error: %v", err)
		}

		if resp.ID != "test-id" {
			t.Errorf("Expected response ID 'test-id', got '%s'", resp.ID)
		}
	})

	t.Run("Retry on retryable error", func(t *testing.T) {
		attempts := 0

		// Create a custom provider that counts attempts
		mockProvider := &CountingProvider{
			name: "test",
			completeFunc: func(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
				attempts++
				if attempts == 1 {
					return nil, errors.New("timeout: connection failed")
				}
				return &CompletionResponse{ID: "success"}, nil
			},
		}

		config := RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 10 * time.Millisecond,
			MaxBackoff:     100 * time.Millisecond,
			Multiplier:     2.0,
		}

		retryable := NewRetryableProvider(mockProvider, config)

		ctx := context.Background()
		req := CompletionRequest{
			Messages:  []Message{},
			Model:     "test-model",
			MaxTokens: 100,
		}

		resp, err := retryable.Complete(ctx, req)
		if err != nil {
			t.Fatalf("Expected success after retry, got error: %v", err)
		}

		if attempts != 2 {
			t.Errorf("Expected 2 attempts, got %d", attempts)
		}

		if resp.ID != "success" {
			t.Errorf("Expected success response")
		}
	})

	t.Run("Max retries exceeded", func(t *testing.T) {
		mockProvider := &MockProvider{
			name: "test",
			err:  errors.New("rate limit: 429 too many requests"),
		}

		config := RetryConfig{
			MaxRetries:     2,
			InitialBackoff: 10 * time.Millisecond,
			MaxBackoff:     100 * time.Millisecond,
			Multiplier:     2.0,
		}

		retryable := NewRetryableProvider(mockProvider, config)

		ctx := context.Background()
		req := CompletionRequest{
			Messages:  []Message{},
			Model:     "test-model",
			MaxTokens: 100,
		}

		_, err := retryable.Complete(ctx, req)
		if err == nil {
			t.Error("Expected error after max retries")
		}
	})

	t.Run("Context cancellation", func(t *testing.T) {
		mockProvider := &MockProvider{
			name: "test",
			err:  errors.New("timeout"),
		}

		config := RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 100 * time.Millisecond,
			MaxBackoff:     1 * time.Second,
			Multiplier:     2.0,
		}

		retryable := NewRetryableProvider(mockProvider, config)

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		req := CompletionRequest{
			Messages:  []Message{},
			Model:     "test-model",
			MaxTokens: 100,
		}

		_, err := retryable.Complete(ctx, req)
		if err == nil {
			t.Error("Expected error due to context cancellation")
		}

		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Expected context deadline exceeded error, got: %v", err)
		}
	})
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"timeout", errors.New("connection timeout"), true},
		{"rate limit", errors.New("rate limit exceeded"), true},
		{"429 status", errors.New("HTTP 429"), true},
		{"503 status", errors.New("Service unavailable: 503"), true},
		{"500 status", errors.New("Internal server error 500"), true},
		{"connection reset", errors.New("connection reset by peer"), true},
		{"temporary failure", errors.New("temporary failure"), true},
		{"non-retryable", errors.New("invalid API key"), false},
		{"auth error", errors.New("unauthorized"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryable(tt.err)
			if result != tt.expected {
				t.Errorf("isRetryable(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}
