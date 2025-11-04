package tool

import (
	"context"
	"fmt"

	"github.com/opencode/opencode-go/internal/lsp"
	"go.lsp.dev/protocol"
)

type LSPHoverRequest struct {
	FilePath  string `json:"filePath"`
	Line      int    `json:"line"`
	Character int    `json:"character"`
}

type LSPHoverResponse struct {
	Contents string `json:"contents"`
	Range    *Range `json:"range,omitempty"`
}

type LSPCompletionRequest struct {
	FilePath  string `json:"filePath"`
	Line      int    `json:"line"`
	Character int    `json:"character"`
}

type LSPCompletionResponse struct {
	Items []CompletionItem `json:"items"`
}

type CompletionItem struct {
	Label         string `json:"label"`
	Kind          string `json:"kind"`
	Detail        string `json:"detail,omitempty"`
	Documentation string `json:"documentation,omitempty"`
}

type LSPDefinitionRequest struct {
	FilePath  string `json:"filePath"`
	Line      int    `json:"line"`
	Character int    `json:"character"`
}

type LSPDefinitionResponse struct {
	Locations []Location `json:"locations"`
}

type LSPReferencesRequest struct {
	FilePath  string `json:"filePath"`
	Line      int    `json:"line"`
	Character int    `json:"character"`
}

type LSPReferencesResponse struct {
	Locations []Location `json:"locations"`
}

type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// LSPHover gets hover information for a position in a file
func (t *ToolExecutor) LSPHover(req LSPHoverRequest) (*LSPHoverResponse, error) {
	if req.FilePath == "" {
		return nil, fmt.Errorf("filePath is required")
	}

	// Detect language
	language := lsp.DetectLanguage(req.FilePath)
	if language == "" {
		return nil, fmt.Errorf("unsupported file type: %s", req.FilePath)
	}

	// Get server config
	config, ok := lsp.GetServerConfig(language)
	if !ok {
		return nil, fmt.Errorf("no LSP server configured for language: %s", language)
	}

	// Start LSP server if not running
	ctx := context.Background()
	if err := t.lspClient.Start(ctx, language, config.Command, config.Args); err != nil {
		return nil, fmt.Errorf("failed to start LSP server: %w", err)
	}

	// Get hover info
	hover, err := t.lspClient.Hover(ctx, language, req.FilePath, req.Line, req.Character)
	if err != nil {
		return nil, err
	}

	// Extract contents
	contents := extractHoverContents(hover)

	var rng *Range
	if hover.Range != nil {
		rng = &Range{
			Start: Position{
				Line:      int(hover.Range.Start.Line),
				Character: int(hover.Range.Start.Character),
			},
			End: Position{
				Line:      int(hover.Range.End.Line),
				Character: int(hover.Range.End.Character),
			},
		}
	}

	return &LSPHoverResponse{
		Contents: contents,
		Range:    rng,
	}, nil
}

// LSPCompletion gets code completions for a position in a file
func (t *ToolExecutor) LSPCompletion(req LSPCompletionRequest) (*LSPCompletionResponse, error) {
	if req.FilePath == "" {
		return nil, fmt.Errorf("filePath is required")
	}

	// Detect language
	language := lsp.DetectLanguage(req.FilePath)
	if language == "" {
		return nil, fmt.Errorf("unsupported file type: %s", req.FilePath)
	}

	// Get server config
	config, ok := lsp.GetServerConfig(language)
	if !ok {
		return nil, fmt.Errorf("no LSP server configured for language: %s", language)
	}

	// Start LSP server if not running
	ctx := context.Background()
	if err := t.lspClient.Start(ctx, language, config.Command, config.Args); err != nil {
		return nil, fmt.Errorf("failed to start LSP server: %w", err)
	}

	// Get completions
	completions, err := t.lspClient.Completion(ctx, language, req.FilePath, req.Line, req.Character)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	items := make([]CompletionItem, 0, len(completions.Items))
	for _, item := range completions.Items {
		items = append(items, CompletionItem{
			Label:         item.Label,
			Kind:          fmt.Sprintf("%d", int(item.Kind)),
			Detail:        item.Detail,
			Documentation: extractCompletionDocumentation(item.Documentation),
		})
	}

	return &LSPCompletionResponse{
		Items: items,
	}, nil
}

