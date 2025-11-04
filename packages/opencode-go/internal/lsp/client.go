package lsp

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
)

// Client manages LSP servers for multiple languages
type Client struct {
	servers map[string]*LanguageServer
	mu      sync.RWMutex
}

// LanguageServer represents a running LSP server instance
type LanguageServer struct {
	cmd    *exec.Cmd
	conn   jsonrpc2.Conn
	cancel context.CancelFunc
}

// NewClient creates a new LSP client
func NewClient() *Client {
	return &Client{
		servers: make(map[string]*LanguageServer),
	}
}

// readWriteCloser wraps separate read and write closers
type readWriteCloser struct {
	io.ReadCloser
	io.WriteCloser
}

func (rwc *readWriteCloser) Close() error {
	err1 := rwc.ReadCloser.Close()
	err2 := rwc.WriteCloser.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

// Start starts an LSP server for a language
func (c *Client) Start(ctx context.Context, language, command string, args []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.servers[language]; exists {
		return nil // Already running
	}

	// Create context for server lifecycle
	serverCtx, cancel := context.WithCancel(ctx)

	// Start language server process
	cmd := exec.CommandContext(serverCtx, command, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Capture stderr for debugging
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("failed to start LSP server: %w", err)
	}

	// Create JSON-RPC connection (needs ReadWriteCloser)
	rwc := &readWriteCloser{ReadCloser: stdout, WriteCloser: stdin}
	stream := jsonrpc2.NewStream(rwc)
	conn := jsonrpc2.NewConn(stream)

	// Initialize server
	initParams := &protocol.InitializeParams{
		Capabilities: protocol.ClientCapabilities{
			TextDocument: &protocol.TextDocumentClientCapabilities{
				Hover: &protocol.HoverTextDocumentClientCapabilities{
					ContentFormat: []protocol.MarkupKind{
						protocol.Markdown,
						protocol.PlainText,
					},
				},
				Completion: &protocol.CompletionTextDocumentClientCapabilities{
					CompletionItem: &protocol.CompletionTextDocumentClientCapabilitiesItem{
						SnippetSupport: true,
					},
				},
				PublishDiagnostics: &protocol.PublishDiagnosticsClientCapabilities{
					RelatedInformation: true,
				},
			},
		},
	}

	var result protocol.InitializeResult
	_, err = conn.Call(serverCtx, "initialize", initParams, &result)
	if err != nil {
		cancel()
		cmd.Process.Kill()
		return fmt.Errorf("LSP initialize failed: %w", err)
	}

	// Send initialized notification
	err = conn.Notify(serverCtx, "initialized", &protocol.InitializedParams{})
	if err != nil {
		cancel()
		cmd.Process.Kill()
		return fmt.Errorf("LSP initialized notification failed: %w", err)
	}

	c.servers[language] = &LanguageServer{
		cmd:    cmd,
		conn:   conn,
		cancel: cancel,
	}

	return nil
}

// Stop stops an LSP server
func (c *Client) Stop(language string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	server, exists := c.servers[language]
	if !exists {
		return nil
	}

	// Send shutdown request
	_, err := server.conn.Call(context.Background(), "shutdown", nil, nil)
	if err != nil {
		// Log error but continue cleanup
		fmt.Fprintf(os.Stderr, "LSP shutdown error for %s: %v\n", language, err)
	}

	// Send exit notification
	server.conn.Notify(context.Background(), "exit", nil)

	// Close connection
	server.conn.Close()

	// Cancel context to kill process
	server.cancel()

	// Wait for process to exit
	server.cmd.Wait()

	delete(c.servers, language)

	return nil
}

// StopAll stops all running LSP servers
func (c *Client) StopAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for language := range c.servers {
		server := c.servers[language]

		// Only call LSP shutdown if connection exists
		if server.conn != nil {
			server.conn.Call(context.Background(), "shutdown", nil, nil)
			server.conn.Notify(context.Background(), "exit", nil)
			server.conn.Close()
		}

		if server.cancel != nil {
			server.cancel()
		}

		if server.cmd != nil {
			server.cmd.Wait()
		}
	}

	c.servers = make(map[string]*LanguageServer)
}

