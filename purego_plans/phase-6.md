# Phase 6: Production Readiness & Release

**Duration**: Weeks 36-40 (5 weeks)  
**Team**: 2-3 engineers  
**Complexity**: 🟡 Medium  
**Dependencies**: Phases 1-5 complete

---

## Objectives

1. Comprehensive testing (unit, integration, e2e)
2. Performance optimization and profiling
3. Memory leak detection and fixes
4. Complete documentation
5. Migration tools and guides
6. Benchmarking vs TypeScript version
7. Security audit
8. Release preparation

**Success Criterion**: Production-ready Go backend ready for public release

---

## Week 36: Comprehensive Testing

### Task 36.1: Test Coverage Analysis

```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Goals**:

- [ ] > 80% code coverage across all packages
- [ ] 100% coverage for critical paths (session, provider, tool)
- [ ] All edge cases covered

### Task 36.2: Integration Test Suite

Create `test/integration/integration_test.go`:

```go
package integration_test

import (
    "context"
    "testing"
    "time"

    "github.com/opencode/opencode-go/internal/server"
    "github.com/opencode/opencode-go/internal/session"
    "github.com/opencode/opencode-go/internal/provider"
    "github.com/opencode/opencode-go/internal/storage"
)

func TestFullSessionWorkflow(t *testing.T) {
    // Set up
    store, _ := storage.New("test.db")
    defer store.Close()

    sessionMgr := session.NewManager(store)

    // Create session
    sess, err := sessionMgr.Create(context.Background(), "test-project", "anthropic", "claude-3-5-sonnet-20241022")
    if err != nil {
        t.Fatal(err)
    }

    // Add user message
    msg, err := sessionMgr.AddMessage(context.Background(), sess.ID, session.RoleUser, []session.ContentBlock{
        {Type: "text", Text: "Hello, world!"},
    })
    if err != nil {
        t.Fatal(err)
    }

    // Generate AI response (requires API key)
    if testing.Short() {
        t.Skip("Skipping AI integration test in short mode")
    }

    err = sessionMgr.Generate(context.Background(), sess.ID, "Hello, world!")
    if err != nil {
        t.Fatal(err)
    }

    // Verify messages
    messages, err := sessionMgr.GetMessages(context.Background(), sess.ID)
    if err != nil {
        t.Fatal(err)
    }

    if len(messages) < 2 {
        t.Fatalf("Expected at least 2 messages, got %d", len(messages))
    }

    // Clean up
    sessionMgr.Delete(context.Background(), sess.ID)
}

func TestConcurrentSessions(t *testing.T) {
    store, _ := storage.New("test_concurrent.db")
    defer store.Close()

    sessionMgr := session.NewManager(store)

    // Create 10 concurrent sessions
    sessions := make([]*session.Session, 10)
    errors := make(chan error, 10)

    for i := 0; i < 10; i++ {
        go func(index int) {
            sess, err := sessionMgr.Create(context.Background(), "test-project", "anthropic", "claude-3-5-sonnet-20241022")
            if err != nil {
                errors <- err
                return
            }
            sessions[index] = sess
            errors <- nil
        }(i)
    }

    // Wait for all
    for i := 0; i < 10; i++ {
        if err := <-errors; err != nil {
            t.Fatal(err)
        }
    }

    // Verify all sessions created
    for _, sess := range sessions {
        if sess == nil {
            t.Fatal("Session creation failed")
        }
    }
}

func TestSSEStreaming(t *testing.T) {
    // Test Server-Sent Events streaming
    // Start server, connect TUI, verify events arrive
    // ...
}

func TestToolExecution(t *testing.T) {
    // Test all tools end-to-end
    // Read, Write, Edit, Glob, Grep, Bash, WebFetch, Task, Todo
    // ...
}
```

### Task 36.3: End-to-End Tests

Create `test/e2e/tui_test.go`:

```go
package e2e_test

import (
    "os/exec"
    "testing"
    "time"
)

func TestTUIConnection(t *testing.T) {
    // Start Go backend
    server := exec.Command("./bin/opencode-go")
    if err := server.Start(); err != nil {
        t.Fatal(err)
    }
    defer server.Process.Kill()

    // Wait for server to start
    time.Sleep(2 * time.Second)

    // Start TUI with Go backend
    tui := exec.Command("./bin/opencode-tui")
    tui.Env = append(tui.Env, "OPENCODE_BACKEND=go")

    if err := tui.Start(); err != nil {
        t.Fatal(err)
    }
    defer tui.Process.Kill()

    // TUI should connect successfully
    time.Sleep(5 * time.Second)

    // Verify TUI is running (check for process)
    // ...
}
```

### Task 36.4: Race Condition Testing

```bash
go test -race ./...
```

Fix all data races. Common issues:

- Concurrent map access
- Unprotected shared state
- Missing mutex locks

---

## Week 37: Performance Optimization

### Task 37.1: CPU Profiling

Create `test/benchmark/benchmark_test.go`:

```go
package benchmark_test

