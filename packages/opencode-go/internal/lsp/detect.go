package lsp

import (
	"path/filepath"
	"strings"
)

// ServerConfig defines the command and args for starting an LSP server
type ServerConfig struct {
	Command string
	Args    []string
}

var languageServers = map[string]ServerConfig{
	"go": {
		Command: "gopls",
		Args:    []string{},
	},
	"typescript": {
		Command: "typescript-language-server",
		Args:    []string{"--stdio"},
	},
	"javascript": {
		Command: "typescript-language-server",
		Args:    []string{"--stdio"},
	},
	"python": {
		Command: "pylsp",
		Args:    []string{},
	},
	"rust": {
		Command: "rust-analyzer",
		Args:    []string{},
	},
}

// DetectLanguage detects the programming language from a file path
func DetectLanguage(filePath string) string {
	ext := strings.TrimPrefix(filepath.Ext(filePath), ".")

	switch ext {
	case "go":
		return "go"
	case "ts", "tsx":
		return "typescript"
	case "js", "jsx":
		return "javascript"
	case "py":
		return "python"
	case "rs":
		return "rust"
	default:
		return ""
	}
}

// GetServerConfig returns the server configuration for a language
func GetServerConfig(language string) (ServerConfig, bool) {
	cfg, ok := languageServers[language]
	return cfg, ok
}

// AddServerConfig adds or updates a server configuration
func AddServerConfig(language string, command string, args []string) {
	languageServers[language] = ServerConfig{
		Command: command,
		Args:    args,
	}
}

// RemoveServerConfig removes a server configuration
func RemoveServerConfig(language string) {
	delete(languageServers, language)
}

// GetSupportedLanguages returns a list of supported languages
func GetSupportedLanguages() []string {
	languages := make([]string, 0, len(languageServers))
	for lang := range languageServers {
		languages = append(languages, lang)
	}
	return languages
}
