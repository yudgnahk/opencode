package tool

import (
	"github.com/opencode/opencode-go/internal/provider"
	"github.com/opencode/opencode-go/internal/storage"
)

// newTestToolExecutor creates a tool executor for testing
// Tests that don't need task spawning can use this
func newTestToolExecutor(workingDir string) *ToolExecutor {
	// Create temporary in-memory storage for tests
	tmpStorage, _ := storage.New(":memory:")
	registry := provider.NewRegistry()

	return NewToolExecutor(workingDir, tmpStorage, registry, "test-project")
}
