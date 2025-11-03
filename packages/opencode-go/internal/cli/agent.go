package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage custom agents",
	Long:  `Create, list, and manage custom AI agents with specific capabilities.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("⚠️  The 'agent' command is coming soon!")
		fmt.Println("This will allow you to manage custom agents.")
	},
}
