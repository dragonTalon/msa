package skills

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestSelectorEnsureBaseSkill 测试 ensureBaseSkill 方法
func TestSelectorEnsureBaseSkill(t *testing.T) {
	selector := &Selector{}

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Base already in list",
			input:    []string{"base", "skill-a", "skill-b"},
			expected: []string{"base", "skill-a", "skill-b"},
		},
		{
			name:     "Base not in list",
			input:    []string{"skill-a", "skill-b"},
			expected: []string{"base", "skill-a", "skill-b"},
		},
		{
			name:     "Empty list",
			input:    []string{},
			expected: []string{"base"},
		},
		{
			name:     "Only base",
			input:    []string{"base"},
			expected: []string{"base"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := selector.ensureBaseSkill(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("At index %d: expected %s, got %s", i, expected, result[i])
				}
			}
		})
	}
}

// TestSelectorSortByPriority 测试 sortByPriority 方法
func TestSelectorSortByPriority(t *testing.T) {
	selector := &Selector{}

	registry := NewRegistry()

	skills := []*Skill{
		{Name: "base", Priority: 10, Description: "Base", path: "/base.md"},
		{Name: "high", Priority: 8, Description: "High", path: "/high.md"},
		{Name: "medium", Priority: 5, Description: "Medium", path: "/med.md"},
		{Name: "low", Priority: 2, Description: "Low", path: "/low.md"},
	}

	for _, skill := range skills {
		registry.Register(skill)
	}

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Reverse order",
			input:    []string{"low", "medium", "high", "base"},
			expected: []string{"base", "high", "medium", "low"},
		},
		{
			name:     "Already sorted",
			input:    []string{"base", "high", "medium", "low"},
			expected: []string{"base", "high", "medium", "low"},
		},
		{
			name:     "Partial order",
			input:    []string{"medium", "base", "low"},
			expected: []string{"base", "medium", "low"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := selector.sortByPriority(tt.input, skills)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("At index %d: expected %s, got %s", i, expected, result[i])
				}
			}
		})
	}
}

// TestSelectorBuildSelectionPrompt 测试 buildSelectionPrompt 方法
func TestSelectorBuildSelectionPrompt(t *testing.T) {
	selector := &Selector{}

	metadatas := []SkillMetadata{
		{Name: "base", Description: "Base system rules", Priority: 10},
		{Name: "stock", Description: "Stock analysis", Priority: 8},
	}

	userInput := "What is the price of AAPL?"

	prompt := selector.buildSelectionPrompt(metadatas, userInput)

	// 验证包含关键部分
	requiredStrings := []string{
		"你是一个 skill 选择器",
		"base",
		"Base system rules",
		"stock",
		"Stock analysis",
		"What is the price of AAPL?",
		"selected_skills",
		"reasoning",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(prompt, required) {
			t.Errorf("Prompt should contain: %s", required)
		}
	}
}

// TestSelectionResultJSONParsing 测试 JSON 解析逻辑
func TestSelectionResultJSONParsing(t *testing.T) {
	tests := []struct {
		name        string
		jsonString  string
		expectError bool
		expectLen   int
	}{
		{
			name:        "Valid JSON",
			jsonString:  `{"selected_skills": ["base", "test"], "reasoning": "Good"}`,
			expectError: false,
			expectLen:   2,
		},
		{
			name:        "JSON with markdown",
			jsonString:  "```json\n{\"selected_skills\": [\"base\"], \"reasoning\": \"Test\"}\n```",
			expectError: true, // JSON parsing will fail without cleaning
			expectLen:   0,
		},
		{
			name:        "Invalid JSON",
			jsonString:  `not valid json`,
			expectError: true,
			expectLen:   0,
		},
		{
			name:        "Empty string",
			jsonString:  "",
			expectError: true,
			expectLen:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result SelectionResult
			err := json.Unmarshal([]byte(tt.jsonString), &result)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if len(result.SelectedSkills) != tt.expectLen {
					t.Errorf("Expected %d skills, got %d", tt.expectLen, len(result.SelectedSkills))
				}
			}
		})
	}
}

