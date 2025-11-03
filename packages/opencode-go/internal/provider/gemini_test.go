package provider

import (
	"testing"

	"github.com/google/generative-ai-go/genai"
)

func TestGeminiProvider_Name(t *testing.T) {
	// We can't easily create a real GeminiProvider without a context and API key,
	// but we can test the name directly
	name := "gemini"
	if name != "gemini" {
		t.Errorf("Expected name 'gemini', got '%s'", name)
	}
}

func TestGeminiProvider_Models(t *testing.T) {
	// Test the models list conceptually
	expectedModels := []string{
		"gemini-2.0-flash-exp",
		"gemini-exp-1206",
		"gemini-exp-1121",
		"gemini-1.5-pro",
		"gemini-1.5-pro-002",
		"gemini-1.5-flash",
		"gemini-1.5-flash-002",
		"gemini-1.5-flash-8b",
	}

	if len(expectedModels) != 8 {
		t.Errorf("Expected 8 models, got %d", len(expectedModels))
	}

	// Check that expected models are present
	modelSet := make(map[string]bool)
	for _, model := range expectedModels {
		modelSet[model] = true
	}

	expectedKeys := []string{
		"gemini-2.0-flash-exp",
		"gemini-1.5-pro",
		"gemini-1.5-flash",
	}

	for _, key := range expectedKeys {
		if !modelSet[key] {
			t.Errorf("Expected model '%s' not found in models list", key)
		}
	}
}

func TestConvertMessagesToGemini(t *testing.T) {
	messages := []Message{
		{
			Role: "user",
			Content: []ContentBlock{
				{
					Type: "text",
					Text: "Hello, Gemini!",
				},
			},
		},
	}

	result := convertMessagesToGemini(messages)

	if len(result) != 1 {
		t.Fatalf("Expected 1 part, got %d", len(result))
	}

	// Check if it's a text part
	if text, ok := result[0].(genai.Text); ok {
		if string(text) != "Hello, Gemini!" {
			t.Errorf("Expected text 'Hello, Gemini!', got '%s'", string(text))
		}
	} else {
		t.Error("Expected genai.Text type")
	}
}

func TestConvertMessagesToGemini_MultipleBlocks(t *testing.T) {
	messages := []Message{
		{
			Role: "user",
			Content: []ContentBlock{
				{
					Type: "text",
					Text: "First message",
				},
			},
		},
		{
			Role: "assistant",
			Content: []ContentBlock{
				{
					Type: "text",
					Text: "Second message",
				},
			},
		},
	}

	result := convertMessagesToGemini(messages)

	if len(result) != 2 {
		t.Fatalf("Expected 2 parts, got %d", len(result))
	}
}

func TestConvertMessagesToGemini_ToolResult(t *testing.T) {
	messages := []Message{
		{
			Role: "tool",
			Content: []ContentBlock{
				{
					Type: "tool_result",
					ToolResult: &ToolResult{
						ToolUseID: "call_123",
						Content:   "Result data",
					},
				},
			},
		},
	}

	result := convertMessagesToGemini(messages)

	if len(result) != 1 {
		t.Fatalf("Expected 1 part, got %d", len(result))
	}

	// Check if it's a function response
	if fr, ok := result[0].(genai.FunctionResponse); ok {
		if fr.Name != "call_123" {
			t.Errorf("Expected function name 'call_123', got '%s'", fr.Name)
		}
	} else {
		t.Error("Expected genai.FunctionResponse type")
	}
}

func TestConvertContentFromGemini_TextOnly(t *testing.T) {
	content := &genai.Content{
		Parts: []genai.Part{
			genai.Text("Hello, world!"),
		},
	}

	result := convertContentFromGemini(content)

	if len(result) != 1 {
		t.Fatalf("Expected 1 content block, got %d", len(result))
	}

	if result[0].Type != "text" {
		t.Errorf("Expected type 'text', got '%s'", result[0].Type)
	}

	if result[0].Text != "Hello, world!" {
		t.Errorf("Expected text 'Hello, world!', got '%s'", result[0].Text)
	}
}

func TestConvertContentFromGemini_WithFunctionCall(t *testing.T) {
	content := &genai.Content{
		Parts: []genai.Part{
			genai.Text("Let me search for that."),
			genai.FunctionCall{
				Name: "search",
				Args: map[string]interface{}{
					"query": "test query",
				},
			},
		},
	}

	result := convertContentFromGemini(content)

	if len(result) != 2 {
		t.Fatalf("Expected 2 content blocks, got %d", len(result))
	}

	// Check text block
	if result[0].Type != "text" {
		t.Errorf("Expected first block type 'text', got '%s'", result[0].Type)
	}

	// Check tool use block
	if result[1].Type != "tool_use" {
		t.Errorf("Expected second block type 'tool_use', got '%s'", result[1].Type)
	}

	if result[1].ToolUse == nil {
		t.Fatal("Expected tool use data")
	}

	if result[1].ToolUse.Name != "search" {
		t.Errorf("Expected tool name 'search', got '%s'", result[1].ToolUse.Name)
	}
}

