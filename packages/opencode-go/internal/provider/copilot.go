package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/shared"
	"github.com/opencode/opencode-go/internal/auth"
)

const (
	copilotTokenURL = "https://api.github.com/copilot_internal/v2/token"
)

type GitHubCopilotProvider struct {
	client            openai.Client
	accessToken       string
	modelsDevRegistry ModelsDevRegistry
	providerID        string
}

func NewGitHubCopilot(accessToken string) *GitHubCopilotProvider {
	return NewGitHubCopilotWithProviderID(accessToken, "github-copilot")
}

func NewGitHubCopilotWithProviderID(accessToken string, providerID string) *GitHubCopilotProvider {
	// Use the standard GitHub Copilot API endpoint
	// This matches the TypeScript implementation which gets this URL from models.dev
	// IMPORTANT: GitHub Copilot uses /chat/completions WITHOUT /v1 prefix
	// But the OpenAI Go SDK automatically appends /v1 to the baseURL!
	// So we set baseURL to the root to counteract this
	baseURL := "https://api.githubcopilot.com"

	// GitHub Copilot requires specific headers for IDE authentication
	// These headers must match what the TypeScript plugin uses
	// Reference: opencode-copilot-auth@0.0.3/index.mjs
	middleware := option.WithMiddleware(func(req *http.Request, next option.MiddlewareNext) (*http.Response, error) {
		// Add required headers to match the TypeScript implementation
		req.Header.Set("User-Agent", "GitHubCopilotChat/0.31.2")
		req.Header.Set("Editor-Version", "vscode/1.104.1")
		req.Header.Set("Editor-Plugin-Version", "copilot-chat/0.31.2")
		req.Header.Set("Copilot-Integration-Id", "vscode-chat")
		req.Header.Set("Openai-Intent", "conversation-edits")

		// Call the next middleware/handler
		return next(req)
	})

	client := openai.NewClient(
		option.WithBaseURL(baseURL),
		option.WithAPIKey(accessToken),
		middleware,
	)

	return &GitHubCopilotProvider{
		client:      client,
		accessToken: accessToken,
		providerID:  providerID,
	}
}

// tokenResponse represents the response from GitHub Copilot token endpoint
type tokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

// ensureValidToken checks if the current token is valid and refreshes it if needed
func (p *GitHubCopilotProvider) ensureValidToken(ctx context.Context) error {
	// Get current auth info
	authInfo, err := auth.Get(p.providerID)
	if err != nil {
		return fmt.Errorf("failed to get auth info: %w", err)
	}

	if authInfo == nil {
		return fmt.Errorf("no auth info found for provider %s", p.providerID)
	}

	oauth, ok := authInfo.(auth.OAuth)
	if !ok {
		return fmt.Errorf("auth info is not OAuth type for provider %s", p.providerID)
	}

	// Check if token needs refresh (expires is in milliseconds)
	now := time.Now().UnixMilli()
	if oauth.Access != "" && oauth.Expires > now {
		// Token is still valid, update in memory if needed
		if p.accessToken != oauth.Access {
			p.accessToken = oauth.Access
			p.updateClientToken(oauth.Access)
		}
		return nil
	}

	// Token expired or missing, refresh it
	fmt.Printf("DEBUG [Copilot]: Token expired or missing, refreshing... Expires=%d, Now=%d\n", oauth.Expires, now)
	return p.refreshToken(ctx, oauth)
}

// refreshToken refreshes the Copilot API token using the GitHub OAuth token
func (p *GitHubCopilotProvider) refreshToken(ctx context.Context, oauth auth.OAuth) error {
	// Create HTTP request to Copilot token endpoint
	req, err := http.NewRequestWithContext(ctx, "GET", copilotTokenURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create token refresh request: %w", err)
	}

	// Add required headers (matching TypeScript implementation)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+oauth.Refresh)
	req.Header.Set("User-Agent", "GitHubCopilotChat/0.31.2")
	req.Header.Set("Editor-Version", "vscode/1.104.1")
	req.Header.Set("Editor-Plugin-Version", "copilot-chat/0.31.2")

	// Make the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	// Update auth info in auth.json (expires_at is in seconds, convert to milliseconds)
	newOAuth := auth.OAuth{
		Type:    auth.TypeOAuth,
		Refresh: oauth.Refresh,
		Access:  tokenResp.Token,
		Expires: tokenResp.ExpiresAt * 1000,
	}

	if err := auth.Set(p.providerID, newOAuth); err != nil {
		return fmt.Errorf("failed to update auth info: %w", err)
	}

	// Update in-memory token
	p.accessToken = tokenResp.Token
	p.updateClientToken(tokenResp.Token)

	fmt.Printf("DEBUG [Copilot]: Token refreshed successfully, new expiry=%d\n", newOAuth.Expires)
	return nil
}

// updateClientToken recreates the OpenAI client with the new token
func (p *GitHubCopilotProvider) updateClientToken(newToken string) {
	baseURL := "https://api.githubcopilot.com"

	middleware := option.WithMiddleware(func(req *http.Request, next option.MiddlewareNext) (*http.Response, error) {
		req.Header.Set("User-Agent", "GitHubCopilotChat/0.31.2")
		req.Header.Set("Editor-Version", "vscode/1.104.1")
		req.Header.Set("Editor-Plugin-Version", "copilot-chat/0.31.2")
		req.Header.Set("Copilot-Integration-Id", "vscode-chat")
		req.Header.Set("Openai-Intent", "conversation-edits")
		return next(req)
	})

	p.client = openai.NewClient(
		option.WithBaseURL(baseURL),
		option.WithAPIKey(newToken),
		middleware,
	)
}

