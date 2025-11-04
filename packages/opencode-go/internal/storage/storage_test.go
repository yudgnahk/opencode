package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Test successful creation
	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	// Verify database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("Database file was not created")
	}
}

func TestNewCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "nested", "dir", "test.db")

	// Test that New creates nested directories
	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage with nested dir: %v", err)
	}
	defer store.Close()

	// Verify directory was created
	if _, err := os.Stat(filepath.Dir(dbPath)); os.IsNotExist(err) {
		t.Fatal("Directory was not created")
	}
}

func TestGetSet(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := New(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	// Test Set
	testKey := "test-key"
	testValue := []byte("test-value")
	err = store.Set("sessions", testKey, testValue)
	if err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	// Test Get
	retrieved, err := store.Get("sessions", testKey)
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}

	if string(retrieved) != string(testValue) {
		t.Fatalf("Expected %s, got %s", testValue, retrieved)
	}
}

func TestGetNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := New(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	// Test getting non-existent key
	retrieved, err := store.Get("sessions", "non-existent")
	if err != nil {
		t.Fatalf("Get should not error on non-existent key: %v", err)
	}

	if retrieved != nil {
		t.Fatalf("Expected nil for non-existent key, got %v", retrieved)
	}
}

func TestDelete(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := New(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	// Set a value
	testKey := "test-key"
	testValue := []byte("test-value")
	err = store.Set("sessions", testKey, testValue)
	if err != nil {
		t.Fatal(err)
	}

	// Delete it
	err = store.Delete([]string{"sessions", testKey})
	if err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}

	// Verify it's gone
	retrieved, err := store.Get("sessions", testKey)
	if err != nil {
		t.Fatal(err)
	}

	if retrieved != nil {
		t.Fatal("Key should have been deleted")
	}
}

func TestGetSetJSON(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := New(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	// Test struct
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	original := TestStruct{
		Name:  "test",
		Value: 42,
	}

	// Set JSON
	err = store.SetJSON("sessions", "test-json", original)
	if err != nil {
		t.Fatalf("Failed to set JSON: %v", err)
	}

	// Get JSON
	var retrieved TestStruct
	err = store.GetJSON("sessions", "test-json", &retrieved)
	if err != nil {
		t.Fatalf("Failed to get JSON: %v", err)
	}

	if retrieved.Name != original.Name || retrieved.Value != original.Value {
		t.Fatalf("Retrieved JSON doesn't match. Expected %+v, got %+v", original, retrieved)
	}
}

func TestGetJSONNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := New(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	type TestStruct struct {
		Name string `json:"name"`
	}

	var result TestStruct
	err = store.GetJSON("sessions", "non-existent", &result)
	if err != nil {
		t.Fatalf("GetJSON should not error on non-existent key: %v", err)
	}

	// result should be unchanged (zero value)
	if result.Name != "" {
		t.Fatal("Result should be zero value for non-existent key")
	}
}

func TestSetJSONInvalidData(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := New(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	// Try to marshal something that can't be marshaled
	invalidData := make(chan int)
	err = store.SetJSON("sessions", "test", invalidData)
	if err == nil {
		t.Fatal("Expected error when marshaling invalid data")
	}
}

func TestList(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := New(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	// Add multiple items
	items := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for k, v := range items {
		err := store.Set("sessions", k, []byte(v))
		if err != nil {
			t.Fatal(err)
		}
	}

	// List all keys
	keys, err := store.List([]string{"sessions"})
	if err != nil {
		t.Fatalf("Failed to list keys: %v", err)
	}

	if len(keys) != len(items) {
		t.Fatalf("Expected %d keys, got %d", len(items), len(keys))
	}

	// Verify all keys are present
	keyMap := make(map[string]bool)
	for _, k := range keys {
		keyMap[k[1]] = true
	}

	for k := range items {
		if !keyMap[k] {
			t.Fatalf("Key %s not found in list", k)
		}
	}
}

func TestListEmptyBucket(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := New(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	// List empty bucket
	keys, err := store.List([]string{"sessions"})
	if err != nil {
		t.Fatalf("Failed to list empty bucket: %v", err)
	}

	if len(keys) != 0 {
		t.Fatalf("Expected 0 keys in empty bucket, got %d", len(keys))
	}
}

func TestBucketsCreated(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := New(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	// Test that all required buckets were created
	buckets := []string{"sessions", "messages", "projects", "config"}
	for _, bucket := range buckets {
		// Try to set a value in each bucket
		err := store.Set(bucket, "test", []byte("test"))
		if err != nil {
			t.Fatalf("Bucket %s was not created: %v", bucket, err)
		}
	}
}

func TestClose(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := New(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	// Close the storage
	err = store.Close()
	if err != nil {
		t.Fatalf("Failed to close storage: %v", err)
	}

	// Verify we can open it again
	store2, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to reopen database: %v", err)
	}
	defer store2.Close()
}

func TestConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := New(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	// Test concurrent writes
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(n int) {
			key := string(rune('a' + n))
			value := []byte{byte(n)}
			err := store.Set("sessions", key, value)
			if err != nil {
				t.Errorf("Concurrent set failed: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all writes succeeded
	keys, err := store.List([]string{"sessions"})
	if err != nil {
		t.Fatal(err)
	}

	if len(keys) != 10 {
		t.Fatalf("Expected 10 keys after concurrent writes, got %d", len(keys))
	}
}
