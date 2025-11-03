# Pure Go TUI Implementation Plan

## Overview

This plan details the migration of OpenCode's TypeScript backend to pure Go for TUI (Terminal User Interface) operation. This will enable distribution as a single binary without requiring Bun/Node.js runtime.

**Objective**: Rewrite ~14,000 lines of TypeScript backend code to Go while keeping the existing 40,000+ lines of Go TUI interface.

**Timeline**: 4-8 months (6 phases)  
**Team Size**: 2-3 experienced Go engineers  
**Risk Level**: 🟡 Medium  
**Effort**: 35-56 weeks of engineering time

---

## Project Structure

```
packages/
├── opencode/          # Current TypeScript backend (will remain for fallback)
├── opencode-go/       # New Go backend (to be created)
│   ├── cmd/
│   │   └── server/    # Main HTTP server entry point
│   ├── internal/
│   │   ├── server/    # HTTP API handlers
│   │   ├── session/   # Session management
│   │   ├── provider/  # AI provider integrations
│   │   ├── tool/      # Tool system
│   │   ├── lsp/       # LSP integration
│   │   ├── file/      # File operations
│   │   ├── project/   # Project management
│   │   ├── config/    # Configuration
│   │   ├── storage/   # Data persistence
│   │   └── mcp/       # MCP integration (optional)
│   ├── pkg/           # Public packages
│   └── go.mod
└── tui/               # Existing Go TUI (minimal changes needed)
```

---

## Migration Strategy

### Approach: Incremental Parallel Development

**Not** a big-bang rewrite. Instead:

1. ✅ Keep TypeScript backend running (production)
2. ✅ Build Go backend in parallel (`packages/opencode-go`)
3. ✅ Add feature flag to switch between backends
4. ✅ Port components incrementally
5. ✅ Test each component thoroughly before moving on
6. ✅ Switch to Go backend when feature parity achieved

### Feature Flag Mechanism

```go
// Environment variable to choose backend
OPENCODE_BACKEND=typescript  // Current default
OPENCODE_BACKEND=go          // New Go backend
```

TUI will detect which backend is available and connect accordingly.

---

## Six-Phase Implementation

### [Phase 1: Foundation & Infrastructure](./purego_plans/phase-1.md) (Weeks 1-4)

**Goal**: Basic HTTP server with stub endpoints

- Set up Go project structure
- Implement HTTP API framework
- Create stub endpoints (48 total)
- Configuration loading system
- Storage infrastructure
- Basic routing and middleware
- Health check endpoints

**Deliverable**: TUI can connect to Go server (endpoints return empty/mock data)

---

### [Phase 2: File Operations & Tools](./purego_plans/phase-2.md) (Weeks 5-10)

**Goal**: Basic file operations and command execution work

- File read/write/edit tools
- Glob pattern matching
- Grep (ripgrep integration)
- Bash command execution
- Project state management
- Path operations and worktree detection
- .gitignore pattern handling

**Deliverable**: Can browse files, search code, and execute shell commands

---

### [Phase 3: Session Management](./purego_plans/phase-3.md) (Weeks 11-16)

**Goal**: Session lifecycle without AI integration

- Session CRUD operations
- Session state persistence
- Message data structures
- Event bus and streaming
- Session history management
- Concurrent session handling
- Context management

**Deliverable**: Can create/manage sessions and store messages (no AI yet)

---

### [Phase 4: AI Provider Integration](./purego_plans/phase-4.md) ✅ **COMPLETE** (Weeks 17-28)

**Goal**: Working AI conversations with tool calling

**Status**: ✅ Completed on 2025-11-03

**Critical Path - Hardest Component**

- ✅ Start with Anthropic Claude (simplest API)
- ✅ Streaming response handling (SSE)
- ✅ Tool/function calling integration
- ✅ Error handling and retries
- ✅ Add OpenAI support
- ✅ Add Google (Gemini) support
- ⚠️ Other providers (Bedrock, Azure, etc.) - Future enhancement
- ✅ Model registry and configuration
- ✅ Rate limiting and quotas

**Deliverable**: Full AI conversations with multi-provider support

**Achievements**:

- 3 providers integrated: Anthropic Claude, OpenAI, Google Gemini
- Provider abstraction layer with unified interface
- SSE streaming for all providers
- Automatic retry logic (3 retries, exponential backoff)
- Tool calling support across providers
- Token usage tracking
- 68 tests (54 unit + 4 integration + 10 additional)
- Complete documentation (540+ lines)

---

