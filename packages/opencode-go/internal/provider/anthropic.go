package provider

import (
	"context"
	"fmt"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/ssestream"
)

type AnthropicProvider struct {
	client anthropic.Client
	apiKey string
}

func NewAnthropic(apiKey string) *AnthropicProvider {
	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &AnthropicProvider{
		client: client,
		apiKey: apiKey,
	}
}

func (p *AnthropicProvider) Name() string {
	return "anthropic"
}

func (p *AnthropicProvider) Models() []string {
	return []string{
		"claude-3-5-sonnet-20241022",
		"claude-3-5-sonnet-20240620",
		"claude-3-opus-20240229",
		"claude-3-sonnet-20240229",
		"claude-3-haiku-20240307",
	}
}

func (p *AnthropicProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	// Convert to Anthropic format
	messages := make([]anthropic.MessageParam, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = convertMessageToAnthropic(msg)
	}

	// Build request
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(req.Model),
		Messages:  messages,
		MaxTokens: int64(req.MaxTokens),
	}

	if req.Temperature > 0 {
		params.Temperature = anthropic.Float(req.Temperature)
	}

	if req.SystemPrompt != "" {
		params.System = []anthropic.TextBlockParam{
			{Text: req.SystemPrompt},
		}
	}

	if len(req.Tools) > 0 {
		params.Tools = convertToolsToAnthropic(req.Tools)
	}

	if len(req.StopSequences) > 0 {
		params.StopSequences = req.StopSequences
	}

	// Make API call
	response, err := p.client.Messages.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("anthropic API error: %w", err)
	}

	// Convert response
	return &CompletionResponse{
		ID:         response.ID,
		Content:    convertContentFromAnthropic(response.Content),
		StopReason: string(response.StopReason),
		Usage: Usage{
			InputTokens:  int(response.Usage.InputTokens),
			OutputTokens: int(response.Usage.OutputTokens),
		},
		ToolCalls: extractToolCalls(response.Content),
	}, nil
}

func (p *AnthropicProvider) Stream(ctx context.Context, req CompletionRequest) (Stream, error) {
	// Convert request (same as Complete)
	messages := make([]anthropic.MessageParam, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = convertMessageToAnthropic(msg)
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(req.Model),
		Messages:  messages,
		MaxTokens: int64(req.MaxTokens),
	}

	if req.Temperature > 0 {
		params.Temperature = anthropic.Float(req.Temperature)
	}

	if req.SystemPrompt != "" {
		params.System = []anthropic.TextBlockParam{
			{Text: req.SystemPrompt},
		}
	}

	if len(req.Tools) > 0 {
		params.Tools = convertToolsToAnthropic(req.Tools)
	}

	if len(req.StopSequences) > 0 {
		params.StopSequences = req.StopSequences
	}

	// Create streaming request
	stream := p.client.Messages.NewStreaming(ctx, params)

	return &AnthropicStream{stream: stream}, nil
}

// AnthropicStream wraps the Anthropic streaming response
type AnthropicStream struct {
	stream *ssestream.Stream[anthropic.MessageStreamEventUnion]
	chunk  StreamChunk
}

func (s *AnthropicStream) Next() bool {
	if !s.stream.Next() {
		return false
	}

	event := s.stream.Current()

	switch eventVariant := event.AsAny().(type) {
	case anthropic.ContentBlockDeltaEvent:
		switch deltaVariant := eventVariant.Delta.AsAny().(type) {
		case anthropic.TextDelta:
			s.chunk = StreamChunk{
				Type:  ChunkTypeText,
				Delta: deltaVariant.Text,
			}
		}

	case anthropic.MessageStopEvent:
		s.chunk = StreamChunk{
			Type: ChunkTypeStop,
		}

	case anthropic.MessageDeltaEvent:
		if eventVariant.Delta.StopReason != "" {
			s.chunk = StreamChunk{
				Type:       ChunkTypeStop,
				StopReason: string(eventVariant.Delta.StopReason),
			}
		}

	case anthropic.ContentBlockStartEvent:
		switch blockVariant := eventVariant.ContentBlock.AsAny().(type) {
		case anthropic.ToolUseBlock:
			s.chunk = StreamChunk{
				Type: ChunkTypeToolCall,
				ToolCall: &ToolCall{
					ID:   blockVariant.ID,
					Name: blockVariant.Name,
				},
			}
		}
	}

	return true
}

func (s *AnthropicStream) Chunk() StreamChunk {
	return s.chunk
}

func (s *AnthropicStream) Error() error {
	return s.stream.Err()
}

func (s *AnthropicStream) Close() error {
	s.stream.Close()
	return nil
}

// Helper functions for conversion
func convertMessageToAnthropic(msg Message) anthropic.MessageParam {
	blocks := make([]anthropic.ContentBlockParamUnion, 0, len(msg.Content))

	for _, block := range msg.Content {
		switch block.Type {
		case "text":
			blocks = append(blocks, anthropic.NewTextBlock(block.Text))
		case "tool_use":
			if block.ToolUse != nil {
				blocks = append(blocks, anthropic.NewToolUseBlock(
					block.ToolUse.ID,
					block.ToolUse.Input,
					block.ToolUse.Name,
				))
			}
		case "tool_result":
			if block.ToolResult != nil {
				// Convert content to string
				contentStr := fmt.Sprintf("%v", block.ToolResult.Content)
				if str, ok := block.ToolResult.Content.(string); ok {
					contentStr = str
				}
				blocks = append(blocks, anthropic.NewToolResultBlock(
					block.ToolResult.ToolUseID,
					contentStr,
					false, // isError
				))
			}
		}
	}

	// Use appropriate constructor based on role
	if msg.Role == "user" {
		return anthropic.NewUserMessage(blocks...)
	}
	return anthropic.NewAssistantMessage(blocks...)
}

func convertContentFromAnthropic(content []anthropic.ContentBlockUnion) []ContentBlock {
	result := make([]ContentBlock, 0, len(content))
	for _, block := range content {
		switch b := block.AsAny().(type) {
		case anthropic.TextBlock:
			result = append(result, ContentBlock{
				Type: "text",
				Text: b.Text,
			})
		case anthropic.ToolUseBlock:
			result = append(result, ContentBlock{
				Type: "tool_use",
				ToolUse: &ToolUse{
					ID:    b.ID,
					Name:  b.Name,
					Input: b.Input,
				},
			})
		}
	}
	return result
}

func convertToolsToAnthropic(tools []Tool) []anthropic.ToolUnionParam {
	result := make([]anthropic.ToolUnionParam, len(tools))
	for i, tool := range tools {
		// Convert InputSchema interface{} to ToolInputSchemaParam
		inputSchema := anthropic.ToolInputSchemaParam{}
		if schema, ok := tool.InputSchema.(map[string]interface{}); ok {
			if props, ok := schema["properties"]; ok {
				inputSchema.Properties = props
			}
			if req, ok := schema["required"].([]string); ok {
				inputSchema.Required = req
			}
		}

		result[i] = anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Name,
				Description: anthropic.String(tool.Description),
				InputSchema: inputSchema,
			},
		}
	}
	return result
}

func extractToolCalls(content []anthropic.ContentBlockUnion) []ToolCall {
	calls := make([]ToolCall, 0)
	for _, block := range content {
		if b, ok := block.AsAny().(anthropic.ToolUseBlock); ok {
			calls = append(calls, ToolCall{
				ID:        b.ID,
				Name:      b.Name,
				Arguments: b.Input,
			})
		}
	}
	return calls
}
