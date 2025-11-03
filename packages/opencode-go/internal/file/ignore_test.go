package file

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gobwas/glob"
)

func TestLoadIgnorePatterns(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a .gitignore file
	gitignoreContent := `
# Comments should be ignored
*.log
node_modules/
dist/
build/

# Empty lines should be ignored

.DS_Store
*.tmp
`
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	ig, err := LoadIgnorePatterns(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load ignore patterns: %v", err)
	}

	// Test that patterns work
	testCases := []struct {
		path         string
		shouldIgnore bool
	}{
		{"test.log", true},         // Matches *.log
		{"debug.log", true},        // Matches *.log
		{"test.txt", false},        // Doesn't match any pattern
		{"node_modules/pkg", true}, // Matches node_modules/
		{"dist/main.js", true},     // Matches dist/
		{"build/out", true},        // Matches build/
		{".DS_Store", true},        // Matches .DS_Store
		{"temp.tmp", true},         // Matches *.tmp
		{"src/main.go", false},     // Doesn't match any pattern
	}

	for _, tc := range testCases {
		result := ig.ShouldIgnore(tc.path)
		if result != tc.shouldIgnore {
			t.Errorf("Path '%s': expected %v, got %v", tc.path, tc.shouldIgnore, result)
		}
	}
}

func TestLoadIgnorePatterns_NoGitignore(t *testing.T) {
	tmpDir := t.TempDir()

	// No .gitignore file exists
	ig, err := LoadIgnorePatterns(tmpDir)
	if err != nil {
		t.Fatalf("Should not error when .gitignore doesn't exist: %v", err)
	}

	// Should still have default patterns
	testCases := []struct {
		path         string
		shouldIgnore bool
	}{
		{"node_modules/pkg", true}, // Default pattern
		{".git/config", true},      // Default pattern
		{"dist/main.js", true},     // Default pattern
		{"build/out", true},        // Default pattern
		{"test.log", true},         // Default pattern
		{"src/main.go", false},     // Not ignored
	}

	for _, tc := range testCases {
		result := ig.ShouldIgnore(tc.path)
		if result != tc.shouldIgnore {
			t.Errorf("Path '%s': expected %v, got %v", tc.path, tc.shouldIgnore, result)
		}
	}
}

