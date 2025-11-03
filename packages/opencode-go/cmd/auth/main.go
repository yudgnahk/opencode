package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/opencode/opencode-go/internal/auth"
	"github.com/opencode/opencode-go/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "list", "ls":
		listAuth()
	case "set":
		if len(os.Args) < 4 {
			fmt.Println("Usage: auth set <provider> <api-key>")
			os.Exit(1)
		}
		setAuth(os.Args[2], os.Args[3])
	case "remove", "rm", "logout":
		if len(os.Args) < 3 {
			fmt.Println("Usage: auth remove <provider>")
			os.Exit(1)
		}
		removeAuth(os.Args[2])
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("OpenCode Auth CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  auth list                    List all stored credentials")
	fmt.Println("  auth set <provider> <key>    Store API key for provider")
	fmt.Println("  auth remove <provider>       Remove credentials for provider")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  auth list")
	fmt.Println("  auth set anthropic sk-ant-xxx")
	fmt.Println("  auth set openai sk-xxx")
	fmt.Println("  auth remove anthropic")
	fmt.Println()
	fmt.Printf("Auth file: %s\n", config.GetDataDir()+"/auth.json")
}

func listAuth() {
	allAuth, err := auth.All()
	if err != nil {
		fmt.Printf("Error reading auth: %v\n", err)
		os.Exit(1)
	}

	if len(allAuth) == 0 {
		fmt.Println("No credentials stored")
		return
	}

	fmt.Printf("Stored credentials (%d):\n", len(allAuth))
	for provider, info := range allAuth {
		fmt.Printf("  - %s (%s)\n", provider, info.GetType())
	}
	fmt.Println()
	fmt.Printf("Auth file: %s\n", config.GetDataDir()+"/auth.json")
}

func setAuth(provider, key string) {
	// Create API auth
	apiAuth := auth.API{
		Type: auth.TypeAPI,
		Key:  key,
	}

	if err := auth.Set(provider, apiAuth); err != nil {
		fmt.Printf("Error setting auth: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Stored API key for %s\n", provider)
}

func removeAuth(provider string) {
	// Check if exists
	existing, err := auth.Get(provider)
	if err != nil {
		fmt.Printf("Error checking auth: %v\n", err)
		os.Exit(1)
	}
	if existing == nil {
		fmt.Printf("No credentials found for %s\n", provider)
		os.Exit(1)
	}

	if err := auth.Remove(provider); err != nil {
		fmt.Printf("Error removing auth: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Removed credentials for %s\n", provider)
}

func prettyPrint(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
}
