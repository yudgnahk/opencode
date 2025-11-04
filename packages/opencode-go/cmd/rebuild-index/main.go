package main

import (
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
	// Initialize config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize storage
	storageDir := filepath.Join(cfg.DataDir, "storage")
	st, err := storage.New(storageDir)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize provider registry (needed for session manager)
	registry, err := provider.NewRegistry(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize provider registry: %v", err)
	}

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
	sessions := mgr.ListLight("", nil)
	fmt.Printf("Index rebuilt successfully! Found %d sessions\n", len(sessions))
}