### [Phase 5: Advanced Features](./purego_plans/phase-5.md) (Weeks 29-35)

**Goal**: Feature parity with TypeScript backend

- LSP (Language Server Protocol) integration
- Session summarization
- Session compaction
- Todo list management
- Session forking and branching
- Revert functionality
- Web fetching (HTML → Markdown)
- Task/sub-agent spawning
- Tree-sitter integration

**Deliverable**: Complete feature parity

---

### [Phase 6: Production Readiness](./purego_plans/phase-6.md) (Weeks 36-40)

**Goal**: Production-ready release

- Comprehensive error handling
- Performance optimization
- Memory profiling and optimization
- Comprehensive testing (unit, integration, e2e)
- Migration tooling
- Documentation
- Benchmarking vs TypeScript version
- Security audit
- Release preparation

**Deliverable**: Production-ready pure Go TUI

---

## Component Complexity Matrix

| Component              | TS Lines   | Go Estimate | Weeks     | Complexity   | Dependencies  |
| ---------------------- | ---------- | ----------- | --------- | ------------ | ------------- |
| **HTTP API Server**    | 3,000      | 3,500       | 2-3       | 🟢 Low-Med   | chi/gin/echo  |
| **Session Management** | 4,100      | 4,500       | 6-10      | 🟠 High      | sync, context |
| **AI Providers**       | 900        | 2,000       | 8-12      | 🔴 Very High | HTTP, SSE     |
| **Tool System**        | 2,300      | 2,000       | 4-6       | 🟡 Medium    | os/exec       |
| **LSP Integration**    | 800        | 1,000       | 3-5       | 🟡 Medium    | JSON-RPC      |
| **File Operations**    | 600        | 500         | 1-2       | 🟢 Low       | os, filepath  |
| **Project Management** | 1,200      | 1,000       | 2-3       | 🟢 Low-Med   | os, git       |
| **Configuration**      | 300        | 300         | 1-2       | 🟢 Low       | encoding/json |
| **Storage**            | 400        | 400         | 1-2       | 🟢 Low       | database/sql  |
| **MCP**                | 400        | 800         | 4-6       | 🟠 High      | JSON-RPC      |
| **Utilities**          | 800        | 600         | 2-3       | 🟢 Low-Med   | stdlib        |
| **TOTAL**              | **14,900** | **16,600**  | **35-56** |              |               |

---

## Key Technical Decisions

### 1. Web Framework

**Decision**: Use **Chi router** (stdlib-based, lightweight)

**Alternatives considered**:

- Gin (fast but heavier)
- Echo (feature-rich but opinionated)
- Fiber (Express-like but v3 breaking changes)
- stdlib net/http (verbose but zero deps)

**Rationale**: Chi is minimal, uses stdlib patterns, and sufficient for API server needs.

### 2. AI Provider Clients

**Decision**: Use official SDKs where available + custom HTTP clients

- **Anthropic**: Use `anthropic-sdk-go` (official)
- **OpenAI**: Use `sashabaranov/go-openai`
- **Google**: Custom HTTP client (no mature SDK)
- **Others**: Custom implementations

**Rationale**: Leverage existing SDKs to avoid reinventing streaming/tool calling logic.

### 3. Session State Storage

**Decision**: **BoltDB** for embedded database

**Alternatives considered**:

- SQLite (heavier, SQL overhead)
- BadgerDB (complex API)
- File-based JSON (no transactions)
- PostgreSQL (external dependency)

**Rationale**: BoltDB is pure Go, embedded, transactional, and perfect for local session storage.

### 4. LSP Client

**Decision**: Custom JSON-RPC implementation using `sourcegraph/jsonrpc2`

**Rationale**: Go LSP clients are limited; building on jsonrpc2 gives flexibility.

### 5. Streaming (SSE)

**Decision**: Use channels + `text/event-stream` response writer

**Pattern**:

```go
func (s *Server) StreamMessages(w http.ResponseWriter, r *http.Request) {
    flusher, ok := w.(http.Flusher)
    ch := s.session.Subscribe(sessionID)

    for msg := range ch {
        fmt.Fprintf(w, "data: %s\n\n", msg)
        flusher.Flush()
    }
}
```

### 6. Concurrency Patterns

**Decision**: Goroutines + channels for async operations

**TypeScript → Go mappings**:

- `async/await` → goroutines + channels/contexts
- `Promise.all()` → `errgroup.Group`
- `EventEmitter` → channels or `sync.Cond`
- `setTimeout` → `time.After()` or `time.Ticker`

