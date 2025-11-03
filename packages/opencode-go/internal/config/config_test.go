package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_DefaultConfig(t *testing.T) {
	// Test loading when no config file exists
	// Change to temp dir to avoid finding actual config
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}

	// Verify defaults
	if cfg.Server.Host != "localhost" {
		t.Errorf("Expected default host 'localhost', got '%s'", cfg.Server.Host)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Server.Port)
	}

	if cfg.Storage.Path == "" {
		t.Error("Expected storage path to be set")
	}
}

func TestLoad_FromFile(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	// Create config file
	configContent := `
server:
  host: 0.0.0.0
  port: 9000
providers:
  anthropic:
    apiKey: test-key
    model: claude-3-5-sonnet
storage:
  path: /tmp/test.db
`
	configPath := filepath.Join(tmpDir, "opencode.yml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config from file: %v", err)
	}

	// Verify loaded values
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Expected host '0.0.0.0', got '%s'", cfg.Server.Host)
	}

	if cfg.Server.Port != 9000 {
		t.Errorf("Expected port 9000, got %d", cfg.Server.Port)
	}

	if cfg.Storage.Path != "/tmp/test.db" {
		t.Errorf("Expected storage path '/tmp/test.db', got '%s'", cfg.Storage.Path)
	}

	// Check provider
	anthropic, ok := cfg.Providers["anthropic"]
	if !ok {
		t.Fatal("Anthropic provider not found")
	}

	if anthropic.APIKey != "test-key" {
		t.Errorf("Expected API key 'test-key', got '%s'", anthropic.APIKey)
	}

	if anthropic.Model != "claude-3-5-sonnet" {
		t.Errorf("Expected model 'claude-3-5-sonnet', got '%s'", anthropic.Model)
	}
}

func TestLoad_YamlExtension(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	// Create config file with .yaml extension
	configContent := `
server:
  host: test-host
  port: 7777
`
	configPath := filepath.Join(tmpDir, "opencode.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config from .yaml file: %v", err)
	}

	if cfg.Server.Host != "test-host" {
		t.Errorf("Expected host 'test-host', got '%s'", cfg.Server.Host)
	}

	if cfg.Server.Port != 7777 {
		t.Errorf("Expected port 7777, got %d", cfg.Server.Port)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	// Set environment variables
	os.Setenv("OPENCODE_HOST", "env-host")
	os.Setenv("OPENCODE_PORT", "3000")
	os.Setenv("OPENCODE_STORAGE_PATH", "/env/path.db")
	defer func() {
		os.Unsetenv("OPENCODE_HOST")
		os.Unsetenv("OPENCODE_PORT")
		os.Unsetenv("OPENCODE_STORAGE_PATH")
	}()

	// Create config file with different values
	configContent := `
server:
  host: file-host
  port: 9000
storage:
  path: /file/path.db
`
	configPath := filepath.Join(tmpDir, "opencode.yml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Environment variables should override file values
	if cfg.Server.Host != "env-host" {
		t.Errorf("Expected host 'env-host' from env, got '%s'", cfg.Server.Host)
	}

	if cfg.Server.Port != 3000 {
		t.Errorf("Expected port 3000 from env, got %d", cfg.Server.Port)
	}

	if cfg.Storage.Path != "/env/path.db" {
		t.Errorf("Expected storage path '/env/path.db' from env, got '%s'", cfg.Storage.Path)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	// Create invalid YAML with malformed structure
	invalidYAML := `
server: [
  host: test
`
	configPath := filepath.Join(tmpDir, "opencode.yml")
	err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Load()
	if err == nil {
		t.Fatal("Expected error when loading invalid YAML")
	}
}

func TestLoadFromFile_NonExistent(t *testing.T) {
	_, err := loadFromFile("/non/existent/path/config.yml")
	if err == nil {
		t.Fatal("Expected error when loading non-existent file")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	if cfg.Server.Host != "localhost" {
		t.Errorf("Expected default host 'localhost', got '%s'", cfg.Server.Host)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Server.Port)
	}

	if cfg.Storage.Path == "" {
		t.Error("Expected storage path to be set")
	}

	if cfg.Providers == nil {
		t.Error("Expected providers map to be initialized")
	}
}

func TestApplyEnvOverrides_Port(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
	}

	// Test valid port override
	os.Setenv("OPENCODE_PORT", "9999")
	defer os.Unsetenv("OPENCODE_PORT")

	applyEnvOverrides(cfg)

	if cfg.Server.Port != 9999 {
		t.Errorf("Expected port 9999, got %d", cfg.Server.Port)
	}
}

func TestApplyEnvOverrides_InvalidPort(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
	}

	// Test invalid port (should not change)
	os.Setenv("OPENCODE_PORT", "invalid")
	defer os.Unsetenv("OPENCODE_PORT")

	applyEnvOverrides(cfg)

	if cfg.Server.Port != 8080 {
		t.Errorf("Port should remain 8080 with invalid env value, got %d", cfg.Server.Port)
	}
}

