package cli

import (
	"fmt"
	"os"

	"github.com/opencode/opencode-go/internal/auth"
	"github.com/opencode/opencode-go/internal/config"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication credentials",
	Long:  `Store and manage API keys and authentication credentials for AI providers.`,
}

var authListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all stored credentials",
	RunE:    runAuthList,
}

var authSetCmd = &cobra.Command{
	Use:   "set <provider> <api-key>",
	Short: "Store API key for a provider",
	Args:  cobra.ExactArgs(2),
	RunE:  runAuthSet,
}

var authRemoveCmd = &cobra.Command{
	Use:     "remove <provider>",
	Aliases: []string{"rm", "logout"},
	Short:   "Remove credentials for a provider",
	Args:    cobra.ExactArgs(1),
	RunE:    runAuthRemove,
}

func init() {
	authCmd.AddCommand(authListCmd)
	authCmd.AddCommand(authSetCmd)
	authCmd.AddCommand(authRemoveCmd)
}

func runAuthList(cmd *cobra.Command, args []string) error {
	allAuth, err := auth.All()
	if err != nil {
		return fmt.Errorf("error reading auth: %w", err)
	}

	if len(allAuth) == 0 {
		fmt.Println("No credentials stored")
		return nil
	}

	fmt.Printf("Stored credentials (%d):\n", len(allAuth))
	for provider, info := range allAuth {
		fmt.Printf("  - %s (%s)\n", provider, info.GetType())
	}
	fmt.Println()
	fmt.Printf("Auth file: %s\n", config.GetDataDir()+"/auth.json")
	return nil
}

func runAuthSet(cmd *cobra.Command, args []string) error {
	provider := args[0]
	key := args[1]

	// Create API auth
	apiAuth := auth.API{
		Type: auth.TypeAPI,
		Key:  key,
	}

	if err := auth.Set(provider, apiAuth); err != nil {
		return fmt.Errorf("error setting auth: %w", err)
	}

	fmt.Printf("✓ Stored API key for %s\n", provider)
	return nil
}

func runAuthRemove(cmd *cobra.Command, args []string) error {
	provider := args[0]

	// Check if exists
	existing, err := auth.Get(provider)
	if err != nil {
		return fmt.Errorf("error checking auth: %w", err)
	}
	if existing == nil {
		fmt.Fprintf(os.Stderr, "No credentials found for %s\n", provider)
		os.Exit(1)
	}

	if err := auth.Remove(provider); err != nil {
		return fmt.Errorf("error removing auth: %w", err)
	}

	fmt.Printf("✓ Removed credentials for %s\n", provider)
	return nil
}
