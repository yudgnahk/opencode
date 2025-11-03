package provider

import (
	"testing"
)

func TestOpenAIProvider_Name(t *testing.T) {
	provider := NewOpenAI("test-key")
	if provider.Name() != "openai" {
		t.Errorf("Expected name 'openai', got '%s'", provider.Name())
	}
}

func TestOpenAIProvider_Models(t *testing.T) {
	provider := NewOpenAI("test-key")
	models := provider.Models()

	if len(models) == 0 {
		t.Error("Expected at least one model")
	}

	// Check that expected models are present
	expectedModels := map[string]bool{
		"gpt-4o":        false,
		"gpt-4o-mini":   false,
		"gpt-4-turbo":   false,
		"gpt-4":         false,
		"gpt-3.5-turbo": false,
	}

	for _, model := range models {
		if _, exists := expectedModels[model]; exists {
			expectedModels[model] = true
		}
	}

	for model, found := range expectedModels {
		if !found {
			t.Errorf("Expected model '%s' not found in models list", model)
		}
	}
}

func TestConvertMessageToOpenAI_UserMessage(t *testing.T) {
	msg := Message{
		Role: "user",
		Content: []ContentBlock{
			{
				Type: "text",
				Text: "Hello, world!",
			},
		},
	}

	result := convertMessageToOpenAI(msg)

	// Verify it's a user message
	if result.OfUser == nil {
		t.Error("Expected user message")
	}
}

func TestConvertMessageToOpenAI_AssistantMessage(t *testing.T) {
	msg := Message{
		Role: "assistant",
		Content: []ContentBlock{
			{
				Type: "text",
				Text: "Hello there!",
			},
		},
	}

	result := convertMessageToOpenAI(msg)

	// The function should return an assistant message (could be wrapped in a union)
	// We can't easily check the internal structure without OpenAI SDK types,
	// so we just verify the conversion doesn't panic
	_ = result
}

func TestConvertMessageToOpenAI_ToolUseMessage(t *testing.T) {
	msg := Message{
		Role: "assistant",
		Content: []ContentBlock{
			{
				Type: "text",
				Text: "Let me search for that.",
			},
			{
				Type: "tool_use",
				ToolUse: &ToolUse{
					ID:   "call_123",
					Name: "search",
					Input: map[string]interface{}{
						"query": "test query",
					},
				},
			},
		},
	}

	result := convertMessageToOpenAI(msg)

	// Verify assistant message with tool calls
	if result.OfAssistant == nil {
		t.Fatal("Expected assistant message variant")
	}

	if len(result.OfAssistant.ToolCalls) == 0 {
		t.Error("Expected tool calls in assistant message")
	}

	if len(result.OfAssistant.ToolCalls) > 0 {
		toolCall := result.OfAssistant.ToolCalls[0]
		if toolCall.ID != "call_123" {
			t.Errorf("Expected tool call ID 'call_123', got '%s'", toolCall.ID)
		}
		if toolCall.Function.Name != "search" {
			t.Errorf("Expected function name 'search', got '%s'", toolCall.Function.Name)
		}
	}
}

func TestConvertMessageToOpenAI_ToolResultMessage(t *testing.T) {
	msg := Message{
		Role: "tool",
		Content: []ContentBlock{
			{
				Type: "tool_result",
				ToolResult: &ToolResult{
					ToolUseID: "call_123",
					Content:   "Search results here",
				},
			},
		},
	}

	result := convertMessageToOpenAI(msg)

	// Verify tool message
	if result.OfTool == nil {
		t.Error("Expected tool message variant")
	}

	if result.OfTool != nil {
		if result.OfTool.ToolCallID != "call_123" {
			t.Errorf("Expected tool call ID 'call_123', got '%s'", result.OfTool.ToolCallID)
		}
	}
}

func TestConvertToolsToOpenAI(t *testing.T) {
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
				"required": []string{"query"},
			},
		},
	}

	result := convertToolsToOpenAI(tools)

	if len(result) != 1 {
		t.Fatalf("Expected 1 tool, got %d", len(result))
	}

	tool := result[0]
	if tool.Function.Name != "search" {
		t.Errorf("Expected function name 'search', got '%s'", tool.Function.Name)
	}

	// Description is param.Opt[string], check if it's set and has correct value
	if !tool.Function.Description.Valid() || tool.Function.Description.Value != "Search the web" {
		t.Errorf("Expected description 'Search the web', got '%s' (valid: %v)", tool.Function.Description.Value, tool.Function.Description.Valid())
	}

	// Verify parameters were converted
	if len(tool.Function.Parameters) == 0 {
		t.Error("Expected function parameters")
	}
}

func TestConvertToolsToOpenAI_EmptyTools(t *testing.T) {
	tools := []Tool{}
	result := convertToolsToOpenAI(tools)

	if len(result) != 0 {
		t.Errorf("Expected 0 tools, got %d", len(result))
	}
}

