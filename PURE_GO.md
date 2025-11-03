# Impact Analysis: Rewriting TypeScript/JavaScript to Go (TUI-Only Version)

## Executive Summary

This document analyzes the impact of rewriting the **minimal TypeScript/JavaScript code required for TUI terminal operation** to pure Go. This focuses only on making the terminal interface work without the web UI, desktop app, console, or other non-essential components.

**Current Architecture**: The TUI (Terminal User Interface) is already written in Go (~40,000 lines), but it depends on a TypeScript/JavaScript backend server (`packages/opencode`) that provides the HTTP API.

**Verdict for TUI-Only**: **MODERATE to HIGH complexity**. Achievable in 4-8 months with 2-3 engineers, but requires rewriting significant core logic.

---

## Current TUI Architecture

### How It Works Now

1. **TypeScript CLI (`packages/opencode`)** starts an HTTP server (Hono-based)
2. **Go TUI (`packages/tui`)** connects to this server via HTTP/SSE
3. TUI makes API calls for all operations (sessions, messages, file operations, etc.)
4. Backend handles:
   - AI provider integration
   - Session management
   - Tool execution (bash, file operations, etc.)
   - LSP integration
   - Project state
   - MCP (Model Context Protocol)

### Key Connection Point

From `packages/opencode/src/cli/cmd/tui.ts`:

```typescript
// TypeScript spawns the Go TUI binary and passes server URL
const proc = Bun.spawn({
  cmd: [...cmd],
  env: {
    OPENCODE_SERVER: server.url.toString(), // HTTP server URL
  },
})
```

From `packages/tui/cmd/opencode/main.go`:

```go
// Go TUI reads server URL and connects
url := os.Getenv("OPENCODE_SERVER")
httpClient := opencode.NewClient(option.WithBaseURL(url))

// Makes API calls like:
httpClient.Project.Current(context.Background(), ...)
httpClient.Agent.List(context.Background(), ...)
httpClient.Session.Create(context.Background(), ...)
```

---

## What TypeScript Code Powers the TUI

### Core Backend Server (`packages/opencode/src/server/server.ts`)

- **Lines of Code**: ~3,000+ lines
- **Purpose**: HTTP API server that TUI depends on
- **Key Features**:
  - 48+ API endpoints
  - Session management
  - Message streaming (SSE)
  - Configuration management
  - Tool registry

### Session Management (`packages/opencode/src/session/`)

- **Lines of Code**: ~4,100 lines
- **Components**:
  - Message handling and streaming
  - Session state management
  - LLM conversation orchestration
  - Prompt engineering
  - Session compaction/summarization
  - Todo list management
  - Retry logic
  - Revert functionality

### Provider Integration (`packages/opencode/src/provider/`)

- **Lines of Code**: ~900 lines
- **Purpose**: AI provider abstraction (Anthropic, OpenAI, Google, etc.)
- **Dependencies**:
  - Vercel AI SDK (`ai` package)
  - Provider-specific SDKs
  - Model registry and configuration

### Tool System (`packages/opencode/src/tool/`)

- **Lines of Code**: ~2,300 lines
- **Tools Implemented**:
  - `bash` - Execute shell commands
  - `read` - Read files
  - `write` - Write files
  - `edit` - Edit files with replacements
  - `glob` - File pattern matching
  - `grep` - Content search
  - `webfetch` - Fetch web content
  - `task` - Spawn sub-agents
  - `todo` - Task management
  - Tool registry and validation

### LSP Integration (`packages/opencode/src/lsp/`)

- **Lines of Code**: ~800 lines
- **Purpose**: Language Server Protocol client
- **Features**:
  - Diagnostics
  - Hover information
  - Code completions
  - Multi-language support via tree-sitter

### Project Management (`packages/opencode/src/project/`)

- **Lines of Code**: ~1,200 lines
- **Purpose**: Project state and instance management
- **Features**:
  - Project initialization
  - State persistence
  - Worktree detection
  - Configuration loading

### File Operations (`packages/opencode/src/file/`)

- **Lines of Code**: ~600 lines
- **Purpose**: File system operations
- **Features**:
  - Ripgrep integration
  - File watching
  - Ignore patterns (.gitignore, etc.)

### MCP (Model Context Protocol) (`packages/opencode/src/mcp/`)

