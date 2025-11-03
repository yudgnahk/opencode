package provider

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type GeminiProvider struct {
	client *genai.Client
	apiKey string
}

func NewGemini(ctx context.Context, apiKey string) (*GeminiProvider, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	return &GeminiProvider{
		client: client,
		apiKey: apiKey,
	}, nil
}

func (p *GeminiProvider) Name() string {
	return "gemini"
}

func (p *GeminiProvider) Models() []string {
	return []string{
		"gemini-2.0-flash-exp",
		"gemini-exp-1206",
		"gemini-exp-1121",
		"gemini-1.5-pro",
		"gemini-1.5-pro-002",
		"gemini-1.5-flash",
		"gemini-1.5-flash-002",
		"gemini-1.5-flash-8b",
	}
}

func (p *GeminiProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	model := p.client.GenerativeModel(req.Model)

	// Configure model
	if req.SystemPrompt != "" {
		model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{genai.Text(req.SystemPrompt)},
			Role:  "system",
		}
	}

	if req.Temperature > 0 {
		model.SetTemperature(float32(req.Temperature))
	}

	if req.MaxTokens > 0 {
		model.SetMaxOutputTokens(int32(req.MaxTokens))
	}

	if len(req.StopSequences) > 0 {
		model.StopSequences = req.StopSequences
	}

	if len(req.Tools) > 0 {
		model.Tools = convertToolsToGemini(req.Tools)
	}

	// Convert messages to Gemini format
	contents := convertMessagesToGemini(req.Messages)

	// Generate content
	response, err := model.GenerateContent(ctx, contents...)
	if err != nil {
		return nil, fmt.Errorf("gemini API error: %w", err)
	}

	// Extract first candidate
	if len(response.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in response")
	}

	candidate := response.Candidates[0]

	// Convert response
	return &CompletionResponse{
		ID:         "", // Gemini doesn't provide response ID
		Content:    convertContentFromGemini(candidate.Content),
		StopReason: convertFinishReasonFromGemini(candidate.FinishReason),
		Usage: Usage{
			InputTokens:  int(response.UsageMetadata.PromptTokenCount),
			OutputTokens: int(response.UsageMetadata.CandidatesTokenCount),
		},
		ToolCalls: extractToolCallsFromGemini(candidate.Content),
	}, nil
}

func (p *GeminiProvider) Stream(ctx context.Context, req CompletionRequest) (Stream, error) {
	model := p.client.GenerativeModel(req.Model)

	// Configure model (same as Complete)
	if req.SystemPrompt != "" {
		model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{genai.Text(req.SystemPrompt)},
			Role:  "system",
		}
	}

	if req.Temperature > 0 {
		model.SetTemperature(float32(req.Temperature))
	}

	if req.MaxTokens > 0 {
		model.SetMaxOutputTokens(int32(req.MaxTokens))
	}

	if len(req.StopSequences) > 0 {
		model.StopSequences = req.StopSequences
	}

	if len(req.Tools) > 0 {
		model.Tools = convertToolsToGemini(req.Tools)
	}

	// Convert messages
	contents := convertMessagesToGemini(req.Messages)

	// Create streaming request
	iter := model.GenerateContentStream(ctx, contents...)

	return &GeminiStream{iter: iter}, nil
}

// GeminiStream wraps the Gemini streaming response
type GeminiStream struct {
	iter  *genai.GenerateContentResponseIterator
	chunk StreamChunk
	err   error
}

func (s *GeminiStream) Next() bool {
	response, err := s.iter.Next()
	if err == iterator.Done {
		return false
	}
	if err != nil {
		s.err = err
		return false
	}

	// Extract content from response
	if len(response.Candidates) > 0 {
		candidate := response.Candidates[0]

		if candidate.Content != nil && len(candidate.Content.Parts) > 0 {
			for _, part := range candidate.Content.Parts {
				switch p := part.(type) {
				case genai.Text:
					s.chunk = StreamChunk{
						Type:  ChunkTypeText,
						Delta: string(p),
					}
					return true
				case genai.FunctionCall:
					s.chunk = StreamChunk{
						Type: ChunkTypeToolCall,
						ToolCall: &ToolCall{
							Name: p.Name,
						},
					}
					return true
				}
			}
		}

		// Check for finish reason
		if candidate.FinishReason != genai.FinishReasonUnspecified {
			s.chunk = StreamChunk{
				Type:       ChunkTypeStop,
				StopReason: convertFinishReasonFromGemini(candidate.FinishReason),
			}
			return true
		}
	}

	// Continue to next chunk if this one had no meaningful data
	return s.Next()
}

