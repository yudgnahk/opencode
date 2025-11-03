package config

import (
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig              `yaml:"server"`
	Providers map[string]ProviderConfig `yaml:"providers"`
	Storage   StorageConfig             `yaml:"storage"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type ProviderConfig struct {
	APIKey  string `yaml:"apiKey"`
	BaseURL string `yaml:"baseUrl"`
	Model   string `yaml:"model"`
}

type StorageConfig struct {
	Path string `yaml:"path"`
}

func Load() (*Config, error) {
	// Try multiple config locations
	configPaths := []string{
		"./opencode.yml",
		"./opencode.yaml",
		filepath.Join(os.Getenv("HOME"), ".config", "opencode", "config.yml"),
		filepath.Join(os.Getenv("HOME"), ".opencode.yml"),
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
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Apply environment variable overrides
	applyEnvOverrides(&cfg)

	return &cfg, nil
}

func defaultConfig() *Config {
	cfg := &Config{
		Server: ServerConfig{
			Host: "127.0.0.1",
			Port: 8080,
		},
		Storage: StorageConfig{
			Path: filepath.Join(os.Getenv("HOME"), ".opencode", "storage", "opencode.db"),
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
func GetDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".opencode", "data")
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

// GetStatePath returns the XDG state directory for opencode
func GetStatePath() string {
	return filepath.Join(getXDGStateHome(), "opencode")
}

// GetConfigPath returns the XDG config directory for opencode
func GetConfigPath() string {
	return filepath.Join(getXDGConfigHome(), "opencode")
}
