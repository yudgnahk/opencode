package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/opencode/opencode-go/internal/auth"
	"github.com/opencode/opencode-go/internal/provider"
	"github.com/opencode/opencode-go/internal/session"
	"github.com/opencode/opencode-go/internal/storage"
)

func main() {
	ctx := context.Background()

	// Get GitHub Copilot token
	authInfo, err := auth.Get("github-copilot")
	if err != nil || authInfo == nil {
		log.Fatal("No GitHub Copilot auth found. Run: opencode auth github-copilot")
	}

	oauthAuth, ok := authInfo.(auth.OAuth)
	if !ok || oauthAuth.Access == "" {
		log.Fatal("Invalid GitHub Copilot auth format")
	}

	fmt.Println("Creating provider registry...")
	registry := provider.NewRegistry()

	// Create GitHub Copilot provider
	copilot := provider.NewGitHubCopilot(oauthAuth.Access)
	retryConfig := provider.DefaultRetryConfig()
	wrapped := provider.NewRetryableProvider(copilot, retryConfig)
	if err := registry.Register(wrapped); err != nil {
		log.Fatalf("Failed to register provider: %v", err)
	}

	// Create storage
	fmt.Println("Creating storage...")
	tmpDir, err := os.MkdirTemp("", "opencode-test-*")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := storage.New(tmpDir)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}

	// Create session manager
	fmt.Println("Creating session manager...")
	sessionMgr := session.NewManager(store, registry)

	// Create completion service
	fmt.Println("Creating completion service...")
	completionSvc := session.NewCompletionService(sessionMgr, registry)

	// Create a session
	fmt.Println("Creating session...")
	sess, err := sessionMgr.Create(ctx, "test-project", "github-copilot", "gpt-4o")
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	fmt.Printf("Session created: %s\n", sess.ID)

	// Send a message (this mimics what the TUI does)
	fmt.Println("\nSending message to completion service...")
	if err := completionSvc.SendMessage(ctx, sess.ID, "Hello, how are you?"); err != nil {
		log.Fatalf("SendMessage failed: %v", err)
	}

	fmt.Println("\n✓ Message sent successfully!")

	// Get messages to verify
	messages, err := sessionMgr.GetMessages(ctx, sess.ID)
	if err != nil {
		log.Fatalf("Failed to get messages: %v", err)
	}

	fmt.Printf("\nMessages in session (%d):\n", len(messages))
	for i, msg := range messages {
		contentJSON, _ := json.MarshalIndent(msg.Content, "  ", "  ")
		fmt.Printf("%d. Role: %s\n  Content: %s\n", i+1, msg.Role, contentJSON)
	}
}
