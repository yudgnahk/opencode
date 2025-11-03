package tool

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

type GrepRequest struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path"`    // Optional
	Include string `json:"include"` // File pattern like "*.go"
}

type GrepResponse struct {
	Matches []GrepMatch `json:"matches"`
}

type GrepMatch struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Text string `json:"text"`
}

func (t *ToolExecutor) Grep(req GrepRequest) (*GrepResponse, error) {
	if req.Pattern == "" {
		return nil, fmt.Errorf("pattern is required")
	}

	// Check if ripgrep is available
	if _, err := exec.LookPath("rg"); err != nil {
		return nil, fmt.Errorf("ripgrep not found: %w", err)
	}

	args := []string{
		"--json",       // JSON output
		"--no-heading", // Don't group by file
		req.Pattern,
	}

	if req.Include != "" {
		args = append(args, "--glob", req.Include)
	}

	if req.Path != "" {
		args = append(args, req.Path)
	}

	cmd := exec.Command("rg", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run command (ignore exit code - rg exits 1 if no matches)
	_ = cmd.Run()

	// Parse JSON output
	var matches []GrepMatch
	scanner := bufio.NewScanner(&stdout)
	for scanner.Scan() {
		var result struct {
			Type string `json:"type"`
			Data struct {
				Path struct {
					Text string `json:"text"`
				} `json:"path"`
				Lines struct {
					Text string `json:"text"`
				} `json:"lines"`
				LineNumber int `json:"line_number"`
			} `json:"data"`
		}

		if err := json.Unmarshal(scanner.Bytes(), &result); err != nil {
			continue
		}

		if result.Type == "match" {
			matches = append(matches, GrepMatch{
				File: result.Data.Path.Text,
				Line: result.Data.LineNumber,
				Text: result.Data.Lines.Text,
			})
		}
	}

	return &GrepResponse{Matches: matches}, nil
}
