package provider

import (
	"context"
	"fmt"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/shared"
)

type OpenRouterProvider struct {
	client            openai.Client
	apiKey            string
	modelsDevRegistry ModelsDevRegistry
}

func NewOpenRouter(apiKey string) *OpenRouterProvider {
	// OpenRouter uses an OpenAI-compatible API at openrouter.ai/api/v1
	client := openai.NewClient(
		option.WithBaseURL("https://openrouter.ai/api/v1"),
		option.WithAPIKey(apiKey),
	)

	return &OpenRouterProvider{
		client: client,
		apiKey: apiKey,
	}
}

func (p *OpenRouterProvider) Name() string {
	return "openrouter"
}

func (p *OpenRouterProvider) Models() []string {
	// Try to get models from models.dev registry first
	if p.modelsDevRegistry != nil {
		models := p.modelsDevRegistry.GetModels("openrouter")
		if len(models) > 0 {
			return models
		}
	}

	// Fallback to hardcoded list
	return []string{
		"anthropic/claude-3.5-sonnet",
		"anthropic/claude-3-opus",
		"anthropic/claude-3-sonnet",
		"anthropic/claude-3-haiku",
		"openai/gpt-4o",
		"openai/gpt-4-turbo",
		"openai/gpt-4",
		"openai/gpt-3.5-turbo",
		"google/gemini-pro-1.5",
		"google/gemini-flash-1.5",
		"meta-llama/llama-3.1-405b-instruct",
		"meta-llama/llama-3.1-70b-instruct",
		"mistralai/mistral-large",
		"cohere/command-r-plus",
	}
}

func (p *OpenRouterProvider) SetModelsDevRegistry(registry ModelsDevRegistry) {
	p.modelsDevRegistry = registry
}

func (p *OpenRouterProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
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

	// Build request
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
		return nil, fmt.Errorf("openrouter API error: %w", err)
	}

	// Extract first choice
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := response.Choices[0]

	// Convert response - reuse OpenAI helpers
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

func (p *OpenRouterProvider) Stream(ctx context.Context, req CompletionRequest) (Stream, error) {
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

	// Build request
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

	// Make streaming API call
	stream := p.client.Chat.Completions.NewStreaming(ctx, params)

	return &OpenAIStream{stream: stream}, nil
}
