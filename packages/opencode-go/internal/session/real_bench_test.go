//go:build manual
// +build manual

package session

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	
	"github.com/opencode/opencode-go/internal/storage"
	"github.com/opencode/opencode-go/internal/provider"
)

// BenchmarkListRealData benchmarks listing sessions from actual storage
// Run with: go test -tags=manual -bench=. -benchmem ./internal/session/ -run=^$
func BenchmarkListRealData(b *testing.B) {
	home, err := os.UserHomeDir()
	if err != nil {
		b.Fatal(err)
	}
	
	storagePath := filepath.Join(home, ".local", "share", "opencode", "storage")
	
	// Check if directory exists
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		b.Skipf("Storage directory does not exist: %s", storagePath)
	}
	
	store, err := storage.New(storagePath)
	if err != nil {
		b.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()
	
	registry := provider.NewRegistry()
	manager := NewManager(store, registry)
	
	ctx := context.Background()
	
	b.Run("List-Real", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := manager.List(ctx, "")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	
	b.Run("ListLight-Real", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := manager.ListLight(ctx, "")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