func (s *GeminiStream) Chunk() StreamChunk {
	return s.chunk
}

func (s *GeminiStream) Error() error {
	if s.err != nil {
		return s.err
	}
	return nil
}

func (s *GeminiStream) Close() error {
	// Iterator doesn't need explicit closing
	return nil
}

// Helper functions for conversion
func convertMessagesToGemini(messages []Message) []genai.Part {
	var parts []genai.Part

	for _, msg := range messages {
		for _, block := range msg.Content {
			switch block.Type {
			case "text":
				parts = append(parts, genai.Text(block.Text))
			case "tool_result":
				if block.ToolResult != nil {
					// Convert tool result to FunctionResponse
					parts = append(parts, genai.FunctionResponse{
						Name:     block.ToolResult.ToolUseID,
						Response: map[string]interface{}{"result": block.ToolResult.Content},
					})
				}
			}
		}
	}

	return parts
}

func convertContentFromGemini(content *genai.Content) []ContentBlock {
	result := make([]ContentBlock, 0)

	if content == nil {
		return result
	}

	for _, part := range content.Parts {
		switch p := part.(type) {
		case genai.Text:
			result = append(result, ContentBlock{
				Type: "text",
				Text: string(p),
			})
		case genai.FunctionCall:
			result = append(result, ContentBlock{
				Type: "tool_use",
				ToolUse: &ToolUse{
					Name:  p.Name,
					Input: p.Args,
				},
			})
		}
	}

	return result
}

func convertToolsToGemini(tools []Tool) []*genai.Tool {
	funcDecls := make([]*genai.FunctionDeclaration, len(tools))

	for i, tool := range tools {
		// Convert InputSchema to Gemini Schema
		schema := &genai.Schema{
			Type: genai.TypeObject,
		}

		if schemaMap, ok := tool.InputSchema.(map[string]interface{}); ok {
			if props, ok := schemaMap["properties"].(map[string]interface{}); ok {
				schema.Properties = convertPropertiesToGemini(props)
			}
			if required, ok := schemaMap["required"].([]interface{}); ok {
				reqStrs := make([]string, len(required))
				for j, r := range required {
					if str, ok := r.(string); ok {
						reqStrs[j] = str
					}
				}
				schema.Required = reqStrs
			}
		}

		funcDecls[i] = &genai.FunctionDeclaration{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  schema,
		}
	}

	return []*genai.Tool{{FunctionDeclarations: funcDecls}}
}

func convertPropertiesToGemini(props map[string]interface{}) map[string]*genai.Schema {
	result := make(map[string]*genai.Schema)

	for name, prop := range props {
		if propMap, ok := prop.(map[string]interface{}); ok {
			schema := &genai.Schema{}

			if typeStr, ok := propMap["type"].(string); ok {
				schema.Type = convertTypeToGemini(typeStr)
			}
			if desc, ok := propMap["description"].(string); ok {
				schema.Description = desc
			}

			result[name] = schema
		}
	}

	return result
}

func convertTypeToGemini(typeStr string) genai.Type {
	switch typeStr {
	case "string":
		return genai.TypeString
	case "number":
		return genai.TypeNumber
	case "integer":
		return genai.TypeInteger
	case "boolean":
		return genai.TypeBoolean
	case "array":
		return genai.TypeArray
	case "object":
		return genai.TypeObject
	default:
		return genai.TypeUnspecified
	}
}

func extractToolCallsFromGemini(content *genai.Content) []ToolCall {
	calls := make([]ToolCall, 0)

	if content == nil {
		return calls
	}

	for _, part := range content.Parts {
		if fc, ok := part.(genai.FunctionCall); ok {
			calls = append(calls, ToolCall{
				Name:      fc.Name,
				Arguments: fc.Args,
			})
		}
	}

	return calls
}

func convertFinishReasonFromGemini(reason genai.FinishReason) string {
	switch reason {
	case genai.FinishReasonStop:
		return "stop"
	case genai.FinishReasonMaxTokens:
		return "max_tokens"
	case genai.FinishReasonSafety:
		return "safety"
	case genai.FinishReasonRecitation:
		return "recitation"
	case genai.FinishReasonOther:
		return "other"
	default:
		return "unknown"
	}
}

// Close closes the Gemini client
func (p *GeminiProvider) Close() error {
	return p.client.Close()
}
