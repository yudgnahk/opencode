package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func setupTestAuthFile(t *testing.T) (string, func()) {
	t.Helper()

	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "auth-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Override authFilePathFunc for testing
	originalFunc := authFilePathFunc
	testFilePath := filepath.Join(tmpDir, "auth.json")
	authFilePathFunc = func() string {
		return testFilePath
	}

	cleanup := func() {
		authFilePathFunc = originalFunc
		os.RemoveAll(tmpDir)
	}

	return testFilePath, cleanup
}

func TestOAuthType(t *testing.T) {
	oauth := OAuth{
		Type:    TypeOAuth,
		Refresh: "refresh-token",
		Access:  "access-token",
		Expires: 1234567890,
	}

	if oauth.GetType() != TypeOAuth {
		t.Errorf("Expected type %s, got %s", TypeOAuth, oauth.GetType())
	}
}

func TestAPIType(t *testing.T) {
	api := API{
		Type: TypeAPI,
		Key:  "api-key",
	}

	if api.GetType() != TypeAPI {
		t.Errorf("Expected type %s, got %s", TypeAPI, api.GetType())
	}
}

func TestWellKnownType(t *testing.T) {
	wellknown := WellKnown{
		Type:  TypeWellKnown,
		Key:   "key",
		Token: "token",
	}

	if wellknown.GetType() != TypeWellKnown {
		t.Errorf("Expected type %s, got %s", TypeWellKnown, wellknown.GetType())
	}
}

func TestGetNonExistentFile(t *testing.T) {
	_, cleanup := setupTestAuthFile(t)
	defer cleanup()

	info, err := Get("anthropic")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if info != nil {
		t.Errorf("Expected nil info for non-existent file, got %v", info)
	}
}

func TestGetNonExistentProvider(t *testing.T) {
	_, cleanup := setupTestAuthFile(t)
	defer cleanup()

	// Create file with some data
	oauth := OAuth{
		Type:    TypeOAuth,
		Refresh: "refresh",
		Access:  "access",
		Expires: 123,
	}
	if err := Set("openai", oauth); err != nil {
		t.Fatalf("Failed to set auth: %v", err)
	}

	// Try to get non-existent provider
	info, err := Get("anthropic")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if info != nil {
		t.Errorf("Expected nil info for non-existent provider, got %v", info)
	}
}

func TestSetAndGetOAuth(t *testing.T) {
	_, cleanup := setupTestAuthFile(t)
	defer cleanup()

	oauth := OAuth{
		Type:    TypeOAuth,
		Refresh: "refresh-token",
		Access:  "access-token",
		Expires: 1234567890,
	}

	if err := Set("anthropic", oauth); err != nil {
		t.Fatalf("Failed to set OAuth: %v", err)
	}

	info, err := Get("anthropic")
	if err != nil {
		t.Fatalf("Failed to get OAuth: %v", err)
	}

	retrieved, ok := info.(OAuth)
	if !ok {
		t.Fatalf("Expected OAuth type, got %T", info)
	}

	if retrieved.Type != TypeOAuth {
		t.Errorf("Expected type %s, got %s", TypeOAuth, retrieved.Type)
	}
	if retrieved.Refresh != "refresh-token" {
		t.Errorf("Expected refresh %s, got %s", "refresh-token", retrieved.Refresh)
	}
	if retrieved.Access != "access-token" {
		t.Errorf("Expected access %s, got %s", "access-token", retrieved.Access)
	}
	if retrieved.Expires != 1234567890 {
		t.Errorf("Expected expires %d, got %d", 1234567890, retrieved.Expires)
	}
}

func TestSetAndGetAPI(t *testing.T) {
	_, cleanup := setupTestAuthFile(t)
	defer cleanup()

	api := API{
		Type: TypeAPI,
		Key:  "sk-test-key-123",
	}

	if err := Set("openai", api); err != nil {
		t.Fatalf("Failed to set API: %v", err)
	}

	info, err := Get("openai")
	if err != nil {
		t.Fatalf("Failed to get API: %v", err)
	}

	retrieved, ok := info.(API)
	if !ok {
		t.Fatalf("Expected API type, got %T", info)
	}

	if retrieved.Type != TypeAPI {
		t.Errorf("Expected type %s, got %s", TypeAPI, retrieved.Type)
	}
	if retrieved.Key != "sk-test-key-123" {
		t.Errorf("Expected key %s, got %s", "sk-test-key-123", retrieved.Key)
	}
}

func TestSetAndGetWellKnown(t *testing.T) {
	_, cleanup := setupTestAuthFile(t)
	defer cleanup()

	wellknown := WellKnown{
		Type:  TypeWellKnown,
		Key:   "project-key",
		Token: "auth-token",
	}

	if err := Set("gemini", wellknown); err != nil {
		t.Fatalf("Failed to set WellKnown: %v", err)
	}

	info, err := Get("gemini")
	if err != nil {
		t.Fatalf("Failed to get WellKnown: %v", err)
	}

	retrieved, ok := info.(WellKnown)
	if !ok {
		t.Fatalf("Expected WellKnown type, got %T", info)
	}

	if retrieved.Type != TypeWellKnown {
		t.Errorf("Expected type %s, got %s", TypeWellKnown, retrieved.Type)
	}
	if retrieved.Key != "project-key" {
		t.Errorf("Expected key %s, got %s", "project-key", retrieved.Key)
	}
	if retrieved.Token != "auth-token" {
		t.Errorf("Expected token %s, got %s", "auth-token", retrieved.Token)
	}
}

func TestAllEmpty(t *testing.T) {
	_, cleanup := setupTestAuthFile(t)
	defer cleanup()

	all, err := All()
	if err != nil {
		t.Fatalf("Failed to get all: %v", err)
	}

	if len(all) != 0 {
		t.Errorf("Expected empty map, got %d items", len(all))
	}
}

