# OpenCode Go Backend

This is the Pure Go implementation of the OpenCode backend server, designed to replace the TypeScript backend while maintaining full API compatibility.

## Overview

This backend provides:

- HTTP API server with Chi router
- 48 stub API endpoints for OpenCode functionality
- BoltDB-based storage system
- Configuration loading from YAML/environment
- CORS support for web clients
- Graceful shutdown handling

## Directory Structure

```
packages/opencode-go/
├── cmd/
│   └── server/
│       └── main.go           # Entry point
├── internal/
│   ├── server/                # HTTP handlers
│   │   └── handlers.go
│   ├── config/                # Configuration
│   │   └── config.go
│   ├── storage/               # Data persistence
│   │   └── storage.go
│   ├── types/                 # Shared types
│   │   └── api.go
│   └── middleware/            # HTTP middleware
├── pkg/                       # Public APIs
├── go.mod
├── go.sum
└── README.md
```

## Requirements

- Go 1.21 or later
- No external system dependencies

## Building

```bash
# Build the server
go build -o opencode-server ./cmd/server

# Or build and run in one step
go run ./cmd/server/main.go
```

## Running

### Basic Usage

```bash
# Run with default configuration
go run ./cmd/server/main.go

# Or use the built binary
./opencode-server
```

The server will start on `localhost:8080` by default.

### Configuration

The server can be configured via:

1. **Configuration file** (checked in order):
   - `./opencode.yml`
   - `./opencode.yaml`
   - `~/.config/opencode/config.yml`
   - `~/.opencode.yml`

Example `opencode.yml`:

```yaml
server:
  host: localhost
  port: 8080

storage:
  path: ~/.opencode/storage/opencode.db

providers:
  anthropic:
    apiKey: your-api-key
    baseUrl: https://api.anthropic.com
    model: claude-3-5-sonnet-20241022
  openai:
    apiKey: your-api-key
    baseUrl: https://api.openai.com/v1
    model: gpt-4
```

2. **Environment variables**:
   - `OPENCODE_HOST` - Server host (default: `localhost`)
   - `OPENCODE_PORT` - Server port (default: `8080`)
   - `OPENCODE_STORAGE_PATH` - Database path
   - `OPENCODE_<PROVIDER>_API_KEY` - Provider API keys (e.g., `OPENCODE_ANTHROPIC_API_KEY`)

### Testing

Test the server with curl:

```bash
# Health check
curl http://localhost:8080/health

# Server status
curl http://localhost:8080/api/status

# List sessions (returns empty array in Phase 1)
curl http://localhost:8080/api/sessions

# Get agents
curl http://localhost:8080/api/agents

# Get tools
curl http://localhost:8080/api/tools
```

## API Endpoints

Phase 1 includes 48 stub endpoints:

### Health & Status

- `GET /health` - Health check
- `GET /api/status` - Server status

### Project

- `GET /api/project/current` - Get current project
- `POST /api/project/init` - Initialize project

### Sessions

- `GET /api/sessions` - List sessions
- `POST /api/sessions` - Create session
- `GET /api/sessions/{id}` - Get session
- `DELETE /api/sessions/{id}` - Delete session
- `POST /api/sessions/{id}/fork` - Fork session

### Messages

- `GET /api/sessions/{id}/messages` - Get messages
- `POST /api/sessions/{id}/messages` - Create message
- `GET /api/sessions/{id}/stream` - Stream messages (SSE)

### Agents

- `GET /api/agents` - List agents
- `GET /api/agents/{id}` - Get agent

### Tools

- `GET /api/tools` - List tools
- `POST /api/tools/{name}/execute` - Execute tool

### Configuration

- `GET /api/config` - Get configuration
- `POST /api/config` - Update configuration

### File Operations

- `POST /api/files/read` - Read file
- `POST /api/files/write` - Write file
- `POST /api/files/edit` - Edit file
- `POST /api/files/glob` - Glob pattern matching
- `POST /api/files/grep` - Grep search

### Bash

- `POST /api/bash/execute` - Execute bash command

### LSP

- `POST /api/lsp/diagnostics` - Get diagnostics
- `POST /api/lsp/hover` - Get hover information

### Auth

- `POST /api/auth/login` - Login
- `POST /api/auth/logout` - Logout

### Todos

- `GET /api/sessions/{id}/todos` - Get todos
- `POST /api/sessions/{id}/todos` - Update todos

## Development Status

### Phase 1 (Current) ✅

- [x] Project structure
- [x] HTTP server with Chi router
- [x] 48 stub API endpoints
- [x] Configuration loading
- [x] BoltDB storage infrastructure
- [x] CORS middleware
- [x] Graceful shutdown

### Phase 2 (Next)

- [ ] File operations implementation
- [ ] Bash execution
- [ ] Project management

### Phase 3

- [ ] Session management
- [ ] Message storage

### Phase 4

- [ ] AI provider integration
- [ ] Streaming responses

### Phase 5

- [ ] LSP integration
- [ ] Advanced features

## Dependencies

- [Chi](https://github.com/go-chi/chi) - HTTP router and middleware
- [BoltDB](https://github.com/etcd-io/bbolt) - Embedded key/value database
- [YAML](https://github.com/go-yaml/yaml) - YAML parser
- [Validator](https://github.com/go-playground/validator) - Struct validation

## Storage

The server uses BoltDB for persistent storage:

- Default location: `~/.opencode/storage/opencode.db`
- Buckets: `sessions`, `messages`, `projects`, `config`
- Automatic bucket creation on startup

## Graceful Shutdown

The server handles `SIGINT` and `SIGTERM` signals for graceful shutdown:

- Stops accepting new requests
- Waits up to 5 seconds for active requests to complete
- Closes database connections
- Exits cleanly

## Notes

- All endpoints in Phase 1 return stub/empty responses
- Full functionality will be implemented in subsequent phases
- API responses are JSON formatted
- CORS is enabled for all origins (configurable for production)
- Storage persists across restarts

## Troubleshooting

### Port already in use

```bash
# Use a different port
OPENCODE_PORT=8081 go run ./cmd/server/main.go
```

### Storage permission errors

```bash
# Ensure storage directory is writable
mkdir -p ~/.opencode/storage
chmod 755 ~/.opencode/storage
```

### Module errors

```bash
# Update dependencies
go mod tidy
go mod download
```

## License

Same as OpenCode project.
