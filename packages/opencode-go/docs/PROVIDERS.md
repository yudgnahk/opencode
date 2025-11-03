# AI Provider Integration

This document describes the AI provider integration in OpenCode Go, including configuration, usage, and implementation details.

## Overview

OpenCode Go supports multiple AI providers through a unified interface:

- **Anthropic** (Claude models)
- **OpenAI** (GPT models)
- **Gemini** (Google Gemini models)

All providers are integrated through a common `Provider` interface, with automatic retry logic, streaming support, and consistent error handling.

## Authentication

OpenCode Go supports multiple authentication methods for AI providers through the `auth` package. Authentication credentials are stored securely in `~/.local/share/opencode/auth.json` with restricted file permissions (0o600).

### Authentication Types

Three authentication types are supported:

#### 1. API Key Authentication

Simple API key for most providers:

```go
auth.Set("openai", auth.API{
    Type: auth.TypeAPI,
    Key:  "sk-your-api-key",
})
```

#### 2. OAuth Authentication

OAuth with refresh/access tokens:

```go
auth.Set("provider", auth.OAuth{
    Type:    auth.TypeOAuth,
    Refresh: "refresh-token",
    Access:  "access-token",
    Expires: 1234567890,
})
```

#### 3. Well-Known Authentication

Authentication with both key and token:

```go
auth.Set("gemini", auth.WellKnown{
    Type:  auth.TypeWellKnown,
    Key:   "project-key",
    Token: "auth-token",
})
```

### Storage Location

Auth credentials are stored in the XDG data directory:

```
~/.local/share/opencode/auth.json
```

Or use `$XDG_DATA_HOME/opencode/auth.json` if `XDG_DATA_HOME` is set.

File format:

```json
{
  "openai": {
    "type": "api",
    "key": "sk-your-api-key"
  },
  "anthropic": {
    "type": "oauth",
    "refresh": "refresh-token",
    "access": "access-token",
    "expires": 1234567890
  },
  "gemini": {
    "type": "wellknown",
    "key": "project-key",
    "token": "auth-token"
  }
}
```

**Security:** The file is created with `0o600` permissions (read/write for owner only) to protect API keys.

### Managing Authentication

#### Get Auth for a Provider

```go
import "github.com/opencode/opencode-go/internal/auth"

// Get authentication for a specific provider
info, err := auth.Get("openai")
if err != nil {
    log.Fatal(err)
}

// Type assert to the appropriate type
if apiAuth, ok := info.(auth.API); ok {
    fmt.Println("API Key:", apiAuth.Key)
}
```

#### Get All Auth Records

```go
allAuth, err := auth.All()
if err != nil {
    log.Fatal(err)
}

for providerID, info := range allAuth {
    fmt.Printf("Provider: %s, Type: %s\n", providerID, info.GetType())
}
```

#### Set Auth for a Provider

```go
// Set API key authentication
apiAuth := auth.API{
    Type: auth.TypeAPI,
    Key:  "sk-your-new-api-key",
}
if err := auth.Set("openai", apiAuth); err != nil {
    log.Fatal(err)
}

// Update existing auth (same function)
newApiAuth := auth.API{
    Type: auth.TypeAPI,
    Key:  "sk-updated-key",
}
if err := auth.Set("openai", newApiAuth); err != nil {
    log.Fatal(err)
}
```

#### Remove Auth for a Provider

```go
if err := auth.Remove("openai"); err != nil {
    log.Fatal(err)
}
```

### Environment Variables (Legacy)

For backward compatibility, providers can still be configured via environment variables:

```bash
export ANTHROPIC_API_KEY="your-anthropic-key"
export OPENAI_API_KEY="your-openai-key"
export GEMINI_API_KEY="your-gemini-key"
```

**Note:** The auth system takes precedence over environment variables when both are present.

## Quick Start

### 1. Configuration

**Option A: Using Auth System (Recommended)**

```go
import "github.com/opencode/opencode-go/internal/auth"

// Set up authentication
auth.Set("anthropic", auth.API{
    Type: auth.TypeAPI,
    Key:  "your-anthropic-key",
})

auth.Set("openai", auth.API{
    Type: auth.TypeAPI,
    Key:  "your-openai-key",
})

auth.Set("gemini", auth.API{
    Type: auth.TypeAPI,
    Key:  "your-gemini-key",
})
```

**Option B: Using Environment Variables**