import (
    "context"
    "testing"

    "github.com/opencode/opencode-go/internal/session"
    "github.com/opencode/opencode-go/internal/storage"
)

func BenchmarkSessionCreate(b *testing.B) {
    store, _ := storage.New("bench.db")
    defer store.Close()

    mgr := session.NewManager(store)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        mgr.Create(context.Background(), "test", "anthropic", "claude-3-5-sonnet-20241022")
    }
}

func BenchmarkMessageAdd(b *testing.B) {
    store, _ := storage.New("bench.db")
    defer store.Close()

    mgr := session.NewManager(store)
    sess, _ := mgr.Create(context.Background(), "test", "anthropic", "claude-3-5-sonnet-20241022")

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        mgr.AddMessage(context.Background(), sess.ID, session.RoleUser, []session.ContentBlock{
            {Type: "text", Text: "Test message"},
        })
    }
}

func BenchmarkEventBusThroughput(b *testing.B) {
    bus := session.NewEventBus()
    ch := bus.Subscribe("test-session")

    go func() {
        for range ch {
            // Consume events
        }
    }()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        bus.Publish(session.Event{
            Type:      session.EventMessageAdded,
            SessionID: "test-session",
            Data:      "test",
        })
    }
}
```

Run profiling:

```bash
go test -bench=. -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

Optimize hot paths identified by profiler.

### Task 37.2: Memory Profiling

```bash
go test -bench=. -memprofile=mem.prof
go tool pprof mem.prof
```

Look for:

- Excessive allocations
- Memory leaks
- Large object retention

### Task 37.3: 24-Hour Soak Test

```go
func TestMemoryLeaks(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping soak test")
    }

    store, _ := storage.New("soak.db")
    defer store.Close()

    mgr := session.NewManager(store)

    // Run for 24 hours
    timeout := time.After(24 * time.Hour)
    ticker := time.NewTicker(1 * time.Second)

    sessions := 0
    messages := 0

    for {
        select {
        case <-timeout:
            t.Logf("Soak test complete: %d sessions, %d messages", sessions, messages)
            return

        case <-ticker.C:
            // Create session
            sess, _ := mgr.Create(context.Background(), "test", "anthropic", "claude-3-5-sonnet-20241022")
            sessions++

            // Add messages
            for i := 0; i < 10; i++ {
                mgr.AddMessage(context.Background(), sess.ID, session.RoleUser, []session.ContentBlock{
                    {Type: "text", Text: "Test"},
                })
                messages++
            }

            // Delete old sessions
            if sessions%100 == 0 {
                mgr.Delete(context.Background(), sess.ID)
            }
        }
    }
}
```

Monitor:

- Memory usage (should remain constant)
- Goroutine count (should not grow)
- File descriptor leaks

### Task 37.4: Optimization Targets

- [ ] API latency < 10ms for local operations
- [ ] Session creation < 5ms
- [ ] Message add < 2ms
- [ ] Event delivery < 1ms
- [ ] Memory usage < 100MB for 10 sessions
- [ ] No goroutine leaks

---

## Week 38: Documentation & Migration

### Task 38.1: API Documentation

Generate OpenAPI spec:

```go
// internal/server/openapi.go
func GenerateOpenAPISpec() string {
    spec := `
openapi: 3.0.0
info:
  title: OpenCode Go Backend API
  version: 1.0.0
paths:
  /api/sessions:
    get:
      summary: List sessions
      responses:
        '200':
          description: List of sessions
    post:
      summary: Create session
      responses:
        '201':
          description: Session created
  # ... all 48 endpoints
`
    return spec
}
```

### Task 38.2: User Guide

Create `docs/MIGRATION_GUIDE.md`:

```markdown
# Migration Guide: TypeScript to Go Backend

## Overview

This guide helps you migrate from the TypeScript backend to the pure Go backend.

## Prerequisites

- Go 1.21+
- Existing OpenCode installation

## Installation

### Option 1: Download Binary

\`\`\`bash

# Download for your platform

curl -o opencode-go https://releases.opencode.ai/go/latest/linux-amd64
chmod +x opencode-go
\`\`\`

### Option 2: Build from Source

\`\`\`bash
cd packages/opencode-go
go build -o opencode-go ./cmd/server
\`\`\`

## Configuration

The Go backend uses the same configuration format:
\`\`\`yaml

# opencode.yml

providers:
anthropic:
apiKey: sk-ant-xxx
openai:
apiKey: sk-xxx

server:
host: localhost
port: 8080
\`\`\`

## Data Migration

Sessions are stored in BoltDB (compatible format):
\`\`\`bash

# Export from TypeScript backend

opencode export-sessions > sessions.json

# Import to Go backend

opencode-go import-sessions sessions.json
\`\`\`

## Feature Comparison

| Feature   | TypeScript | Go  | Notes               |
| --------- | ---------- | --- | ------------------- |
| Sessions  | ✅         | ✅  | Full parity         |
| Streaming | ✅         | ✅  | SSE-based           |
| Tools     | ✅         | ✅  | All tools supported |
| LSP       | ✅         | ✅  | Same protocol       |
| MCP       | ✅         | ⚠️  | Optional in Go      |

## Performance Improvements

- 🚀 50% faster startup time
- 💾 30% lower memory usage
- ⚡ 2x faster API responses

## Troubleshooting

...
```

### Task 38.3: Developer Guide

Create `docs/DEVELOPMENT.md`:

```markdown
# Development Guide: Go Backend

## Architecture

\`\`\`
cmd/server/ - Entry point
internal/
server/ - HTTP handlers
session/ - Session management
provider/ - AI providers
tool/ - Tool implementations
lsp/ - LSP client
storage/ - Data persistence
\`\`\`

## Building

\`\`\`bash
go build ./cmd/server
\`\`\`

## Testing

\`\`\`bash
go test ./...
go test -race ./...
go test -cover ./...
\`\`\`

## Adding a New Provider

1. Implement `provider.Provider` interface
2. Register in `provider/registry.go`
3. Add tests

## Adding a New Tool

1. Create file in `internal/tool/`
2. Add to tool executor
3. Add API endpoint
4. Add tests
```

### Task 38.4: Migration Script

Create `scripts/migrate.sh`:

```bash
#!/bin/bash
set -e

echo "OpenCode Migration: TypeScript → Go Backend"
echo "==========================================="

# Check prerequisites
command -v opencode >/dev/null 2>&1 || { echo "opencode not found"; exit 1; }
command -v opencode-go >/dev/null 2>&1 || { echo "opencode-go not found"; exit 1; }

# Export data
echo "Exporting sessions..."
opencode export-sessions > /tmp/sessions.json

echo "Exporting configuration..."
opencode export-config > /tmp/config.yml

# Import to Go backend
echo "Importing to Go backend..."
opencode-go import-sessions /tmp/sessions.json
opencode-go import-config /tmp/config.yml

# Verify
echo "Verifying migration..."
opencode-go verify-data

echo "✅ Migration complete!"
echo "You can now use: OPENCODE_BACKEND=go opencode tui"
```

---

## Week 39: Security & Release Prep

### Task 39.1: Security Audit

Review:

- [ ] No hardcoded API keys
- [ ] Environment variables for secrets
- [ ] Input validation on all endpoints
- [ ] SQL injection prevention (BoltDB is safe)
- [ ] Path traversal prevention in file operations
- [ ] Command injection prevention in bash tool
- [ ] XSS prevention (not applicable - TUI only)
- [ ] Rate limiting on AI provider calls

### Task 39.2: Security Fixes

Example: Path traversal prevention

```go
func (t *ToolExecutor) Read(req ReadRequest) (*ReadResponse, error) {
    // Prevent path traversal
    absPath, err := filepath.Abs(req.FilePath)
    if err != nil {
        return nil, err
    }

    // Ensure path is within project
    if !strings.HasPrefix(absPath, t.projectPath) {
        return nil, fmt.Errorf("access denied: path outside project")
    }

    // Continue with read...
}
```

### Task 39.3: Release Checklist

Create `RELEASE_CHECKLIST.md`:

```markdown
# Go Backend Release Checklist

## Pre-Release

- [ ] All tests passing
- [ ] No race conditions
- [ ] Code coverage >80%
- [ ] Documentation complete
- [ ] CHANGELOG.md updated
- [ ] Version bumped

## Build

- [ ] Build for linux-amd64
- [ ] Build for darwin-amd64
- [ ] Build for darwin-arm64
- [ ] Build for windows-amd64
- [ ] All binaries <50MB

## Testing

- [ ] Manual testing on all platforms
- [ ] TUI connection test
- [ ] AI provider tests (all)
- [ ] Tool tests (all)
- [ ] Migration test

## Release

- [ ] Tag version (v1.0.0-go)
- [ ] Create GitHub release
- [ ] Upload binaries
- [ ] Update documentation site
- [ ] Announce on Discord/Twitter
```

