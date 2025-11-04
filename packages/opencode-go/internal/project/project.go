package project

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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

	// Detect git repository first
	gitInfo := detectGit(absPath)

	// Generate ID from git root if available
	var projectID string
	if gitInfo != nil {
		projectID = generateID(gitInfo.Root)
	} else {
		projectID = "global"
	}

	project := &Project{
		ID:   projectID,
		Path: absPath,
		Name: name,
		Git:  gitInfo,
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

func generateID(gitRoot string) string {
	// Match TypeScript behavior: git rev-list --max-parents=0 --all
	cmd := exec.Command("git", "rev-list", "--max-parents=0", "--all")
	cmd.Dir = gitRoot
	output, err := cmd.Output()
	if err != nil {
		return "global"
	}

	// Split by newlines, filter empty, trim, and sort
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var commits []string
	for _, line := range lines {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			commits = append(commits, trimmed)
		}
	}

	if len(commits) == 0 {
		return "global"
	}

	sort.Strings(commits)
	return commits[0]
}
