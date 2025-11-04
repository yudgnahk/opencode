//go:build manual
// +build manual

package session

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	
	"github.com/opencode/opencode-go/internal/storage"
	"github.com/opencode/opencode-go/internal/provider"
)

// TestListRealData tests listing sessions from actual ~/.local/share/opencode/storage
// Run with: go test -tags=manual -v ./internal/session/ -run TestListRealData
func TestListRealData(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	
	storagePath := filepath.Join(home, ".local", "share", "opencode", "storage")
	
	t.Logf("Using storage path: %s", storagePath)
	
	// Check if directory exists
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		t.Skipf("Storage directory does not exist: %s", storagePath)
	}
	
	store, err := storage.New(storagePath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()
	
	registry := provider.NewRegistry()
	manager := NewManager(store, registry)
	
	ctx := context.Background()
	
	// Test ListLight (fast)
	t.Run("ListLight", func(t *testing.T) {
		items, err := manager.ListLight(ctx, "")
		if err != nil {
			t.Fatalf("Error listing sessions: %v", err)
		}
		
		t.Logf("Found %d sessions using ListLight", len(items))
		
		if len(items) > 0 {
			t.Logf("\nFirst 3 sessions:")
			for i, item := range items {
				if i >= 3 {
					break
				}
				data, _ := json.MarshalIndent(item, "  ", "  ")
				t.Logf("%d. %s", i+1, data)
			}
		}
	})
	
	// Test List (full)
	t.Run("List", func(t *testing.T) {
		sessions, err := manager.List(ctx, "")
		if err != nil {
			t.Fatalf("Error listing sessions: %v", err)
		}
		
		t.Logf("Found %d sessions using List", len(sessions))
		
		if len(sessions) > 0 {
			t.Logf("First session has %d message IDs", len(sessions[0].MessageIDs))
		}
	})
}
