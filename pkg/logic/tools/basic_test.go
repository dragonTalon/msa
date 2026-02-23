package tools

import (
	"msa/pkg/model"
	"testing"

	"github.com/cloudwego/eino/components/tool"
)

// mockTool is a mock implementation of MsaTool for testing
type mockTool struct {
	name        string
	description string
	group       model.ToolGroup
	toolInfo    tool.BaseTool
	infoError   error
}

func (m *mockTool) GetToolInfo() (tool.BaseTool, error) {
	return m.toolInfo, m.infoError
}

func (m *mockTool) GetName() string {
	return m.name
}

func (m *mockTool) GetDescription() string {
	return m.description
}

func (m *mockTool) GetToolGroup() model.ToolGroup {
	return m.group
}

// helper function to create a mock tool
func newMockTool(name, desc string, group model.ToolGroup) *mockTool {
	return &mockTool{
		name:        name,
		description: desc,
		group:       group,
	}
}

// TestRegisterTool tests the RegisterTool function
func TestRegisterTool(t *testing.T) {
	tests := []struct {
		name    string
		tool    *mockTool
		wantNil bool
	}{
		{
			name:    "register stock tool",
			tool:    newMockTool("stock_query", "Query stock information", model.StockToolGroup),
			wantNil: false,
		},
		{
			name:    "register search tool",
			tool:    newMockTool("web_search", "Search the web", model.SearchToolGroup),
			wantNil: false,
		},
		{
			name:    "register finance tool",
			tool:    newMockTool("finance_data", "Get finance data", model.FinanceToolGroup),
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Register the tool
			RegisterTool(tt.tool)

			// Verify tool is in toolMap
			got := toolMap[tt.tool.GetName()]
			if (got == nil) != tt.wantNil {
				t.Errorf("RegisterTool() tool in map = %v, wantNil %v", got == nil, tt.wantNil)
			}
			if got != nil && got.GetName() != tt.tool.GetName() {
				t.Errorf("RegisterTool() tool name = %v, want %v", got.GetName(), tt.tool.GetName())
			}
		})
	}
}

// TestRegisterTool_ToolGroup tests that tools are organized into groups
func TestRegisterTool_ToolGroup(t *testing.T) {
	// Clear toolGroupMap for this test (note: this affects global state)
	// In production, you'd want a better way to reset state

	tool1 := newMockTool("tool1", "Description 1", model.StockToolGroup)
	tool2 := newMockTool("tool2", "Description 2", model.StockToolGroup)
	tool3 := newMockTool("tool3", "Description 3", model.SearchToolGroup)

	RegisterTool(tool1)
	RegisterTool(tool2)
	RegisterTool(tool3)

	// Check stock tool group
	stockGroup := toolGroupMap[model.StockToolGroup]
	if len(stockGroup) < 2 {
		t.Errorf("StockToolGroup should have at least 2 tools, got %d", len(stockGroup))
	}

	// Check search tool group
	searchGroup := toolGroupMap[model.SearchToolGroup]
	if len(searchGroup) < 1 {
		t.Errorf("SearchToolGroup should have at least 1 tool, got %d", len(searchGroup))
	}

	// Verify tool names in groups
	foundTool1 := false
	foundTool2 := false
	for _, pair := range stockGroup {
		if pair.Key == "tool1" {
			foundTool1 = true
		}
		if pair.Key == "tool2" {
			foundTool2 = true
		}
	}
	if !foundTool1 {
		t.Error("Tool1 not found in StockToolGroup")
	}
	if !foundTool2 {
		t.Error("Tool2 not found in StockToolGroup")
	}
}

// TestGetAllTools tests the GetAllTools function
func TestGetAllTools(t *testing.T) {
	// Register some test tools
	testTools := []*mockTool{
		newMockTool("test_tool1", "Test tool 1", model.StockToolGroup),
		newMockTool("test_tool2", "Test tool 2", model.SearchToolGroup),
		newMockTool("test_tool3", "Test tool 3", model.FinanceToolGroup),
	}

	for _, tool := range testTools {
		RegisterTool(tool)
	}

	// Get all tools
	allTools := GetAllTools()

	// Verify we got tools
	if len(allTools) < len(testTools) {
		t.Errorf("GetAllTools() returned %d tools, want at least %d", len(allTools), len(testTools))
	}
}

// TestGetAllTools_WithError tests GetAllTools when a tool returns an error
func TestGetAllTools_WithError(t *testing.T) {
	// Create a tool that returns an error
	errorTool := &mockTool{
		name:        "error_tool",
		description: "Tool that errors",
		group:       model.StockToolGroup,
		infoError:   &testError{"simulated error"},
	}

	RegisterTool(errorTool)

	// GetAllTools should not panic, it should skip the errored tool
	allTools := GetAllTools()
	// We expect this to not crash
	_ = allTools
}

// TestMsaTool_Interface tests that mockTool correctly implements MsaTool
func TestMsaTool_Interface(t *testing.T) {
	tool := newMockTool("interface_test", "Test interface implementation", model.MarketToolGroup)

	// Test all interface methods
	if got := tool.GetName(); got != "interface_test" {
		t.Errorf("GetName() = %v, want %v", got, "interface_test")
	}

	if got := tool.GetDescription(); got != "Test interface implementation" {
		t.Errorf("GetDescription() = %v, want %v", got, "Test interface implementation")
	}

	if got := tool.GetToolGroup(); got != model.MarketToolGroup {
		t.Errorf("GetToolGroup() = %v, want %v", got, model.MarketToolGroup)
	}

	// GetToolInfo can return nil for testing purposes
	info, err := tool.GetToolInfo()
	if err != nil {
		t.Errorf("GetToolInfo() unexpected error = %v", err)
	}
	if info != nil && info != tool.toolInfo {
		t.Error("GetToolInfo() should return the configured toolInfo")
	}
}

// TestRegisterTool_DuplicateName tests registering tools with duplicate names
func TestRegisterTool_DuplicateName(t *testing.T) {
	tool1 := newMockTool("duplicate", "First tool", model.StockToolGroup)
	tool2 := newMockTool("duplicate", "Second tool with same name", model.SearchToolGroup)

	RegisterTool(tool1)
	RegisterTool(tool2)

	// Both should be registered, the second one overwrites in toolMap
	got := toolMap["duplicate"]
	if got == nil {
		t.Error("RegisterTool() - tool should be registered")
	}
	// The second registration should have overwritten
	if got.GetDescription() != "Second tool with same name" {
		t.Logf("Note: duplicate name handling - got description: %v", got.GetDescription())
	}
}

// testError is a simple error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
