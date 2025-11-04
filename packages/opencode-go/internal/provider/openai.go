package provider

import (
	"context"
	"encoding/json"
	"fmt"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/packages/ssestream"
	"github.com/openai/openai-go/shared"
)

type OpenAIProvider struct {
	client            openai.Client
	apiKey            string
	modelsDevRegistry ModelsDevRegistry
}

func NewOpenAI(apiKey string) *OpenAIProvider {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &OpenAIProvider{
		client: client,
		apiKey: apiKey,
	}
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}

func (p *OpenAIProvider) Models() []string {
	// Try to get models from models.dev registry first
	if p.modelsDevRegistry != nil {
		models := p.modelsDevRegistry.GetModels("openai")
		if len(models) > 0 {
			return models
		}
	}

	// Fallback to hardcoded list
	return []string{
		"gpt-4o",
		"gpt-4o-2024-11-20",
		"gpt-4o-2024-08-06",
		"gpt-4o-2024-05-13",
		"gpt-4o-mini",
		"gpt-4o-mini-2024-07-18",
		"gpt-4-turbo",
		"gpt-4-turbo-2024-04-09",
		"gpt-4",
		"gpt-3.5-turbo",
	}
}

func (p *OpenAIProvider) SetModelsDevRegistry(registry ModelsDevRegistry) {
	p.modelsDevRegistry = registry
}

func (p *OpenAIProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	// Convert to OpenAI format
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(req.Messages)+1)

	// Add system message if provided
	if req.SystemPrompt != "" {
		messages = append(messages, openai.SystemMessage(req.SystemPrompt))
	}

	// Convert request messages
	for _, msg := range req.Messages {
		messages = append(messages, convertMessageToOpenAI(msg))
	}

	// Build request - use model from request
	params := openai.ChatCompletionNewParams{
		Model:    shared.ChatModel(req.Model),
		Messages: messages,
	}

	if req.MaxTokens > 0 {
		params.MaxTokens = param.NewOpt(int64(req.MaxTokens))
	}

	if req.Temperature > 0 {
		params.Temperature = param.NewOpt(req.Temperature)
	}

	if len(req.Tools) > 0 {
		params.Tools = convertToolsToOpenAI(req.Tools)
	}

	if len(req.StopSequences) > 0 {
		params.Stop = openai.ChatCompletionNewParamsStopUnion{
			OfStringArray: req.StopSequences,
		}
	}

	// Make API call
	response, err := p.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("openai API error: %w", err)
	}

	// Extract first choice
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := response.Choices[0]

	// Convert response
	return &CompletionResponse{
		ID:         response.ID,
		Content:    convertContentFromOpenAI(choice.Message),
		StopReason: string(choice.FinishReason),
		Usage: Usage{
			InputTokens:  int(response.Usage.PromptTokens),
			OutputTokens: int(response.Usage.CompletionTokens),
		},
		ToolCalls: extractToolCallsFromOpenAI(choice.Message.ToolCalls),
	}, nil
}

func (p *OpenAIProvider) Stream(ctx context.Context, req CompletionRequest) (Stream, error) {
	// Convert request (same as Complete)
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(req.Messages)+1)

	if req.SystemPrompt != "" {
		messages = append(messages, openai.SystemMessage(req.SystemPrompt))
	}

	for _, msg := range req.Messages {
		messages = append(messages, convertMessageToOpenAI(msg))
	}

	params := openai.ChatCompletionNewParams{
		Model:    shared.ChatModel(req.Model),
		Messages: messages,
	}

	if req.MaxTokens > 0 {
		params.MaxTokens = param.NewOpt(int64(req.MaxTokens))
	}

	if req.Temperature > 0 {
		params.Temperature = param.NewOpt(req.Temperature)
	}

	if len(req.Tools) > 0 {
		params.Tools = convertToolsToOpenAI(req.Tools)
	}

	if len(req.StopSequences) > 0 {
		params.Stop = openai.ChatCompletionNewParamsStopUnion{
			OfStringArray: req.StopSequences,
		}
	}

	// Create streaming request
	stream := p.client.Chat.Completions.NewStreaming(ctx, params)

	return &OpenAIStream{stream: stream}, nil
}

// OpenAIStream wraps the OpenAI streaming response
type OpenAIStream struct {
	stream *ssestream.Stream[openai.ChatCompletionChunk]
	chunk  StreamChunk
}

