package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var attachCmd = &cobra.Command{
	Use:   "attach",
	Short: "Attach to a running OpenCode server",
	Long:  `Connect to and display information about a running OpenCode server.`,
	Run: func(cmd *cobra.Command, args []string) {
		serverURL := os.Getenv("OPENCODE_SERVER")
		if serverURL == "" {
			serverURL = "http://localhost:8124"
		}

		ctx := context.Background()

		// Check server health
		healthResp, err := http.Get(serverURL + "/health")
		if err != nil {
			fmt.Printf("❌ Failed to connect to server at %s: %v\n", serverURL, err)
			fmt.Println("\n💡 Make sure the server is running with: opencode serve")
			os.Exit(1)
		}
		defer healthResp.Body.Close()

		if healthResp.StatusCode != http.StatusOK {
			fmt.Printf("❌ Server returned error status: %d\n", healthResp.StatusCode)
			os.Exit(1)
		}

		healthBody, _ := io.ReadAll(healthResp.Body)
		var healthData map[string]any
		json.Unmarshal(healthBody, &healthData)

		fmt.Printf("✅ Connected to OpenCode server at %s\n\n", serverURL)
		fmt.Println("Server Status:")
		fmt.Printf("  Status: %v\n", healthData["status"])

		// Get models
		req, _ := http.NewRequestWithContext(ctx, "GET", serverURL+"/models", nil)
		modelsResp, err := http.DefaultClient.Do(req)
		if err == nil && modelsResp.StatusCode == http.StatusOK {
			defer modelsResp.Body.Close()
			modelsBody, _ := io.ReadAll(modelsResp.Body)
			var modelsData map[string]any
			json.Unmarshal(modelsBody, &modelsData)

			if models, ok := modelsData["models"].([]any); ok {
				fmt.Printf("  Available models: %d\n", len(models))
			}
		}

		// Get sessions
		req, _ = http.NewRequestWithContext(ctx, "GET", serverURL+"/sessions", nil)
		sessionsResp, err := http.DefaultClient.Do(req)
		if err == nil && sessionsResp.StatusCode == http.StatusOK {
			defer sessionsResp.Body.Close()
			sessionsBody, _ := io.ReadAll(sessionsResp.Body)
			var sessionsData []any
			json.Unmarshal(sessionsBody, &sessionsData)

			fmt.Printf("  Active sessions: %d\n", len(sessionsData))
		}

		fmt.Println("\n📊 Server is running and accepting connections")
	},
}
