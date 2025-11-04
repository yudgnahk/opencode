package tool

import (
	"github.com/opencode/opencode-go/internal/lsp"
)

// ToolExecutor handles all tool operations
type ToolExecutor struct {
	workingDir string
	lspClient  *lsp.Client
}

// NewToolExecutor creates a new tool executor
func NewToolExecutor(workingDir string) *ToolExecutor {
	return &ToolExecutor{
		workingDir: workingDir,
		lspClient:  lsp.NewClient(),
	}
}

// GetLSPClient returns the LSP client
func (t *ToolExecutor) GetLSPClient() *lsp.Client {
	return t.lspClient
}

// StopLSPServers stops all LSP servers on cleanup
func (t *ToolExecutor) StopLSPServers() {
	if t.lspClient != nil {
		t.lspClient.StopAll()
	}
}
