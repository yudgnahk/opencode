# Phase 5: Advanced Features - Kickoff

**Start Date**: November 3, 2025  
**Target Duration**: 7 weeks (Weeks 29-35)  
**Team Size**: 2 engineers  
**Complexity**: 🟡 Medium-High

---

## Executive Summary

Phase 5 represents the final push toward **feature parity** with the TypeScript backend. With Phase 4's AI provider integration complete, we now focus on advanced features that enhance the developer experience and enable sophisticated workflows.

**Goal**: Complete feature parity with TypeScript backend

---

## Phase 4 Recap

✅ **Successfully Completed**:

- 3 AI providers integrated (Anthropic, OpenAI, Gemini)
- Streaming responses via SSE
- Tool calling support
- Provider abstraction layer
- Retry logic with exponential backoff
- 68 comprehensive tests
- 540+ lines of documentation

**Key Learnings**:

- Provider abstraction layer works well for multiple APIs
- Streaming implementation requires careful chunk handling
- Retry logic is critical for production reliability
- Comprehensive testing caught many edge cases

---

## Phase 5 Objectives

### 1. LSP Integration (Weeks 29-30) 🎯

**Priority**: High  
**Complexity**: Medium-High

Integrate Language Server Protocol to provide:

- Real-time diagnostics
- Hover information
- Code completions
- Go-to-definition
- Find references

**Why This Matters**: Enables OpenCode to provide intelligent code insights without relying on external tools.

**Technical Approach**:

- Use `go.lsp.dev/jsonrpc2` for JSON-RPC communication
- Support multiple language servers (gopls, typescript-language-server, pylsp, rust-analyzer)
- Manage LSP server lifecycle (start/stop per language)
- Handle asynchronous notifications (diagnostics)

**Success Criteria**:

- [ ] LSP servers can be started/stopped dynamically
- [ ] Diagnostics are retrieved and displayed
- [ ] Hover information works
- [ ] Code completions work
- [ ] Supports at least 4 languages (Go, TypeScript, Python, Rust)

### 2. Session Summarization (Week 31) 📝

**Priority**: Medium  
**Complexity**: Low-Medium

Automatically summarize long conversation sessions to:

- Generate session titles
- Create executive summaries
- Help users understand session context at a glance

**Technical Approach**:

- Use existing AI providers to generate summaries
- Extract conversation text from messages
- Send to AI with summarization prompt
- Update session title with concise summary

**Success Criteria**:

- [ ] Sessions can be summarized on-demand
- [ ] Session titles auto-update with first summary
- [ ] Summaries are concise (< 100 characters for titles)
- [ ] Works with all providers

### 3. Todo Management (Week 32) ✅

**Priority**: High  
**Complexity**: Low

Implement task tracking system within sessions:

- Create todos
- Update todo status (pending/in_progress/completed/cancelled)
- Set priority (high/medium/low)
- Persist todos with session state

**Technical Approach**:

- Store todos as JSON in BoltDB
- Emit events for todo changes
- Integrate with session lifecycle
- Provide API endpoints for CRUD operations

**Success Criteria**:

- [ ] Todos can be created, read, updated, deleted
- [ ] Todo status transitions work correctly
- [ ] Todos persist across sessions
- [ ] Events are emitted for UI updates

### 4. Web Fetching (Week 33) 🌐

**Priority**: Medium-High  
**Complexity**: Medium

Fetch web content and convert to usable formats:

- HTTP GET requests with configurable timeout
- HTML → Markdown conversion
- Support text, markdown, and HTML output formats
- Handle redirects and errors gracefully

**Technical Approach**:

- Use `net/http` for requests
- Use `github.com/JohannesKaufmann/html-to-markdown` for conversion
- Implement timeout handling (max 120 seconds)
- Add User-Agent headers

**Success Criteria**:

- [ ] URLs can be fetched successfully
- [ ] HTML converts to Markdown correctly
- [ ] Timeouts work properly
- [ ] Error handling is robust
- [ ] Headers are preserved

