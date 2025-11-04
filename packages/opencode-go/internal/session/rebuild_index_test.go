// go:build manual

package session

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/opencode/opencode-go/internal/config"
	"github.com/opencode/opencode-go/internal/provider"
	"github.com/opencode/opencode-go/internal/storage"
)

// TestRebuildIndex manually rebuilds the session index from actual session files
// Run with: go test -tags=manual -v ./internal/session/ -run TestRebuildIndex
func TestRebuildIndex(t *testing.T) {
	// Get storage directory
	storageDir := filepath.Join(config.Path.Data, "storage")

	// Initialize storage
	st, err := storage.New(storageDir)
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	// Delete old index file to force rebuild
	indexPath := filepath.Join(storageDir, "_index", "sessions.json")
	if err := os.Remove(indexPath); err != nil && !os.IsNotExist(err) {
		t.Logf("Warning: Failed to remove old index: %v", err)
	} else {
		t.Log("Old index removed")
	}

	// Initialize provider registry
	registry := provider.NewRegistry()

	// Initialize session manager (this will rebuild the index)
	t.Log("Rebuilding session index...")
	mgr := NewManager(st, registry)

	// Get session count to verify
	ctx := context.Background()
	sessions, err := mgr.ListLight(ctx, "")
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}

	t.Logf("Index rebuilt successfully! Found %d sessions", len(sessions))

	// Print first few sessions to verify schema
	for i, sess := range sessions {
		if i >= 3 {
			break
		}
		t.Logf("Session %d: ID=%s, Created=%d, Updated=%d, Version=%s, Title=%s",
			i+1, sess.ID, sess.Created, sess.Updated, sess.Version, sess.Title)
	}
}
