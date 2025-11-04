package provider

import (
	"context"
	"fmt"
	"math"
	"time"
)

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Multiplier     float64
}

// DefaultRetryConfig returns sensible defaults for API retries
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		Multiplier:     2.0,
	}
}

// RetryableProvider wraps a provider with retry logic
type RetryableProvider struct {
	provider Provider
	config   RetryConfig
}

// NewRetryableProvider creates a provider with retry capability
func NewRetryableProvider(provider Provider, config RetryConfig) *RetryableProvider {
	return &RetryableProvider{
		provider: provider,
		config:   config,
	}
}

func (p *RetryableProvider) Name() string {
	return p.provider.Name()
}

func (p *RetryableProvider) Models() []string {
	return p.provider.Models()
}

func (p *RetryableProvider) SetModelsDevRegistry(registry ModelsDevRegistry) {
	p.provider.SetModelsDevRegistry(registry)
}

func (p *RetryableProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	var lastErr error
	backoff := p.config.InitialBackoff

	for attempt := 0; attempt <= p.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}

			// Exponential backoff with max limit
			backoff = time.Duration(math.Min(
				float64(backoff)*p.config.Multiplier,
				float64(p.config.MaxBackoff),
			))
		}

		resp, err := p.provider.Complete(ctx, req)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Don't retry on context cancellation
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Check if error is retryable
		if !isRetryable(err) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (p *RetryableProvider) Stream(ctx context.Context, req CompletionRequest) (Stream, error) {
	var lastErr error
	backoff := p.config.InitialBackoff

	for attempt := 0; attempt <= p.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}

			backoff = time.Duration(math.Min(
				float64(backoff)*p.config.Multiplier,
				float64(p.config.MaxBackoff),
			))
		}

		stream, err := p.provider.Stream(ctx, req)
		if err == nil {
			return stream, nil
		}

		lastErr = err

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if !isRetryable(err) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// isRetryable determines if an error should trigger a retry
// This is a simple heuristic - can be enhanced based on provider-specific errors
func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Add logic to check for rate limiting, timeouts, temporary errors, etc.
	// For now, we'll be conservative and not retry most errors
	msg := err.Error()

	// Common retryable patterns
	retryablePatterns := []string{
		"timeout",
		"rate limit",
		"429",
		"503",
		"500",
		"connection reset",
		"connection refused",
		"temporary",
	}

	for _, pattern := range retryablePatterns {
		if contains(msg, pattern) {
			return true
		}
	}

	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
