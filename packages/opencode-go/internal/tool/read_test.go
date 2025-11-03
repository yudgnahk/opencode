package tool

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRead(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Write test content
	content := "line1\nline2\nline3\nline4\nline5\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	executor := NewToolExecutor(tmpDir)

	tests := []struct {
		name    string
		req     ReadRequest
		wantErr bool
		check   func(resp *ReadResponse) bool
	}{
		{
			name: "read entire file",
			req: ReadRequest{
				FilePath: testFile,
			},
			wantErr: false,
			check: func(resp *ReadResponse) bool {
				return len(resp.Lines) == 5 && resp.Total == 5
			},
		},
		{
			name: "read with offset",
			req: ReadRequest{
				FilePath: testFile,
				Offset:   2,
			},
			wantErr: false,
			check: func(resp *ReadResponse) bool {
				return len(resp.Lines) == 3 && resp.Total == 5
			},
		},
		{
			name: "read with limit",
			req: ReadRequest{
				FilePath: testFile,
				Limit:    2,
			},
			wantErr: false,
			check: func(resp *ReadResponse) bool {
				// Total equals limit when we break early
				return len(resp.Lines) == 2 && resp.Total == 2
			},
		},
		{
			name: "read with offset and limit",
			req: ReadRequest{
				FilePath: testFile,
				Offset:   1,
				Limit:    2,
			},
			wantErr: false,
			check: func(resp *ReadResponse) bool {
				// Total equals offset+limit when we break early
				return len(resp.Lines) == 2 && resp.Total == 3
			},
		},
		{
			name: "read nonexistent file",
			req: ReadRequest{
				FilePath: filepath.Join(tmpDir, "nonexistent.txt"),
			},
			wantErr: true,
		},
		{
			name: "missing filepath",
			req: ReadRequest{
				FilePath: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := executor.Read(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil && !tt.check(resp) {
				t.Errorf("Read() response check failed: %+v", resp)
			}
		})
	}
}

func TestReadLongLines(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "long.txt")

	// Create line longer than 2000 chars
	longLine := make([]byte, 2500)
	for i := range longLine {
		longLine[i] = 'a'
	}
	content := string(longLine) + "\n"

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	executor := NewToolExecutor(tmpDir)
	resp, err := executor.Read(ReadRequest{FilePath: testFile})
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	// Line should be truncated to 2000 chars + "..." + newline
	if len(resp.Content) > 2010 {
		t.Errorf("Long line not truncated properly, got length %d", len(resp.Content))
	}
}
