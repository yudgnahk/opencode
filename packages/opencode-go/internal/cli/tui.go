package cli

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/opencode/opencode-go/internal/config"
	"github.com/spf13/cobra"
)

var (
	tuiModel   string
	tuiPrompt  string
	tuiAgent   string
	tuiSession string
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Start the interactive terminal UI",
	Long:  `Launches the OpenCode terminal UI for interactive coding sessions.`,
	RunE:  runTui,
}

func init() {
	tuiCmd.Flags().StringVarP(&tuiModel, "model", "m", "", "model to use (provider/model)")
	tuiCmd.Flags().StringVarP(&tuiPrompt, "prompt", "p", "", "prompt to begin with")
	tuiCmd.Flags().StringVar(&tuiAgent, "agent", "", "agent to begin with")
	tuiCmd.Flags().StringVarP(&tuiSession, "session", "s", "", "session ID to continue")
}

func runTui(cmd *cobra.Command, args []string) error {
	// Get server URL from env or config
	serverURL := os.Getenv("OPENCODE_SERVER")
	if serverURL == "" {
		cfg, err := config.Load()
		if err == nil {
			serverURL = fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)
		} else {
			serverURL = "http://localhost:8080"
		}
	}

	// Check if server is running
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(serverURL + "/health")
	if err != nil || resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Error: OpenCode server is not running at %s\n", serverURL)
		fmt.Fprintln(os.Stderr, "\nPlease start the server first:")
		fmt.Fprintln(os.Stderr, "  opencode serve")
		fmt.Fprintln(os.Stderr, "\nOr set OPENCODE_SERVER environment variable to point to a running server.")
		return fmt.Errorf("server not reachable")
	}

	// Find TUI binary - try multiple locations
	tuiPath := ""
	possiblePaths := []string{
		// Relative to project root when running from bin/opencode
		filepath.Join("packages", "tui", "cmd", "opencode", "dist", "tui"),
		// Relative to current directory (development)
		filepath.Join("..", "..", "tui", "cmd", "opencode", "dist", "tui"),
		// In PATH
		"tui",
	}

	for _, path := range possiblePaths {
		if absPath, err := filepath.Abs(path); err == nil {
			if _, err := os.Stat(absPath); err == nil {
				tuiPath = absPath
				break
			}
		}
	}

	// If not found, try to find it via the executable path
	if tuiPath == "" {
		execPath, err := os.Executable()
		if err == nil {
			execDir := filepath.Dir(execPath)
			// If running from bin/, go up to project root
			projectRoot := filepath.Join(execDir, "..")
			testPath := filepath.Join(projectRoot, "packages", "tui", "cmd", "opencode", "dist", "tui")
			if _, err := os.Stat(testPath); err == nil {
				tuiPath = testPath
			}
		}
	}

	if tuiPath == "" {
		return fmt.Errorf("TUI binary not found. Please run 'make build-tui' first")
	}

	// Build TUI command args
	tuiArgs := []string{}
	if tuiModel != "" {
		tuiArgs = append(tuiArgs, "--model", tuiModel)
	}
	if tuiPrompt != "" {
		tuiArgs = append(tuiArgs, "--prompt", tuiPrompt)
	}
	if tuiAgent != "" {
		tuiArgs = append(tuiArgs, "--agent", tuiAgent)
	}
	if tuiSession != "" {
		tuiArgs = append(tuiArgs, "--session", tuiSession)
	}

	// Launch TUI
	tuiCmd := exec.Command(tuiPath, tuiArgs...)
	tuiCmd.Stdin = os.Stdin
	tuiCmd.Stdout = os.Stdout
	tuiCmd.Stderr = os.Stderr
	tuiCmd.Env = append(os.Environ(), fmt.Sprintf("OPENCODE_SERVER=%s", serverURL))

	return tuiCmd.Run()
}
