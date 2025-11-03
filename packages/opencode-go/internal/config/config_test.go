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
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Expected default host '127.0.0.1', got '%s'", cfg.Server.Host)
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

	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Expected default host '127.0.0.1', got '%s'", cfg.Server.Host)
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
			Host: "127.0.0.1",
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
			Host: "127.0.0.1",
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

	// Should be absolute path
	if !filepath.IsAbs(dir) {
		t.Errorf("Expected absolute path, got '%s'", dir)
	}

	// Should end with opencode (matches XDG structure)
	if filepath.Base(dir) != "opencode" {
		t.Errorf("Expected path to end with 'opencode', got '%s'", dir)
	}
}

func TestXDGPaths(t *testing.T) {
	// Test that all XDG paths are set
	if Path.Data == "" {
		t.Error("Path.Data should be set")
	}
	if Path.Config == "" {
		t.Error("Path.Config should be set")
	}
	if Path.State == "" {
		t.Error("Path.State should be set")
	}
	if Path.Cache == "" {
		t.Error("Path.Cache should be set")
	}
	if Path.Log == "" {
		t.Error("Path.Log should be set")
	}
	if Path.Bin == "" {
		t.Error("Path.Bin should be set")
	}

	// Verify paths end with opencode
	paths := map[string]string{
		"Data":   Path.Data,
		"Config": Path.Config,
		"State":  Path.State,
		"Cache":  Path.Cache,
	}

	for name, path := range paths {
		if filepath.Base(path) != "opencode" {
			t.Errorf("Expected %s path to end with 'opencode', got '%s'", name, path)
		}
	}
}

func TestInitialize(t *testing.T) {
	// Use temp dir for testing
	tmpDir := t.TempDir()

	// Override XDG paths temporarily
	oldDataHome := os.Getenv("XDG_DATA_HOME")
	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	oldStateHome := os.Getenv("XDG_STATE_HOME")
	oldCacheHome := os.Getenv("XDG_CACHE_HOME")

	os.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "data"))
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "config"))
	os.Setenv("XDG_STATE_HOME", filepath.Join(tmpDir, "state"))
	os.Setenv("XDG_CACHE_HOME", filepath.Join(tmpDir, "cache"))

	defer func() {
		os.Setenv("XDG_DATA_HOME", oldDataHome)
		os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
		os.Setenv("XDG_STATE_HOME", oldStateHome)
		os.Setenv("XDG_CACHE_HOME", oldCacheHome)
	}()

	// Reinitialize Path struct with new env vars
	Path.Data = filepath.Join(getXDGDataHome(), app)
	Path.Bin = filepath.Join(getXDGDataHome(), app, "bin")
	Path.Log = filepath.Join(getXDGDataHome(), app, "log")
	Path.Cache = filepath.Join(getXDGCacheHome(), app)
	Path.Config = filepath.Join(getXDGConfigHome(), app)
	Path.State = filepath.Join(getXDGStateHome(), app)

	// Test Initialize
	err := Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Verify directories were created
	dirs := []string{
		Path.Data,
		Path.Config,
		Path.State,
		Path.Log,
		Path.Bin,
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Directory not created: %s", dir)
		}
	}
}

func TestLoad_XDGConfigDir(t *testing.T) {
	// This test verifies loading from XDG_CONFIG_HOME/opencode
	tmpDir := t.TempDir()

	// Override XDG_CONFIG_HOME for testing
	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldConfigHome)

	// Reinitialize Path.Config with new env var
	Path.Config = filepath.Join(getXDGConfigHome(), app)

	configDir := Path.Config
	configPath := filepath.Join(configDir, "config.json")

	// Create config directory and file
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	testConfig := `{
  "server": {
    "host": "test-config-host",
    "port": 5555
  }
}`
	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to different temp dir to avoid finding local config
	workDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(workDir)
	defer os.Chdir(originalWd)

	// Now we should be able to load JSON config
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load JSON config from XDG_CONFIG_HOME: %v", err)
	}

	// Verify loaded values from JSON
	if cfg.Server.Host != "test-config-host" {
		t.Errorf("Expected host 'test-config-host', got '%s'", cfg.Server.Host)
	}

	if cfg.Server.Port != 5555 {
		t.Errorf("Expected port 5555, got %d", cfg.Server.Port)
	}
}

