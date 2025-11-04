package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/opencode/opencode-go/internal/auth"
	"github.com/opencode/opencode-go/internal/modelsdev"
	"github.com/opencode/opencode-go/internal/provider"
	"github.com/spf13/cobra"
)

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "List available AI models",
	Long:  `Display a list of all available AI models from configured providers.`,
	RunE:  runModels,
}

func runModels(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Initialize models.dev registry
	modelsDevReg := modelsdev.NewRegistry(modelsdev.GetDefaultCacheDir())
	if err := modelsDevReg.Load(); err != nil {
		// Log warning but continue with fallback
		fmt.Printf("Warning: failed to load models.dev: %v\n", err)
	}

	// Get all provider credentials
	allAuth, err := auth.All()
	if err != nil {
		return fmt.Errorf("failed to get authentication: %w", err)
	}

	// Create registry and register providers
	registry := provider.NewRegistry()

	// Register providers based on available credentials
	for providerID, authInfo := range allAuth {
		var p provider.Provider
		var regErr error

		// Extract API key based on auth type
		var apiKey string
		switch info := authInfo.(type) {
		case auth.API:
			apiKey = info.Key
		case auth.WellKnown:
			apiKey = info.Key
		case auth.OAuth:
			// OAuth providers need special handling, skip for now
			continue
		default:
			continue
		}

		// Create provider based on type
		normalizedID := strings.ToLower(providerID)
		switch normalizedID {
		case "anthropic":
			p = provider.NewAnthropic(apiKey)
		case "openai":
			p = provider.NewOpenAI(apiKey)
		case "gemini", "google":
			p, regErr = provider.NewGemini(ctx, apiKey)
		default:
			continue
		}

		if regErr != nil {
			continue
		}

		// Set models.dev registry on the provider
		p.SetModelsDevRegistry(modelsDevReg)

		if err := registry.Register(p); err != nil {
			continue
		}
	}

	// Get all models from registry
	allModels := registry.ListModels()

	if len(allModels) == 0 {
		fmt.Println("No providers configured.")
		fmt.Println()
		fmt.Println("Use 'opencode auth set <provider> <api-key>' to configure providers.")
		fmt.Println()
		fmt.Println("Supported providers:")
		fmt.Println("  - anthropic")
		fmt.Println("  - openai")
		fmt.Println("  - google")
		return nil
	}

	// Sort provider names for consistent output
	providerNames := make([]string, 0, len(allModels))
	for name := range allModels {
		providerNames = append(providerNames, name)
	}
	sort.Strings(providerNames)

	// Display models
	for _, providerName := range providerNames {
		models := allModels[providerName]
		if len(models) == 0 {
			continue
		}

		fmt.Printf("%s:\n", providerName)
		for _, model := range models {
			fmt.Printf("  %s/%s\n", providerName, model)
		}
		fmt.Println()
	}

	return nil
}