```bash
# Anthropic (Claude)
export ANTHROPIC_API_KEY="your-anthropic-key"

# OpenAI (GPT)
export OPENAI_API_KEY="your-openai-key"

# Gemini (Google)
export GEMINI_API_KEY="your-gemini-key"
```

### 2. Start the Server

```bash
go run ./cmd/server/main.go
```

Providers with valid API keys will be automatically initialized and registered.

### 3. Use the API

Create a session and send messages:

```bash
# Create a session with Claude
curl -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{"provider": "anthropic", "model": "claude-3-5-sonnet-20241022"}'

# Send a message (non-streaming)
curl -X POST http://localhost:8080/api/sessions/{id}/complete \
  -H "Content-Type: application/json" \
  -d '{"content": "Hello, how are you?"}'

# Send a message (streaming)
curl -X POST http://localhost:8080/api/sessions/{id}/complete/stream \
  -H "Content-Type: application/json" \
  -d '{"content": "Tell me a story"}' \
  --no-buffer
```

## Provider Details

### Anthropic (Claude)

**Provider Name:** `anthropic`

**Available Models:**

- `claude-3-5-sonnet-20241022` (recommended)
- `claude-3-5-haiku-20241022`
- `claude-3-opus-20240229`
- `claude-3-sonnet-20240229`
- `claude-3-haiku-20240307`

**Environment Variables:**

- `ANTHROPIC_API_KEY` - Your Anthropic API key (required)

**Features:**

- Streaming support
- Tool/function calling
- System prompts
- Multi-modal (text + images)
- Token usage tracking

**Example:**

```bash
curl -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "anthropic",
    "model": "claude-3-5-sonnet-20241022",
    "name": "My Claude Session"
  }'
```

### OpenAI (GPT)

**Provider Name:** `openai`

**Available Models:**

- `gpt-4-turbo-preview`
- `gpt-4-turbo`
- `gpt-4`
- `gpt-4-32k`
- `gpt-3.5-turbo`

**Environment Variables:**

- `OPENAI_API_KEY` - Your OpenAI API key (required)

**Features:**

- Streaming support
- Function calling
- System messages
- Token usage tracking

**Example:**

```bash
curl -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "openai",
    "model": "gpt-4-turbo",
    "name": "My GPT Session"
  }'
```

### Gemini (Google)

**Provider Name:** `gemini`

**Available Models:**

- `gemini-1.5-pro`
- `gemini-1.5-flash`
- `gemini-1.0-pro`

**Environment Variables:**

- `GEMINI_API_KEY` - Your Google AI API key (required)

**Features:**

- Streaming support
- Function calling
- Multi-modal (text + images)
- Token usage tracking

**Example:**

```bash
curl -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "gemini",
    "model": "gemini-1.5-pro",
    "name": "My Gemini Session"
  }'
```

## API Endpoints

### POST /api/sessions/{id}/complete

Send a message and get a complete AI response (non-streaming).

**Request:**

```json
{
  "content": "Your message here"
}
```

**Response:**

```json
{
  "id": "msg_abc123",
  "sessionId": "sess_xyz789",
  "role": "assistant",
  "content": [
    {
      "type": "text",
      "text": "AI response here"
    }
  ],
  "createdAt": "2024-01-01T00:00:00Z",
  "metadata": {
    "usage": {
      "inputTokens": 100,
      "outputTokens": 50
    },
    "stopReason": "end_turn"
  }
}
```

**Example:**

```bash
curl -X POST http://localhost:8080/api/sessions/sess_xyz789/complete \
  -H "Content-Type: application/json" \
  -d '{"content": "Explain quantum computing"}'
```

### POST /api/sessions/{id}/complete/stream

Send a message and stream the AI response using Server-Sent Events (SSE).

**Request:**

```json
{
  "content": "Your message here"
}
```

**Response:** SSE stream with events:

```
event: chunk
data: {"type":"text","delta":"Hello"}

event: chunk
data: {"type":"text","delta":" world"}

event: chunk
data: {"type":"stop","stopReason":"end_turn"}
```

**Chunk Types:**

- `text` - Text delta
- `tool_call` - Function/tool call
- `stop` - End of stream

**Example:**

```bash
curl -X POST http://localhost:8080/api/sessions/sess_xyz789/complete/stream \
  -H "Content-Type: application/json" \
  -d '{"content": "Write a poem"}' \
  --no-buffer
```

**JavaScript Example:**