---

## Risk Management

### High-Risk Areas

#### 1. AI Provider Integration (🔴 Critical)

**Risks**:

- Streaming implementation bugs
- Tool calling format mismatches
- Provider API changes
- Rate limiting issues

**Mitigation**:

- Start with one provider (Anthropic - simplest)
- Extensive testing with real API calls
- Implement retry logic from day one
- Version lock provider SDKs
- Create provider abstraction layer

#### 2. Session State Management (🟠 High)

**Risks**:

- Race conditions in concurrent access
- Message ordering issues
- State corruption
- Memory leaks from unclosed channels

**Mitigation**:

- Use mutexes liberally
- Context cancellation for cleanup
- Thorough race detector testing (`go test -race`)
- Memory profiling (`pprof`)

#### 3. Feature Parity Gaps (🟡 Medium)

**Risks**:

- Missing edge cases from TypeScript version
- Behavioral differences
- Breaking existing workflows

**Mitigation**:

- Maintain comprehensive test suite
- Side-by-side testing (TS vs Go)
- Feature flag for gradual rollout
- Keep TypeScript backend as fallback

### Testing Strategy

#### Unit Tests

- Each component >80% coverage
- Table-driven tests for tools
- Mock AI provider responses

#### Integration Tests

- Real API calls to providers (with test accounts)
- File operation workflows
- Session lifecycle tests

#### End-to-End Tests

- Full TUI → Go backend → AI provider flows
- Automated TUI interaction tests
- Performance benchmarks

#### Migration Tests

- Data migration from TS to Go storage
- Session compatibility
- Configuration migration

---

## Success Criteria

### Functional Requirements

- [x] Phase 1-3: Foundation, File Operations, Session Management ✅
- [x] Phase 4: Multi-provider AI support (Anthropic, OpenAI, Google) ✅
- [x] Streaming responses work correctly ✅
- [x] Tool calling works for all providers ✅
- [ ] Phase 5: LSP integration functional
- [ ] Phase 5: All tools functional (read, write, edit, glob, grep, bash, webfetch, task, todo)
- [ ] Phase 6: All 48 API endpoints implemented
- [ ] Phase 6: TUI works identically with Go backend
- [ ] Phase 6: Session management feature-complete

### Performance Requirements

- [ ] API latency ≤ TypeScript version
- [ ] Memory usage ≤ TypeScript version
- [ ] Startup time < 1 second
- [ ] Supports 10+ concurrent sessions
- [ ] No memory leaks (24-hour soak test)

### Quality Requirements

- [ ] > 80% test coverage
- [ ] Zero race conditions (`go test -race` passes)
- [ ] All linters pass (`golangci-lint`)
- [ ] Documentation complete
- [ ] Migration guide written

### Distribution Requirements

- [ ] Single binary for Linux, macOS, Windows
- [ ] Binary size < 50MB
- [ ] No external runtime dependencies
- [ ] Installation script updated
- [ ] Homebrew formula (if applicable)

---

## Timeline & Milestones

### Month 1-2: Foundation (Phases 1-2)

**Weeks 1-4**: HTTP server, routing, configuration  
**Weeks 5-10**: File operations, basic tools

**Milestone**: TUI can browse files and execute commands with Go backend

### Month 3-4: Core Logic (Phase 3)

**Weeks 11-16**: Session management, state persistence, messaging

**Milestone**: Sessions work without AI integration

### Month 5-6: AI Integration (Phase 4)

**Weeks 17-28**: Provider integrations, streaming, tool calling

**Milestone**: AI conversations work with multiple providers

### Month 7: Advanced Features (Phase 5)

**Weeks 29-35**: LSP, summarization, advanced session features

**Milestone**: Feature parity achieved

### Month 8: Polish (Phase 6)

**Weeks 36-40**: Testing, optimization, documentation, release prep

**Milestone**: Production release

---

## Resource Requirements

### Team Composition

- **2 Senior Go Engineers** (full-time)
  - Deep Go experience (3+ years)
  - HTTP/REST API development
  - Concurrency patterns
  - Experience with AI APIs preferred
- **1 Full-Stack Engineer** (50% time)
  - TypeScript expertise (to reference current implementation)
  - Testing and quality assurance
  - Documentation

### Infrastructure

- **Development**: Local machines
- **Testing**: GitHub Actions CI/CD
- **AI Provider Access**: Test accounts with API credits
  - Anthropic Claude API
  - OpenAI API
  - Google Gemini API
- **Monitoring**: Basic logging and metrics

