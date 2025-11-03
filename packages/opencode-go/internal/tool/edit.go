package tool

import (
	"fmt"
	"os"
	"strings"
)

type EditRequest struct {
	FilePath   string `json:"filePath"`
	OldString  string `json:"oldString"`
	NewString  string `json:"newString"`
	ReplaceAll bool   `json:"replaceAll"`
}

type EditResponse struct {
	Success  bool `json:"success"`
	Replaced int  `json:"replaced"`
}

func (t *ToolExecutor) Edit(req EditRequest) (*EditResponse, error) {
	if req.FilePath == "" {
		return nil, fmt.Errorf("filePath is required")
	}
	if req.OldString == "" {
		return nil, fmt.Errorf("oldString is required")
	}
	if req.NewString == "" {
		return nil, fmt.Errorf("newString is required")
	}

	// Read file
	content, err := os.ReadFile(req.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	contentStr := string(content)

	// Count occurrences
	count := strings.Count(contentStr, req.OldString)

	if count == 0 {
		return nil, fmt.Errorf("oldString not found in content")
	}

	if count > 1 && !req.ReplaceAll {
		return nil, fmt.Errorf("oldString found multiple times and requires more code context to uniquely identify the intended match")
	}

	// Perform replacement
	var newContent string
	replaced := 1
	if req.ReplaceAll {
		newContent = strings.ReplaceAll(contentStr, req.OldString, req.NewString)
		replaced = count
	} else {
		newContent = strings.Replace(contentStr, req.OldString, req.NewString, 1)
	}

	// Write back
	if err := os.WriteFile(req.FilePath, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &EditResponse{
		Success:  true,
		Replaced: replaced,
	}, nil
}