func (s *OpenAIStream) Next() bool {
	if !s.stream.Next() {
		return false
	}

	chunk := s.stream.Current()

	// OpenAI streams deltas in chunks
	if len(chunk.Choices) > 0 {
		delta := chunk.Choices[0].Delta

		// Text delta
		if delta.Content != "" {
			s.chunk = StreamChunk{
				Type:  ChunkTypeText,
				Delta: delta.Content,
			}
			return true
		}

		// Tool call delta
		if len(delta.ToolCalls) > 0 {
			tc := delta.ToolCalls[0]
			s.chunk = StreamChunk{
				Type: ChunkTypeToolCall,
				ToolCall: &ToolCall{
					ID:   tc.ID,
					Name: tc.Function.Name,
				},
			}
			return true
		}

		// Stop reason
		if chunk.Choices[0].FinishReason != "" {
			s.chunk = StreamChunk{
				Type:       ChunkTypeStop,
				StopReason: string(chunk.Choices[0].FinishReason),
			}
			return true
		}
	}

	// Continue to next chunk if this one had no meaningful data
	return s.Next()
}

func (s *OpenAIStream) Chunk() StreamChunk {
	return s.chunk
}

func (s *OpenAIStream) Error() error {
	return s.stream.Err()
}

func (s *OpenAIStream) Close() error {
	s.stream.Close()
	return nil
}

// Helper functions for conversion
func convertMessageToOpenAI(msg Message) openai.ChatCompletionMessageParamUnion {
	// Handle tool results
	if msg.Role == "tool" {
		for _, block := range msg.Content {
			if block.Type == "tool_result" && block.ToolResult != nil {
				contentStr := fmt.Sprintf("%v", block.ToolResult.Content)
				if str, ok := block.ToolResult.Content.(string); ok {
					contentStr = str
				}
				return openai.ToolMessage(contentStr, block.ToolResult.ToolUseID)
			}
		}
	}

	// Handle assistant messages with tool calls
	if msg.Role == "assistant" {
		var textContent string
		var toolCalls []openai.ChatCompletionMessageToolCallParam

		for _, block := range msg.Content {
			switch block.Type {
			case "text":
				textContent = block.Text
			case "tool_use":
				if block.ToolUse != nil {
					args, _ := json.Marshal(block.ToolUse.Input)
					toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCallParam{
						ID: block.ToolUse.ID,
						Function: openai.ChatCompletionMessageToolCallFunctionParam{
							Name:      block.ToolUse.Name,
							Arguments: string(args),
						},
					})
				}
			}
		}

		// Create assistant message with tool calls if present
		if len(toolCalls) > 0 {
			return openai.ChatCompletionMessageParamUnion{
				OfAssistant: &openai.ChatCompletionAssistantMessageParam{
					Content: openai.ChatCompletionAssistantMessageParamContentUnion{
						OfString: param.NewOpt(textContent),
					},
					ToolCalls: toolCalls,
				},
			}
		}
		return openai.AssistantMessage(textContent)
	}

	// Handle user messages
	content := ""
	for _, block := range msg.Content {
		if block.Type == "text" {
			content = block.Text
			break
		}
	}

	return openai.UserMessage(content)
}

func convertContentFromOpenAI(msg openai.ChatCompletionMessage) []ContentBlock {
	result := make([]ContentBlock, 0)

	// Add text content
	if msg.Content != "" {
		result = append(result, ContentBlock{
			Type: "text",
			Text: msg.Content,
		})
	}

	// Add tool calls
	for _, tc := range msg.ToolCalls {
		var input interface{}
		_ = json.Unmarshal([]byte(tc.Function.Arguments), &input)

		result = append(result, ContentBlock{
			Type: "tool_use",
			ToolUse: &ToolUse{
				ID:    tc.ID,
				Name:  tc.Function.Name,
				Input: input,
			},
		})
	}

	return result
}

func convertToolsToOpenAI(tools []Tool) []openai.ChatCompletionToolParam {
	result := make([]openai.ChatCompletionToolParam, len(tools))
	for i, tool := range tools {
		// Convert interface{} to map[string]any for FunctionParameters
		params := shared.FunctionParameters{}
		if schemaMap, ok := tool.InputSchema.(map[string]interface{}); ok {
			for k, v := range schemaMap {
				params[k] = v
			}
		}

		result[i] = openai.ChatCompletionToolParam{
			Function: shared.FunctionDefinitionParam{
				Name:        tool.Name,
				Description: param.NewOpt(tool.Description),
				Parameters:  params,
			},
		}
	}
	return result
}

func extractToolCallsFromOpenAI(toolCalls []openai.ChatCompletionMessageToolCall) []ToolCall {
	calls := make([]ToolCall, len(toolCalls))
	for i, tc := range toolCalls {
		var args interface{}
		_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)

		calls[i] = ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: args,
		}
	}
	return calls
}