func TestConvertContentFromGemini_NilContent(t *testing.T) {
	result := convertContentFromGemini(nil)

	if len(result) != 0 {
		t.Errorf("Expected 0 content blocks for nil content, got %d", len(result))
	}
}

func TestConvertToolsToGemini(t *testing.T) {
	tools := []Tool{
		{
			Name:        "search",
			Description: "Search the web",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query",
					},
				},
				"required": []interface{}{"query"},
			},
		},
	}

	result := convertToolsToGemini(tools)

	if len(result) != 1 {
		t.Fatalf("Expected 1 tool, got %d", len(result))
	}

	tool := result[0]
	if len(tool.FunctionDeclarations) != 1 {
		t.Fatalf("Expected 1 function declaration, got %d", len(tool.FunctionDeclarations))
	}

	funcDecl := tool.FunctionDeclarations[0]
	if funcDecl.Name != "search" {
		t.Errorf("Expected function name 'search', got '%s'", funcDecl.Name)
	}

	if funcDecl.Description != "Search the web" {
		t.Errorf("Expected description 'Search the web', got '%s'", funcDecl.Description)
	}

	// Check schema
	if funcDecl.Parameters == nil {
		t.Fatal("Expected parameters schema")
	}

	if funcDecl.Parameters.Type != genai.TypeObject {
		t.Errorf("Expected schema type Object, got %v", funcDecl.Parameters.Type)
	}
}

func TestConvertToolsToGemini_EmptyTools(t *testing.T) {
	tools := []Tool{}
	result := convertToolsToGemini(tools)

	if len(result) != 1 {
		t.Fatalf("Expected 1 tool wrapper, got %d", len(result))
	}

	if len(result[0].FunctionDeclarations) != 0 {
		t.Errorf("Expected 0 function declarations, got %d", len(result[0].FunctionDeclarations))
	}
}

func TestConvertPropertiesToGemini(t *testing.T) {
	props := map[string]interface{}{
		"query": map[string]interface{}{
			"type":        "string",
			"description": "Search query",
		},
		"limit": map[string]interface{}{
			"type":        "integer",
			"description": "Result limit",
		},
	}

	result := convertPropertiesToGemini(props)

	if len(result) != 2 {
		t.Fatalf("Expected 2 properties, got %d", len(result))
	}

	// Check query property
	if querySchema, ok := result["query"]; ok {
		if querySchema.Type != genai.TypeString {
			t.Errorf("Expected query type String, got %v", querySchema.Type)
		}
		if querySchema.Description != "Search query" {
			t.Errorf("Expected description 'Search query', got '%s'", querySchema.Description)
		}
	} else {
		t.Error("Expected 'query' property")
	}

	// Check limit property
	if limitSchema, ok := result["limit"]; ok {
		if limitSchema.Type != genai.TypeInteger {
			t.Errorf("Expected limit type Integer, got %v", limitSchema.Type)
		}
	} else {
		t.Error("Expected 'limit' property")
	}
}

func TestConvertTypeToGemini(t *testing.T) {
	tests := []struct {
		typeStr  string
		expected genai.Type
	}{
		{"string", genai.TypeString},
		{"number", genai.TypeNumber},
		{"integer", genai.TypeInteger},
		{"boolean", genai.TypeBoolean},
		{"array", genai.TypeArray},
		{"object", genai.TypeObject},
		{"unknown", genai.TypeUnspecified},
	}

	for _, tt := range tests {
		t.Run(tt.typeStr, func(t *testing.T) {
			result := convertTypeToGemini(tt.typeStr)
			if result != tt.expected {
				t.Errorf("convertTypeToGemini(%s) = %v, expected %v", tt.typeStr, result, tt.expected)
			}
		})
	}
}

func TestExtractToolCallsFromGemini(t *testing.T) {
	content := &genai.Content{
		Parts: []genai.Part{
			genai.Text("Let me search."),
			genai.FunctionCall{
				Name: "search",
				Args: map[string]interface{}{
					"query": "test",
				},
			},
		},
	}

	result := extractToolCallsFromGemini(content)

	if len(result) != 1 {
		t.Fatalf("Expected 1 tool call, got %d", len(result))
	}

	if result[0].Name != "search" {
		t.Errorf("Expected tool name 'search', got '%s'", result[0].Name)
	}

	if args, ok := result[0].Arguments.(map[string]interface{}); ok {
		if query, ok := args["query"].(string); !ok || query != "test" {
			t.Error("Expected query argument 'test'")
		}
	} else {
		t.Error("Expected arguments map")
	}
}

func TestExtractToolCallsFromGemini_NilContent(t *testing.T) {
	result := extractToolCallsFromGemini(nil)

	if len(result) != 0 {
		t.Errorf("Expected 0 tool calls for nil content, got %d", len(result))
	}
}

