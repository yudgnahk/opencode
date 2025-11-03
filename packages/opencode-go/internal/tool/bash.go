package tool

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

type BashRequest struct {
	Command     string `json:"command"`
	Timeout     int    `json:"timeout"`     // milliseconds
	Description string `json:"description"` // Optional description
}

type BashResponse struct {
	Output   string `json:"output"`
	Error    string `json:"error"`
	ExitCode int    `json:"exitCode"`
	Duration int64  `json:"duration"` // milliseconds
}

func (t *ToolExecutor) Bash(req BashRequest) (*BashResponse, error) {
	if req.Command == "" {
		return nil, fmt.Errorf("command is required")
	}

	// Default timeout: 2 minutes
	timeout := 120000
	if req.Timeout > 0 {
		timeout = req.Timeout
	}
	if timeout > 600000 { // Max 10 minutes
		timeout = 600000
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
	defer cancel()

	// Execute command with shell
	cmd := exec.CommandContext(ctx, "sh", "-c", req.Command)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start).Milliseconds()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("command timed out after %dms", timeout)
		} else {
			return nil, fmt.Errorf("command execution failed: %w", err)
		}
	}

	// Truncate output if too large (30,000 chars max)
	output := stdout.String()
	if len(output) > 30000 {
		output = output[:30000] + "\n[Output truncated...]"
	}

	errOutput := stderr.String()
	if len(errOutput) > 30000 {
		errOutput = errOutput[:30000] + "\n[Error output truncated...]"
	}

	return &BashResponse{
		Output:   output,
		Error:    errOutput,
		ExitCode: exitCode,
		Duration: duration,
	}, nil
}
