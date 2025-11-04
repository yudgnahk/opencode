package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	
	"github.com/opencode/opencode-go/internal/provider"
)

func main() {
	authData, err := os.ReadFile("/Users/kha@dnanexus.com/.local/share/opencode/auth.json")
	if err != nil {
		log.Fatal("Failed to read auth file:", err)
	}
	
	var auth map[string]struct {
		Access string
	}
	err = json.Unmarshal(authData, &auth)
	if err != nil {
		log.Fatal("Failed to parse auth file:", err)
	}
	
	token := auth["github-copilot"].Access
	if token == "" {
		log.Fatal("No GitHub Copilot token found")
	}
	
	fmt.Println("Creating GitHub Copilot provider...")
	p := provider.NewGitHubCopilot(token)
	
	fmt.Println("Sending test message...")
	resp, err := p.Complete(context.Background(), provider.CompletionRequest{
		Model: "claude-3.5-sonnet",
		Messages: []provider.Message{
			{
				Role: "user",
				Content: []provider.ContentBlock{
					{Type: "text", Text: "Say hello in exactly 5 words"},
				},
			},
		},
	})
	
	if err != nil {
		log.Fatal("Error:", err)
	}
	
	fmt.Printf("Success! Response: %+v\n", resp)
}