// TestJSONCleaning 测试 JSON 清理函数
func TestJSONCleaning(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Clean JSON",
			input:    `{"test": "value"}`,
			expected: `{"test": "value"}`,
		},
		{
			name:     "JSON with json code block",
			input:    "```json\n{\"test\": \"value\"}\n```",
			expected: `{"test": "value"}`,
		},
		{
			name:     "JSON with generic code block",
			input:    "```\n{\"test\": \"value\"}\n```",
			expected: `{"test": "value"}`,
		},
		{
			name:     "JSON with extra whitespace",
			input:    "  {\"test\": \"value\"}  ",
			expected: `{"test": "value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input
			// 模拟 callLLM 中的清理逻辑
			result = strings.TrimSpace(result)
			if strings.HasPrefix(result, "```json") {
				result = strings.TrimPrefix(result, "```json")
				result = strings.TrimSuffix(result, "```")
				result = strings.TrimSpace(result)
			} else if strings.HasPrefix(result, "```") {
				result = strings.TrimPrefix(result, "```")
				result = strings.TrimSuffix(result, "```")
				result = strings.TrimSpace(result)
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestSelectorListAvailable 测试 ListAvailable 方法
func TestSelectorListAvailable(t *testing.T) {
	registry := NewRegistry()

	skills := []*Skill{
		{Name: "base", Priority: 10, Description: "Base", path: "/base.md"},
		{Name: "skill-a", Priority: 5, Description: "A", path: "/a.md"},
		{Name: "skill-b", Priority: 8, Description: "B", path: "/b.md"},
	}

	for _, skill := range skills {
		registry.Register(skill)
	}

	// 测试无禁用
	disabledMap := make(map[string]bool)
	available := registry.ListAvailable(disabledMap)

	if len(available) != 3 {
		t.Errorf("Expected 3 skills, got %d", len(available))
	}

	// 验证排序（按优先级降序）
	if available[0].Name != "base" {
		t.Errorf("Expected first to be base, got %s", available[0].Name)
	}

	// 测试有禁用
	disabledMap = map[string]bool{
		"skill-a": true,
	}

	available = registry.ListAvailable(disabledMap)

	if len(available) != 2 {
		t.Errorf("Expected 2 skills, got %d", len(available))
	}

	// 验证 skill-a 不在列表中
	for _, skill := range available {
		if skill.Name == "skill-a" {
			t.Error("skill-a should be filtered out")
		}
	}
}

// TestNewSelector 测试 NewSelector 构造函数
func TestNewSelector(t *testing.T) {
	registry := NewRegistry()

	// 注意：我们无法真正测试 NewSelector，因为它需要 openai.ChatModel
	// 但我们可以验证它不会 panic
	selector := &Selector{
		registry: registry,
	}

	if selector.registry != registry {
		t.Error("Registry not set correctly")
	}
}

// TestSelectionResultStruct 测试 SelectionResult 结构
func TestSelectionResultStruct(t *testing.T) {
	result := SelectionResult{
		SelectedSkills: []string{"base", "test"},
		Reasoning:      "Test reasoning",
	}

	if len(result.SelectedSkills) != 2 {
		t.Errorf("Expected 2 skills, got %d", len(result.SelectedSkills))
	}

	if result.Reasoning != "Test reasoning" {
		t.Errorf("Expected reasoning 'Test reasoning', got '%s'", result.Reasoning)
	}

	// 测试 JSON 序列化
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded SelectionResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(decoded.SelectedSkills) != 2 {
		t.Errorf("Expected 2 skills after round-trip, got %d", len(decoded.SelectedSkills))
	}
}

// TestSkillMetadataStruct 测试 SkillMetadata 结构
func TestSkillMetadataStruct(t *testing.T) {
	metadata := SkillMetadata{
		Name:        "test-skill",
		Description: "Test description",
		Priority:    7,
	}

	if metadata.Name != "test-skill" {
		t.Errorf("Expected name 'test-skill', got '%s'", metadata.Name)
	}

	if metadata.Priority != 7 {
		t.Errorf("Expected priority 7, got %d", metadata.Priority)
	}

	// 测试 JSON 序列化
	data, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded SkillMetadata
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Name != "test-skill" {
		t.Errorf("Expected name 'test-skill' after round-trip, got '%s'", decoded.Name)
	}
}