func (p *GitHubCopilotProvider) Name() string {
	return "github-copilot"
}

func (p *GitHubCopilotProvider) Models() []string {
	// Hardcoded list of actually supported models (source of truth)
	supportedModels := []string{
		"gpt-4o",
		"gpt-4o-2024-08-06",
		"gpt-4",
		"gpt-4-turbo",
		"gpt-3.5-turbo",
		"o1-preview",
		"o1-mini",
		"claude-3.5-sonnet",
	}

	// Try to get models from models.dev registry and filter them
	if p.modelsDevRegistry != nil {
		registryModels := p.modelsDevRegistry.GetModels("github-copilot")
		if len(registryModels) > 0 {
			// Filter registry models to only include actually supported ones
			filteredModels := []string{}
			supportedSet := make(map[string]bool)
			for _, m := range supportedModels {
				supportedSet[m] = true
			}

			for _, m := range registryModels {
				if supportedSet[m] {
					filteredModels = append(filteredModels, m)
				}
			}

			// Only use filtered list if it has models, otherwise fallback
			if len(filteredModels) > 0 {
				return filteredModels
			}
		}
	}

	// Fallback to hardcoded list
	return supportedModels
}

func (p *GitHubCopilotProvider) SetModelsDevRegistry(registry ModelsDevRegistry) {
	p.modelsDevRegistry = registry
}

func (p *GitHubCopilotProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	fmt.Printf("DEBUG [Copilot Complete]: Starting - Model=%s, Messages=%d\n", req.Model, len(req.Messages))

	// Ensure token is valid before making API call
	if err := p.ensureValidToken(ctx); err != nil {
		fmt.Printf("DEBUG [Copilot Complete]: Token validation failed - Error=%v\n", err)
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Convert to OpenAI format
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(req.Messages)+1)

	// Add system message if provided
	if req.SystemPrompt != "" {
		messages = append(messages, openai.SystemMessage(req.SystemPrompt))
		fmt.Printf("DEBUG [Copilot Complete]: Added system prompt\n")
	}

	// Convert request messages
	for i, msg := range req.Messages {
		messages = append(messages, convertMessageToOpenAI(msg))
		fmt.Printf("DEBUG [Copilot Complete]: Converted message %d - Role=%s\n", i, msg.Role)
	}

	// Build request
	params := openai.ChatCompletionNewParams{
		Model:    shared.ChatModel(req.Model),
		Messages: messages,
	}

	if req.MaxTokens > 0 {
		params.MaxTokens = param.NewOpt(int64(req.MaxTokens))
		fmt.Printf("DEBUG [Copilot Complete]: MaxTokens=%d\n", req.MaxTokens)
	}

	if req.Temperature > 0 {
		params.Temperature = param.NewOpt(req.Temperature)
		fmt.Printf("DEBUG [Copilot Complete]: Temperature=%f\n", req.Temperature)
	}

	if len(req.Tools) > 0 {
		params.Tools = convertToolsToOpenAI(req.Tools)
		fmt.Printf("DEBUG [Copilot Complete]: Tools=%d\n", len(req.Tools))
	}

	if len(req.StopSequences) > 0 {
		params.Stop = openai.ChatCompletionNewParamsStopUnion{
			OfStringArray: req.StopSequences,
		}
		fmt.Printf("DEBUG [Copilot Complete]: StopSequences=%d\n", len(req.StopSequences))
	}

	// Make API call
	fmt.Printf("DEBUG [Copilot Complete]: Calling GitHub Copilot API...\n")
	response, err := p.client.Chat.Completions.New(ctx, params)
	if err != nil {
		fmt.Printf("DEBUG [Copilot Complete]: API call failed - Error=%v\n", err)
		return nil, fmt.Errorf("github copilot API error: %w", err)
	}
	fmt.Printf("DEBUG [Copilot Complete]: API call succeeded - ResponseID=%s\n", response.ID)

	// Extract first choice
	if len(response.Choices) == 0 {
		fmt.Printf("DEBUG [Copilot Complete]: No choices in response\n")
		return nil, fmt.Errorf("no choices in response")
	}

	choice := response.Choices[0]
	fmt.Printf("DEBUG [Copilot Complete]: Got choice - FinishReason=%s\n", choice.FinishReason)

	// Convert response
	content := []ContentBlock{}
	if choice.Message.Content != "" {
		content = append(content, ContentBlock{
			Type: "text",
			Text: choice.Message.Content,
		})
		fmt.Printf("DEBUG [Copilot Complete]: Content length=%d\n", len(choice.Message.Content))
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
		fmt.Printf("DEBUG [Copilot Complete]: ToolCalls=%d\n", len(toolCalls))
	}

	fmt.Printf("DEBUG [Copilot Complete]: Returning response - InputTokens=%d, OutputTokens=%d\n",
		response.Usage.PromptTokens, response.Usage.CompletionTokens)

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
	// Ensure token is valid before making API call
	if err := p.ensureValidToken(ctx); err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

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