// Hover gets hover information for a position in a file
func (c *Client) Hover(ctx context.Context, language, filePath string, line, character int) (*protocol.Hover, error) {
	c.mu.RLock()
	server, exists := c.servers[language]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("LSP server not running for %s", language)
	}

	uri := protocol.DocumentURI("file://" + filePath)

	// Open document if not already open
	if err := c.openDocument(ctx, server, uri, language, filePath); err != nil {
		return nil, err
	}

	params := &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: uri},
			Position: protocol.Position{
				Line:      uint32(line),
				Character: uint32(character),
			},
		},
	}

	var hover protocol.Hover
	_, err := server.conn.Call(ctx, "textDocument/hover", params, &hover)
	if err != nil {
		return nil, fmt.Errorf("hover request failed: %w", err)
	}

	return &hover, nil
}

// Completion gets code completions for a position in a file
func (c *Client) Completion(ctx context.Context, language, filePath string, line, character int) (*protocol.CompletionList, error) {
	c.mu.RLock()
	server, exists := c.servers[language]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("LSP server not running for %s", language)
	}

	uri := protocol.DocumentURI("file://" + filePath)

	// Open document if not already open
	if err := c.openDocument(ctx, server, uri, language, filePath); err != nil {
		return nil, err
	}

	params := &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: uri},
			Position: protocol.Position{
				Line:      uint32(line),
				Character: uint32(character),
			},
		},
	}

	var completions protocol.CompletionList
	_, err := server.conn.Call(ctx, "textDocument/completion", params, &completions)
	if err != nil {
		return nil, fmt.Errorf("completion request failed: %w", err)
	}

	return &completions, nil
}

// Definition finds the definition of a symbol
func (c *Client) Definition(ctx context.Context, language, filePath string, line, character int) ([]protocol.Location, error) {
	c.mu.RLock()
	server, exists := c.servers[language]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("LSP server not running for %s", language)
	}

	uri := protocol.DocumentURI("file://" + filePath)

	// Open document if not already open
	if err := c.openDocument(ctx, server, uri, language, filePath); err != nil {
		return nil, err
	}

	params := &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: uri},
			Position: protocol.Position{
				Line:      uint32(line),
				Character: uint32(character),
			},
		},
	}

	var locations []protocol.Location
	_, err := server.conn.Call(ctx, "textDocument/definition", params, &locations)
	if err != nil {
		return nil, fmt.Errorf("definition request failed: %w", err)
	}

	return locations, nil
}

// References finds all references to a symbol
func (c *Client) References(ctx context.Context, language, filePath string, line, character int) ([]protocol.Location, error) {
	c.mu.RLock()
	server, exists := c.servers[language]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("LSP server not running for %s", language)
	}

	uri := protocol.DocumentURI("file://" + filePath)

	// Open document if not already open
	if err := c.openDocument(ctx, server, uri, language, filePath); err != nil {
		return nil, err
	}

	params := &protocol.ReferenceParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: uri},
			Position: protocol.Position{
				Line:      uint32(line),
				Character: uint32(character),
			},
		},
		Context: protocol.ReferenceContext{
			IncludeDeclaration: true,
		},
	}

	var locations []protocol.Location
	_, err := server.conn.Call(ctx, "textDocument/references", params, &locations)
	if err != nil {
		return nil, fmt.Errorf("references request failed: %w", err)
	}

	return locations, nil
}

// openDocument opens a document in the LSP server if not already open
func (c *Client) openDocument(ctx context.Context, server *LanguageServer, uri protocol.DocumentURI, language, filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	err = server.conn.Notify(ctx, "textDocument/didOpen", &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        uri,
			LanguageID: protocol.LanguageIdentifier(language),
			Version:    1,
			Text:       string(content),
		},
	})

	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to open document: %w", err)
	}

	return nil
}

// CloseDocument closes a document in the LSP server
func (c *Client) CloseDocument(ctx context.Context, language, filePath string) error {
	c.mu.RLock()
	server, exists := c.servers[language]
	c.mu.RUnlock()

	if !exists {
		return fmt.Errorf("LSP server not running for %s", language)
	}

	uri := protocol.DocumentURI("file://" + filePath)

	err := server.conn.Notify(ctx, "textDocument/didClose", &protocol.DidCloseTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri},
	})

	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to close document: %w", err)
	}

	return nil
}