- **Lines of Code**: ~400 lines
- **Purpose**: Integration with MCP servers
- **Dependencies**: `@modelcontextprotocol/sdk`

### Agent System (`packages/opencode/src/agent/`)

- **Lines of Code**: ~300 lines
- **Purpose**: Agent definitions and management

### Other Core Systems

- **Configuration** (`src/config/`): ~300 lines
- **Authentication** (`src/auth/`): ~200 lines
- **Storage** (`src/storage/`): ~400 lines
- **Bus/Events** (`src/bus/`): ~200 lines
- **Global state** (`src/global/`): ~200 lines

### Total Core Backend for TUI

**~14,000 lines of TypeScript** that would need to be rewritten

---

## Required Components for Pure Go TUI

### 1. HTTP API Server (Must Rewrite)

**Current**: Hono (TypeScript web framework)  
**Go Alternative**:

- **gin** / **echo** / **fiber** / **chi** - Similar web frameworks
- **stdlib net/http** - More verbose but no dependencies

**Effort**: 2-3 weeks  
**Complexity**: 🟢 LOW to MEDIUM

The API surface is well-defined (48 endpoints). Go web frameworks are mature. The main work is translating TypeScript handlers to Go.

### 2. Session Management (Must Rewrite)

**Current**: Complex TypeScript implementation with:

- Message streaming
- State management
- Compaction
- Summarization
- Todo tracking

**Go Alternative**: Reimplement from scratch

**Effort**: 6-10 weeks  
**Complexity**: 🟠 HIGH

This is complex business logic with:

- Concurrent message handling
- State persistence
- Message history management
- Streaming to multiple clients

**Challenges**:

- TypeScript's async/await patterns need Go goroutine equivalents
- State management architecture decisions
- Message format compatibility

### 3. AI Provider Integration (Must Rewrite)

**Current**: Vercel AI SDK (TypeScript) with provider abstractions

**Go Alternatives**:

- **Direct API calls** to OpenAI, Anthropic, Google APIs
- **go-openai** - OpenAI Go client
- **anthropic-sdk-go** - Anthropic Go client
- **Custom HTTP clients** for each provider

**Effort**: 8-12 weeks  
**Complexity**: 🔴 VERY HIGH

This is the hardest part:

- Vercel AI SDK is sophisticated and handles many edge cases
- Streaming responses
- Function/tool calling
- Error handling and retries
- Provider-specific quirks
- Model registry and configuration

**Challenges**:

- AI SDK abstracts away complexity
- Need to handle streaming SSE properly
- Tool calling formats differ by provider
- Go libraries less mature than TypeScript versions

### 4. Tool System (Must Rewrite)

**Current**: ~2,300 lines implementing all tools

**Go Alternative**: Reimplement each tool

**Effort**: 4-6 weeks  
**Complexity**: 🟡 MEDIUM

Most tools are straightforward:

- `bash`: Already have `exec.Command` in Go
- `read`/`write`: Standard library file operations
- `edit`: String manipulation and file I/O
- `glob`: Use `filepath.Match` or `github.com/gobwas/glob`
- `grep`: Use `ripgrep` via exec or reimplement

**Easier in Go**:

- File operations are simpler
- exec.Command is built-in
- No need for Bun-specific APIs

**Harder in Go**:

- `webfetch`: Need HTTP client + HTML parsing library
- `task`: Sub-agent spawning and management
- Dynamic tool registration (current plugin system)

### 5. LSP Integration (Must Rewrite)

**Current**: TypeScript LSP client using `vscode-jsonrpc`

**Go Alternative**:

- **go-lsp** - Language Server Protocol for Go
- **gopls** packages - Can reuse parts
- **Custom JSON-RPC client**

**Effort**: 3-5 weeks  
**Complexity**: 🟡 MEDIUM

LSP is a well-defined protocol. Go has libraries for this. Main work is:

- Setting up LSP client
- Managing language server processes
- Handling diagnostics and completions

### 6. File System Operations (Easy to Rewrite)

**Current**: Ripgrep integration, file watching

**Go Alternative**:

- **os** package for basic operations
- **exec.Command** to call ripgrep
- **fsnotify** for file watching

**Effort**: 1-2 weeks  
**Complexity**: 🟢 LOW

Go excels at file operations. This is straightforward.

