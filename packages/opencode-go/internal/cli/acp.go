package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var acpCmd = &cobra.Command{
	Use:   "acp",
	Short: "ACP (Agent Communication Protocol) commands",
	Long:  `Manage ACP protocol for agent-to-agent communication.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("⚠️  The 'acp' command is coming soon!")
		fmt.Println("This will provide ACP protocol features.")
	},
}