### 5. Task/Sub-Agent Spawning (Week 34) 🤖

**Priority**: High  
**Complexity**: Medium-High

Enable spawning of sub-agents for complex tasks:

- Create new sessions for sub-tasks
- Execute prompts in sub-agent context
- Return results to parent session
- Support different agent types (general, codex, specialized)

**Technical Approach**:

- Reuse existing session management
- Create child sessions with specific configuration
- Execute AI completions in child context
- Extract and return final result

**Success Criteria**:

- [ ] Sub-agents can be spawned from main session
- [ ] Sub-agent results are returned to parent
- [ ] Sub-sessions are tracked properly
- [ ] Different agent types are supported
- [ ] Error handling prevents parent session corruption

### 6. Session Revert & Branching (Week 35) 🔄

**Priority**: Medium  
**Complexity**: Low-Medium

Enable session history management:

- Revert session to any previous message
- Branch from a specific message (fork)
- Delete messages after revert point
- Preserve session state consistency

**Technical Approach**:

- Find message index in session
- Remove subsequent messages
- Delete message data from storage
- Emit revert events
- Branch is alias for Fork (already implemented)

**Success Criteria**:

- [ ] Sessions can be reverted to any message
- [ ] Messages are properly deleted
- [ ] Session state remains consistent
- [ ] Events are emitted for UI updates
- [ ] Branching creates new session from point

### 7. Tree-sitter Integration (Optional) 🌲

**Priority**: Low  
**Complexity**: High

If time permits, add AST parsing:

- Parse code into Abstract Syntax Trees
- Enable structural code analysis
- Support multiple languages
- Provide code structure information

**Technical Approach**:

- Use `github.com/smacker/go-tree-sitter`
- Add language-specific parsers
- Create AST query API
- Integrate with code tools

**Success Criteria**:

- [ ] Code can be parsed into AST
- [ ] Structural queries work
- [ ] Supports major languages (Go, TypeScript, Python)

---

## Implementation Priority

### Week 29-30: LSP Integration

**Focus**: Get intelligent code insights working

**Tasks**:

1. Install LSP dependencies
2. Create LSP client with JSON-RPC
3. Implement server lifecycle management
4. Add diagnostics retrieval
5. Add hover information
6. Add code completions
7. Create language detection system
8. Configure common language servers
9. Write comprehensive tests
10. Test with real LSP servers

**Deliverable**: LSP working for at least 4 languages

### Week 31: Session Summarization

**Focus**: Auto-generate session context

**Tasks**:

1. Create summarization function
2. Extract text from messages
3. Call AI provider with summary prompt
4. Update session titles
5. Add summarization API endpoint
6. Test with various conversation lengths
7. Handle edge cases (empty sessions, very long sessions)

**Deliverable**: Sessions auto-summarize on demand

### Week 32: Todo Management

**Focus**: Task tracking within sessions

**Tasks**:

1. Define Todo data structures
2. Implement CRUD operations
3. Add persistence layer
4. Create API endpoints
5. Implement event emission
6. Add status transitions
7. Test todo lifecycle
8. Integration tests with sessions

**Deliverable**: Full todo system working

### Week 33: Web Fetching

**Focus**: External content retrieval

**Tasks**:

1. Install HTML-to-Markdown library
2. Create WebFetch tool
3. Implement HTTP request handling
4. Add HTML → Markdown conversion
5. Support multiple output formats
6. Add timeout configuration
7. Handle errors and redirects
8. Test with real websites

**Deliverable**: Web pages can be fetched and converted

### Week 34: Task Spawning

**Focus**: Sub-agent execution

**Tasks**:

1. Create Task tool structure
2. Implement sub-session creation
3. Execute prompts in sub-context
4. Extract results
5. Handle errors gracefully
6. Support agent type configuration
7. Test sub-agent workflows
8. Integration tests with AI providers

**Deliverable**: Sub-agents work end-to-end

