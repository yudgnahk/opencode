package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/opencode/opencode-go/internal/config"
	"github.com/opencode/opencode-go/internal/provider"
	"github.com/opencode/opencode-go/internal/session"
	"github.com/opencode/opencode-go/internal/storage"
)

func main() {
	// Initialize config (not strictly needed but loads XDG paths)
	_, err := config.Load()
	if err != nil {
		log.Printf("Warning: Failed to load config: %v", err)
	}

	// Initialize storage using XDG data directory
	storageDir := filepath.Join(config.Path.Data, "storage")
	st, err := storage.New(storageDir)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer st.Close()

	// Initialize provider registry (needed for session manager)
	registry := provider.NewRegistry()

	// Delete old index file
	indexPath := filepath.Join(storageDir, "_index", "sessions.json")
	if err := os.Remove(indexPath); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Failed to remove old index: %v", err)
	} else {
		fmt.Println("Old index removed")
	}

	// Initialize session manager (this will rebuild the index)
	fmt.Println("Rebuilding session index...")
	mgr := session.NewManager(st, registry)

	// Get session count to verify
	ctx := context.Background()
	sessions, err := mgr.ListLight(ctx, "")
	if err != nil {
		log.Fatalf("Failed to list sessions: %v", err)
	}
	fmt.Printf("Index rebuilt successfully! Found %d sessions\n", len(sessions))
}