### 7. Configuration & Storage (Easy to Rewrite)

**Current**: TypeScript config loading and storage

**Go Alternative**:

- **encoding/json** for JSON
- **BurntSushi/toml** for TOML
- **os** package for paths

**Effort**: 1-2 weeks  
**Complexity**: 🟢 LOW

Simple data persistence and configuration loading.

### 8. MCP Integration (Must Rewrite or Skip)

**Current**: `@modelcontextprotocol/sdk` (TypeScript)

**Go Alternative**:

- Implement MCP protocol from scratch
- Or **skip MCP support initially**

**Effort**: 4-6 weeks (or skip)  
**Complexity**: 🟠 HIGH (or skip)

MCP is relatively new. No mature Go SDK exists. Would need to implement protocol or skip this feature.

**Recommendation**: Skip initially, add later if needed

### 9. Project & State Management (Easy to Rewrite)

**Current**: TypeScript project management

**Go Alternative**: Standard Go code

**Effort**: 2-3 weeks  
**Complexity**: 🟢 LOW to MEDIUM

---

## Components NOT Needed for TUI-Only

These can be completely ignored:

- ❌ Desktop UI (`packages/desktop`) - 8,000 lines
- ❌ Web docs (`packages/web`) - 5,000 lines
- ❌ UI library (`packages/ui`) - 6,000 lines
- ❌ Console (`packages/console`) - 10,000 lines
- ❌ Infrastructure (`infra/`) - 1,000 lines
- ❌ VS Code extension (`sdks/vscode`) - 1,000 lines
- ❌ Slack integration (`packages/slack`) - 300 lines
- ❌ JavaScript SDK (`packages/sdk/js`) - 2,000 lines

**Total Eliminated**: ~33,000 lines (60% of TypeScript code)

---

## Rewrite Effort Breakdown (TUI-Only)

| Component               | Lines (TS)  | Effort            | Complexity    | Priority       |
| ----------------------- | ----------- | ----------------- | ------------- | -------------- |
| HTTP API Server         | ~3,000      | 2-3 weeks         | 🟢 Low-Medium | Critical       |
| Session Management      | ~4,100      | 6-10 weeks        | 🟠 High       | Critical       |
| AI Provider Integration | ~900        | 8-12 weeks        | 🔴 Very High  | Critical       |
| Tool System             | ~2,300      | 4-6 weeks         | 🟡 Medium     | Critical       |
| LSP Integration         | ~800        | 3-5 weeks         | 🟡 Medium     | High           |
| File Operations         | ~600        | 1-2 weeks         | 🟢 Low        | High           |
| Project Management      | ~1,200      | 2-3 weeks         | 🟢 Low-Medium | High           |
| Configuration           | ~300        | 1-2 weeks         | 🟢 Low        | High           |
| Storage                 | ~400        | 1-2 weeks         | 🟢 Low        | High           |
| MCP Integration         | ~400        | 4-6 weeks (skip?) | 🟠 High       | Low (optional) |
| Other utilities         | ~800        | 2-3 weeks         | 🟢 Low-Medium | Medium         |
| **TOTAL**               | **~14,900** | **35-56 weeks**   |               |                |

**Realistic Timeline**: 4-8 months with 2-3 experienced Go engineers

---

## Key Challenges

### 1. AI SDK Replacement (Biggest Challenge)

The Vercel AI SDK is extremely sophisticated:

- Handles streaming from multiple providers
- Manages function/tool calling
- Handles errors and retries
- Abstracts provider differences

**In Go, you'd need to**:

- Write HTTP clients for each provider
- Handle Server-Sent Events (SSE) streaming
- Parse and format tool calls per provider
- Implement retry logic
- Handle rate limiting

**Mitigation**:

- Start with one provider (e.g., Anthropic)
- Use existing Go SDKs where available
- Copy patterns from AI SDK source code

### 2. Session State Management

Complex concurrent state management:

- Multiple sessions running simultaneously
- Message streaming to TUI
- State persistence
- Session forking and branching

**In Go**:

- Use goroutines and channels
- Mutex locks for shared state
- Context for cancellation

### 3. TypeScript to Go Patterns

Common TypeScript patterns that need translation:

- `async/await` → goroutines + channels/contexts
- Promises → channels or sync primitives
- Dynamic typing → interfaces and type assertions
- JSON schema validation → struct tags + validation library

