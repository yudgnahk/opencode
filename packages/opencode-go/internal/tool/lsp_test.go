package tool

import (
	"testing"
)

func TestLSPHover(t *testing.T) {
	executor := newTestToolExecutor("/tmp")

	// Test with missing filePath
	_, err := executor.LSPHover(LSPHoverRequest{
		Line:      10,
		Character: 5,
	})
	if err == nil {
		t.Error("LSPHover should fail with missing filePath")
	}

	// Test with unsupported file type
	_, err = executor.LSPHover(LSPHoverRequest{
		FilePath:  "/tmp/test.xyz",
		Line:      10,
		Character: 5,
	})
	if err == nil {
		t.Error("LSPHover should fail with unsupported file type")
	}
}

func TestLSPCompletion(t *testing.T) {
	executor := newTestToolExecutor("/tmp")

	// Test with missing filePath
	_, err := executor.LSPCompletion(LSPCompletionRequest{
		Line:      10,
		Character: 5,
	})
	if err == nil {
		t.Error("LSPCompletion should fail with missing filePath")
	}

	// Test with unsupported file type
	_, err = executor.LSPCompletion(LSPCompletionRequest{
		FilePath:  "/tmp/test.abc",
		Line:      10,
		Character: 5,
	})
	if err == nil {
		t.Error("LSPCompletion should fail with unsupported file type")
	}
}

func TestLSPDefinition(t *testing.T) {
	executor := newTestToolExecutor("/tmp")

	// Test with missing filePath
	_, err := executor.LSPDefinition(LSPDefinitionRequest{
		Line:      10,
		Character: 5,
	})
	if err == nil {
		t.Error("LSPDefinition should fail with missing filePath")
	}

	// Test with unsupported file type
	_, err = executor.LSPDefinition(LSPDefinitionRequest{
		FilePath:  "/tmp/test.unknown",
		Line:      10,
		Character: 5,
	})
	if err == nil {
		t.Error("LSPDefinition should fail with unsupported file type")
	}
}

func TestLSPReferences(t *testing.T) {
	executor := newTestToolExecutor("/tmp")

	// Test with missing filePath
	_, err := executor.LSPReferences(LSPReferencesRequest{
		Line:      10,
		Character: 5,
	})
	if err == nil {
		t.Error("LSPReferences should fail with missing filePath")
	}

	// Test with unsupported file type
	_, err = executor.LSPReferences(LSPReferencesRequest{
		FilePath:  "/tmp/test.txt",
		Line:      10,
		Character: 5,
	})
	if err == nil {
		t.Error("LSPReferences should fail with unsupported file type")
	}
}

func TestStopLSPServers(t *testing.T) {
	executor := newTestToolExecutor("/tmp")

	// Should not panic
	executor.StopLSPServers()

	// Verify LSP client exists
	if executor.GetLSPClient() == nil {
		t.Error("LSP client should not be nil")
	}
}
