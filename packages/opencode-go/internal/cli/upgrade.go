package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade OpenCode to the latest version",
	Long:  `Check for and install the latest version of OpenCode.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("⚠️  The 'upgrade' command is coming soon!")
		fmt.Println("This will allow you to upgrade OpenCode to the latest version.")
		fmt.Printf("Current version: %s\n", version)
	},
}
