package project

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Project struct {
	ID   string   `json:"id"`
	Path string   `json:"path"`
	Name string   `json:"name"`
	Git  *GitInfo `json:"git,omitempty"`
}

type GitInfo struct {
	Root     string `json:"root"`
	Branch   string `json:"branch"`
	Worktree bool   `json:"worktree"`
}

func Detect(path string) (*Project, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	name := filepath.Base(absPath)

	project := &Project{
		ID:   generateID(absPath),
		Path: absPath,
		Name: name,
	}

	// Detect git repository
	if gitInfo := detectGit(absPath); gitInfo != nil {
		project.Git = gitInfo
	}

	return project, nil
}

func detectGit(path string) *GitInfo {
	gitDir := findGitDir(path)
	if gitDir == "" {
		return nil
	}

	// Check if worktree
	worktree := false
	if info, err := os.Stat(filepath.Join(gitDir, "commondir")); err == nil && !info.IsDir() {
		worktree = true
	}

	// Get current branch
	branch := getCurrentBranch(gitDir)

	return &GitInfo{
		Root:     filepath.Dir(gitDir),
		Branch:   branch,
		Worktree: worktree,
	}
}

func findGitDir(path string) string {
	current := path
	for {
		gitPath := filepath.Join(current, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return gitPath
		}

		parent := filepath.Dir(current)
		if parent == current {
			return ""
		}
		current = parent
	}
}

func getCurrentBranch(gitDir string) string {
	headFile := filepath.Join(gitDir, "HEAD")
	content, err := os.ReadFile(headFile)
	if err != nil {
		return "unknown"
	}

	// Parse "ref: refs/heads/main"
	line := strings.TrimSpace(string(content))
	if strings.HasPrefix(line, "ref: ") {
		ref := strings.TrimPrefix(line, "ref: ")
		return filepath.Base(ref)
	}

	return "detached"
}

func generateID(path string) string {
	hash := sha256.Sum256([]byte(path))
	return fmt.Sprintf("%x", hash[:8])
}