// LSPDefinition finds the definition of a symbol
func (t *ToolExecutor) LSPDefinition(req LSPDefinitionRequest) (*LSPDefinitionResponse, error) {
	if req.FilePath == "" {
		return nil, fmt.Errorf("filePath is required")
	}

	// Detect language
	language := lsp.DetectLanguage(req.FilePath)
	if language == "" {
		return nil, fmt.Errorf("unsupported file type: %s", req.FilePath)
	}

	// Get server config
	config, ok := lsp.GetServerConfig(language)
	if !ok {
		return nil, fmt.Errorf("no LSP server configured for language: %s", language)
	}

	// Start LSP server if not running
	ctx := context.Background()
	if err := t.lspClient.Start(ctx, language, config.Command, config.Args); err != nil {
		return nil, fmt.Errorf("failed to start LSP server: %w", err)
	}

	// Get definitions
	definitions, err := t.lspClient.Definition(ctx, language, req.FilePath, req.Line, req.Character)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	locations := make([]Location, 0, len(definitions))
	for _, loc := range definitions {
		locations = append(locations, Location{
			URI: string(loc.URI),
			Range: Range{
				Start: Position{
					Line:      int(loc.Range.Start.Line),
					Character: int(loc.Range.Start.Character),
				},
				End: Position{
					Line:      int(loc.Range.End.Line),
					Character: int(loc.Range.End.Character),
				},
			},
		})
	}

	return &LSPDefinitionResponse{
		Locations: locations,
	}, nil
}

// LSPReferences finds all references to a symbol
func (t *ToolExecutor) LSPReferences(req LSPReferencesRequest) (*LSPReferencesResponse, error) {
	if req.FilePath == "" {
		return nil, fmt.Errorf("filePath is required")
	}

	// Detect language
	language := lsp.DetectLanguage(req.FilePath)
	if language == "" {
		return nil, fmt.Errorf("unsupported file type: %s", req.FilePath)
	}

	// Get server config
	config, ok := lsp.GetServerConfig(language)
	if !ok {
		return nil, fmt.Errorf("no LSP server configured for language: %s", language)
	}

	// Start LSP server if not running
	ctx := context.Background()
	if err := t.lspClient.Start(ctx, language, config.Command, config.Args); err != nil {
		return nil, fmt.Errorf("failed to start LSP server: %w", err)
	}

	// Get references
	references, err := t.lspClient.References(ctx, language, req.FilePath, req.Line, req.Character)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	locations := make([]Location, 0, len(references))
	for _, loc := range references {
		locations = append(locations, Location{
			URI: string(loc.URI),
			Range: Range{
				Start: Position{
					Line:      int(loc.Range.Start.Line),
					Character: int(loc.Range.Start.Character),
				},
				End: Position{
					Line:      int(loc.Range.End.Line),
					Character: int(loc.Range.End.Character),
				},
			},
		})
	}

	return &LSPReferencesResponse{
		Locations: locations,
	}, nil
}

// Helper functions

func extractHoverContents(hover *protocol.Hover) string {
	if hover == nil {
		return ""
	}

	// Handle MarkupContent
	if hover.Contents.Kind != "" {
		return hover.Contents.Value
	}

	// Handle string content
	if hover.Contents.Value != "" {
		return hover.Contents.Value
	}

	return ""
}

func extractCompletionDocumentation(doc any) string {
	if doc == nil {
		return ""
	}

	switch v := doc.(type) {
	case string:
		return v
	case protocol.MarkupContent:
		return v.Value
	default:
		return ""
	}
}
