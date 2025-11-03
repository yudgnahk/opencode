# Phase 1: Foundation & Infrastructure

**Duration**: Weeks 1-4 (1 month)  
**Team**: 2 engineers  
**Complexity**: 🟢 Low-Medium  
**Dependencies**: None

---

## Objectives

1. Set up Go project structure
2. Create HTTP API framework with 48 stub endpoints
3. Implement configuration loading
4. Set up storage infrastructure
5. Basic routing and middleware
6. Health check and connectivity

**Success Criterion**: TUI can connect to Go backend and make API calls (receiving empty/mock responses)

---

## Detailed Tasks

### Week 1: Project Setup & Structure

#### Task 1.1: Initialize Go Module

- [ ] Create `packages/opencode-go` directory
- [ ] Initialize Go module: `go mod init github.com/opencode/opencode-go`
- [ ] Set up directory structure:
  ```
  packages/opencode-go/
  ├── cmd/
  │   └── server/
  │       └── main.go           # Entry point
  ├── internal/
  │   ├── server/                # HTTP handlers
  │   ├── config/                # Configuration
  │   ├── storage/               # Data persistence
  │   ├── types/                 # Shared types
  │   └── middleware/            # HTTP middleware
  ├── pkg/                       # Public APIs
  ├── go.mod
  ├── go.sum
  └── README.md
  ```

#### Task 1.2: Add Core Dependencies

- [ ] Install Chi router: `go get github.com/go-chi/chi/v5`
- [ ] Install Chi middleware: `go get github.com/go-chi/cors`
- [ ] Install validation: `go get github.com/go-playground/validator/v10`
- [ ] Install YAML parser: `go get gopkg.in/yaml.v3`
- [ ] Create `go.mod` with require statements

#### Task 1.3: Basic Server Setup

Create `cmd/server/main.go`:

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/opencode/opencode-go/internal/server"
)

func main() {
    r := chi.NewRouter()

    // Create server
    srv := server.New()
    srv.RegisterRoutes(r)

    // Start HTTP server
    httpServer := &http.Server{
        Addr:    ":8080",
        Handler: r,
    }

    // Graceful shutdown
    go func() {
        if err := httpServer.ListenAndServe(); err != nil {
            log.Fatal(err)
        }
    }()

    // Wait for interrupt
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    httpServer.Shutdown(ctx)
}
```

---

### Week 2: API Endpoint Structure

#### Task 2.1: Define API Types

Create `internal/types/api.go`:

```go
package types

// Session represents a chat session
type Session struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}

// Message represents a conversation message
type Message struct {
    ID        string                 `json:"id"`
    Role      string                 `json:"role"`
    Content   []ContentBlock         `json:"content"`
    CreatedAt time.Time              `json:"createdAt"`
}

// ContentBlock represents message content
type ContentBlock struct {
    Type string `json:"type"`
    Text string `json:"text,omitempty"`
}

// Project represents workspace project info
type Project struct {
    ID   string `json:"id"`
    Path string `json:"path"`
    Name string `json:"name"`
}

// Agent represents an AI agent configuration
type Agent struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
}

// Tool represents an available tool
type Tool struct {
    Name        string      `json:"name"`
    Description string      `json:"description"`
    Parameters  interface{} `json:"parameters"`
}

// Config represents OpenCode configuration
type Config struct {
    Providers map[string]ProviderConfig `json:"providers"`
    Agents    []Agent                   `json:"agents"`
}

type ProviderConfig struct {
    APIKey string `json:"apiKey"`
    Model  string `json:"model"`
}
```

#### Task 2.2: Create Stub Handlers

Create `internal/server/handlers.go` with stub implementations:

```go
package server

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
)

type Server struct {
    // Will add fields later
}

func New() *Server {
    return &Server{}
}

