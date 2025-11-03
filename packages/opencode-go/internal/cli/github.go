package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var githubCmd = &cobra.Command{
	Use:   "github",
	Short: "GitHub integration commands",
	Long:  `Integrate with GitHub for issue tracking, pull requests, and more.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("⚠️  The 'github' command is coming soon!")
		fmt.Println("This will provide GitHub integration features.")
	},
}
