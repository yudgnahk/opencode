package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "0.1.0-go"

var rootCmd = &cobra.Command{
	Use:   "opencode",
	Short: "OpenCode - AI-powered coding assistant",
	Long: `
 ▗▄▖ ▗▄▄▖ ▗▄▄▄▖▗▖  ▗▖ ▗▄▄▖ ▗▄▖ ▗▄▄▄ ▗▄▄▄▖
▐▌ ▐▌▐▌ ▐▌▐▌   ▐▛▚▖▐▌▐▌   ▐▌ ▐▌▐▌  █▐▌   
▐▌ ▐▌▐▛▀▘ ▐▛▀▀▘▐▌ ▝▜▌▐▌   ▐▌ ▐▌▐▌  █▐▛▀▀▘
▝▚▄▞▘▐▌   ▐▙▄▄▖▐▌  ▐▌▝▚▄▄▖▝▚▄▞▘▐▙▄▄▀▐▙▄▄▖

OpenCode is an AI-powered coding assistant that helps you write,
understand, and debug code faster.`,
	Version: version,
}

var (
	printLogs bool
	logLevel  string
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&printLogs, "print-logs", false, "print logs to stderr")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "INFO", "log level (DEBUG, INFO, WARN, ERROR)")

	// Add subcommands
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(modelsCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(attachCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(agentCmd)
	rootCmd.AddCommand(upgradeCmd)
	rootCmd.AddCommand(debugCmd)
	rootCmd.AddCommand(mcpCmd)
	rootCmd.AddCommand(githubCmd)
	rootCmd.AddCommand(acpCmd)
}

func Execute() error {
	return rootCmd.Execute()
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("OpenCode %s\n", version)
	},
}