```javascript
const eventSource = new EventSource("/api/sessions/sess_xyz789/complete/stream")

eventSource.addEventListener("chunk", (event) => {
  const chunk = JSON.parse(event.data)

  if (chunk.type === "text") {
    console.log(chunk.delta)
  } else if (chunk.type === "stop") {
    console.log("Stream ended:", chunk.stopReason)
    eventSource.close()
  }
})

eventSource.addEventListener("error", (event) => {
  const error = JSON.parse(event.data)
  console.error("Error:", error.error)
  eventSource.close()
})
```

## Advanced Usage

### Tool/Function Calling

The provider interface supports tool calling for all providers:

```go
tools := []provider.Tool{
    {
        Name:        "get_weather",
        Description: "Get current weather for a location",
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "location": map[string]interface{}{
                    "type":        "string",
                    "description": "City name",
                },
            },
            "required": []string{"location"},
        },
    },
}

req := provider.CompletionRequest{
    Messages:    messages,
    Model:       "claude-3-5-sonnet-20241022",
    Tools:       tools,
    MaxTokens:   4096,
    Temperature: 0.7,
}

resp, err := prov.Complete(ctx, req)
```

### Retry Logic

All providers are automatically wrapped with retry logic:

- **Max retries:** 3 attempts
- **Initial delay:** 1 second
- **Max delay:** 10 seconds
- **Backoff multiplier:** 2x (exponential backoff)

Retryable errors:

- Rate limit errors (429)
- Timeout errors
- Connection errors
- Server errors (500+)

### Token Usage Tracking

All responses include token usage information:

```go
resp, err := prov.Complete(ctx, req)
if err != nil {
    return err
}

fmt.Printf("Input tokens: %d\n", resp.Usage.InputTokens)
fmt.Printf("Output tokens: %d\n", resp.Usage.OutputTokens)
fmt.Printf("Total tokens: %d\n",
    resp.Usage.InputTokens + resp.Usage.OutputTokens)
```

### Custom Temperature and Max Tokens

```bash
# In your completion request, these are set by the CompletionService
# Default values:
# - MaxTokens: 4096
# - Temperature: 0.7
```

## Provider Interface

All providers implement the `Provider` interface:

```go
type Provider interface {
    // Complete generates a non-streaming completion
    Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)

    // Stream generates a streaming completion
    Stream(ctx context.Context, req CompletionRequest) (Stream, error)

    // Name returns provider name
    Name() string

    // Models returns available models
    Models() []string
}
```

### Unified Message Format

The provider interface uses a unified message format:

```go
type Message struct {
    Role    string         `json:"role"`    // "user", "assistant", "system"
    Content []ContentBlock `json:"content"`
}

type ContentBlock struct {
    Type       string      `json:"type"` // "text", "tool_use", "tool_result"
    Text       string      `json:"text,omitempty"`
    ToolUse    *ToolUse    `json:"toolUse,omitempty"`
    ToolResult *ToolResult `json:"toolResult,omitempty"`
}
```

Each provider converts this format to/from its native SDK format automatically.

## Error Handling

### Common Errors

**Provider not found:**

```json
{
  "error": "provider anthropic not found"
}
```

_Solution:_ Ensure `ANTHROPIC_API_KEY` environment variable is set.

**Invalid API key:**

```json
{
  "error": "completion failed: authentication error"
}
```

_Solution:_ Check that your API key is valid and has the necessary permissions.

**Rate limit exceeded:**

```json
{
  "error": "completion failed: rate limit exceeded"
}
```

_Solution:_ The retry logic will automatically retry with exponential backoff. If this persists, wait or upgrade your API plan.

**Model not available:**

```json
{
  "error": "model not found"
}
```

_Solution:_ Use one of the models listed in the provider's `Models()` method.

### Error Response Format

All errors return a consistent format:

```json
{
  "error": "detailed error message",
  "code": "error_code",
  "details": {}
}
```

## Testing

### Unit Tests

Provider implementations include comprehensive unit tests:

```bash
# Run all provider tests
go test ./internal/provider/...

# Run specific provider tests
go test ./internal/provider -run TestOpenAI
go test ./internal/provider -run TestGemini
go test ./internal/provider -run TestAnthropic

# Run with coverage
go test ./internal/provider/... -cover
```

### Integration Tests

Test the complete flow with mock providers:

```bash
# Run integration tests
go test ./internal/session -run TestIntegration
```

### Manual Testing

Use the provided SSE client for testing streaming:

```go
import "github.com/opencode/opencode-go/test"

client := test.NewSSEClient("http://localhost:8080")
events, err := client.Connect("/api/sessions/{id}/complete/stream")
```