### Budget Estimate

- **Engineering time**: 2.5 FTE × 6 months = 15 person-months
- **AI API testing**: $500-1,000/month × 6 months = $3,000-6,000
- **Total**: ~$150,000-250,000 (depends on engineer rates)

---

## Dependencies & Prerequisites

### External Dependencies (Go Modules)

#### Critical

- `github.com/go-chi/chi/v5` - HTTP router
- `github.com/anthropics/anthropic-sdk-go` - Anthropic client
- `github.com/sashabaranov/go-openai` - OpenAI client
- `go.etcd.io/bbolt` - Embedded database
- `sourcegraph/jsonrpc2` - LSP foundation

#### Important

- `github.com/fsnotify/fsnotify` - File watching
- `gopkg.in/yaml.v3` - YAML parsing
- `github.com/go-playground/validator/v10` - Validation
- `golang.org/x/sync/errgroup` - Concurrency utilities

#### Optional

- `github.com/smacker/go-tree-sitter` - Tree-sitter bindings
- `github.com/JohannesKaufmann/html-to-markdown` - HTML conversion
- `github.com/gobwas/glob` - Glob patterns

### Internal Dependencies

- **TUI package** (`packages/tui`) - Consumer of Go backend API
- **Go SDK** (`packages/sdk/go`) - May need updates for new types

### Tool Requirements

- Go 1.21+
- ripgrep (for grep tool)
- LSP servers (for testing LSP integration)

---

## Migration Checklist

See detailed checklists in individual phase files:

- [Phase 1 Checklist](./purego_plans/phase-1.md#checklist)
- [Phase 2 Checklist](./purego_plans/phase-2.md#checklist)
- [Phase 3 Checklist](./purego_plans/phase-3.md#checklist)
- [Phase 4 Checklist](./purego_plans/phase-4.md#checklist)
- [Phase 5 Checklist](./purego_plans/phase-5.md#checklist)
- [Phase 6 Checklist](./purego_plans/phase-6.md#checklist)

---

## Rollback Plan

### If Migration Fails

**Always keep TypeScript backend operational**:

1. Feature flag allows instant rollback
2. No breaking changes to TUI interface
3. Session data format compatible (JSON)
4. Configuration format unchanged

### Rollback Triggers

- Critical bugs not fixable within 1 week
- Performance regressions >50%
- Provider API compatibility issues
- Data corruption in Go backend

### Rollback Procedure

1. Set `OPENCODE_BACKEND=typescript`
2. Stop Go server process
3. Start TypeScript server
4. TUI reconnects automatically
5. No data loss (sessions stored in compatible format)

---

## Post-Migration Plan

### Phase 7: MCP Integration (Optional)

**Weeks 41-48**: If MCP support needed

- Implement Model Context Protocol from spec
- No mature Go SDK exists
- Build custom JSON-RPC client

### Phase 8: Performance Optimization

- Profile production usage
- Optimize hot paths
- Reduce memory allocations
- Connection pooling

### Phase 9: Plugin System (Future)

- Design Go plugin architecture
- Enable community extensions
- Tool registry for custom tools

---

## Communication Plan

### Weekly Updates

- Progress against phase milestones
- Blockers and risks
- Demos of completed components

### Phase Completion Reviews

- Demo to stakeholders
- Testing report
- Performance benchmarks
- Go/no-go decision for next phase

### Documentation

- Architecture decision records (ADRs)
- API documentation (OpenAPI spec)
- Migration guide for users
- Developer guide for contributors

---

## Conclusion

This plan provides a structured, incremental approach to rewriting OpenCode's TypeScript backend to pure Go for TUI operation. The phased approach minimizes risk while delivering value progressively.

**Key Success Factors**:

1. ✅ Incremental development (not big-bang)
2. ✅ Feature flag for safe rollback
3. ✅ Focus on AI provider integration (hardest part)
4. ✅ Comprehensive testing at each phase
5. ✅ Keep TypeScript backend as safety net
6. ✅ Clear success criteria and milestones

**Expected Outcome**: Single-binary Go distribution for TUI users without Bun/Node.js dependency, better performance, and simpler deployment.

---

## Appendix

### Reference Links

- [TypeScript Backend Code](../packages/opencode)
- [Go TUI Code](../packages/tui)
- [PURE_GO Impact Analysis](../PURE_GO.md)

### Related Documents

- Phase-specific implementation details in `./purego_plans/`
- Architecture decision records (to be created)
- API compatibility documentation (to be created)