### 4. Dependency Management

TypeScript has rich ecosystem:

- tree-sitter bindings
- Zod for validation
- gray-matter for frontmatter
- turndown for HTML→Markdown

**Go alternatives**:

- tree-sitter: `tree-sitter-go` (different API)
- Validation: `go-playground/validator`
- Frontmatter: `go-yaml` + custom parsing
- HTML→Markdown: `gomarkdown` or custom

---

## What You Keep (Already Go)

✅ **TUI Interface** - All 40,000+ lines stay as-is  
✅ **Go SDK** - Already exists  
✅ **Terminal rendering** - Bubble Tea framework  
✅ **UI components** - All TUI components in Go  
✅ **Input handling** - Keyboard/mouse in Go

**This is huge**: 70% of terminal user experience is already Go!

---

## Phased Migration Strategy

### Phase 1: API Foundation (Weeks 1-4)

**Goal**: Basic HTTP server with stub endpoints

- ✅ Set up Go web server (Chi/Gin/Echo)
- ✅ Implement API structure (48 endpoints as stubs)
- ✅ Basic routing and middleware
- ✅ Configuration loading
- ✅ Storage setup

**Deliverable**: TUI can connect but nothing works yet

### Phase 2: Basic Operations (Weeks 5-10)

**Goal**: File operations and project management work

- ✅ File read/write/edit tools
- ✅ Glob and grep tools
- ✅ Bash execution
- ✅ Project state management
- ✅ Path operations

**Deliverable**: Can browse files and execute commands

### Phase 3: Session Foundation (Weeks 11-16)

**Goal**: Basic session management without AI

- ✅ Session CRUD operations
- ✅ Session state persistence
- ✅ Message data structures
- ✅ Event streaming to TUI

**Deliverable**: Can create sessions and store messages

### Phase 4: AI Integration (Weeks 17-28)

**Goal**: Working AI conversation (hardest part)

- ✅ Provider integration (start with one: Anthropic)
- ✅ Streaming responses
- ✅ Tool calling integration
- ✅ Model configuration
- ✅ Add more providers (OpenAI, Google, etc.)

**Deliverable**: Can have AI conversations and use tools

### Phase 5: Advanced Features (Weeks 29-35)

**Goal**: Feature parity with TypeScript version

- ✅ LSP integration
- ✅ Session summarization
- ✅ Session compaction
- ✅ Todo management
- ✅ Session forking/revert
- ✅ Web fetching

**Deliverable**: Full feature parity

### Phase 6: Polish & Optimization (Weeks 36-40)

**Goal**: Production ready

- ✅ Error handling improvements
- ✅ Performance optimization
- ✅ Testing and edge cases
- ✅ Documentation
- ✅ Migration tools

**Deliverable**: Production-ready pure Go TUI

---

## Benefits of Pure Go TUI

### 1. Single Binary ✅

- No Bun/Node.js required
- Simpler installation
- Smaller download size
- No JavaScript runtime overhead

### 2. Better Resource Usage ✅

- Lower memory footprint
- Better CPU efficiency for backend operations
- Native concurrency (goroutines)

### 3. Easier Deployment ✅

- Cross-compile for all platforms
- Static binary (no dependencies)
- Simpler Docker images

### 4. Development Benefits ✅

- Strong typing throughout
- Better concurrency primitives
- Standard library batteries included
- Consistent toolchain

### 5. Performance Gains ✅

- Faster startup time
- Lower latency API calls
- Better resource management
- Native execution (no JIT)

---

## Trade-offs & Losses

### 1. Development Velocity (Short-term)

- 4-8 months of migration work
- Features frozen during migration
- Risk of bugs during transition
- Testing overhead

### 2. Ecosystem

- Vercel AI SDK is more mature than Go alternatives
- Some TypeScript libraries have no Go equivalent
- Tree-sitter Go bindings are less polished

### 3. Dynamic Features

- Plugin system would need redesign
- No dynamic code loading (or use Go plugins with limitations)

### 4. Existing Code

- Throw away 14,000 lines of working TypeScript
- Re-introduce bugs that were already fixed
- Lose institutional knowledge in current codebase

---

## Effort Comparison

