package lsp

import (
	"testing"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{
			name:     "Go file",
			filePath: "/path/to/file.go",
			expected: "go",
		},
		{
			name:     "TypeScript file",
			filePath: "/path/to/file.ts",
			expected: "typescript",
		},
		{
			name:     "TSX file",
			filePath: "/path/to/component.tsx",
			expected: "typescript",
		},
		{
			name:     "JavaScript file",
			filePath: "/path/to/script.js",
			expected: "javascript",
		},
		{
			name:     "JSX file",
			filePath: "/path/to/component.jsx",
			expected: "javascript",
		},
		{
			name:     "Python file",
			filePath: "/path/to/script.py",
			expected: "python",
		},
		{
			name:     "Rust file",
			filePath: "/path/to/main.rs",
			expected: "rust",
		},
		{
			name:     "Unknown extension",
			filePath: "/path/to/file.txt",
			expected: "",
		},
		{
			name:     "No extension",
			filePath: "/path/to/file",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectLanguage(tt.filePath)
			if result != tt.expected {
				t.Errorf("DetectLanguage(%q) = %q, want %q", tt.filePath, result, tt.expected)
			}
		})
	}
}

func TestGetServerConfig(t *testing.T) {
	tests := []struct {
		name         string
		language     string
		expectFound  bool
		expectedCmd  string
		expectedArgs []string
	}{
		{
			name:         "Go language",
			language:     "go",
			expectFound:  true,
			expectedCmd:  "gopls",
			expectedArgs: []string{},
		},
		{
			name:         "TypeScript language",
			language:     "typescript",
			expectFound:  true,
			expectedCmd:  "typescript-language-server",
			expectedArgs: []string{"--stdio"},
		},
		{
			name:        "Unknown language",
			language:    "unknown",
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, found := GetServerConfig(tt.language)
			if found != tt.expectFound {
				t.Errorf("GetServerConfig(%q) found = %v, want %v", tt.language, found, tt.expectFound)
			}
			if found && cfg.Command != tt.expectedCmd {
				t.Errorf("GetServerConfig(%q) command = %q, want %q", tt.language, cfg.Command, tt.expectedCmd)
			}
			if found && len(cfg.Args) != len(tt.expectedArgs) {
				t.Errorf("GetServerConfig(%q) args length = %d, want %d", tt.language, len(cfg.Args), len(tt.expectedArgs))
			}
		})
	}
}

func TestAddAndRemoveServerConfig(t *testing.T) {
	testLang := "testlang"
	testCmd := "test-lsp"
	testArgs := []string{"--test"}

	// Add a new config
	AddServerConfig(testLang, testCmd, testArgs)

	// Verify it was added
	cfg, found := GetServerConfig(testLang)
	if !found {
		t.Errorf("AddServerConfig failed: config not found for %q", testLang)
	}
	if cfg.Command != testCmd {
		t.Errorf("AddServerConfig: command = %q, want %q", cfg.Command, testCmd)
	}

	// Remove the config
	RemoveServerConfig(testLang)

	// Verify it was removed
	_, found = GetServerConfig(testLang)
	if found {
		t.Errorf("RemoveServerConfig failed: config still found for %q", testLang)
	}
}

func TestGetSupportedLanguages(t *testing.T) {
	langs := GetSupportedLanguages()

	// Verify we have at least the core languages
	expectedLangs := []string{"go", "typescript", "javascript", "python", "rust"}

	for _, expected := range expectedLangs {
		found := false
		for _, lang := range langs {
			if lang == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetSupportedLanguages missing expected language: %q", expected)
		}
	}
}
