package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/opencode/opencode-go/internal/config"
	"github.com/opencode/opencode-go/internal/provider"
	"github.com/opencode/opencode-go/internal/session"
	"github.com/opencode/opencode-go/internal/storage"
	"github.com/spf13/cobra"
)

var (
	exportFormat string
	exportOutput string
)

var exportCmd = &cobra.Command{
	Use:   "export <session-id>",
	Short: "Export session data",
	Long:  `Export a session's conversation history and metadata to JSON or Markdown.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sessionID := args[0]
		ctx := context.Background()

		// Get storage path from XDG data directory
		storageDir := filepath.Join(config.GetDataDir(), "storage")
		store, err := storage.New(storageDir)
		if err != nil {
			fmt.Printf("Error opening storage: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()

		// Load session
		registry := provider.NewRegistry()
		sessionMgr := session.NewManager(store, registry)
		sess, err := sessionMgr.Get(ctx, sessionID)
		if err != nil {
			fmt.Printf("Error loading session: %v\n", err)
			os.Exit(1)
		}

		// Load messages
		messages, err := sessionMgr.GetMessages(ctx, sessionID)
		if err != nil {
			fmt.Printf("Error loading messages: %v\n", err)
			os.Exit(1)
		}

		// Export based on format
		var output []byte
		switch exportFormat {
		case "json":
			output, err = exportJSON(sess, messages)
		case "markdown", "md":
			output, err = exportMarkdown(sess, messages)
		default:
			fmt.Printf("Unknown export format: %s (supported: json, markdown)\n", exportFormat)
			os.Exit(1)
		}

		if err != nil {
			fmt.Printf("Error exporting session: %v\n", err)
			os.Exit(1)
		}

		// Write output
		if exportOutput == "" || exportOutput == "-" {
			fmt.Println(string(output))
		} else {
			err := os.WriteFile(exportOutput, output, 0644)
			if err != nil {
				fmt.Printf("Error writing to file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✅ Session exported to %s\n", exportOutput)
		}
	},
}

func init() {
	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "json", "Export format (json, markdown)")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file (default: stdout)")
}

func exportJSON(sess *session.Session, messages []*session.Message) ([]byte, error) {
	data := map[string]any{
		"session":  sess,
		"messages": messages,
	}
	return json.MarshalIndent(data, "", "  ")
}

func exportMarkdown(sess *session.Session, messages []*session.Message) ([]byte, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Session: %s\n\n", sess.ID))
	sb.WriteString(fmt.Sprintf("**Created:** %s\n", time.UnixMilli(sess.Time.Created).Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("**Updated:** %s\n", time.UnixMilli(sess.Time.Updated).Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("**Version:** %s\n\n", sess.Version))

	if sess.Title != "" {
		sb.WriteString(fmt.Sprintf("**Title:** %s\n\n", sess.Title))
	}

	sb.WriteString("---\n\n")
	sb.WriteString("## Conversation\n\n")

	for _, msg := range messages {
		role := string(msg.Role)
		if role == "user" {
			sb.WriteString("### 👤 User\n\n")
		} else if role == "assistant" {
			sb.WriteString("### 🤖 Assistant\n\n")
		} else {
			sb.WriteString(fmt.Sprintf("### %s\n\n", strings.Title(role)))
		}

		// Handle content blocks
		for _, block := range msg.Content {
			if block.Type == "text" {
				if text, ok := block.Data.(string); ok {
					sb.WriteString(text)
					sb.WriteString("\n\n")
				}
			} else if block.Type == "tool_use" {
				if toolData, ok := block.Data.(map[string]any); ok {
					if name, ok := toolData["name"].(string); ok {
						sb.WriteString(fmt.Sprintf("**Tool:** %s\n\n", name))
					}
				}
			}
		}

		sb.WriteString("---\n\n")
	}

	return []byte(sb.String()), nil
}
