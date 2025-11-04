# LSP Integration

## Overview

The OpenCode Go backend includes Language Server Protocol (LSP) integration to provide code intelligence features such as hover information, code completion, go-to-definition, and find references.

## Architecture

### Components

1. **LSP Client** (`internal/lsp/client.go`)
   - Manages multiple language server instances
   - Handles LSP lifecycle (initialization, shutdown)
   - Provides methods for LSP operations

2. **Language Detection** (`internal/lsp/detect.go`)
   - Maps file extensions to language identifiers
   - Provides server configuration for each language

3. **Tool Integration** (`internal/tool/lsp.go`)
   - Exposes LSP functionality as tool methods
   - Handles request/response formatting

4. **HTTP API** (`internal/server/handlers.go`)
   - REST endpoints for LSP operations

## Supported Languages

| Language   | Server                     | File Extensions |
| ---------- | -------------------------- | --------------- |
| Go         | gopls                      | .go             |
| TypeScript | typescript-language-server | .ts, .tsx       |
| JavaScript | typescript-language-server | .js, .jsx       |
| Python     | pylsp                      | .py             |
| Rust       | rust-analyzer              | .rs             |

## API Endpoints

### POST /lsp/hover

Get hover information for a position in a file.

**Request:**

```json
{
  "filePath": "/path/to/file.go",
  "line": 10,
  "character": 5
}
```

**Response:**

```json
{
  "contents": "func Println(a ...interface{}) (n int, err error)",
  "range": {
    "start": { "line": 10, "character": 4 },
    "end": { "line": 10, "character": 11 }
  }
}
```

### POST /lsp/completion

Get code completions for a position in a file.

**Request:**

```json
{
  "filePath": "/path/to/file.go",
  "line": 10,
  "character": 5
}
```

**Response:**

```json
{
  "items": [
    {
      "label": "Println",
      "kind": "3",
      "detail": "func(a ...interface{}) (n int, err error)",
      "documentation": "Println formats using the default formats..."
    }
  ]
}
```

### POST /lsp/definition

Find the definition of a symbol.

**Request:**

```json
{
  "filePath": "/path/to/file.go",
  "line": 10,
  "character": 5
}
```

**Response:**

```json
{
  "locations": [
    {
      "uri": "file:///path/to/definition.go",
      "range": {
        "start": { "line": 20, "character": 5 },
        "end": { "line": 20, "character": 15 }
      }
    }
  ]
}
```

### POST /lsp/references

Find all references to a symbol.

**Request:**

```json
{
  "filePath": "/path/to/file.go",
  "line": 10,
  "character": 5
}
```

**Response:**

```json
{
  "locations": [
    {
      "uri": "file:///path/to/file1.go",
      "range": {
        "start": { "line": 15, "character": 2 },
        "end": { "line": 15, "character": 10 }
      }
    },
    {
      "uri": "file:///path/to/file2.go",
      "range": {
        "start": { "line": 8, "character": 4 },
        "end": { "line": 8, "character": 12 }
      }
    }
  ]
}
```

## Usage

### Programmatic Usage

```go
import (
    "context"
    "github.com/opencode/opencode-go/internal/lsp"
)

// Create LSP client
client := lsp.NewClient()

// Start a language server
ctx := context.Background()
err := client.Start(ctx, "go", "gopls", []string{})

// Get hover information
hover, err := client.Hover(ctx, "go", "/path/to/file.go", 10, 5)

// Get completions
completions, err := client.Completion(ctx, "go", "/path/to/file.go", 10, 5)

// Find definition
definitions, err := client.Definition(ctx, "go", "/path/to/file.go", 10, 5)

// Find references
references, err := client.References(ctx, "go", "/path/to/file.go", 10, 5)

// Stop a server
client.Stop("go")

// Stop all servers
client.StopAll()
```

### Via Tool Executor

```go
import (
    "github.com/opencode/opencode-go/internal/tool"
)

executor := tool.NewToolExecutor("/working/directory")

// Get hover information
resp, err := executor.LSPHover(tool.LSPHoverRequest{
    FilePath:  "/path/to/file.go",
    Line:      10,
    Character: 5,
})

// Clean up when done
executor.StopLSPServers()
```

## Server Lifecycle

Language servers are started automatically on first request for a given language:

1. File extension is detected from the file path
2. Language is mapped using `DetectLanguage()`
3. Server configuration is retrieved using `GetServerConfig()`
4. If server not running, `Start()` is called
5. LSP operation is performed
6. Server remains running for subsequent requests

Servers are stopped:

- Explicitly via `Stop(language)` or `StopAll()`
- Automatically when the tool executor is cleaned up

## Configuration

### Adding a New Language

To add support for a new language:

1. Update `internal/lsp/detect.go`:

```go
func DetectLanguage(filePath string) string {
    ext := strings.TrimPrefix(filepath.Ext(filePath), ".")

    switch ext {
    // ... existing cases ...
    case "rb":
        return "ruby"
    default:
        return ""
    }
}
```

2. Add server configuration:

```go
func GetServerConfig(language string) (ServerConfig, bool) {
    configs := map[string]ServerConfig{
        // ... existing configs ...
        "ruby": {
            Command: "solargraph",
            Args:    []string{"stdio"},
        },
    }

    cfg, ok := configs[language]
    return cfg, ok
}
```

### Dynamic Configuration

Server configurations can be modified at runtime:

```go
// Add a new server configuration
lsp.AddServerConfig("custom", lsp.ServerConfig{
    Command: "custom-lsp-server",
    Args:    []string{"--stdio"},
})

// Remove a configuration
lsp.RemoveServerConfig("custom")

// Get all supported languages
languages := lsp.GetSupportedLanguages()
```

## Requirements

Language servers must be installed separately:

- **Go**: `go install golang.org/x/tools/gopls@latest`
- **TypeScript/JavaScript**: `npm install -g typescript-language-server typescript`
- **Python**: `pip install python-lsp-server`
- **Rust**: Install via rustup: `rustup component add rust-analyzer`

## Testing

### Unit Tests

```bash
# Run LSP client tests
go test ./internal/lsp/...

# Run LSP tool tests
go test ./internal/tool/... -run LSP
```

### Integration Tests

Integration tests require actual language servers to be installed:

```bash
# Run all tests including integration tests
go test ./internal/lsp/... -tags=integration
```

## Error Handling

Common errors:

1. **Server not found**: The language server executable is not in PATH
   - Solution: Install the required language server

2. **Unsupported file type**: No language mapping for file extension
   - Solution: Add language configuration

3. **Server initialization failed**: LSP server failed to start
   - Solution: Check server logs (stderr is captured)

4. **Server not running**: Operation attempted on stopped server
   - Solution: Server will auto-start on first request

## Performance Considerations

- Language servers are kept running to avoid initialization overhead
- Each language has a single server instance shared across files
- Document lifecycle is managed automatically (open/close)
- Servers can be stopped explicitly to free resources

## Future Enhancements

- [ ] Diagnostics support (publish diagnostics notification handler)
- [ ] Code actions (quick fixes, refactorings)
- [ ] Formatting support
- [ ] Rename symbol
- [ ] Document symbols
- [ ] Workspace symbols
- [ ] Server restart on crash
- [ ] Server health monitoring
- [ ] Configuration from project settings