| Approach         | Time         | Engineers | Risk         | Result                       |
| ---------------- | ------------ | --------- | ------------ | ---------------------------- |
| **Keep Current** | 0 months     | 0         | 🟢 None      | Hybrid (TS backend + Go TUI) |
| **Pure Go TUI**  | 4-8 months   | 2-3       | 🟡 Medium    | Single Go binary for TUI     |
| **Full Pure Go** | 36-60 months | 5-10      | 🔴 Very High | Impossible (VS Code)         |

---

## Recommendation

### For TUI-Only Pure Go: ✅ **FEASIBLE & RECOMMENDED**

**Pros**:

- Achievable in reasonable timeframe (4-8 months)
- Single binary distribution for terminal users
- Better performance and resource usage
- Eliminates Bun/Node.js dependency
- TUI is already 70% Go

**Cons**:

- Significant upfront investment
- Feature freeze during migration
- Risk of regressions
- Lose TypeScript AI SDK advantages

### Suggested Approach: **Incremental Migration**

Rather than big-bang rewrite:

1. **Keep current TypeScript backend running**
2. **Start new Go backend in parallel** (`packages/opencode-go`)
3. **Add feature flag** to choose backend
4. **Port one component at a time** (file ops → sessions → AI)
5. **Test incrementally** with both backends
6. **Switch over when ready**

This allows:

- ✅ No downtime
- ✅ Test as you go
- ✅ Fall back if issues arise
- ✅ Learn Go patterns incrementally

### Timeline

- **Month 1-2**: Basic server + file operations
- **Month 3-4**: Session management
- **Month 5-6**: AI provider integration
- **Month 7-8**: Advanced features + polish

### Success Criteria

- [ ] TUI works with pure Go backend
- [ ] Feature parity with TypeScript version
- [ ] Performance equal or better
- [ ] Single binary distribution
- [ ] All tests passing
- [ ] Documentation updated

---

## Conclusion (TUI-Only Focus)

Rewriting the **TypeScript backend to pure Go for TUI operation is FEASIBLE** and offers real benefits:

1. **Scope is manageable**: ~14,000 lines (not 55,000)
2. **TUI is already Go**: 70% of work is done
3. **Timeline is reasonable**: 4-8 months vs 3-5 years
4. **Real benefits**: Single binary, better performance, simpler deployment
5. **Risk is moderate**: Not touching web UIs, VS Code extension, etc.

**Key Decision Points**:

- ✅ **Do this if**: You want simpler distribution and don't need web UIs
- ✅ **Do this if**: You want pure Go for TUI terminal users
- ❌ **Don't do this if**: You need web UI, console, desktop app
- ❌ **Don't do this if**: Current hybrid architecture works well

**Best Path Forward**:

1. Start with **incremental migration**
2. Keep TypeScript backend as fallback
3. Port **file operations first** (easiest)
4. Then **sessions**, then **AI providers** (hardest)
5. Test continuously with feature flags
6. Switch over when confident

The TUI-focused pure Go rewrite is **significantly more achievable** than full rewrite, with **measurable benefits** for terminal users who don't need the web components.

---

## Appendix: Core Components Breakdown

### Must Rewrite (Critical Path)

| Component  | TS Lines | Go Effort  | Status       |
| ---------- | -------- | ---------- | ------------ |
| Server/API | 3,000    | 2-3 weeks  | 🟢 Easy      |
| Sessions   | 4,100    | 6-10 weeks | 🟠 Hard      |
| Providers  | 900      | 8-12 weeks | 🔴 Very Hard |
| Tools      | 2,300    | 4-6 weeks  | 🟡 Medium    |
| LSP        | 800      | 3-5 weeks  | 🟡 Medium    |
| Files      | 600      | 1-2 weeks  | 🟢 Easy      |
| Project    | 1,200    | 2-3 weeks  | 🟢 Easy      |
| Config     | 300      | 1-2 weeks  | 🟢 Easy      |
| Storage    | 400      | 1-2 weeks  | 🟢 Easy      |

### Can Skip (Not Essential for TUI)

- MCP integration (400 lines) - Optional
- Agent system complexities - Simplified version
- Authentication - Simplified for local use
- Snapshot sharing - Can add later

### Already Have (Keep As-Is)

- TUI interface - 40,000 lines Go ✅
- Go SDK - 4,000 lines Go ✅
- Terminal rendering - All done ✅