func TestLoadIgnorePatterns_DefaultPatterns(t *testing.T) {
	tmpDir := t.TempDir()

	ig, err := LoadIgnorePatterns(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Verify default patterns are added
	defaultPaths := []string{
		"node_modules/index.js",
		".git/HEAD",
		"dist/bundle.js",
		"build/output",
		"error.log",
	}

	for _, path := range defaultPaths {
		if !ig.ShouldIgnore(path) {
			t.Errorf("Default pattern should ignore '%s'", path)
		}
	}
}

func TestLoadIgnorePatterns_CommentsAndEmptyLines(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .gitignore with comments and empty lines
	gitignoreContent := `
# This is a comment
*.test

# Another comment

# Empty line above
*.bak

`
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	ig, err := LoadIgnorePatterns(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Only actual patterns should work
	if !ig.ShouldIgnore("file.test") {
		t.Error("Should ignore *.test")
	}

	if !ig.ShouldIgnore("file.bak") {
		t.Error("Should ignore *.bak")
	}

	if ig.ShouldIgnore("file.txt") {
		t.Error("Should not ignore *.txt")
	}
}

func TestLoadIgnorePatterns_GlobPatterns(t *testing.T) {
	tmpDir := t.TempDir()

	gitignoreContent := `
*.js
**/*.js
src/**/*.test.*
*.md
foo/bar/**
`
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	ig, err := LoadIgnorePatterns(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load ignore patterns: %v", err)
	}

	testCases := []struct {
		path         string
		shouldIgnore bool
	}{
		{"main.js", true},                  // Matches *.js
		{"src/app.js", true},               // Matches **/*.js
		{"dist/bundle.js", true},           // Matches **/*.js (also default pattern)
		{"src/utils/helper.test.ts", true}, // Matches src/**/*.test.*
		{"README.md", true},                // Matches *.md
		{"foo/bar/baz.txt", true},          // Matches foo/bar/**
		{"main.go", false},                 // No match
	}

	for _, tc := range testCases {
		result := ig.ShouldIgnore(tc.path)
		if result != tc.shouldIgnore {
			t.Errorf("Path '%s': expected %v, got %v", tc.path, tc.shouldIgnore, result)
		}
	}
}

func TestAddPattern(t *testing.T) {
	ig := &IgnorePatterns{
		patterns: make([]glob.Glob, 0),
	}

	// Add valid pattern
	ig.addPattern("*.txt")

	if len(ig.patterns) != 1 {
		t.Fatalf("Expected 1 pattern, got %d", len(ig.patterns))
	}

	if !ig.ShouldIgnore("test.txt") {
		t.Error("Should ignore test.txt")
	}
}

func TestAddPattern_InvalidPattern(t *testing.T) {
	ig := &IgnorePatterns{
		patterns: make([]glob.Glob, 0),
	}

	// Add invalid pattern (should be silently ignored)
	ig.addPattern("[invalid")

	// Should have no patterns
	if len(ig.patterns) != 0 {
		t.Fatalf("Expected 0 patterns for invalid pattern, got %d", len(ig.patterns))
	}
}

func TestShouldIgnore_NoPatterns(t *testing.T) {
	ig := &IgnorePatterns{
		patterns: make([]glob.Glob, 0),
	}

	// No patterns means nothing is ignored
	if ig.ShouldIgnore("anything.txt") {
		t.Error("Should not ignore anything with no patterns")
	}
}

func TestShouldIgnore_MultipleMatches(t *testing.T) {
	ig := &IgnorePatterns{
		patterns: make([]glob.Glob, 0),
	}

	ig.addPattern("*.log")
	ig.addPattern("*.txt")
	ig.addPattern("test.*")

	// File matches multiple patterns
	if !ig.ShouldIgnore("test.log") {
		t.Error("Should ignore test.log (matches both *.log and test.*)")
	}

	if !ig.ShouldIgnore("test.txt") {
		t.Error("Should ignore test.txt (matches both *.txt and test.*)")
	}

	if !ig.ShouldIgnore("debug.log") {
		t.Error("Should ignore debug.log (matches *.log)")
	}
}

func TestLoadIgnorePatterns_DirectoryTraversal(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested .gitignore
	nestedDir := filepath.Join(tmpDir, "nested")
	err := os.MkdirAll(nestedDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	gitignoreContent := `
*.ignore
**/*.ignore
`
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	err = os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Load from root
	ig, err := LoadIgnorePatterns(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Should work for nested paths too
	if !ig.ShouldIgnore("nested/test.ignore") {
		t.Error("Should ignore nested/test.ignore")
	}

	// Should also work for root level
	if !ig.ShouldIgnore("test.ignore") {
		t.Error("Should ignore test.ignore")
	}
}

func TestLoadIgnorePatterns_RealWorldExample(t *testing.T) {
	tmpDir := t.TempDir()

	// Real-world .gitignore example
	gitignoreContent := `
# Dependencies
node_modules/**
vendor/**
*.gem
*.egg-info/**

# Build outputs
dist/**
build/**
*.o
*.so
*.dylib

# IDE
.vscode/**
.idea/**
*.swp
*.swo

# OS files
.DS_Store
Thumbs.db

# Logs
*.log
**/*.log
logs/**

# Environment
.env
.env.local
`
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	ig, err := LoadIgnorePatterns(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Test various paths
	testCases := []struct {
		path         string
		shouldIgnore bool
	}{
		{"node_modules/react/index.js", true},
		{"vendor/pkg.go", true},
		{"dist/bundle.js", true},
		{"build/main.o", true},
		{".vscode/settings.json", true},
		{".idea/workspace.xml", true},
		{".DS_Store", true},
		{"app.log", true},
		{".env", true},
		{".env.local", true},
		{"src/main.go", false},
		{"README.md", false},
	}

	for _, tc := range testCases {
		result := ig.ShouldIgnore(tc.path)
		if result != tc.shouldIgnore {
			t.Errorf("Path '%s': expected %v, got %v", tc.path, tc.shouldIgnore, result)
		}
	}
}