### Task 39.4: Build Pipeline

Create `.github/workflows/build-go.yml`:

```yaml
name: Build Go Backend

on:
  push:
    tags:
      - "v*-go"

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        goos: [linux, darwin, windows]
        goarch: [amd64, arm64]
        exclude:
          - goos: windows
            goarch: arm64

    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-go@v4
        with:
          go-version: "1.21"

      - name: Build
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
        run: |
          cd packages/opencode-go
          go build -o opencode-go-${{ matrix.goos }}-${{ matrix.goarch }} ./cmd/server

      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: opencode-go-${{ matrix.goos }}-${{ matrix.goarch }}
          path: packages/opencode-go/opencode-go-${{ matrix.goos }}-${{ matrix.goarch }}
```

---

## Week 40: Final Testing & Launch

### Task 40.1: Beta Testing

- [ ] Internal testing with team
- [ ] External beta with 10-20 users
- [ ] Collect feedback
- [ ] Fix critical bugs

### Task 40.2: Performance Comparison

Run benchmarks vs TypeScript:

```bash
# Startup time
time opencode tui --backend=typescript &  # Kill after start
time opencode tui --backend=go &

# Memory usage
ps aux | grep opencode

# API latency
curl -w "@curl-format.txt" http://localhost:8080/api/sessions
```

Create comparison table:
| Metric | TypeScript | Go | Improvement |
|--------|-----------|-----|-------------|
| Startup | 2.5s | 0.8s | 68% faster |
| Memory | 150MB | 95MB | 37% less |
| API latency | 15ms | 6ms | 60% faster |
| Binary size | 80MB (Bun) | 45MB | 44% smaller |

### Task 40.3: Documentation Site Update

Update https://opencode.ai/docs:

- Add "Go Backend" section
- Migration guide
- Performance comparison
- FAQ

### Task 40.4: Launch

1. **Soft Launch**: Feature flag in TUI (`OPENCODE_BACKEND=go`)
2. **Announcement**: Blog post, Discord, Twitter
3. **Monitor**: Error reporting, user feedback
4. **Iterate**: Fix issues, optimize
5. **Default**: Make Go backend default after 2-4 weeks

---

## Success Metrics

### Code Quality

- [ ] > 80% test coverage
- [ ] Zero race conditions
- [ ] Zero memory leaks
- [ ] All linters pass

### Performance

- [ ] Startup time <1s
- [ ] Memory usage <100MB (10 sessions)
- [ ] API latency <10ms
- [ ] Binary size <50MB

### Reliability

- [ ] All tests passing
- [ ] 24-hour soak test passes
- [ ] No panics or crashes
- [ ] Graceful error handling

### Usability

- [ ] Complete documentation
- [ ] Migration guide
- [ ] Release notes
- [ ] Troubleshooting guide

---

## Phase 6 Checklist

### Testing

- [ ] Achieve >80% code coverage
- [ ] Write integration tests
- [ ] Write e2e tests
- [ ] Run race detector (fix all issues)
- [ ] 24-hour soak test
- [ ] Load testing (10+ sessions)

### Performance

- [ ] CPU profiling
- [ ] Memory profiling
- [ ] Optimize hot paths
- [ ] Fix memory leaks
- [ ] Benchmark vs TypeScript

### Documentation

- [ ] API documentation (OpenAPI)
- [ ] User migration guide
- [ ] Developer guide
- [ ] CHANGELOG
- [ ] Release notes

### Security

- [ ] Security audit
- [ ] Fix vulnerabilities
- [ ] Input validation
- [ ] Path traversal prevention
- [ ] Command injection prevention

### Release

- [ ] Build for all platforms
- [ ] Create release checklist
- [ ] Set up CI/CD pipeline
- [ ] Beta testing
- [ ] Launch plan

### Post-Release

- [ ] Monitor errors
- [ ] Collect feedback
- [ ] Fix critical bugs
- [ ] Performance tuning
- [ ] Make default backend

---

## Conclusion

After Phase 6, you will have:
✅ A production-ready pure Go backend  
✅ Single-binary distribution (no Bun/Node.js)  
✅ Complete feature parity with TypeScript  
✅ Better performance (faster, less memory)  
✅ Comprehensive testing and documentation  
✅ Migration path for existing users

**Timeline Summary**:

- Phase 1-2: Foundation (10 weeks)
- Phase 3: Sessions (6 weeks)
- Phase 4: AI Integration (12 weeks) ⚠️ HARD
- Phase 5: Advanced Features (7 weeks)
- Phase 6: Production (5 weeks)

**Total: 40 weeks (8-10 months) with 2-3 engineers**

The TUI will work identically, but powered by pure Go! 🎉
