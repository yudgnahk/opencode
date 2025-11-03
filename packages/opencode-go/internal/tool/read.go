package tool

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type ReadRequest struct {
	FilePath string `json:"filePath"`
	Offset   int    `json:"offset"`
	Limit    int    `json:"limit"`
}

type ReadResponse struct {
	Content string   `json:"content"`
	Lines   []string `json:"lines"`
	Total   int      `json:"total"`
}

func (t *ToolExecutor) Read(req ReadRequest) (*ReadResponse, error) {
	if req.FilePath == "" {
		return nil, fmt.Errorf("filePath is required")
	}

	file, err := os.Open(req.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Default limit
	if req.Limit == 0 {
		req.Limit = 2000
	}

	scanner := bufio.NewScanner(file)
	// Set max token size to handle long lines (up to 2000 chars)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 2000)

	var lines []string
	lineNum := 0

	for scanner.Scan() {
		if lineNum >= req.Offset && lineNum < req.Offset+req.Limit {
			line := scanner.Text()
			// Truncate lines longer than 2000 characters
			if len(line) > 2000 {
				line = line[:2000] + "..."
			}
			// Format with line numbers (like cat -n)
			lines = append(lines, fmt.Sprintf("%5d\t%s", lineNum+1, line))
		}
		lineNum++
		if lineNum >= req.Offset+req.Limit {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	content := strings.Join(lines, "\n")

	return &ReadResponse{
		Content: content,
		Lines:   lines,
		Total:   lineNum,
	}, nil
}