### Week 35: Session Revert & Polish

**Focus**: History management and cleanup

**Tasks**:

1. Implement session revert
2. Handle message deletion
3. Test state consistency
4. Implement branching (alias Fork)
5. Event emission
6. Edge case testing
7. Code review and refactoring
8. Documentation updates
9. Integration testing
10. Performance testing

**Deliverable**: Session history fully manageable

---

## Dependencies & Prerequisites

### External Go Modules Required

**LSP**:

- `go.lsp.dev/jsonrpc2` - JSON-RPC 2.0 implementation
- `go.lsp.dev/protocol` - LSP protocol types

**Web Fetching**:

- `github.com/JohannesKaufmann/html-to-markdown` - HTML conversion

**Tree-sitter (Optional)**:

- `github.com/smacker/go-tree-sitter` - Tree-sitter bindings
- Language-specific parsers

### System Requirements

**LSP Servers** (must be installed):

- `gopls` (Go)
- `typescript-language-server` (TypeScript/JavaScript)
- `pylsp` (Python)
- `rust-analyzer` (Rust)

**Testing**:

- Real language servers for integration tests
- Web access for fetch tests
- AI provider API keys

---

## Testing Strategy

### Unit Tests

**Coverage Target**: >80% for each component

**Focus Areas**:

- LSP client initialization and lifecycle
- Todo CRUD operations
- Summarization logic
- Web fetch request handling
- Task spawning session creation
- Session revert logic

### Integration Tests

**Real-World Scenarios**:

- LSP with actual language servers
- Summarization with AI providers
- Web fetch with real URLs
- Sub-agent task execution
- Session revert workflows

### Manual Testing

**Via TUI**:

- Interactive LSP hover/completions
- Create and manage todos
- Fetch web pages
- Spawn sub-tasks
- Revert sessions

---

## Risk Management

### High-Risk Areas

#### 1. LSP Integration 🟠

**Risks**:

- Language server crashes
- JSON-RPC protocol issues
- Async notification handling
- Process management complexity

**Mitigation**:

- Robust error handling
- Process lifecycle monitoring
- Timeout handling
- Graceful degradation (work without LSP)

#### 2. Task Spawning 🟡

**Risks**:

- Sub-session state pollution
- Resource leaks from unclosed sessions
- Parent-child relationship confusion
- Error propagation

**Mitigation**:

- Clear session hierarchy
- Automatic cleanup on error
- Resource monitoring
- Comprehensive error handling

#### 3. Web Fetching 🟡

**Risks**:

- Malicious URLs
- Large file downloads
- Timeout issues
- SSL/TLS errors

**Mitigation**:

- Size limits on responses
- Strict timeout enforcement
- SSL verification
- User-Agent identification

---

## Success Metrics

### Functional Metrics

- [ ] LSP provides diagnostics for 4+ languages
- [ ] Session summarization works for all AI providers
- [ ] Todos persist across sessions
- [ ] Web pages fetch and convert correctly
- [ ] Sub-agents execute and return results
- [ ] Sessions revert without data loss

### Performance Metrics

- [ ] LSP response time < 500ms
- [ ] Summarization completes in < 5s
- [ ] Web fetch completes within timeout
- [ ] Sub-agent spawning < 1s overhead
- [ ] Session revert < 100ms

### Quality Metrics

- [ ] > 80% test coverage
- [ ] Zero memory leaks
- [ ] All linters pass
- [ ] Documentation complete

---

## Deliverables

By end of Phase 5, we will have:

1. ✅ **LSP Client**: Working diagnostics, hover, completions
2. ✅ **Summarization**: Auto-generate session context
3. ✅ **Todo System**: Full task tracking
4. ✅ **Web Fetch**: HTML → Markdown conversion
5. ✅ **Task Spawning**: Sub-agent execution
6. ✅ **Session Revert**: History management
7. ✅ **Complete Tests**: Unit + integration + manual
8. ✅ **Documentation**: Updated for all features