func (s *Server) RegisterRoutes(r chi.Router) {
    // Health & Status
    r.Get("/health", s.handleHealth)
    r.Get("/api/status", s.handleStatus)

    // Project
    r.Get("/api/project/current", s.handleProjectCurrent)
    r.Post("/api/project/init", s.handleProjectInit)

    // Sessions
    r.Get("/api/sessions", s.handleSessionsList)
    r.Post("/api/sessions", s.handleSessionsCreate)
    r.Get("/api/sessions/{id}", s.handleSessionsGet)
    r.Delete("/api/sessions/{id}", s.handleSessionsDelete)
    r.Post("/api/sessions/{id}/fork", s.handleSessionsFork)

    // Messages
    r.Get("/api/sessions/{id}/messages", s.handleMessagesGet)
    r.Post("/api/sessions/{id}/messages", s.handleMessagesCreate)
    r.Get("/api/sessions/{id}/stream", s.handleMessagesStream)

    // Agents
    r.Get("/api/agents", s.handleAgentsList)
    r.Get("/api/agents/{id}", s.handleAgentsGet)

    // Tools
    r.Get("/api/tools", s.handleToolsList)
    r.Post("/api/tools/{name}/execute", s.handleToolsExecute)

    // Configuration
    r.Get("/api/config", s.handleConfigGet)
    r.Post("/api/config", s.handleConfigUpdate)

    // File operations
    r.Post("/api/files/read", s.handleFilesRead)
    r.Post("/api/files/write", s.handleFilesWrite)
    r.Post("/api/files/edit", s.handleFilesEdit)
    r.Post("/api/files/glob", s.handleFilesGlob)
    r.Post("/api/files/grep", s.handleFilesGrep)

    // Bash
    r.Post("/api/bash/execute", s.handleBashExecute)

    // LSP
    r.Post("/api/lsp/diagnostics", s.handleLSPDiagnostics)
    r.Post("/api/lsp/hover", s.handleLSPHover)

    // Auth
    r.Post("/api/auth/login", s.handleAuthLogin)
    r.Post("/api/auth/logout", s.handleAuthLogout)

    // Todos
    r.Get("/api/sessions/{id}/todos", s.handleTodosGet)
    r.Post("/api/sessions/{id}/todos", s.handleTodosUpdate)
}

// Stub implementations (return empty responses)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(map[string]interface{}{
        "version": "0.1.0-go",
        "backend": "go",
    })
}

// Project stubs
func (s *Server) handleProjectCurrent(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement in Phase 2
    json.NewEncoder(w).Encode(map[string]string{"id": "stub"})
}

func (s *Server) handleProjectInit(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement in Phase 2
    w.WriteHeader(http.StatusCreated)
}

// Session stubs
func (s *Server) handleSessionsList(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement in Phase 3
    json.NewEncoder(w).Encode([]interface{}{})
}

func (s *Server) handleSessionsCreate(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement in Phase 3
    json.NewEncoder(w).Encode(map[string]string{"id": "stub-session"})
}

func (s *Server) handleSessionsGet(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement in Phase 3
    sessionID := chi.URLParam(r, "id")
    json.NewEncoder(w).Encode(map[string]string{"id": sessionID})
}

func (s *Server) handleSessionsDelete(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement in Phase 3
    w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleSessionsFork(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement in Phase 3
    json.NewEncoder(w).Encode(map[string]string{"id": "stub-fork"})
}

// Message stubs
func (s *Server) handleMessagesGet(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement in Phase 3
    json.NewEncoder(w).Encode([]interface{}{})
}

func (s *Server) handleMessagesCreate(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement in Phase 4
    json.NewEncoder(w).Encode(map[string]string{"id": "stub-message"})
}

func (s *Server) handleMessagesStream(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement in Phase 4 (SSE streaming)
    w.Header().Set("Content-Type", "text/event-stream")
    w.WriteHeader(http.StatusOK)
}

// Agent stubs
func (s *Server) handleAgentsList(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode([]interface{}{})
}

func (s *Server) handleAgentsGet(w http.ResponseWriter, r *http.Request) {
    agentID := chi.URLParam(r, "id")
    json.NewEncoder(w).Encode(map[string]string{"id": agentID})
}

// Tool stubs
func (s *Server) handleToolsList(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode([]interface{}{})
}

func (s *Server) handleToolsExecute(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement in Phase 2
    json.NewEncoder(w).Encode(map[string]string{"result": "stub"})
}

// Config stubs
func (s *Server) handleConfigGet(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(map[string]interface{}{})
}

func (s *Server) handleConfigUpdate(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
}

// File operation stubs
func (s *Server) handleFilesRead(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement in Phase 2
    json.NewEncoder(w).Encode(map[string]string{"content": ""})
}

func (s *Server) handleFilesWrite(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement in Phase 2
    w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleFilesEdit(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement in Phase 2
    w.WriteHeader(http.StatusOK)
}

func (s *Server) handleFilesGlob(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement in Phase 2
    json.NewEncoder(w).Encode([]string{})
}

func (s *Server) handleFilesGrep(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement in Phase 2
    json.NewEncoder(w).Encode([]interface{}{})
}

// Bash stub
func (s *Server) handleBashExecute(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement in Phase 2
    json.NewEncoder(w).Encode(map[string]string{"output": ""})
}

// LSP stubs
func (s *Server) handleLSPDiagnostics(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement in Phase 5
    json.NewEncoder(w).Encode([]interface{}{})
}

func (s *Server) handleLSPHover(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement in Phase 5
    json.NewEncoder(w).Encode(map[string]string{})
}

// Auth stubs
func (s *Server) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(map[string]string{"token": "stub-token"})
}

func (s *Server) handleAuthLogout(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusNoContent)
}

