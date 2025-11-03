package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Manage MCP (Model Context Protocol) servers",
	Long:  `Install, configure, and manage MCP servers for extended functionality.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("⚠️  The 'mcp' command is coming soon!")
		fmt.Println("This will allow you to manage MCP servers.")
	},
}
