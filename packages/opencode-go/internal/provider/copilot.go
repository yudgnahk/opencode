package provider

import (
	"context"
	"fmt"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/shared"
)

type GitHubCopilotProvider struct {
	client            openai.Client
	accessToken       string
	modelsDevRegistry ModelsDevRegistry
}

func NewGitHubCopilot(accessToken string) *GitHubCopilotProvider {
	// Parse access token to extract proxy endpoint
	// Access token format: tid=...;ol=...;exp=...;sku=...;proxy-ep=proxy.business.githubcopilot.com;...
	// We need to extract the proxy-ep value
	proxyEndpoint := "https://api.githubcopilot.com"

	// Parse access token for proxy endpoint
	// Simple parsing - in production you'd want more robust parsing
	if len(accessToken) > 0 {
		// Look for proxy-ep= in the access token
		start := 0
		for i := 0; i < len(accessToken)-9; i++ {
			if accessToken[i:i+9] == "proxy-ep=" {
				start = i + 9
				end := start
				for end < len(accessToken) && accessToken[end] != ';' {
					end++
				}
				if end > start {
					proxyEndpoint = "https://" + accessToken[start:end]
				}
				break
			}
		}
	}

	client := openai.NewClient(
		option.WithBaseURL(proxyEndpoint+"/v1"),
		option.WithAPIKey(accessToken),
	)

	return &GitHubCopilotProvider{
		client:      client,
		accessToken: accessToken,
	}
}

func (p *GitHubCopilotProvider) Name() string {
	return "github-copilot"
}

func (p *GitHubCopilotProvider) Models() []string {
	// Try to get models from models.dev registry first
	if p.modelsDevRegistry != nil {
		models := p.modelsDevRegistry.GetModels("github-copilot")
		if len(models) > 0 {
			return models
		}
	}

	// Fallback to hardcoded list
	return []string{
		"gpt-4o",
		"gpt-4o-2024-08-06",
		"gpt-4",
		"gpt-4-turbo",
		"gpt-3.5-turbo",
		"o1-preview",
		"o1-mini",
		"claude-3.5-sonnet",
	}
}

func (p *GitHubCopilotProvider) SetModelsDevRegistry(registry ModelsDevRegistry) {
	p.modelsDevRegistry = registry
}

func (p *GitHubCopilotProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
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
		return nil, fmt.Errorf("github copilot API error: %w", err)
	}

	// Extract first choice
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := response.Choices[0]

	// Convert response
	content := []ContentBlock{}
	if choice.Message.Content != "" {
		content = append(content, ContentBlock{
			Type: "text",
			Text: choice.Message.Content,
		})
	}

	// Handle tool calls
	var toolCalls []ToolCall
	if len(choice.Message.ToolCalls) > 0 {
		for _, tc := range choice.Message.ToolCalls {
			toolCalls = append(toolCalls, ToolCall{
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments,
			})
		}
	}

	return &CompletionResponse{
		ID:         response.ID,
		Content:    content,
		StopReason: string(choice.FinishReason),
		Usage: Usage{
			InputTokens:  int(response.Usage.PromptTokens),
			OutputTokens: int(response.Usage.CompletionTokens),
		},
		ToolCalls: toolCalls,
	}, nil
}

func (p *GitHubCopilotProvider) Stream(ctx context.Context, req CompletionRequest) (Stream, error) {
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
