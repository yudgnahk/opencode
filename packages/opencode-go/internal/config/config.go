package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"gopkg.in/yaml.v3"
)

const app = "opencode"

// Path contains all XDG base directory paths for opencode
// This matches the TypeScript implementation in packages/opencode/src/global/index.ts
var Path = struct {
	Data   string
	Bin    string
	Log    string
	Cache  string
	Config string
	State  string
}{
	Data:   filepath.Join(getXDGDataHome(), app),
	Bin:    filepath.Join(getXDGDataHome(), app, "bin"),
	Log:    filepath.Join(getXDGDataHome(), app, "log"),
	Cache:  filepath.Join(getXDGCacheHome(), app),
	Config: filepath.Join(getXDGConfigHome(), app),
	State:  filepath.Join(getXDGStateHome(), app),
}

type Config struct {
	Server    ServerConfig              `yaml:"server" json:"server"`
	Providers map[string]ProviderConfig `yaml:"providers" json:"providers"`
	Storage   StorageConfig             `yaml:"storage" json:"storage"`
}

type ServerConfig struct {
	Host string `yaml:"host" json:"host"`
	Port int    `yaml:"port" json:"port"`
}

type ProviderConfig struct {
	APIKey  string `yaml:"apiKey" json:"apiKey"`
	BaseURL string `yaml:"baseUrl" json:"baseUrl"`
	Model   string `yaml:"model" json:"model"`
}

type StorageConfig struct {
	Path string `yaml:"path" json:"path"`
}

// CACHE_VERSION should be incremented when cache format changes
// This matches the TypeScript implementation in packages/opencode/src/global/index.ts
const CACHE_VERSION = "9"

// Initialize creates all necessary directories
// This matches the TypeScript implementation which creates directories on startup
func Initialize() error {
	dirs := []string{
		Path.Data,
		Path.Config,
		Path.State,
		Path.Log,
		Path.Bin,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Handle cache versioning - clear cache if version has changed
	if err := handleCacheVersion(); err != nil {
		// Log error but don't fail initialization
		// The TypeScript version also continues on cache clear errors
	}

	return nil
}

// handleCacheVersion checks cache version and clears if changed
// This matches the TypeScript cache version handling
func handleCacheVersion() error {
	versionFile := filepath.Join(Path.Cache, "version")

	// Read current version
	currentVersion, err := os.ReadFile(versionFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// If version matches, nothing to do
	if string(currentVersion) == CACHE_VERSION {
		return nil
	}

	// Version changed or doesn't exist - clear cache
	entries, err := os.ReadDir(Path.Cache)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Remove all cache contents
	for _, entry := range entries {
		path := filepath.Join(Path.Cache, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			// Continue on error, similar to TypeScript
			continue
		}
	}

	// Recreate cache directory and write new version
	if err := os.MkdirAll(Path.Cache, 0755); err != nil {
		return err
	}

	return os.WriteFile(versionFile, []byte(CACHE_VERSION), 0644)
}

func Load() (*Config, error) {
	// Try multiple config locations
	// This matches the TypeScript config loading order
	configPaths := []string{
		filepath.Join(Path.Config, "config.json"),
		filepath.Join(Path.Config, "opencode.json"),
		filepath.Join(Path.Config, "opencode.jsonc"),
		"./opencode.json",
		"./opencode.jsonc",
		filepath.Join(Path.Config, "config.yml"),
		"./opencode.yml",
		"./opencode.yaml",
	}

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			return loadFromFile(path)
		}
	}

	// Return default config if none found
	return defaultConfig(), nil
}

func loadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	ext := filepath.Ext(path)

	// Parse based on file extension
	if ext == ".json" || ext == ".jsonc" {
		// Handle JSONC (JSON with comments)
		if ext == ".jsonc" {
			data = stripJSONComments(data)
		}
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
	} else {
		// Default to YAML parsing
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
	}

	// Apply environment variable overrides
	applyEnvOverrides(&cfg)

	return &cfg, nil
}

// stripJSONComments removes single-line (//) and multi-line (/* */) comments from JSON
// This allows JSONC (JSON with Comments) to be parsed as regular JSON
func stripJSONComments(data []byte) []byte {
	text := string(data)

	// Remove single-line comments
	singleLineCommentRegex := regexp.MustCompile(`//.*`)
	text = singleLineCommentRegex.ReplaceAllString(text, "")

	// Remove multi-line comments
	multiLineCommentRegex := regexp.MustCompile(`(?s)/\*.*?\*/`)
	text = multiLineCommentRegex.ReplaceAllString(text, "")

	return []byte(text)
}

func defaultConfig() *Config {
	cfg := &Config{
		Server: ServerConfig{
			Host: "127.0.0.1",
			Port: 8080,
		},
		Storage: StorageConfig{
			// Storage directory in XDG_DATA_HOME/opencode/storage
			// (usually ~/.local/share/opencode/storage)
			// This matches the TypeScript version structure
			Path: filepath.Join(Path.Data, "storage"),
		},
		Providers: make(map[string]ProviderConfig),
	}

	// Apply environment variable overrides
	applyEnvOverrides(cfg)

	return cfg
}

func applyEnvOverrides(cfg *Config) {
	// Server overrides
	if host := os.Getenv("OPENCODE_HOST"); host != "" {
		cfg.Server.Host = host
	}

	if portStr := os.Getenv("OPENCODE_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.Server.Port = port
		}
	}

	// Storage overrides
	if storagePath := os.Getenv("OPENCODE_STORAGE_PATH"); storagePath != "" {
		cfg.Storage.Path = storagePath
	}

	// Provider API keys from environment
	// Example: OPENCODE_ANTHROPIC_API_KEY, OPENCODE_OPENAI_API_KEY
	for providerName := range cfg.Providers {
		envKey := "OPENCODE_" + providerName + "_API_KEY"
		if apiKey := os.Getenv(envKey); apiKey != "" {
			provider := cfg.Providers[providerName]
			provider.APIKey = apiKey
			cfg.Providers[providerName] = provider
		}
	}
}

// GetDataDir returns the path to the OpenCode data directory
// This is where auth.json and other data files are stored
// Matches TypeScript Global.Path.data
func GetDataDir() string {
	return Path.Data
}

// GetStatePath returns the XDG state directory for opencode
// Matches TypeScript Global.Path.state
func GetStatePath() string {
	return Path.State
}

// GetConfigPath returns the XDG config directory for opencode
// Matches TypeScript Global.Path.config
func GetConfigPath() string {
	return Path.Config
}

// GetCachePath returns the XDG cache directory for opencode
// Matches TypeScript Global.Path.cache
func GetCachePath() string {
	return Path.Cache
}

// GetLogPath returns the log directory for opencode
// Matches TypeScript Global.Path.log
func GetLogPath() string {
	return Path.Log
}

// GetBinPath returns the bin directory for opencode
// Matches TypeScript Global.Path.bin
func GetBinPath() string {
	return Path.Bin
}

// XDG base directory paths
func getXDGDataHome() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return xdg
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share")
}

func getXDGConfigHome() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return xdg
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config")
}

func getXDGStateHome() string {
	if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
		return xdg
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "state")
}

func getXDGCacheHome() string {
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return xdg
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache")
}
