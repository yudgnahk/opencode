package tool

import (
	"github.com/opencode/opencode-go/internal/lsp"
	"github.com/opencode/opencode-go/internal/provider"
	"github.com/opencode/opencode-go/internal/storage"
)

// ToolExecutor handles all tool operations
type ToolExecutor struct {
	workingDir   string
	lspClient    *lsp.Client
	taskExecutor *TaskExecutor
}

// NewToolExecutor creates a new tool executor
func NewToolExecutor(workingDir string, storage *storage.Storage, providerRegistry *provider.Registry, projectID string) *ToolExecutor {
	return &ToolExecutor{
		workingDir:   workingDir,
		lspClient:    lsp.NewClient(),
		taskExecutor: NewTaskExecutor(storage, providerRegistry, projectID),
	}
}

// GetLSPClient returns the LSP client
func (t *ToolExecutor) GetLSPClient() *lsp.Client {
	return t.lspClient
}

// GetTaskExecutor returns the task executor
func (t *ToolExecutor) GetTaskExecutor() *TaskExecutor {
	return t.taskExecutor
}

// StopLSPServers stops all LSP servers on cleanup
func (t *ToolExecutor) StopLSPServers() {
	if t.lspClient != nil {
		t.lspClient.StopAll()
	}
}
