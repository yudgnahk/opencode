package tool

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGlob(t *testing.T) {
	tmpDir := t.TempDir()
	executor := NewToolExecutor(tmpDir)

	// Create test files
	files := []string{
		"file1.txt",
		"file2.txt",
		"test.go",
		"main.go",
		"dir1/file3.txt",
		"dir1/test.go",
		"dir2/main.go",
	}

	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		os.MkdirAll(filepath.Dir(path), 0755)
		os.WriteFile(path, []byte("test"), 0644)
	}

	tests := []struct {
		name      string
		req       GlobRequest
		wantErr   bool
		wantCount int
		wantMatch func(files []string) bool
	}{
		{
			name: "match all txt files",
			req: GlobRequest{
				Pattern: "*.txt",
				Path:    tmpDir,
			},
			wantErr:   false,
			wantCount: 2,
			wantMatch: func(files []string) bool {
				return len(files) == 2
			},
		},
		{
			name: "match all go files recursively",
			req: GlobRequest{
				Pattern: "**/*.go",
				Path:    tmpDir,
			},
			wantErr:   false,
			wantCount: 2, // Only matches files in subdirectories
			wantMatch: func(files []string) bool {
				return len(files) == 2
			},
		},
		{
			name: "match files in specific directory",
			req: GlobRequest{
				Pattern: "dir1/*",
				Path:    tmpDir,
			},
			wantErr:   false,
			wantCount: 2,
			wantMatch: func(files []string) bool {
				return len(files) == 2
			},
		},
		{
			name: "no matches",
			req: GlobRequest{
				Pattern: "*.md",
				Path:    tmpDir,
			},
			wantErr:   false,
			wantCount: 0,
			wantMatch: func(files []string) bool {
				return len(files) == 0
			},
		},
		{
			name: "missing pattern",
			req: GlobRequest{
				Pattern: "",
				Path:    tmpDir,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := executor.Glob(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Glob() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if resp == nil {
					t.Error("Glob() returned nil response")
					return
				}
				if tt.wantMatch != nil && !tt.wantMatch(resp.Files) {
					t.Errorf("Glob() files = %v, count want %d", resp.Files, tt.wantCount)
				}
			}
		})
	}
}
