package search

import (
	"context"
	"encoding/json"
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
	// 新行为：不返回 error，而是返回包含错误信息的 JSON
	if err != nil {
		t.Errorf("WebSearch() should not return error, got: %v", err)
	}
	if result == "" {
		t.Error("WebSearch() with empty query should return error JSON result")
	}
	// 验证返回的 JSON 包含失败状态
	var response model.ToolResult
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		t.Errorf("Failed to parse result JSON: %v", err)
	}
	if response.Success {
		t.Error("Expected success to be false")
	}
	if response.ErrorMsg == "" {
		t.Error("Expected error message in response")
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

// TestWebSearchData_Struct tests the WebSearchData struct
func TestWebSearchData_Struct(t *testing.T) {
	// Verify the struct compiles correctly
	data := &model.WebSearchData{
		Query:      "test",
		Results:    nil,
		RequestID:  "123",
		UsedEngine: "google",
	}

	if data.Query != "test" {
		t.Error("WebSearchData struct not working correctly")
	}
}

// TestFetchPageContent_EmptyURL tests FetchPageContent with empty URL
func TestFetchPageContent_EmptyURL(t *testing.T) {
	ctx := context.Background()
	params := &model.FetchPageParams{
		URL: "",
	}

	result, err := FetchPageContent(ctx, params)
	// 新行为：不返回 error，而是返回包含错误信息的 JSON
	if err != nil {
		t.Errorf("FetchPageContent() should not return error, got: %v", err)
	}
	if result == "" {
		t.Error("FetchPageContent() with empty URL should return error JSON result")
	}
	// 验证返回的 JSON 包含失败状态
	var response model.ToolResult
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		t.Errorf("Failed to parse result JSON: %v", err)
	}
	if response.Success {
		t.Error("Expected success to be false")
	}
	if response.ErrorMsg == "" {
		t.Error("Expected error message in response")
	}
}

// TestFetchPageContent_InvalidURL tests FetchPageContent with invalid URL
func TestFetchPageContent_InvalidURL(t *testing.T) {
	ctx := context.Background()
	params := &model.FetchPageParams{
		URL: "not-a-valid-url",
	}

	result, err := FetchPageContent(ctx, params)
	// 新行为：不返回 error，而是返回包含错误信息的 JSON
	if err != nil {
		t.Errorf("FetchPageContent() should not return error, got: %v", err)
	}
	if result == "" {
		t.Error("FetchPageContent() with invalid URL should return error JSON result")
	}
	// 验证返回的 JSON 包含失败状态
	var response model.ToolResult
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		t.Errorf("Failed to parse result JSON: %v", err)
	}
	if response.Success {
		t.Error("Expected success to be false")
	}
}

// TestFetchPageData_Struct tests the FetchPageData struct
func TestFetchPageData_Struct(t *testing.T) {
	// Verify the struct compiles correctly
	data := &model.FetchPageData{
		URL:         "https://example.com",
		Title:       "Test Page",
		Content:     "Test content",
		HasMore:     false,
		TotalLength: 12,
	}

	if data.URL != "https://example.com" {
		t.Error("FetchPageData struct not working correctly")
	}
}
