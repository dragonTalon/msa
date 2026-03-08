package config

import (
	"testing"

	"github.com/charmbracelet/bubbletea"
)

// TestConfigModel_Init 测试配置模型初始化
func TestConfigModel_Init(t *testing.T) {
	model := NewConfigModel()

	if model == nil {
		t.Fatal("NewConfigModel() returned nil")
	}

	if model.state == nil {
		t.Error("ConfigModel.state should not be nil")
	}

	if len(model.state.Items) == 0 {
		t.Error("ConfigModel should have items")
	}

	if model.state.SelectedIndex != 0 {
		t.Errorf("SelectedIndex should be 0, got %d", model.state.SelectedIndex)
	}
}

// TestConfigModel_UpdateNavigation 测试导航
func TestConfigModel_UpdateNavigation(t *testing.T) {
	model := NewConfigModel()

	// 测试向下移动
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	m, ok := newModel.(*ConfigModel)
	if !ok {
		t.Fatal("Update should return *ConfigModel")
	}
	if m.state.SelectedIndex != 1 {
		t.Errorf("After KeyDown, SelectedIndex should be 1, got %d", m.state.SelectedIndex)
	}

	// 测试向上移动
	model = m
	newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	m, ok = newModel.(*ConfigModel)
	if !ok {
		t.Fatal("Update should return *ConfigModel")
	}
	if m.state.SelectedIndex != 0 {
		t.Errorf("After KeyUp, SelectedIndex should be 0, got %d", m.state.SelectedIndex)
	}
}

// TestConfigModel_Items 测试配置项
func TestConfigModel_Items(t *testing.T) {
	model := NewConfigModel()

	// 检查是否有必需的配置项
	hasProvider := false
	hasAPIKey := false
	hasBaseURL := false

	for _, item := range model.state.Items {
		if item.Key == "provider" {
			hasProvider = true
		}
		if item.Key == "apikey" {
			hasAPIKey = true
		}
		if item.Key == "baseurl" {
			hasBaseURL = true
		}
	}

	if !hasProvider {
		t.Error("ConfigModel should have provider item")
	}
	if !hasAPIKey {
		t.Error("ConfigModel should have apikey item")
	}
	if !hasBaseURL {
		t.Error("ConfigModel should have baseurl item")
	}
}

// TestEditorModel_Init 测试编辑器初始化
func TestEditorModel_Init(t *testing.T) {
	editor := NewEditorModel("Test Title", "Test Label", "test value", false)

	if editor.title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got '%s'", editor.title)
	}

	if editor.label != "Test Label" {
		t.Errorf("Expected label 'Test Label', got '%s'", editor.label)
	}

	if editor.value != "test value" {
		t.Errorf("Expected value 'test value', got '%s'", editor.value)
	}

	if editor.IsAborted() {
		t.Error("New editor should not be aborted")
	}
}

// TestProviderSelectorModel_Init 测试 Provider 选择器初始化
func TestProviderSelectorModel_Init(t *testing.T) {
	selector := NewProviderSelectorModel("siliconflow")

	if selector == nil {
		t.Fatal("NewProviderSelectorModel() returned nil")
	}

	if len(selector.providers) == 0 {
		t.Error("ProviderSelector should have providers")
	}

	if selector.index < 0 || selector.index >= len(selector.providers) {
		t.Errorf("Invalid index: %d", selector.index)
	}
}

// TestStyles_Exist 测试样式定义
func TestStyles_Exist(t *testing.T) {
	// 测试样式是否定义（编译时检查）
	_ = TitleStyle
	_ = LabelStyle
	_ = ValueStyle
	_ = SelectedStyle
	_ = ErrorStyle
	_ = WarningStyle
	_ = SuccessStyle
	_ = HelpStyle
}

// TestMaskAPIKey 测试 API Key 隐藏
func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		expected string
	}{
		{
			name:     "short key",
			apiKey:   "sk-123",
			expected: "sk-123",
		},
		{
			name:     "normal key",
			apiKey:   "sk-1234567890abcdefghijklmnop",
			expected: "sk-1...mnop",
		},
		{
			name:     "exact 8 chars",
			apiKey:   "sk-12345",
			expected: "sk-12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskAPIKey(tt.apiKey)
			if result != tt.expected {
				t.Errorf("MaskAPIKey(%q) = %q, want %q", tt.apiKey, result, tt.expected)
			}
		})
	}
}
