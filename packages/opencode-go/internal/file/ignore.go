package file

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
)

type IgnorePatterns struct {
	patterns []glob.Glob
}

func LoadIgnorePatterns(projectPath string) (*IgnorePatterns, error) {
	ig := &IgnorePatterns{
		patterns: make([]glob.Glob, 0),
	}

	// Load .gitignore
	gitignorePath := filepath.Join(projectPath, ".gitignore")
	_ = ig.loadFromFile(gitignorePath) // Ignore errors if file doesn't exist

	// Add default patterns
	ig.addPattern("node_modules/**")
	ig.addPattern(".git/**")
	ig.addPattern("dist/**")
	ig.addPattern("build/**")
	ig.addPattern("*.log")

	return ig, nil
}

func (ig *IgnorePatterns) loadFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		ig.addPattern(line)
	}

	return scanner.Err()
}

func (ig *IgnorePatterns) addPattern(pattern string) {
	g, err := glob.Compile(pattern, '/')
	if err == nil {
		ig.patterns = append(ig.patterns, g)
	}
}

func (ig *IgnorePatterns) ShouldIgnore(path string) bool {
	for _, pattern := range ig.patterns {
		if pattern.Match(path) {
			return true
		}
	}
	return false
}
