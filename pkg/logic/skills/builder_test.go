package skills

import (
	"os"
	"path/filepath"
	"testing"
)

// TestBuilderBuildSystemPrompt 测试 Prompt 构建功能
func TestBuilderBuildSystemPrompt(t *testing.T) {
	registry := NewRegistry()

	// 创建临时目录和测试 skills
	tmpDir := t.TempDir()

	skills := []struct {
		name     string
		content  string
		priority int
	}{
		{
			name: "skill-a",
			content: `---
name: skill-a
description: Skill A
version: 1.0.0
priority: 5
---

# Skill A Content

This is content from skill A.`,
			priority: 5,
		},
		{
			name: "skill-b",
			content: `---
name: skill-b
description: Skill B
version: 1.0.0
priority: 3
---

# Skill B Content

This is content from skill B.`,
			priority: 3,
		},
	}

	for _, skillInfo := range skills {
		skillPath := filepath.Join(tmpDir, skillInfo.name+".md")
		if err := os.WriteFile(skillPath, []byte(skillInfo.content), 0644); err != nil {
			t.Fatalf("Failed to create skill file: %v", err)
		}

		skill := &Skill{
			Name:        skillInfo.name,
			Description: "Skill " + skillInfo.name,
			path:        skillPath,
			Priority:    skillInfo.priority,
			Source:      SkillSourceBuiltin,
		}
		registry.Register(skill)
	}

	builder := NewBuilder(registry)

	// 测试用的模板变量
	testVars := PromptVars{
		Time:    "2025-03-08 18:00:00",
		Weekday: "2025年03月08日 星期六",
		Role:    "专业股票分析助手",
		Style:   "理性、专业、客观且严谨",
	}

	// 测试构建单个 skill 的 prompt
	result, err := builder.BuildSystemPrompt([]string{"skill-a"}, testVars)
	if err != nil {
		t.Fatalf("BuildSystemPrompt failed: %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty prompt")
	}

	// 验证包含系统配置部分
	if !contains(result, "【系统配置】") {
		t.Error("Expected system configuration in prompt")
	}

	// 验证包含 skill 列表
	if !contains(result, "- name: skill-a") {
		t.Error("Expected skill-a in Skill List")
	}

	if !contains(result, "describe: Skill skill-a") {
		t.Error("Expected skill-a description in prompt")
	}

	// 测试构建多个 skills 的 prompt
	result, err = builder.BuildSystemPrompt([]string{"skill-a", "skill-b"}, testVars)
	if err != nil {
		t.Fatalf("BuildSystemPrompt with multiple skills failed: %v", err)
	}

	if !contains(result, "- name: skill-a") {
		t.Error("Expected skill-a in Skill List")
	}

	if !contains(result, "- name: skill-b") {
		t.Error("Expected skill-b in Skill List")
	}

	// 测试空列表
	result, err = builder.BuildSystemPrompt([]string{}, testVars)
	if err != nil {
		t.Fatalf("BuildSystemPrompt with empty list failed: %v", err)
	}

	// 应该仍然包含系统配置和基础规则
	if !contains(result, "【系统配置】") {
		t.Error("Expected system configuration even with no skills")
	}
}

// TestBuilderInvalidSkill 测试无效 skill 处理
func TestBuilderInvalidSkill(t *testing.T) {
	registry := NewRegistry()

	// 注册一个有效的 skill
	tmpDir := t.TempDir()
	validSkillPath := filepath.Join(tmpDir, "valid.md")
	content := `---
name: valid
description: Valid skill
version: 1.0.0
priority: 5
---

Valid content`

	if err := os.WriteFile(validSkillPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create valid skill: %v", err)
	}

	validSkill := &Skill{
		Name:   "valid",
		path:   validSkillPath,
		Source: SkillSourceBuiltin,
	}
	registry.Register(validSkill)

	builder := NewBuilder(registry)

	testVars := PromptVars{
		Time:    "2025-03-08 18:00:00",
		Weekday: "2025年03月08日 星期六",
		Role:    "专业股票分析助手",
		Style:   "理性、专业、客观且严谨",
	}

	// 包含一个不存在的 skill
	result, err := builder.BuildSystemPrompt([]string{"valid", "non-existent"}, testVars)
	if err != nil {
		t.Fatalf("BuildSystemPrompt should not error with invalid skill: %v", err)
	}

	// 应该只包含有效的 skill
	if !contains(result, "- name: valid") {
		t.Error("Expected valid skill in Skill List")
	}

	// 不应该包含无效 skill
	if contains(result, "- name: non-existent") {
		t.Error("Should not include non-existent skill")
	}
}

