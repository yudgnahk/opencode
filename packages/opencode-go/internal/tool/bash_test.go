package tool

import (
	"strings"
	"testing"
)

func TestBash(t *testing.T) {
	tmpDir := t.TempDir()
	executor := newTestToolExecutor(tmpDir)

	tests := []struct {
		name    string
		req     BashRequest
		wantErr bool
		check   func(resp *BashResponse) bool
	}{
		{
			name: "simple echo",
			req: BashRequest{
				Command:     "echo hello",
				Description: "echo test",
			},
			wantErr: false,
			check: func(resp *BashResponse) bool {
				return strings.TrimSpace(resp.Output) == "hello" && resp.ExitCode == 0
			},
		},
		{
			name: "command with error",
			req: BashRequest{
				Command:     "ls /nonexistent",
				Description: "error test",
			},
			wantErr: false,
			check: func(resp *BashResponse) bool {
				return resp.ExitCode != 0 && resp.Error != ""
			},
		},
		{
			name: "with timeout",
			req: BashRequest{
				Command:     "echo quick",
				Description: "timeout test",
				Timeout:     5000,
			},
			wantErr: false,
			check: func(resp *BashResponse) bool {
				return resp.ExitCode == 0
			},
		},
		{
			name: "missing command",
			req: BashRequest{
				Command:     "",
				Description: "empty command",
			},
			wantErr: true,
		},
		{
			name: "timeout exceeds max",
			req: BashRequest{
				Command:     "echo test",
				Description: "long timeout",
				Timeout:     700000,
			},
			wantErr: false, // Timeout is capped at 600000, not an error
			check: func(resp *BashResponse) bool {
				return resp.ExitCode == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := executor.Bash(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if resp == nil {
					t.Error("Bash() returned nil response")
					return
				}
				if tt.check != nil && !tt.check(resp) {
					t.Errorf("Bash() check failed: %+v", resp)
				}
			}
		})
	}
}