---

## What Feature Parity Means

At the end of Phase 5, the Go backend will have:

**✅ All Core Features**:

- AI conversations with multiple providers ✅ (Phase 4)
- File operations (read, write, edit) ✅ (Phase 2)
- Search tools (glob, grep) ✅ (Phase 2)
- Bash execution ✅ (Phase 2)
- Session management ✅ (Phase 3)
- LSP integration 🎯 (Phase 5)
- Todo management 🎯 (Phase 5)
- Web fetching 🎯 (Phase 5)
- Task spawning 🎯 (Phase 5)
- Session revert/branch 🎯 (Phase 5)

**⏭️ Phase 6 Will Add**:

- Production hardening
- Performance optimization
- Complete API endpoint coverage
- Migration tooling
- Release preparation

---

## Communication Plan

### Weekly Updates

Every Monday:

- Progress against weekly goals
- Completed features
- Blockers or risks
- Next week's focus

### Phase Completion Review

End of Week 35:

- Demo all Phase 5 features
- Testing report
- Performance benchmarks
- Go/no-go for Phase 6

---

## Getting Started

### Immediate Next Steps

1. **Review Phase 5 plan** in `purego_plans/phase-5.md`
2. **Set up LSP dependencies** (`go get` commands)
3. **Install language servers** (gopls, typescript-language-server, etc.)
4. **Begin Week 29 tasks** - LSP client implementation
5. **Create feature branches** for each major component

### Development Environment

**Required Tools**:

```bash
# Install LSP servers
go install golang.org/x/tools/gopls@latest
npm install -g typescript-language-server
pip install python-lsp-server
# rust-analyzer from rustup

# Install Go dependencies
cd packages/opencode-go
go get go.lsp.dev/jsonrpc2
go get go.lsp.dev/protocol
go get github.com/JohannesKaufmann/html-to-markdown
```

**Testing Setup**:

```bash
# Ensure AI provider API keys are set
export ANTHROPIC_API_KEY="..."
export OPENAI_API_KEY="..."
export GEMINI_API_KEY="..."

# Run existing tests to verify Phase 4 baseline
go test ./...
```

---

## Questions & Answers

### Q: Why is LSP first priority?

A: LSP integration is foundational for intelligent code assistance. Getting this working early enables testing with real code scenarios throughout the rest of Phase 5.

### Q: Do we need all language servers?

A: Start with 2-3 (Go, TypeScript) for initial implementation. Add others incrementally as time permits.

### Q: What if Phase 5 takes longer than 7 weeks?

A: Features are prioritized. LSP, Todo, and Task spawning are must-haves. Revert and tree-sitter can slide to Phase 6 if needed.

### Q: How does this differ from TypeScript backend?

A: Functionally identical. Implementation differs (Go vs TypeScript), but user experience and API contract remain the same.

---

## Resources

### Documentation

- [LSP Specification](https://microsoft.github.io/language-server-protocol/)
- [go.lsp.dev Documentation](https://pkg.go.dev/go.lsp.dev)
- [Tree-sitter Documentation](https://tree-sitter.github.io/tree-sitter/)

### Code References

- TypeScript backend: `packages/opencode/src/`
- Go backend (current): `packages/opencode-go/internal/`
- Phase 5 plan: `purego_plans/phase-5.md`

### Team Contacts

- Technical lead: TBD
- Phase 4 context: See `packages/opencode-go/docs/PROVIDERS.md`
- Questions: Create issue or discussion

---

## Conclusion

Phase 5 is the **feature parity milestone**. By completing these advanced features, the Go backend will match the TypeScript implementation's capabilities, setting the stage for production release in Phase 6.

**Key Focus**: Quality over speed. Each feature should be thoroughly tested and documented before moving to the next.

**Let's build something great!** 🚀

---

**Approved by**: TBD  
**Last Updated**: November 3, 2025  
**Next Review**: Week 29 (Start of LSP implementation)