func TestLoad_JSONConfig(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	// Create JSON config file
	configContent := `{
  "server": {
    "host": "json-host",
    "port": 8888
  },
  "providers": {
    "openai": {
      "apiKey": "json-key",
      "model": "gpt-4"
    }
  }
}`
	configPath := filepath.Join(tmpDir, "opencode.json")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load JSON config: %v", err)
	}

	if cfg.Server.Host != "json-host" {
		t.Errorf("Expected host 'json-host', got '%s'", cfg.Server.Host)
	}

	if cfg.Server.Port != 8888 {
		t.Errorf("Expected port 8888, got %d", cfg.Server.Port)
	}

	openai, ok := cfg.Providers["openai"]
	if !ok {
		t.Fatal("OpenAI provider not found")
	}

	if openai.APIKey != "json-key" {
		t.Errorf("Expected API key 'json-key', got '%s'", openai.APIKey)
	}
}

func TestLoad_JSONCConfig(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	// Create JSONC config file with comments
	configContent := `{
  // Server configuration
  "server": {
    "host": "jsonc-host", // This is a comment
    "port": 7777
  },
  /* Provider configuration
     with multiple lines */
  "providers": {
    "anthropic": {
      "apiKey": "jsonc-key",
      "model": "claude-3" // Latest model
    }
  }
}`
	configPath := filepath.Join(tmpDir, "opencode.jsonc")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load JSONC config: %v", err)
	}

	if cfg.Server.Host != "jsonc-host" {
		t.Errorf("Expected host 'jsonc-host', got '%s'", cfg.Server.Host)
	}

	if cfg.Server.Port != 7777 {
		t.Errorf("Expected port 7777, got %d", cfg.Server.Port)
	}

	anthropic, ok := cfg.Providers["anthropic"]
	if !ok {
		t.Fatal("Anthropic provider not found")
	}

	if anthropic.APIKey != "jsonc-key" {
		t.Errorf("Expected API key 'jsonc-key', got '%s'", anthropic.APIKey)
	}

	if anthropic.Model != "claude-3" {
		t.Errorf("Expected model 'claude-3', got '%s'", anthropic.Model)
	}
}

func TestCacheVersioning(t *testing.T) {
	tmpDir := t.TempDir()

	// Set up temporary cache directory
	oldCacheHome := os.Getenv("XDG_CACHE_HOME")
	os.Setenv("XDG_CACHE_HOME", tmpDir)
	defer os.Setenv("XDG_CACHE_HOME", oldCacheHome)

	// Reinitialize Path.Cache with new env var
	Path.Cache = filepath.Join(getXDGCacheHome(), app)

	// Create cache directory with old version
	if err := os.MkdirAll(Path.Cache, 0755); err != nil {
		t.Fatal(err)
	}

	// Write old version file
	versionFile := filepath.Join(Path.Cache, "version")
	if err := os.WriteFile(versionFile, []byte("0"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a dummy cache file
	dummyFile := filepath.Join(Path.Cache, "dummy.cache")
	if err := os.WriteFile(dummyFile, []byte("test data"), 0644); err != nil {
		t.Fatal(err)
	}

	// Run Initialize which should handle cache versioning
	if err := Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Verify version was updated
	newVersion, err := os.ReadFile(versionFile)
	if err != nil {
		t.Fatalf("Failed to read version file: %v", err)
	}

	if string(newVersion) != CACHE_VERSION {
		t.Errorf("Expected cache version '%s', got '%s'", CACHE_VERSION, string(newVersion))
	}

	// Verify dummy cache file was removed
	if _, err := os.Stat(dummyFile); !os.IsNotExist(err) {
		t.Error("Expected dummy cache file to be removed")
	}
}

func TestCacheVersioning_SameVersion(t *testing.T) {
	tmpDir := t.TempDir()

	// Set up temporary cache directory
	oldCacheHome := os.Getenv("XDG_CACHE_HOME")
	os.Setenv("XDG_CACHE_HOME", tmpDir)
	defer os.Setenv("XDG_CACHE_HOME", oldCacheHome)

	// Reinitialize Path.Cache with new env var
	Path.Cache = filepath.Join(getXDGCacheHome(), app)

	// Create cache directory with current version
	if err := os.MkdirAll(Path.Cache, 0755); err != nil {
		t.Fatal(err)
	}

	// Write current version file
	versionFile := filepath.Join(Path.Cache, "version")
	if err := os.WriteFile(versionFile, []byte(CACHE_VERSION), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a dummy cache file
	dummyFile := filepath.Join(Path.Cache, "dummy.cache")
	if err := os.WriteFile(dummyFile, []byte("test data"), 0644); err != nil {
		t.Fatal(err)
	}

	// Run Initialize
	if err := Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Verify dummy cache file still exists (not cleared)
	if _, err := os.Stat(dummyFile); os.IsNotExist(err) {
		t.Error("Cache file should not be removed when version matches")
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
