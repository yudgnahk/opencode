package tool

import (
	"fmt"
	"os"
	"path/filepath"
)

type WriteRequest struct {
	FilePath string `json:"filePath"`
	Content  string `json:"content"`
}

type WriteResponse struct {
	Success bool   `json:"success"`
	Path    string `json:"path"`
}

func (t *ToolExecutor) Write(req WriteRequest) (*WriteResponse, error) {
	if req.FilePath == "" {
		return nil, fmt.Errorf("filePath is required")
	}
	if req.Content == "" {
		return nil, fmt.Errorf("content is required")
	}

	// Ensure directory exists
	dir := filepath.Dir(req.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(req.FilePath, []byte(req.Content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &WriteResponse{
		Success: true,
		Path:    req.FilePath,
	}, nil
}
