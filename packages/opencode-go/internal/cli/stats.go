package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show usage statistics",
	Long:  `Display statistics about your OpenCode usage.`,
	RunE:  runStats,
}

func runStats(cmd *cobra.Command, args []string) error {
	fmt.Println("Usage statistics - Coming soon!")
	return nil
}
