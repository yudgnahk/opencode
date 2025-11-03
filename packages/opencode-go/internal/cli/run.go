package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/opencode/opencode-go/internal/auth"
	"github.com/opencode/opencode-go/internal/config"
	"github.com/opencode/opencode-go/internal/provider"
	"github.com/opencode/opencode-go/internal/session"
	"github.com/opencode/opencode-go/internal/storage"
	"github.com/spf13/cobra"
)

var (
	runModel    string
	runSession  string
	runContinue bool
	runShare    bool
	runAgent    string
	runFormat   string
	runFiles    []string
	runTitle    string
	runCommand  string
)

var runCmd = &cobra.Command{
	Use:   "run [message...]",
	Short: "Run OpenCode with a message",
	Long: `Run OpenCode with a message. This is the main interactive command that 
sends a prompt to an AI model and displays the response.

Examples:
  opencode run "explain this code" -f main.go
  opencode run "fix the bug" --continue
  opencode run --session abc123 "add feature"
  cat input.txt | opencode run "summarize this"`,
	RunE: runRun,
}

func init() {
	runCmd.Flags().StringVarP(&runModel, "model", "m", "", "model to use (provider/model)")
	runCmd.Flags().StringVarP(&runSession, "session", "s", "", "session ID to continue")
	runCmd.Flags().BoolVarP(&runContinue, "continue", "c", false, "continue the last session")
	runCmd.Flags().BoolVar(&runShare, "share", false, "share the session")
	runCmd.Flags().StringVar(&runAgent, "agent", "", "agent to use")
	runCmd.Flags().StringVar(&runFormat, "format", "default", "output format (default, json)")
	runCmd.Flags().StringSliceVarP(&runFiles, "file", "f", []string{}, "file(s) to attach")
	runCmd.Flags().StringVar(&runTitle, "title", "", "title for the session")
	runCmd.Flags().StringVar(&runCommand, "command", "", "command to run, use message for args")
}

func runRun(cmd *cobra.Command, args []string) error {
	message := strings.Join(args, " ")

	// Read from stdin if piped
	stat, err := os.Stdin.Stat()
	if err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
		stdinBytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		stdinContent := strings.TrimSpace(string(stdinBytes))
		if stdinContent != "" {
			if message != "" {
				message = message + "\n" + stdinContent
			} else {
				message = stdinContent
			}
		}
	}

	if message == "" && runCommand == "" {
		return fmt.Errorf("you must provide a message or a command")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize storage
	store, err := storage.New(cfg.Storage.Path)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer store.Close()

	// Create session manager
	sessionMgr := session.NewManager(store)

	// Create provider registry
	registry := provider.NewRegistry()

	// Initialize providers
	ctx := context.Background()
	if err := initializeProviders(ctx, registry); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to initialize some providers: %v\n", err)
	}

	// Get or create session
	var sess *session.Session
	if runContinue {
		// Find last session
		sessions, err := sessionMgr.List(ctx, "")
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}
		if len(sessions) > 0 {
			sess = sessions[0] // First is most recent
		}
	} else if runSession != "" {
		sess, err = sessionMgr.Get(ctx, runSession)
		if err != nil {
			return fmt.Errorf("session not found: %w", err)
		}
	}

	// Create new session if needed
	if sess == nil {
		// Determine model
		providerID, modelID := "anthropic", "claude-3-5-sonnet-20241022"
		if runModel != "" {
			parts := strings.Split(runModel, "/")
			if len(parts) != 2 {
				return fmt.Errorf("invalid model format, use provider/model")
			}
			providerID, modelID = parts[0], parts[1]
		}

		sess, err = sessionMgr.Create(ctx, "", providerID, modelID)
		if err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}

		// Update title if provided
		title := runTitle
		if title == "" && message != "" {
			title = message
			if len(title) > 50 {
				title = title[:50] + "..."
			}
		}
		if title != "" {
			sessionMgr.Update(ctx, sess.ID, map[string]interface{}{
				"title": title,
			})
		}
	}

	fmt.Printf("Session: %s\n", sess.ID)
	if runShare {
		fmt.Printf("~  https://opencode.ai/s/%s\n", sess.ID[len(sess.ID)-8:])
	}

	// Create completion service
	completionSvc := session.NewCompletionService(sessionMgr, registry)

	// Send message
	if err := completionSvc.SendMessage(ctx, sess.ID, message); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	// Get and display response
	messages, err := sessionMgr.GetMessages(ctx, sess.ID)
	if err != nil {
		return fmt.Errorf("failed to get messages: %w", err)
	}

	// Display last assistant message
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == session.RoleAssistant {
			fmt.Println()
			for _, content := range messages[i].Content {
				if content.Type == "text" {
					fmt.Println(content.Text)
				}
			}
			fmt.Println()
			break
		}
	}

	return nil
}

func initializeProviders(ctx context.Context, registry *provider.Registry) error {
	retryConfig := provider.DefaultRetryConfig()

	// Helper to get API key from auth system or env var
	getAPIKey := func(providerID, envVar string) string {
		// Check auth system first
		authInfo, err := auth.Get(providerID)
		if err == nil && authInfo != nil {
			if apiAuth, ok := authInfo.(auth.API); ok && apiAuth.Key != "" {
				return apiAuth.Key
			}
		}
		// Fall back to environment variable
		return os.Getenv(envVar)
	}

	// Initialize Anthropic provider with retry
	if apiKey := getAPIKey("anthropic", "ANTHROPIC_API_KEY"); apiKey != "" {
		anthropic := provider.NewAnthropic(apiKey)
		wrapped := provider.NewRetryableProvider(anthropic, retryConfig)
		if err := registry.Register(wrapped); err != nil {
			return fmt.Errorf("failed to register anthropic: %w", err)
		}
	}

	// Initialize OpenAI provider with retry
	if apiKey := getAPIKey("openai", "OPENAI_API_KEY"); apiKey != "" {
		openai := provider.NewOpenAI(apiKey)
		wrapped := provider.NewRetryableProvider(openai, retryConfig)
		if err := registry.Register(wrapped); err != nil {
			return fmt.Errorf("failed to register openai: %w", err)
		}
	}

	// Initialize Gemini provider with retry
	if apiKey := getAPIKey("gemini", "GEMINI_API_KEY"); apiKey != "" {
		gemini, err := provider.NewGemini(ctx, apiKey)
		if err != nil {
			return fmt.Errorf("failed to create gemini provider: %w", err)
		}
		wrapped := provider.NewRetryableProvider(gemini, retryConfig)
		if err := registry.Register(wrapped); err != nil {
			return fmt.Errorf("failed to register gemini: %w", err)
		}
	}

	return nil
}