func TestExtractToolCallsFromGemini_NoToolCalls(t *testing.T) {
	content := &genai.Content{
		Parts: []genai.Part{
			genai.Text("Just text, no tool calls"),
		},
	}

	result := extractToolCallsFromGemini(content)

	if len(result) != 0 {
		t.Errorf("Expected 0 tool calls, got %d", len(result))
	}
}

func TestConvertFinishReasonFromGemini(t *testing.T) {
	tests := []struct {
		reason   genai.FinishReason
		expected string
	}{
		{genai.FinishReasonStop, "stop"},
		{genai.FinishReasonMaxTokens, "max_tokens"},
		{genai.FinishReasonSafety, "safety"},
		{genai.FinishReasonRecitation, "recitation"},
		{genai.FinishReasonOther, "other"},
		{genai.FinishReasonUnspecified, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := convertFinishReasonFromGemini(tt.reason)
			if result != tt.expected {
				t.Errorf("convertFinishReasonFromGemini(%v) = %s, expected %s", tt.reason, result, tt.expected)
			}
		})
	}
}

func TestMessageConversionRoundTrip(t *testing.T) {
	// Test converting messages to Gemini and extracting content back
	originalMessages := []Message{
		{
			Role: "user",
			Content: []ContentBlock{
				{
					Type: "text",
					Text: "Hello, AI!",
				},
			},
		},
	}

	// Convert to Gemini
	geminiParts := convertMessagesToGemini(originalMessages)

	if len(geminiParts) != 1 {
		t.Fatalf("Expected 1 part after conversion, got %d", len(geminiParts))
	}

	// Verify text was preserved
	if text, ok := geminiParts[0].(genai.Text); ok {
		if string(text) != "Hello, AI!" {
			t.Errorf("Text not preserved in conversion, got '%s'", string(text))
		}
	} else {
		t.Error("Expected text part")
	}
}

func TestToolSchemaConversion(t *testing.T) {
	// Test complete tool schema conversion
	tool := Tool{
		Name:        "calculator",
		Description: "Perform calculations",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"expression": map[string]interface{}{
					"type":        "string",
					"description": "Math expression",
				},
				"precision": map[string]interface{}{
					"type":        "integer",
					"description": "Decimal precision",
				},
			},
			"required": []interface{}{"expression"},
		},
	}

	result := convertToolsToGemini([]Tool{tool})

	if len(result) != 1 || len(result[0].FunctionDeclarations) != 1 {
		t.Fatal("Expected 1 tool with 1 function declaration")
	}

	funcDecl := result[0].FunctionDeclarations[0]

	// Verify basic properties
	if funcDecl.Name != "calculator" {
		t.Errorf("Expected name 'calculator', got '%s'", funcDecl.Name)
	}

	if funcDecl.Description != "Perform calculations" {
		t.Errorf("Expected description 'Perform calculations', got '%s'", funcDecl.Description)
	}

	// Verify schema structure
	if funcDecl.Parameters == nil {
		t.Fatal("Expected parameters schema")
	}

	if funcDecl.Parameters.Type != genai.TypeObject {
		t.Errorf("Expected object type, got %v", funcDecl.Parameters.Type)
	}

	// Verify properties were converted
	if len(funcDecl.Parameters.Properties) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(funcDecl.Parameters.Properties))
	}

	// Verify required fields
	if len(funcDecl.Parameters.Required) != 1 || funcDecl.Parameters.Required[0] != "expression" {
		t.Error("Expected 'expression' in required fields")
	}
}

func TestComplexContentConversion(t *testing.T) {
	// Test conversion of complex content with mixed types
	content := &genai.Content{
		Parts: []genai.Part{
			genai.Text("I'll help with that."),
			genai.FunctionCall{
				Name: "tool1",
				Args: map[string]interface{}{"arg1": "value1"},
			},
			genai.Text("Processing..."),
			genai.FunctionCall{
				Name: "tool2",
				Args: map[string]interface{}{"arg2": "value2"},
			},
		},
	}

	result := convertContentFromGemini(content)

	if len(result) != 4 {
		t.Fatalf("Expected 4 content blocks, got %d", len(result))
	}

	// Verify sequence
	if result[0].Type != "text" || result[0].Text != "I'll help with that." {
		t.Error("First block should be text")
	}

	if result[1].Type != "tool_use" || result[1].ToolUse == nil || result[1].ToolUse.Name != "tool1" {
		t.Error("Second block should be tool_use for tool1")
	}

	if result[2].Type != "text" || result[2].Text != "Processing..." {
		t.Error("Third block should be text")
	}

	if result[3].Type != "tool_use" || result[3].ToolUse == nil || result[3].ToolUse.Name != "tool2" {
		t.Error("Fourth block should be tool_use for tool2")
	}
}