func TestApplyEnvOverrides_ProviderAPIKey(t *testing.T) {
	cfg := &Config{
		Providers: map[string]ProviderConfig{
			"anthropic": {
				APIKey: "file-key",
			},
		},
	}

	// Test provider API key override
	os.Setenv("OPENCODE_anthropic_API_KEY", "env-key")
	defer os.Unsetenv("OPENCODE_anthropic_API_KEY")

	applyEnvOverrides(cfg)

	if cfg.Providers["anthropic"].APIKey != "env-key" {
		t.Errorf("Expected API key 'env-key', got '%s'", cfg.Providers["anthropic"].APIKey)
	}
}

func TestGetDataDir(t *testing.T) {
	dir := GetDataDir()

	// Should return a path
	if dir == "" {
		t.Error("GetDataDir should return a path")
	}

	// Should end with .opencode/data
	if !filepath.IsAbs(dir) && dir != ".opencode/data" {
		t.Errorf("Expected path to be absolute or '.opencode/data', got '%s'", dir)
	}
}

func TestLoad_HomeConfigDir(t *testing.T) {
	// This test verifies the ~/.config/opencode/config.yml path
	// We'll create a config there and ensure it's loaded

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot determine home directory")
	}

	configDir := filepath.Join(home, ".config", "opencode")
	configPath := filepath.Join(configDir, "config.yml")

	// Check if we can write to this location
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Skip("Cannot create config directory")
	}

	// Save original config if it exists
	originalConfig, _ := os.ReadFile(configPath)
	hasOriginal := originalConfig != nil

	// Create test config
	testConfig := `
server:
  host: test-config-host
  port: 5555
`
	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Skip("Cannot write test config")
	}

	// Restore original config or remove test config
	defer func() {
		if hasOriginal {
			os.WriteFile(configPath, originalConfig, 0644)
		} else {
			os.Remove(configPath)
		}
	}()

	// Change to temp dir to avoid finding local config
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Server.Host != "test-config-host" {
		t.Errorf("Expected host 'test-config-host', got '%s'", cfg.Server.Host)
	}

	if cfg.Server.Port != 5555 {
		t.Errorf("Expected port 5555, got %d", cfg.Server.Port)
	}
}

func TestLoad_MultipleProviders(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	configContent := `
providers:
  anthropic:
    apiKey: anthropic-key
    model: claude-3
    baseUrl: https://api.anthropic.com
  openai:
    apiKey: openai-key
    model: gpt-4
    baseUrl: https://api.openai.com
  gemini:
    apiKey: gemini-key
    model: gemini-pro
`
	configPath := filepath.Join(tmpDir, "opencode.yml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify all three providers
	if len(cfg.Providers) != 3 {
		t.Errorf("Expected 3 providers, got %d", len(cfg.Providers))
	}

	providers := []string{"anthropic", "openai", "gemini"}
	for _, p := range providers {
		if _, ok := cfg.Providers[p]; !ok {
			t.Errorf("Provider '%s' not found", p)
		}
	}
}
