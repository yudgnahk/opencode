package tool

// ToolExecutor handles all tool operations
type ToolExecutor struct {
	workingDir string
}

// NewToolExecutor creates a new tool executor
func NewToolExecutor(workingDir string) *ToolExecutor {
	return &ToolExecutor{
		workingDir: workingDir,
	}
}