// Todo stubs
func (s *Server) handleTodosGet(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement in Phase 5
    json.NewEncoder(w).Encode([]interface{}{})
}

func (s *Server) handleTodosUpdate(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement in Phase 5
    w.WriteHeader(http.StatusOK)
}
```

---

### Week 3: Configuration System

#### Task 3.1: Configuration Types

Create `internal/config/config.go`:

```go
package config

import (
    "os"
    "path/filepath"

    "gopkg.in/yaml.v3"
)

type Config struct {
    Server    ServerConfig              `yaml:"server"`
    Providers map[string]ProviderConfig `yaml:"providers"`
    Storage   StorageConfig             `yaml:"storage"`
}

type ServerConfig struct {
    Host string `yaml:"host"`
    Port int    `yaml:"port"`
}

type ProviderConfig struct {
    APIKey  string `yaml:"apiKey"`
    BaseURL string `yaml:"baseUrl"`
    Model   string `yaml:"model"`
}

type StorageConfig struct {
    Path string `yaml:"path"`
}

func Load() (*Config, error) {
    // Try multiple config locations
    configPaths := []string{
        "./opencode.yml",
        "./opencode.yaml",
        filepath.Join(os.Getenv("HOME"), ".config", "opencode", "config.yml"),
        filepath.Join(os.Getenv("HOME"), ".opencode.yml"),
    }

    for _, path := range configPaths {
        if _, err := os.Stat(path); err == nil {
            return loadFromFile(path)
        }
    }

    // Return default config if none found
    return defaultConfig(), nil
}

func loadFromFile(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}

func defaultConfig() *Config {
    return &Config{
        Server: ServerConfig{
            Host: "localhost",
            Port: 8080,
        },
        Storage: StorageConfig{
            Path: filepath.Join(os.Getenv("HOME"), ".opencode", "storage"),
        },
    }
}
```

#### Task 3.2: Environment Variable Overrides

- [ ] Support `OPENCODE_PORT` environment variable
- [ ] Support `OPENCODE_HOST` environment variable
- [ ] Support `OPENCODE_STORAGE_PATH` environment variable
- [ ] Support provider API keys from environment

---

### Week 4: Storage Infrastructure

#### Task 4.1: Storage Interface

Create `internal/storage/storage.go`:

```go
package storage

import (
    "encoding/json"
    "os"
    "path/filepath"

    bolt "go.etcd.io/bbolt"
)

type Storage struct {
    db *bolt.DB
}

func New(path string) (*Storage, error) {
    // Ensure directory exists
    if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
        return nil, err
    }

    db, err := bolt.Open(path, 0600, nil)
    if err != nil {
        return nil, err
    }

    // Create buckets
    err = db.Update(func(tx *bolt.Tx) error {
        buckets := []string{"sessions", "messages", "projects", "config"}
        for _, bucket := range buckets {
            if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
                return err
            }
        }
        return nil
    })

    if err != nil {
        db.Close()
        return nil, err
    }

    return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
    return s.db.Close()
}

// Generic get/set methods
func (s *Storage) Get(bucket, key string) ([]byte, error) {
    var data []byte
    err := s.db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(bucket))
        data = b.Get([]byte(key))
        return nil
    })
    return data, err
}

func (s *Storage) Set(bucket, key string, value []byte) error {
    return s.db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(bucket))
        return b.Put([]byte(key), value)
    })
}

func (s *Storage) Delete(bucket, key string) error {
    return s.db.Update(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(bucket))
        return b.Delete([]byte(key))
    })
}