## Performance Considerations

### Request Timeouts

Default timeouts are set at the HTTP server level:

- **Read timeout:** 15 seconds
- **Write timeout:** 15 seconds (increased for streaming)
- **Idle timeout:** 60 seconds

For long-running completions, use streaming to avoid timeouts.

### Concurrent Requests

The provider registry is thread-safe and supports concurrent requests:

```go
// Safe to call from multiple goroutines
for i := 0; i < 10; i++ {
    go func() {
        prov, _ := registry.Get("anthropic")
        resp, _ := prov.Complete(ctx, req)
        // ...
    }()
}
```

### Memory Usage

- Non-streaming completions buffer the entire response in memory
- Streaming completions process chunks incrementally
- For large responses (>10KB), prefer streaming

## Best Practices

### 1. Choose the Right Provider

- **Anthropic (Claude):** Best for reasoning, analysis, long-form content
- **OpenAI (GPT):** Best for general tasks, well-documented, wide adoption
- **Gemini:** Best for multi-modal tasks, cost-effective

### 2. Use Streaming for Long Responses

```bash
# Instead of this (waits for complete response):
curl -X POST .../complete -d '{"content": "Write a long article"}'

# Use this (streams chunks as they arrive):
curl -X POST .../complete/stream -d '{"content": "Write a long article"}'
```

### 3. Handle Errors Gracefully

The retry logic handles transient errors, but you should still handle permanent failures:

```go
resp, err := prov.Complete(ctx, req)
if err != nil {
    if isAuthError(err) {
        // API key issue - notify admin
    } else if isRateLimitError(err) {
        // Rate limit - backoff and retry later
    }
    return err
}
```

### 4. Monitor Token Usage

Track token usage to estimate costs:

```go
totalInputTokens += resp.Usage.InputTokens
totalOutputTokens += resp.Usage.OutputTokens

// Log or store for billing/analytics
```

### 5. Use Appropriate Models

- Start with smaller models (e.g., `claude-3-5-haiku-20241022`, `gpt-3.5-turbo`)
- Upgrade to larger models only when needed
- Test performance vs. cost trade-offs

## Troubleshooting

### Provider not initializing

**Problem:** Provider doesn't show up in registry

**Solution:**

1. Check environment variable is set: `echo $ANTHROPIC_API_KEY`
2. Check server logs for initialization errors
3. Verify API key format (no extra spaces/newlines)

### Streaming not working

**Problem:** SSE stream not receiving events

**Solution:**

1. Ensure client supports SSE (EventSource API)
2. Use `--no-buffer` with curl
3. Check proxy/CDN settings (some buffer SSE)
4. Verify `Content-Type: text/event-stream` header

### High latency

**Problem:** Completions taking too long

**Solution:**

1. Use streaming to see partial results faster
2. Reduce `maxTokens` to limit response length
3. Use smaller/faster models
4. Check network latency to provider API

### Rate limiting

**Problem:** Getting rate limit errors frequently

**Solution:**

1. Retry logic automatically handles this
2. Add client-side rate limiting/queuing
3. Upgrade API plan for higher limits
4. Use multiple API keys with load balancing

## Migration Guide

### From TypeScript Backend

The Go provider interface is similar to the TypeScript version:

**TypeScript:**

```typescript
const response = await provider.complete({
  messages: [...],
  model: 'claude-3-5-sonnet-20241022',
  maxTokens: 4096
});
```

**Go:**

```go
resp, err := provider.Complete(ctx, provider.CompletionRequest{
  Messages:  []provider.Message{...},
  Model:     "claude-3-5-sonnet-20241022",
  MaxTokens: 4096,
})
```

Key differences:

- Go uses context.Context for cancellation
- Go uses explicit error returns (no try/catch)
- Go uses struct tags for JSON serialization

## Future Enhancements

Planned features:

- [ ] Provider health check endpoint
- [ ] Model listing endpoint (`GET /api/providers`)
- [ ] Cost estimation per request
- [ ] Response caching
- [ ] Request/response logging
- [ ] Custom retry configurations per provider
- [ ] Provider-specific configuration (base URL, timeout)

## Support

For issues or questions:

- GitHub Issues: https://github.com/sst/opencode/issues
- Documentation: https://opencode.ai/docs
- Provider-specific docs:
  - Anthropic: https://docs.anthropic.com
  - OpenAI: https://platform.openai.com/docs
  - Gemini: https://ai.google.dev/docs
