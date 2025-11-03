package tool

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEdit(t *testing.T) {
	tmpDir := t.TempDir()
	executor := NewToolExecutor(tmpDir)

	tests := []struct {
		name       string
		initial    string
		req        EditRequest
		wantErr    bool
		wantResult string
	}{
		{
			name:    "replace single occurrence",
			initial: "hello world\nhello universe",
			req: EditRequest{
				FilePath:  filepath.Join(tmpDir, "single.txt"),
				OldString: "world",
				NewString: "everyone",
			},
			wantErr:    false,
			wantResult: "hello everyone\nhello universe",
		},
		{
			name:    "replace all occurrences",
			initial: "hello world\nhello world again",
			req: EditRequest{
				FilePath:   filepath.Join(tmpDir, "multiple.txt"),
				OldString:  "world",
				NewString:  "everyone",
				ReplaceAll: true,
			},
			wantErr:    false,
			wantResult: "hello everyone\nhello everyone again",
		},
		{
			name:    "string not found",
			initial: "hello world",
			req: EditRequest{
				FilePath:  filepath.Join(tmpDir, "notfound.txt"),
				OldString: "universe",
				NewString: "everyone",
			},
			wantErr: true,
		},
		{
			name:    "multiple occurrences without replaceAll",
			initial: "hello world\nhello world again",
			req: EditRequest{
				FilePath:  filepath.Join(tmpDir, "ambiguous.txt"),
				OldString: "world",
				NewString: "everyone",
			},
			wantErr: true,
		},
		{
			name:    "missing filepath",
			initial: "",
			req: EditRequest{
				FilePath:  "",
				OldString: "old",
				NewString: "new",
			},
			wantErr: true,
		},
		{
			name:    "missing oldString",
			initial: "content",
			req: EditRequest{
				FilePath:  filepath.Join(tmpDir, "test.txt"),
				OldString: "",
				NewString: "new",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file if needed
			if tt.initial != "" {
				if err := os.WriteFile(tt.req.FilePath, []byte(tt.initial), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			resp, err := executor.Edit(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Edit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if resp == nil {
					t.Error("Edit() returned nil response")
					return
				}
				// Read the file and verify content
				content, err := os.ReadFile(tt.req.FilePath)
				if err != nil {
					t.Errorf("Failed to read edited file: %v", err)
					return
				}
				if string(content) != tt.wantResult {
					t.Errorf("Edit() result = %q, want %q", string(content), tt.wantResult)
				}
			}
		})
	}
}
