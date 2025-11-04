package tool

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWrite(t *testing.T) {
	tmpDir := t.TempDir()
	executor := newTestToolExecutor(tmpDir)

	tests := []struct {
		name    string
		req     WriteRequest
		wantErr bool
		check   func(filePath string) bool
	}{
		{
			name: "write new file",
			req: WriteRequest{
				FilePath: filepath.Join(tmpDir, "new.txt"),
				Content:  "hello world",
			},
			wantErr: false,
			check: func(filePath string) bool {
				content, err := os.ReadFile(filePath)
				return err == nil && string(content) == "hello world"
			},
		},
		{
			name: "write to nested directory",
			req: WriteRequest{
				FilePath: filepath.Join(tmpDir, "nested", "dir", "file.txt"),
				Content:  "nested content",
			},
			wantErr: false,
			check: func(filePath string) bool {
				content, err := os.ReadFile(filePath)
				return err == nil && string(content) == "nested content"
			},
		},
		{
			name: "overwrite existing file",
			req: WriteRequest{
				FilePath: filepath.Join(tmpDir, "existing.txt"),
				Content:  "new content",
			},
			wantErr: false,
			check: func(filePath string) bool {
				content, err := os.ReadFile(filePath)
				return err == nil && string(content) == "new content"
			},
		},
		{
			name: "missing filepath",
			req: WriteRequest{
				FilePath: "",
				Content:  "content",
			},
			wantErr: true,
		},
		{
			name: "empty content",
			req: WriteRequest{
				FilePath: filepath.Join(tmpDir, "empty.txt"),
				Content:  "",
			},
			wantErr: true,
		},
	}

	// Create file for overwrite test
	existingFile := filepath.Join(tmpDir, "existing.txt")
	os.WriteFile(existingFile, []byte("old content"), 0644)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := executor.Write(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if resp == nil {
					t.Error("Write() returned nil response")
					return
				}
				if tt.check != nil && !tt.check(tt.req.FilePath) {
					t.Errorf("Write() file check failed for %s", tt.req.FilePath)
				}
			}
		})
	}
}
