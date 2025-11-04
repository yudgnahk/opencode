package lsp

import (
	"context"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.servers == nil {
		t.Fatal("NewClient created client with nil servers map")
	}
	if len(client.servers) != 0 {
		t.Errorf("NewClient created client with %d servers, want 0", len(client.servers))
	}
}

func TestStartServerNotFound(t *testing.T) {
	client := NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Try to start a server with a command that doesn't exist
	err := client.Start(ctx, "testlang", "nonexistent-command-12345", []string{})
	if err == nil {
		t.Error("Start with nonexistent command should fail")
	}

	// Verify server was not added
	client.mu.RLock()
	_, exists := client.servers["testlang"]
	client.mu.RUnlock()

	if exists {
		t.Error("Failed start should not add server to map")
	}
}

func TestStartServerTwice(t *testing.T) {
	client := NewClient()

	// Mock a server by adding directly to the map
	client.mu.Lock()
	_, cancel := context.WithCancel(context.Background())
	client.servers["testlang"] = &LanguageServer{
		cancel: cancel,
	}
	client.mu.Unlock()

	// Try to start the same server again
	err := client.Start(context.Background(), "testlang", "some-command", []string{})
	if err != nil {
		t.Errorf("Starting already-running server should not error, got: %v", err)
	}

	// Cleanup
	cancel()
	client.mu.Lock()
	delete(client.servers, "testlang")
	client.mu.Unlock()
}

func TestStopNonexistentServer(t *testing.T) {
	client := NewClient()

	// Stopping a server that doesn't exist should not error
	err := client.Stop("nonexistent")
	if err != nil {
		t.Errorf("Stop nonexistent server should not error, got: %v", err)
	}
}

func TestStopAll(t *testing.T) {
	client := NewClient()

	// Add some mock servers
	_, cancel1 := context.WithCancel(context.Background())
	_, cancel2 := context.WithCancel(context.Background())

	client.mu.Lock()
	client.servers["lang1"] = &LanguageServer{cancel: cancel1}
	client.servers["lang2"] = &LanguageServer{cancel: cancel2}
	client.mu.Unlock()

	// Stop all
	client.StopAll()

	// Verify all servers were removed
	client.mu.RLock()
	count := len(client.servers)
	client.mu.RUnlock()

	if count != 0 {
		t.Errorf("StopAll left %d servers, want 0", count)
	}
}

func TestHoverServerNotRunning(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	_, err := client.Hover(ctx, "go", "/path/to/file.go", 10, 5)
	if err == nil {
		t.Error("Hover with no running server should fail")
	}
}

func TestCompletionServerNotRunning(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	_, err := client.Completion(ctx, "go", "/path/to/file.go", 10, 5)
	if err == nil {
		t.Error("Completion with no running server should fail")
	}
}

func TestDefinitionServerNotRunning(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	_, err := client.Definition(ctx, "go", "/path/to/file.go", 10, 5)
	if err == nil {
		t.Error("Definition with no running server should fail")
	}
}

func TestReferencesServerNotRunning(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	_, err := client.References(ctx, "go", "/path/to/file.go", 10, 5)
	if err == nil {
		t.Error("References with no running server should fail")
	}
}

func TestCloseDocumentServerNotRunning(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	err := client.CloseDocument(ctx, "go", "/path/to/file.go")
	if err == nil {
		t.Error("CloseDocument with no running server should fail")
	}
}

// TestReadWriteCloser tests that readWriteCloser properly combines readers and writers
// Note: Integration tests will cover actual usage with real LSP servers.