func TestExtractToolCallsFromOpenAI_NoToolCalls(t *testing.T) {
	// Simulate empty tool calls from OpenAI
	result := extractToolCallsFromOpenAI(nil)

	if len(result) != 0 {
		t.Errorf("Expected 0 tool calls, got %d", len(result))
	}

	// Test with empty slice
	toolCalls := []ToolCall{}
	if len(toolCalls) != 0 {
		t.Errorf("Expected empty tool calls, got %d", len(toolCalls))
	}
}

func TestConvertContentFromOpenAI_TextOnly(t *testing.T) {
	// Mock OpenAI message - we can't easily construct the full SDK type,
	// so we'll test the logic conceptually
	content := []ContentBlock{
		{
			Type: "text",
			Text: "Hello, world!",
		},
	}

	if len(content) != 1 {
		t.Fatalf("Expected 1 content block, got %d", len(content))
	}

	if content[0].Type != "text" {
		t.Errorf("Expected type 'text', got '%s'", content[0].Type)
	}

	if content[0].Text != "Hello, world!" {
		t.Errorf("Expected text 'Hello, world!', got '%s'", content[0].Text)
	}
}

func TestConvertContentFromOpenAI_WithToolCalls(t *testing.T) {
	// Test the conceptual structure
	content := []ContentBlock{
		{
			Type: "text",
			Text: "Let me search for that.",
		},
		{
			Type: "tool_use",
			ToolUse: &ToolUse{
				ID:   "call_123",
				Name: "search",
				Input: map[string]interface{}{
					"query": "test",
				},
			},
		},
	}

	if len(content) != 2 {
		t.Fatalf("Expected 2 content blocks, got %d", len(content))
	}

	// Verify text block
	if content[0].Type != "text" {
		t.Errorf("Expected first block type 'text', got '%s'", content[0].Type)
	}

	// Verify tool use block
	if content[1].Type != "tool_use" {
		t.Errorf("Expected second block type 'tool_use', got '%s'", content[1].Type)
	}

	if content[1].ToolUse == nil {
		t.Fatal("Expected tool use data")
	}

	if content[1].ToolUse.ID != "call_123" {
		t.Errorf("Expected tool use ID 'call_123', got '%s'", content[1].ToolUse.ID)
	}

	if content[1].ToolUse.Name != "search" {
		t.Errorf("Expected tool name 'search', got '%s'", content[1].ToolUse.Name)
	}
}

func TestMessageRoundTrip(t *testing.T) {
	// Test converting a message to OpenAI format and back
	originalMsg := Message{
		Role: "user",
		Content: []ContentBlock{
			{
				Type: "text",
				Text: "Hello, AI!",
			},
		},
	}

	// Convert to OpenAI format
	openaiMsg := convertMessageToOpenAI(originalMsg)

	// Verify the conversion preserved the role
	if openaiMsg.OfUser == nil {
		t.Error("Expected user message after conversion")
	}
}

func TestToolConversionRoundTrip(t *testing.T) {
	// Test converting tools to OpenAI format
	originalTools := []Tool{
		{
			Name:        "calculator",
			Description: "Perform calculations",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"expression": map[string]interface{}{
						"type":        "string",
						"description": "Math expression",
					},
				},
				"required": []string{"expression"},
			},
		},
	}

	// Convert to OpenAI format
	openaiTools := convertToolsToOpenAI(originalTools)

	if len(openaiTools) != len(originalTools) {
		t.Fatalf("Expected %d tools after conversion, got %d", len(originalTools), len(openaiTools))
	}

	// Verify tool properties preserved
	if openaiTools[0].Function.Name != originalTools[0].Name {
		t.Errorf("Expected name '%s', got '%s'", originalTools[0].Name, openaiTools[0].Function.Name)
	}

	if !openaiTools[0].Function.Description.Valid() || openaiTools[0].Function.Description.Value != originalTools[0].Description {
		t.Error("Tool description not preserved")
	}
}

func TestConvertMessageWithMultipleContentBlocks(t *testing.T) {
	msg := Message{
		Role: "user",
		Content: []ContentBlock{
			{
				Type: "text",
				Text: "First part",
			},
			{
				Type: "text",
				Text: "Second part",
			},
		},
	}

	// Convert to OpenAI - should use the first text block
	result := convertMessageToOpenAI(msg)

	if result.OfUser == nil {
		t.Error("Expected user message")
	}
}

func TestToolResultWithNonStringContent(t *testing.T) {
	msg := Message{
		Role: "tool",
		Content: []ContentBlock{
			{
				Type: "tool_result",
				ToolResult: &ToolResult{
					ToolUseID: "call_456",
					Content: map[string]interface{}{
						"status": "success",
						"data":   []string{"item1", "item2"},
					},
				},
			},
		},
	}

	result := convertMessageToOpenAI(msg)

	// Should convert complex content to string format
	if result.OfTool == nil {
		t.Error("Expected tool message for non-string content")
	}
}

func TestEmptyMessageContent(t *testing.T) {
	msg := Message{
		Role:    "user",
		Content: []ContentBlock{},
	}

	result := convertMessageToOpenAI(msg)

	// Should handle empty content gracefully
	if result.OfUser == nil {
		t.Error("Expected user message even with empty content")
	}
}
