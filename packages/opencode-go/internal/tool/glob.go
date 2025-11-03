package tool

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/gobwas/glob"
)

type GlobRequest struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path"` // Optional, defaults to CWD
}

type GlobResponse struct {
	Files []string `json:"files"`
}

func (t *ToolExecutor) Glob(req GlobRequest) (*GlobResponse, error) {
	if req.Pattern == "" {
		return nil, fmt.Errorf("pattern is required")
	}

	basePath := req.Path
	if basePath == "" {
		basePath = "."
	}

	// Compile glob pattern
	g, err := glob.Compile(req.Pattern, '/')
	if err != nil {
		return nil, err
	}

	var files []string

	// Walk directory tree
	err = filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if path matches pattern
		relPath, _ := filepath.Rel(basePath, path)
		if g.Match(relPath) {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort by modification time (newest first)
	sort.Slice(files, func(i, j int) bool {
		infoI, errI := os.Stat(files[i])
		infoJ, errJ := os.Stat(files[j])
		if errI != nil || errJ != nil {
			return false
		}
		return infoI.ModTime().After(infoJ.ModTime())
	})

	return &GlobResponse{Files: files}, nil
}
