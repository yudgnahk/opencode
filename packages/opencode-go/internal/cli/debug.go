package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Debug commands for troubleshooting",
	Long:  `Various debug commands to help troubleshoot issues with OpenCode.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("⚠️  The 'debug' command is coming soon!")
		fmt.Println("This will provide various debugging utilities.")
	},
}
