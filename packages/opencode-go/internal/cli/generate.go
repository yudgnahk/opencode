package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate OpenAPI specs or other artifacts",
	Long:  `Generate OpenAPI specifications or other code artifacts from your project.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("⚠️  The 'generate' command is coming soon!")
		fmt.Println("This will generate OpenAPI specs and other artifacts from your code.")
	},
}
