package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "List available AI models",
	Long:  `Display a list of all available AI models from configured providers.`,
	RunE:  runModels,
}

func runModels(cmd *cobra.Command, args []string) error {
	fmt.Println("Available AI models:")
	fmt.Println()
	fmt.Println("Anthropic:")
	fmt.Println("  - claude-3-5-sonnet-20241022 (recommended)")
	fmt.Println("  - claude-3-5-haiku-20241022")
	fmt.Println("  - claude-3-opus-20240229")
	fmt.Println()
	fmt.Println("OpenAI:")
	fmt.Println("  - gpt-4o")
	fmt.Println("  - gpt-4o-mini")
	fmt.Println("  - gpt-4-turbo")
	fmt.Println()
	fmt.Println("Google:")
	fmt.Println("  - gemini-2.0-flash-exp")
	fmt.Println("  - gemini-1.5-pro")
	fmt.Println()
	fmt.Println("Use 'opencode auth set <provider> <api-key>' to configure providers.")
	return nil
}
