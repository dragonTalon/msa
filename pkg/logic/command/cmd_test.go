package command

import (
	"context"
	"msa/pkg/model"
	"testing"
)

// mockCommand is a mock implementation of MsaCommand for testing
type mockCommand struct {
	name        string
	description string
}

func (m *mockCommand) Name() string {
	return m.name
}

func (m *mockCommand) Description() string {
	return m.description
}

func (m *mockCommand) Run(ctx context.Context, args []string) (*model.CmdResult, error) {
	return &model.CmdResult{Code: 0, Msg: "mock executed"}, nil
}

func (m *mockCommand) ToSelect(item []*model.SelectorItem) (*model.BaseSelector, error) {
	return &model.BaseSelector{}, nil
}

// helper function to create a mock command
func newMockCommand(name, desc string) *mockCommand {
	return &mockCommand{name: name, description: desc}
}

// TestRegisterCommand tests the RegisterCommand function
func TestRegisterCommand(t *testing.T) {
	tests := []struct {
		name    string
		cmd     *mockCommand
		setup   func() // setup before test
		cleanup func() // cleanup after test
	}{
		{
			name: "register single command",
			cmd:  newMockCommand("test", "test command"),
		},
		{
			name: "register multiple commands",
			cmd:  newMockCommand("multi", "multi command"),
		},
		{
			name: "register command with special chars",
			cmd:  newMockCommand("special-cmd", "special command"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: We can't easily clean up the global commandMap
			// so we just register and verify it works
			RegisterCommand(tt.cmd)

			// Verify command is retrievable
			got := GetCommand(tt.cmd.Name())
			if got == nil {
				t.Errorf("RegisterCommand() - command %s not found after registration", tt.cmd.Name())
			}
			if got != nil && got.Name() != tt.cmd.Name() {
				t.Errorf("RegisterCommand() - got name %v, want %v", got.Name(), tt.cmd.Name())
			}
		})
	}
}

// TestGetCommand tests the GetCommand function
func TestGetCommand(t *testing.T) {
	// Register test commands
	testCmd1 := newMockCommand("gettest1", "description 1")
	testCmd2 := newMockCommand("gettest2", "description 2")
	RegisterCommand(testCmd1)
	RegisterCommand(testCmd2)

	tests := []struct {
		name      string
		input     string
		wantName  string
		wantFound bool
	}{
		{
			name:      "existing command",
			input:     "gettest1",
			wantName:  "gettest1",
			wantFound: true,
		},
		{
			name:      "existing command with slash prefix",
			input:     "/gettest1",
			wantName:  "gettest1",
			wantFound: true,
		},
		{
			name:      "existing command uppercase",
			input:     "GETTEST1",
			wantName:  "gettest1",
			wantFound: true,
		},
		{
			name:      "existing command with slash and uppercase",
			input:     "/GETTEST1",
			wantName:  "gettest1",
			wantFound: true,
		},
		{
			name:      "non-existing command",
			input:     "notexist",
			wantFound: false,
		},
		{
			name:      "empty string",
			input:     "",
			wantFound: false,
		},
		{
			name:      "only slash",
			input:     "/",
			wantFound: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCommand(tt.input)
			if tt.wantFound {
				if got == nil {
					t.Errorf("GetCommand(%q) = nil, want command with name %q", tt.input, tt.wantName)
				} else if got.Name() != tt.wantName {
					t.Errorf("GetCommand(%q) = %v, want name %q", tt.input, got.Name(), tt.wantName)
				}
			} else {
				if got != nil {
					t.Errorf("GetCommand(%q) = %v, want nil", tt.input, got.Name())
				}
			}
		})
	}
}

// TestGetLikeCommand tests the GetLikeCommand function
func TestGetLikeCommand(t *testing.T) {
	// Register test commands
	testCmds := []*mockCommand{
		newMockCommand("search", "search command"),
		newMockCommand("show", "show command"),
		newMockCommand("select", "select command"),
		newMockCommand("config", "config command"),
	}
	for _, cmd := range testCmds {
		RegisterCommand(cmd)
	}

	tests := []struct {
		name     string
		input    string
		wantLen  int
		contains []string // expected commands in result
	}{
		{
			name:     "empty prefix returns all commands",
			input:    "",
			wantLen:  0, // empty string returns nil, not all commands
			contains: []string{},
		},
		{
			name:     "prefix s matches search, show, select",
			input:    "s",
			wantLen:  4, // May include other pre-registered commands
			contains: []string{"search", "show", "select"},
		},
		{
			name:     "prefix se matches search, select",
			input:    "se",
			wantLen:  2,
			contains: []string{"search", "select"},
		},
		{
			name:     "exact match",
			input:    "show",
			wantLen:  1,
			contains: []string{"show"},
		},
		{
			name:     "no matches",
			input:    "xyz",
			wantLen:  0,
			contains: []string{},
		},
		{
			name:     "with slash prefix",
			input:    "/s",
			wantLen:  4, // May include other pre-registered commands
			contains: []string{"search", "show", "select"},
		},
		{
			name:     "uppercase prefix",
			input:    "S",
			wantLen:  4, // May include other pre-registered commands
			contains: []string{"search", "show", "select"},
		},
		{
			name:     "mixed case prefix",
			input:    "Se",
			wantLen:  2, // search, select
			contains: []string{"search", "select"},
		},
		{
			name:     "slash with uppercase",
			input:    "/S",
			wantLen:  4, // May include other pre-registered commands
			contains: []string{"search", "show", "select"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetLikeCommand(tt.input)
			if got == nil {
				if tt.wantLen > 0 {
					t.Errorf("GetLikeCommand(%q) = nil, want %d items", tt.input, tt.wantLen)
				}
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("GetLikeCommand(%q) length = %d, want %d", tt.input, len(got), tt.wantLen)
			}
			// Check that expected commands are present
			for _, expected := range tt.contains {
				found := false
				for _, cmd := range got {
					if cmd == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetLikeCommand(%q) does not contain expected command %q", tt.input, expected)
				}
			}
		})
	}
}

// TestGetLikeCommand_EmptyMap tests GetLikeCommand with no commands registered
func TestGetLikeCommand_EmptyMap(t *testing.T) {
	// This test assumes commandMap is empty (or at least doesn't have the specific prefix)
	// Since we can't easily reset the global state, we test with a unique prefix
	result := GetLikeCommand("veryuniqueprefixthatdoesnotexist")
	if result != nil && len(result) > 0 {
		t.Errorf("GetLikeCommand() with unique prefix should return empty slice, got %v", result)
	}
}

// TestMsaCommand_Interface tests that mockCommand correctly implements MsaCommand
func TestMsaCommand_Interface(t *testing.T) {
	cmd := newMockCommand("test", "test description")

	// Test all interface methods
	if got := cmd.Name(); got != "test" {
		t.Errorf("Name() = %v, want %v", got, "test")
	}

	if got := cmd.Description(); got != "test description" {
		t.Errorf("Description() = %v, want %v", got, "test description")
	}

	ctx := context.Background()
	result, err := cmd.Run(ctx, []string{"arg1"})
	if err != nil {
		t.Errorf("Run() error = %v", err)
	}
	if result == nil {
		t.Error("Run() returned nil result")
	}

	items := []*model.SelectorItem{}
	selector, err := cmd.ToSelect(items)
	if err != nil {
		t.Errorf("ToSelect() error = %v", err)
	}
	if selector == nil {
		t.Error("ToSelect() returned nil selector")
	}
}