// Helper for JSON marshaling
func (s *Storage) GetJSON(bucket, key string, v interface{}) error {
    data, err := s.Get(bucket, key)
    if err != nil {
        return err
    }
    if data == nil {
        return nil
    }
    return json.Unmarshal(data, v)
}

func (s *Storage) SetJSON(bucket, key string, v interface{}) error {
    data, err := json.Marshal(v)
    if err != nil {
        return err
    }
    return s.Set(bucket, key, data)
}

// List all keys in bucket
func (s *Storage) List(bucket string) ([]string, error) {
    var keys []string
    err := s.db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(bucket))
        return b.ForEach(func(k, v []byte) error {
            keys = append(keys, string(k))
            return nil
        })
    })
    return keys, err
}
```

#### Task 4.2: Initialize Storage

- [ ] Install BoltDB: `go get go.etcd.io/bbolt`
- [ ] Create storage buckets (sessions, messages, projects, config)
- [ ] Add storage to Server struct
- [ ] Initialize storage in main.go

---

## Testing

### Unit Tests

- [ ] Test configuration loading from multiple locations
- [ ] Test configuration defaults
- [ ] Test storage CRUD operations
- [ ] Test JSON marshaling in storage

### Integration Tests

- [ ] Start server and verify all endpoints return 200 (or appropriate status)
- [ ] Test health endpoint
- [ ] Verify API endpoints are accessible

### Manual Testing

- [ ] Start Go server: `go run cmd/server/main.go`
- [ ] Test with curl:
  ```bash
  curl http://localhost:8080/health
  curl http://localhost:8080/api/status
  curl http://localhost:8080/api/sessions
  ```
- [ ] Modify TUI to connect to Go backend (feature flag)
- [ ] Verify TUI can connect (will get empty responses)

---

## Deliverables

- [x] `packages/opencode-go` directory structure
- [x] Working HTTP server with Chi router
- [x] 48 stub API endpoints
- [x] Configuration loading system
- [x] BoltDB storage infrastructure
- [x] Basic middleware (CORS, logging)
- [x] Health check endpoint
- [x] README with build/run instructions

---

## Success Metrics

- [ ] Server starts without errors
- [ ] All 48 endpoints return appropriate HTTP status codes
- [ ] Configuration loads from file and environment
- [ ] Storage can persist data across restarts
- [ ] TUI can connect to Go backend
- [ ] No panics or crashes during basic operations

---

## Blockers & Risks

### Risk: API Contract Mismatch

**Mitigation**: Reference TypeScript server implementation closely. Use OpenAPI spec if available.

### Risk: TUI Compatibility

**Mitigation**: Keep TypeScript server running. Add feature flag to TUI.

---

## Phase 1 Checklist

### Setup

- [ ] Create `packages/opencode-go` directory
- [ ] Initialize Go module
- [ ] Set up directory structure
- [ ] Add dependencies (Chi, BoltDB, YAML, validator)
- [ ] Create README with instructions

### API Server

- [ ] Implement main.go with Chi router
- [ ] Add graceful shutdown
- [ ] Implement health check endpoint
- [ ] Create server struct and interface
- [ ] Define 48 API routes
- [ ] Implement stub handlers (return empty data)

### Configuration

- [ ] Define Config struct
- [ ] Implement config loading from file
- [ ] Support multiple config locations
- [ ] Add environment variable overrides
- [ ] Create default config

### Storage

- [ ] Install BoltDB
- [ ] Create Storage struct
- [ ] Implement basic CRUD operations
- [ ] Add JSON helper methods
- [ ] Create storage buckets
- [ ] Test persistence across restarts

### Types

- [ ] Define Session type
- [ ] Define Message type
- [ ] Define Project type
- [ ] Define Agent type
- [ ] Define Tool type
- [ ] Define Config types

### Testing

- [ ] Unit tests for config loading
- [ ] Unit tests for storage operations
- [ ] Integration test for server startup
- [ ] Manual testing with curl
- [ ] TUI connection test (with feature flag)

### Documentation

- [ ] README with build instructions
- [ ] README with run instructions
- [ ] API endpoint list
- [ ] Configuration documentation

---

## Next Phase

**Phase 2**: File Operations & Tools (Weeks 5-10)

- Implement read/write/edit tools
- Add glob and grep functionality
- Bash command execution
- Project management