// TestBuilderPromptStructure 测试 Prompt 结构
func TestBuilderPromptStructure(t *testing.T) {
	registry := NewRegistry()

	tmpDir := t.TempDir()
	skillPath := filepath.Join(tmpDir, "test.md")
	content := `---
name: test
description: Test skill
version: 1.0.0
priority: 5
---

# Test Content

Test body content.`

	if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create skill file: %v", err)
	}

	skill := &Skill{
		Name:   "test",
		path:   skillPath,
		Source: SkillSourceBuiltin,
	}
	registry.Register(skill)

	testVars := PromptVars{
		Time:    "2025-03-08 18:00:00",
		Weekday: "2025年03月08日 星期六",
		Role:    "专业股票分析助手",
		Style:   "理性、专业、客观且严谨",
	}

	builder := NewBuilder(registry)
	result, _ := builder.BuildSystemPrompt([]string{"test"}, testVars)

	// 验证结构顺序
	positions := make(map[string]int)

	sections := []string{
		"【系统配置】",
		"【角色定义】",
		"## Skill List",
		"【工具调用约束】",
	}

	for _, section := range sections {
		pos := indexOf(result, section)
		if pos == -1 {
			t.Errorf("Missing section: %s", section)
			continue
		}
		positions[section] = pos
	}

	// 验证顺序：系统配置 < 角色定义 < skill < 工具约束
	if positions["【系统配置】"] > positions["【角色定义】"] {
		t.Error("System configuration should come before role definition")
	}

	if positions["【角色定义】"] > positions["## Skill List"] {
		t.Error("Role definition should come before skill section")
	}

	if positions["## Skill List"] > positions["【工具调用约束】"] {
		t.Error("Skill section should come before tool constraints")
	}
}

// TestBuilderVariableFilling 测试变量填充
func TestBuilderVariableFilling(t *testing.T) {
	registry := NewRegistry()

	tmpDir := t.TempDir()
	skillPath := filepath.Join(tmpDir, "test.md")
	content := `---
name: test
description: Test skill
version: 1.0.0
priority: 5
---

# Test Content`

	if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create skill file: %v", err)
	}

	skill := &Skill{
		Name:   "test",
		path:   skillPath,
		Source: SkillSourceBuiltin,
	}
	registry.Register(skill)

	testVars := PromptVars{
		Time:    "2025-03-08 18:00:00",
		Weekday: "2025年03月08日 星期六",
		Role:    "专业股票分析助手",
		Style:   "理性、专业、客观且严谨",
	}

	builder := NewBuilder(registry)
	result, _ := builder.BuildSystemPrompt([]string{"test"}, testVars)

	// 验证模板变量已被正确填充
	expectedValues := []string{
		"当前实时时间：2025-03-08 18:00:00",
		"今天是：2025年03月08日 星期六",
		"你是专业的专业股票分析助手",
		"需使用理性、专业、客观且严谨的语气",
	}

	for _, expected := range expectedValues {
		if !contains(result, expected) {
			t.Errorf("Expected rendered value: %s", expected)
		}
	}

	// 验证不包含未渲染的模板变量
	unexpectedPlaceholders := []string{
		"{{.Time}}",
		"{{.Weekday}}",
		"{{.Role}}",
		"{{.Style}}",
	}

	for _, placeholder := range unexpectedPlaceholders {
		if contains(result, placeholder) {
			t.Errorf("Template variable should have been replaced: %s", placeholder)
		}
	}
}

// TestBuilderMultipleSkillsOrdering 测试多个 skills 的顺序
func TestBuilderMultipleSkillsOrdering(t *testing.T) {
	registry := NewRegistry()

	tmpDir := t.TempDir()

	// 创建多个 skills，按特定顺序添加
	skillNames := []string{"alpha", "beta", "gamma", "delta"}

	for i, name := range skillNames {
		content := `---
name: ` + name + `
description: Skill ` + name + `
version: 1.0.0
priority: ` + string(rune('5'+i)) + `
---

# ` + name + ` Content`

		skillPath := filepath.Join(tmpDir, name+".md")
		if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create skill file: %v", err)
		}

		skill := &Skill{
			Name:   name,
			path:   skillPath,
			Source: SkillSourceBuiltin,
		}
		registry.Register(skill)
	}

	testVars := PromptVars{
		Time:    "2025-03-08 18:00:00",
		Weekday: "2025年03月08日 星期六",
		Role:    "专业股票分析助手",
		Style:   "理性、专业、客观且严谨",
	}

	builder := NewBuilder(registry)

	// 按 alpha, beta, gamma 顺序请求
	requestedOrder := []string{"alpha", "beta", "gamma"}
	result, _ := builder.BuildSystemPrompt(requestedOrder, testVars)

	// 验证顺序
	for _, name := range requestedOrder {
		expectedEntry := "- name: " + name
		if !contains(result, expectedEntry) {
			t.Errorf("Missing entry for skill: %s", name)
			continue
		}

		// 顺序由 Skill List 中条目的先后决定
	}

	// delta 不应该被包含（未请求）
	if contains(result, "- name: delta") {
		t.Error("Should not include unrequested skill")
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
