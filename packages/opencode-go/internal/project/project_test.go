package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetect(t *testing.T) {
	tmpDir := t.TempDir()

	project, err := Detect(tmpDir)
	if err != nil {
		t.Fatalf("Failed to detect project: %v", err)
	}

	if project == nil {
		t.Fatal("Expected project, got nil")
	}

	// Check basic fields
	if project.Path == "" {
		t.Error("Project path is empty")
	}

	if project.Name == "" {
		t.Error("Project name is empty")
	}

	if project.ID == "" {
		t.Error("Project ID is empty")
	}

	// Should have absolute path
	if !filepath.IsAbs(project.Path) {
		t.Errorf("Expected absolute path, got: %s", project.Path)
	}

	// Name should be the directory name
	expectedName := filepath.Base(tmpDir)
	if project.Name != expectedName {
		t.Errorf("Expected name %s, got %s", expectedName, project.Name)
	}
}

func TestDetect_WithGit(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create HEAD file with branch reference
	headContent := "ref: refs/heads/main\n"
	err = os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte(headContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	project, err := Detect(tmpDir)
	if err != nil {
		t.Fatalf("Failed to detect project: %v", err)
	}

	// Should detect git
	if project.Git == nil {
		t.Fatal("Expected git info, got nil")
	}

	if project.Git.Root != tmpDir {
		t.Errorf("Expected git root %s, got %s", tmpDir, project.Git.Root)
	}

	if project.Git.Branch != "main" {
		t.Errorf("Expected branch main, got %s", project.Git.Branch)
	}

	if project.Git.Worktree {
		t.Error("Expected worktree to be false")
	}
}

func TestDetect_NoGit(t *testing.T) {
	tmpDir := t.TempDir()

	project, err := Detect(tmpDir)
	if err != nil {
		t.Fatalf("Failed to detect project: %v", err)
	}

	// Should not have git info
	if project.Git != nil {
		t.Error("Expected no git info for non-git directory")
	}
}

func TestDetect_RelativePath(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp dir
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Detect with relative path
	project, err := Detect(".")
	if err != nil {
		t.Fatalf("Failed to detect project: %v", err)
	}

	// Should resolve to absolute path
	if !filepath.IsAbs(project.Path) {
		t.Errorf("Expected absolute path, got: %s", project.Path)
	}

	// On macOS, tmpDir might have symlinks (e.g., /var -> /private/var)
	// So we evaluate both and compare
	expectedPath, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		expectedPath = tmpDir
	}

	actualPath, err := filepath.EvalSymlinks(project.Path)
	if err != nil {
		actualPath = project.Path
	}

	if actualPath != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, actualPath)
	}
}

func TestDetectGit(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create HEAD file
	headContent := "ref: refs/heads/develop\n"
	err = os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte(headContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	gitInfo := detectGit(tmpDir)
	if gitInfo == nil {
		t.Fatal("Expected git info, got nil")
	}

	if gitInfo.Root != tmpDir {
		t.Errorf("Expected root %s, got %s", tmpDir, gitInfo.Root)
	}

	if gitInfo.Branch != "develop" {
		t.Errorf("Expected branch develop, got %s", gitInfo.Branch)
	}

	if gitInfo.Worktree {
		t.Error("Expected worktree to be false")
	}
}

func TestDetectGit_NoGit(t *testing.T) {
	tmpDir := t.TempDir()

	gitInfo := detectGit(tmpDir)
	if gitInfo != nil {
		t.Error("Expected nil git info for non-git directory")
	}
}

func TestDetectGit_Worktree(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create commondir file (indicates worktree)
	err = os.WriteFile(filepath.Join(gitDir, "commondir"), []byte("../main/.git\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create HEAD file
	headContent := "ref: refs/heads/feature\n"
	err = os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte(headContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	gitInfo := detectGit(tmpDir)
	if gitInfo == nil {
		t.Fatal("Expected git info, got nil")
	}

	if !gitInfo.Worktree {
		t.Error("Expected worktree to be true")
	}

	if gitInfo.Branch != "feature" {
		t.Errorf("Expected branch feature, got %s", gitInfo.Branch)
	}
}

func TestFindGitDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Find from root
	found := findGitDir(tmpDir)
	if found != gitDir {
		t.Errorf("Expected %s, got %s", gitDir, found)
	}
}

func TestFindGitDir_Nested(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git directory at root
	gitDir := filepath.Join(tmpDir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create nested directory
	nestedDir := filepath.Join(tmpDir, "src", "app")
	err = os.MkdirAll(nestedDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Find from nested directory should find root .git
	found := findGitDir(nestedDir)
	if found != gitDir {
		t.Errorf("Expected %s, got %s", gitDir, found)
	}
}

func TestFindGitDir_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	// No .git directory
	found := findGitDir(tmpDir)
	if found != "" {
		t.Errorf("Expected empty string, got %s", found)
	}
}

func TestGetCurrentBranch(t *testing.T) {
	tmpDir := t.TempDir()

	testCases := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "main branch",
			content:  "ref: refs/heads/main\n",
			expected: "main",
		},
		{
			name:     "develop branch",
			content:  "ref: refs/heads/develop\n",
			expected: "develop",
		},
		{
			name:     "feature branch",
			content:  "ref: refs/heads/feature/new-feature\n",
			expected: "new-feature",
		},
		{
			name:     "detached HEAD",
			content:  "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0\n",
			expected: "detached",
		},
		{
			name:     "no whitespace",
			content:  "ref: refs/heads/master",
			expected: "master",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a unique git dir for this test
			gitDir := filepath.Join(tmpDir, tc.name)
			err := os.MkdirAll(gitDir, 0755)
			if err != nil {
				t.Fatal(err)
			}

			// Write HEAD file
			headFile := filepath.Join(gitDir, "HEAD")
			err = os.WriteFile(headFile, []byte(tc.content), 0644)
			if err != nil {
				t.Fatal(err)
			}

			branch := getCurrentBranch(gitDir)
			if branch != tc.expected {
				t.Errorf("Expected branch %s, got %s", tc.expected, branch)
			}
		})
	}
}

func TestGetCurrentBranch_NoHEAD(t *testing.T) {
	tmpDir := t.TempDir()

	// No HEAD file
	branch := getCurrentBranch(tmpDir)
	if branch != "unknown" {
		t.Errorf("Expected unknown, got %s", branch)
	}
}

func TestGenerateID(t *testing.T) {
	// Test with non-git directories (should return "global")
	id1 := generateID("/tmp/no-git")
	if id1 != "global" {
		t.Errorf("Expected 'global' for non-git path, got %s", id1)
	}

	// Test deterministic behavior for same path
	id2 := generateID("/tmp/no-git")
	if id1 != id2 {
		t.Errorf("Expected same ID for same path, got %s and %s", id1, id2)
	}
}

func TestGenerateID_Deterministic(t *testing.T) {
	path := "/tmp/test/project"

	// Generate ID multiple times
	id1 := generateID(path)
	id2 := generateID(path)
	id3 := generateID(path)

	// Should be the same every time
	if id1 != id2 || id2 != id3 {
		t.Errorf("Expected deterministic ID, got %s, %s, %s", id1, id2, id3)
	}
}
