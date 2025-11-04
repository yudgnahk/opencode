package session

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/opencode/opencode-go/internal/provider"
	"github.com/opencode/opencode-go/internal/storage"
)

// BenchmarkList compares performance of List vs ListLight
func BenchmarkList(b *testing.B) {
	// Create temp directory for test storage
	tempDir, err := os.MkdirTemp("", "opencode-bench-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize storage and manager
	store, err := storage.New(tempDir)
	if err != nil {
		b.Fatal(err)
	}
	defer store.Close()

	registry := provider.NewRegistry()
	manager := NewManager(store, registry)

	// Create test sessions
	ctx := context.Background()
	numSessions := 100
	projectID := "test-project"

	for i := 0; i < numSessions; i++ {
		_, err := manager.Create(ctx, projectID, "/test/dir", "1.0.0")
		if err != nil {
			b.Fatal(err)
		}
	}

	b.Run("List", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := manager.List(ctx, projectID)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ListLight", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := manager.ListLight(ctx, projectID)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkListScaling tests performance with different numbers of sessions
func BenchmarkListScaling(b *testing.B) {
	sizes := []int{10, 50, 100, 500}

	for _, size := range sizes {
		b.Run(filepath.Join("sessions", string(rune('0'+size/100)), string(rune('0'+(size/10)%10)), string(rune('0'+size%10))), func(b *testing.B) {
			// Create temp directory for test storage
			tempDir, err := os.MkdirTemp("", "opencode-bench-*")
			if err != nil {
				b.Fatal(err)
			}
			defer os.RemoveAll(tempDir)

			// Initialize storage and manager
			store, err := storage.New(tempDir)
			if err != nil {
				b.Fatal(err)
			}
			defer store.Close()

			registry := provider.NewRegistry()
			manager := NewManager(store, registry)

			// Create test sessions
			ctx := context.Background()
			projectID := "test-project"

			for i := 0; i < size; i++ {
				_, err := manager.Create(ctx, projectID, "/test/dir", "1.0.0")
				if err != nil {
					b.Fatal(err)
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := manager.ListLight(ctx, projectID)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