func TestAllMultipleProviders(t *testing.T) {
	_, cleanup := setupTestAuthFile(t)
	defer cleanup()

	// Set multiple providers
	oauth := OAuth{Type: TypeOAuth, Refresh: "r1", Access: "a1", Expires: 123}
	api := API{Type: TypeAPI, Key: "key1"}
	wellknown := WellKnown{Type: TypeWellKnown, Key: "k1", Token: "t1"}

	if err := Set("anthropic", oauth); err != nil {
		t.Fatalf("Failed to set OAuth: %v", err)
	}
	if err := Set("openai", api); err != nil {
		t.Fatalf("Failed to set API: %v", err)
	}
	if err := Set("gemini", wellknown); err != nil {
		t.Fatalf("Failed to set WellKnown: %v", err)
	}

	// Get all
	all, err := All()
	if err != nil {
		t.Fatalf("Failed to get all: %v", err)
	}

	if len(all) != 3 {
		t.Errorf("Expected 3 items, got %d", len(all))
	}

	// Verify each provider
	if _, ok := all["anthropic"].(OAuth); !ok {
		t.Errorf("Expected OAuth for anthropic")
	}
	if _, ok := all["openai"].(API); !ok {
		t.Errorf("Expected API for openai")
	}
	if _, ok := all["gemini"].(WellKnown); !ok {
		t.Errorf("Expected WellKnown for gemini")
	}
}

func TestUpdateExisting(t *testing.T) {
	_, cleanup := setupTestAuthFile(t)
	defer cleanup()

	// Set initial value
	api1 := API{Type: TypeAPI, Key: "key1"}
	if err := Set("openai", api1); err != nil {
		t.Fatalf("Failed to set initial API: %v", err)
	}

	// Update with new value
	api2 := API{Type: TypeAPI, Key: "key2"}
	if err := Set("openai", api2); err != nil {
		t.Fatalf("Failed to update API: %v", err)
	}

	// Verify updated value
	info, err := Get("openai")
	if err != nil {
		t.Fatalf("Failed to get API: %v", err)
	}

	retrieved, ok := info.(API)
	if !ok {
		t.Fatalf("Expected API type, got %T", info)
	}

	if retrieved.Key != "key2" {
		t.Errorf("Expected updated key %s, got %s", "key2", retrieved.Key)
	}
}

func TestRemove(t *testing.T) {
	_, cleanup := setupTestAuthFile(t)
	defer cleanup()

	// Set multiple providers
	api1 := API{Type: TypeAPI, Key: "key1"}
	api2 := API{Type: TypeAPI, Key: "key2"}
	if err := Set("openai", api1); err != nil {
		t.Fatalf("Failed to set openai: %v", err)
	}
	if err := Set("anthropic", api2); err != nil {
		t.Fatalf("Failed to set anthropic: %v", err)
	}

	// Remove one provider
	if err := Remove("openai"); err != nil {
		t.Fatalf("Failed to remove openai: %v", err)
	}

	// Verify removed
	info, err := Get("openai")
	if err != nil {
		t.Fatalf("Failed to get openai: %v", err)
	}
	if info != nil {
		t.Errorf("Expected nil after removal, got %v", info)
	}

	// Verify other still exists
	info, err = Get("anthropic")
	if err != nil {
		t.Fatalf("Failed to get anthropic: %v", err)
	}
	if info == nil {
		t.Errorf("Expected anthropic to still exist")
	}
}

func TestRemoveNonExistent(t *testing.T) {
	_, cleanup := setupTestAuthFile(t)
	defer cleanup()

	// Remove non-existent provider should not error
	if err := Remove("non-existent"); err != nil {
		t.Errorf("Expected no error removing non-existent, got %v", err)
	}
}

func TestFilePermissions(t *testing.T) {
	filepath, cleanup := setupTestAuthFile(t)
	defer cleanup()

	api := API{Type: TypeAPI, Key: "test-key"}
	if err := Set("test", api); err != nil {
		t.Fatalf("Failed to set auth: %v", err)
	}

	// Check file permissions
	info, err := os.Stat(filepath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	mode := info.Mode().Perm()
	expected := os.FileMode(0o600)
	if mode != expected {
		t.Errorf("Expected file permissions %o, got %o", expected, mode)
	}
}

func TestJSONFormat(t *testing.T) {
	filepath, cleanup := setupTestAuthFile(t)
	defer cleanup()

	// Set auth data
	api := API{Type: TypeAPI, Key: "test-key"}
	if err := Set("openai", api); err != nil {
		t.Fatalf("Failed to set auth: %v", err)
	}

	// Read raw JSON
	data, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Verify it's valid JSON with proper formatting
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	// Verify structure
	if _, ok := parsed["openai"]; !ok {
		t.Errorf("Expected 'openai' key in JSON")
	}

	// Verify indentation (2 spaces, matching TypeScript)
	expected := `{
  "openai": {
    "type": "api",
    "key": "test-key"
  }
}`
	if string(data) != expected {
		t.Errorf("JSON format mismatch.\nExpected:\n%s\nGot:\n%s", expected, string(data))
	}
}

func TestConcurrentAccess(t *testing.T) {
	_, cleanup := setupTestAuthFile(t)
	defer cleanup()

	// Test concurrent reads and writes
	done := make(chan bool, 10)

	for i := 0; i < 5; i++ {
		go func(id int) {
			api := API{Type: TypeAPI, Key: "key"}
			if err := Set("test", api); err != nil {
				t.Errorf("Concurrent set failed: %v", err)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		go func() {
			if _, err := Get("test"); err != nil {
				t.Errorf("Concurrent get failed: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
