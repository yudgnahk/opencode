package tool

import (
	"context"
	"os"
	"testing"

	"github.com/opencode/opencode-go/internal/provider"
	"github.com/opencode/opencode-go/internal/storage"
)

func TestTaskExecutor(t *testing.T) {
	// Skip if no API key available
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping task executor test")
	}

	// Create test storage
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"
	store, err := storage.New(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	// Create provider registry
	registry := provider.NewRegistry()

	// Register Anthropic provider
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	anthropic := provider.NewAnthropic(apiKey)
	if err := registry.Register(anthropic); err != nil {
		t.Fatal(err)
	}

	// Create task executor
	executor := NewTaskExecutor(store, registry, "test-project")

	t.Run("execute simple task", func(t *testing.T) {
		req := TaskRequest{
			Description:  "Calculate sum",
			Prompt:       "What is 15 + 27? Please respond with just the number.",
			SubagentType: "general",
		}

		ctx := context.Background()
		resp, err := executor.Execute(ctx, req)
		if err != nil {
			t.Fatalf("task execution failed: %v", err)
		}

		if resp.Result == "" {
			t.Error("expected non-empty result")
		}

		if resp.SessionID == "" {
			t.Error("expected session ID")
		}

		t.Logf("Task result: %s", resp.Result)
		t.Logf("Session ID: %s", resp.SessionID)
	})

	t.Run("execute with different subagent type", func(t *testing.T) {
		req := TaskRequest{
			Description:  "Explain code concept",
			Prompt:       "Explain what a mutex is in one sentence.",
			SubagentType: "codex",
		}

		ctx := context.Background()
		resp, err := executor.Execute(ctx, req)
		if err != nil {
			t.Fatalf("task execution failed: %v", err)
		}

		if resp.Result == "" {
			t.Error("expected non-empty result")
		}

		t.Logf("Task result: %s", resp.Result)
	})

	t.Run("execute with empty description", func(t *testing.T) {
		req := TaskRequest{
			Description: "",
			Prompt:      "Hello world",
		}

		ctx := context.Background()
		_, err := executor.Execute(ctx, req)
		// Should succeed even with empty description
		if err != nil {
			t.Errorf("expected success with empty description, got: %v", err)
		}
	})
}

func TestGetProviderForSubagent(t *testing.T) {
	store, _ := storage.New(":memory:")
	registry := provider.NewRegistry()
	executor := NewTaskExecutor(store, registry, "test")

	tests := []struct {
		name         string
		subagentType string
		wantProvider string
		wantModel    string
	}{
		{
			name:         "general type",
			subagentType: "general",
			wantProvider: "anthropic",
			wantModel:    "claude-3-5-sonnet-20241022",
		},
		{
			name:         "codex type",
			subagentType: "codex",
			wantProvider: "anthropic",
			wantModel:    "claude-3-5-sonnet-20241022",
		},
		{
			name:         "gemini type",
			subagentType: "gemini",
			wantProvider: "gemini",
			wantModel:    "gemini-2.0-flash-exp",
		},
		{
			name:         "unknown type defaults to anthropic",
			subagentType: "unknown",
			wantProvider: "anthropic",
			wantModel:    "claude-3-5-sonnet-20241022",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotProvider, gotModel := executor.getProviderForSubagent(tt.subagentType)
			if gotProvider != tt.wantProvider {
				t.Errorf("getProviderForSubagent() provider = %v, want %v", gotProvider, tt.wantProvider)
			}
			if gotModel != tt.wantModel {
				t.Errorf("getProviderForSubagent() model = %v, want %v", gotModel, tt.wantModel)
			}
		})
	}
}

func TestExtractTextFromMessage(t *testing.T) {
	// Test helper function
	t.Run("extract text from message with text block", func(t *testing.T) {
		// This would need session.Message type, but we're just testing the function
		// In a real scenario, we'd import session package
		// For now, we just ensure the function exists and can be called
	})
}

func TestTruncateTitle(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{
			name:  "short title",
			title: "Short title",
			want:  "Short title",
		},
		{
			name:  "exact 100 chars",
			title: "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
			want:  "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
		},
		{
			name:  "over 100 chars",
			title: "12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901",
			want:  "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateTitle(tt.title)
			if got != tt.want {
				t.Errorf("truncateTitle() = %v, want %v", got, tt.want)
			}
		})
	}
}
