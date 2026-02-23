package search

import (
	"context"
	"msa/pkg/model"
	"testing"
)

// TestSearchTool_Interface tests that SearchTool correctly implements MsaTool
func TestSearchTool_Interface(t *testing.T) {
	tool := NewSearchTool(nil)

	// Test GetName
	if got := tool.GetName(); got != "web_search" {
		t.Errorf("GetName() = %v, want %v", got, "web_search")
	}

	// Test GetDescription
	desc := tool.GetDescription()
	if desc == "" {
		t.Error("GetDescription() should not be empty")
	}

	// Test GetToolGroup
	if got := tool.GetToolGroup(); got != model.SearchToolGroup {
		t.Errorf("GetToolGroup() = %v, want %v", got, model.SearchToolGroup)
	}
}

// TestSearchTool_GetToolInfo tests the GetToolInfo method
func TestSearchTool_GetToolInfo(t *testing.T) {
	tool := NewSearchTool(nil)

	info, err := tool.GetToolInfo()
	if err != nil {
		t.Errorf("GetToolInfo() error = %v", err)
	}
	if info == nil {
		t.Error("GetToolInfo() should return a valid tool.BaseTool")
	}
}

// TestNewSearchTool tests the NewSearchTool constructor
func TestNewSearchTool(t *testing.T) {
	tool := NewSearchTool(nil)
	if tool == nil {
		t.Error("NewSearchTool() should not return nil")
	}
}

// TestNewSearchToolWithEngine tests the NewSearchToolWithEngine constructor
func TestNewSearchToolWithEngine(t *testing.T) {
	tool := NewSearchToolWithEngine(nil)
	if tool == nil {
		t.Error("NewSearchToolWithEngine() should not return nil")
	}
}

// TestWebSearch_EmptyQuery tests WebSearch with empty query
func TestWebSearch_EmptyQuery(t *testing.T) {
	ctx := context.Background()
	params := &model.WebSearchParams{
		Query: "",
	}

	result, err := WebSearch(ctx, params)
	if err == nil {
		t.Error("WebSearch() with empty query should return error")
	}
	if result != "" {
		t.Error("WebSearch() with empty query should return empty result")
	}
}

// TestWebSearch_ValidQuery_Note documents that real search testing requires mocking
func TestWebSearch_ValidQuery_Note(t *testing.T) {
	t.Skip("Skipping real search test - requires HTTP/mock setup")

	// To properly test WebSearch, we need to:
	// 1. Mock the search router (internal.SearchRouter)
	// 2. Set up test search results
	// 3. Test various scenarios (success, failure, empty results)

	ctx := context.Background()
	params := &model.WebSearchParams{
		Query: "test query",
	}

	// This would require mocking defaultSearchRouter
	_, _ = WebSearch(ctx, params)
}

// TestGetDefaultRouter tests the GetDefaultRouter function
func TestGetDefaultRouter(t *testing.T) {
	router := GetDefaultRouter()
	// May be nil, just verify function doesn't panic
	_ = router
}

// TestSetDefaultRouter tests the SetDefaultRouter function
func TestSetDefaultRouter(t *testing.T) {
	// Note: We can't easily create a SearchRouter without accessing internal package
	// This test verifies the function exists and doesn't panic with nil
	SetDefaultRouter(nil)
}

// TestWebSearchParams_Struct tests the WebSearchParams struct
func TestWebSearchParams_Struct(t *testing.T) {
	// Verify the struct compiles correctly
	params := &model.WebSearchParams{
		Query: "test query",
	}

	if params.Query != "test query" {
		t.Error("WebSearchParams struct not working correctly")
	}
}

// TestWebSearchResponse_Struct tests the WebSearchResponse struct
func TestWebSearchResponse_Struct(t *testing.T) {
	// Verify the struct compiles correctly
	response := &model.WebSearchResponse{
		Query:     "test",
		Results:   nil,
		RequestID: "123",
		Status:    "success",
		Error:     "",
		Message:   "test message",
	}

	if response.Query != "test" {
		t.Error("WebSearchResponse struct not working correctly")
	}
}
