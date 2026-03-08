package skills

import (
	"os"
	"path/filepath"
	"testing"
)

// TestIntegrationFullWorkflow 测试完整的工作流程
func TestIntegrationFullWorkflow(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	// 创建内置 skills 目录
	builtinDir := filepath.Join(tmpDir, "builtin")
	if err := os.Mkdir(builtinDir, 0755); err != nil {
		t.Fatalf("Failed to create builtin dir: %v", err)
	}

	// 创建用户 skills 目录
	userDir := filepath.Join(tmpDir, "user")
	if err := os.Mkdir(userDir, 0755); err != nil {
		t.Fatalf("Failed to create user dir: %v", err)
	}

	// 创建一些测试 skills
	skills := []struct {
		dir      string
		name     string
		content  string
		priority int
	}{
		{
			dir:  builtinDir,
			name: "base",
			content: `---
name: base
description: Base system rules
version: 1.0.0
priority: 10
---

# Base System Rules

This is the base system content.`,
			priority: 10,
		},
		{
			dir:  builtinDir,
			name: "stock",
			content: `---
name: stock
description: Stock analysis
version: 1.0.0
priority: 8
---

# Stock Analysis

Stock analysis content.`,
			priority: 8,
		},
		{
			dir:  userDir,
			name: "custom",
			content: `---
name: custom
description: Custom user skill
version: 1.0.0
priority: 6
---

# Custom Skill

Custom skill content.`,
			priority: 6,
		},
	}

	for _, skillInfo := range skills {
		skillSubDir := filepath.Join(skillInfo.dir, skillInfo.name)
		if err := os.Mkdir(skillSubDir, 0755); err != nil {
			t.Fatalf("Failed to create skill dir: %v", err)
		}

		skillPath := filepath.Join(skillSubDir, "SKILL.md")
		if err := os.WriteFile(skillPath, []byte(skillInfo.content), 0644); err != nil {
			t.Fatalf("Failed to write skill file: %v", err)
		}
	}

	// 创建 registry 和 loader
	registry := NewRegistry()
	loader := NewLoader(builtinDir, userDir, registry)

	// 加载所有 skills
	if err := loader.LoadAll(); err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	// 验证所有 skills 被加载
	allSkills := registry.ListAll()
	if len(allSkills) != 3 {
		t.Fatalf("Expected 3 skills, got %d", len(allSkills))
	}

	// 验证排序（按优先级降序）
	expectedOrder := []string{"base", "stock", "custom"}
	for i, expectedName := range expectedOrder {
		if allSkills[i].Name != expectedName {
			t.Errorf("At index %d: expected %s, got %s", i, expectedName, allSkills[i].Name)
		}
	}

	// 测试 Builder
	builder := NewBuilder(registry)
	testVars := PromptVars{
		Time:    "2025-03-08 18:00:00",
		Weekday: "2025年03月08日 星期六",
		Role:    "专业股票分析助手",
		Style:   "理性、专业、客观且严谨",
	}
	prompt, err := builder.BuildSystemPrompt([]string{"base", "stock"}, testVars)
	if err != nil {
		t.Fatalf("BuildSystemPrompt failed: %v", err)
	}

	// 验证 prompt 包含关键内容
	requiredStrings := []string{
		"【系统配置】",
		"## Skill List",
		"- name: base",
		"- name: stock",
	}

	for _, required := range requiredStrings {
		if !containsString(prompt, required) {
			t.Errorf("Prompt should contain: %s", required)
		}
	}

	// 测试 Manager
	manager := &Manager{
		registry: registry,
		loader:   loader,
	}

	// 测试获取 skill
	baseSkill, err := manager.GetSkill("base")
	if err != nil || baseSkill == nil {
		t.Error("Failed to get base skill")
	}

	if baseSkill.Priority != 10 {
		t.Errorf("Expected priority 10, got %d", baseSkill.Priority)
	}

	// 测试列出 skills
	listedSkills := manager.ListSkills()
	if len(listedSkills) != 3 {
		t.Errorf("Expected 3 listed skills, got %d", len(listedSkills))
	}
}

// TestIntegrationInvalidSkillFiles 测试包含无效 skill 文件的场景
func TestIntegrationInvalidSkillFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建一个有效的 skill
	validSkillDir := filepath.Join(tmpDir, "valid")
	if err := os.Mkdir(validSkillDir, 0755); err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}

	validSkill := `---
name: valid
description: Valid skill
version: 1.0.0
priority: 5
---

Valid content`

	if err := os.WriteFile(filepath.Join(validSkillDir, "SKILL.md"), []byte(validSkill), 0644); err != nil {
		t.Fatalf("Failed to write valid skill: %v", err)
	}

	// 创建一个无效的 skill（格式错误）
	invalidSkillDir := filepath.Join(tmpDir, "invalid")
	if err := os.Mkdir(invalidSkillDir, 0755); err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}

	invalidSkill := `---
name: invalid
description: Invalid skill
version: 1.0.0
priority: invalid_number
---

Content`

	if err := os.WriteFile(filepath.Join(invalidSkillDir, "SKILL.md"), []byte(invalidSkill), 0644); err != nil {
		t.Fatalf("Failed to write invalid skill: %v", err)
	}

	// 加载
	registry := NewRegistry()
	loader := NewLoader(tmpDir, "", registry)

	err := loader.LoadAll()
	// 不应该返回错误
	if err != nil {
		t.Errorf("LoadAll should not return error even with invalid skills, got: %v", err)
	}

	// 只有有效的 skill 应该被加载
	allSkills := registry.ListAll()
	if len(allSkills) != 1 {
		t.Errorf("Expected 1 valid skill, got %d", len(allSkills))
	}

	if len(allSkills) > 0 && allSkills[0].Name != "valid" {
		t.Errorf("Expected 'valid' skill, got '%s'", allSkills[0].Name)
	}
}

// TestIntegrationSkillSourceTracking 测试技能来源追踪
func TestIntegrationSkillSourceTracking(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建内置 skill
	builtinDir := filepath.Join(tmpDir, "builtin")
	builtinSkillDir := filepath.Join(builtinDir, "builtin-skill")
	if err := os.MkdirAll(builtinSkillDir, 0755); err != nil {
		t.Fatalf("Failed to create builtin dir: %v", err)
	}

	builtinContent := `---
name: builtin-skill
description: Builtin skill
version: 1.0.0
priority: 5
---

Builtin content`

	if err := os.WriteFile(filepath.Join(builtinSkillDir, "SKILL.md"), []byte(builtinContent), 0644); err != nil {
		t.Fatalf("Failed to write builtin skill: %v", err)
	}

	// 创建用户 skill
	userDir := filepath.Join(tmpDir, "user")
	userSkillDir := filepath.Join(userDir, "user-skill")
	if err := os.MkdirAll(userSkillDir, 0755); err != nil {
		t.Fatalf("Failed to create user dir: %v", err)
	}

	userContent := `---
name: user-skill
description: User skill
version: 1.0.0
priority: 5
---

User content`

	if err := os.WriteFile(filepath.Join(userSkillDir, "SKILL.md"), []byte(userContent), 0644); err != nil {
		t.Fatalf("Failed to write user skill: %v", err)
	}

	// 加载
	registry := NewRegistry()
	loader := NewLoader(builtinDir, userDir, registry)

	if err := loader.LoadAll(); err != nil {
		t.Fatalf("LoadAll failed: %v", err)
	}

	allSkills := registry.ListAll()

	// 验证来源正确
	builtinFound := false
	userFound := false

	for _, skill := range allSkills {
		if skill.Name == "builtin-skill" {
			if skill.Source != SkillSourceBuiltin {
				t.Errorf("Expected builtin source, got %v", skill.Source)
			}
			builtinFound = true
		}
		if skill.Name == "user-skill" {
			if skill.Source != SkillSourceUser {
				t.Errorf("Expected user source, got %v", skill.Source)
			}
			userFound = true
		}
	}

	if !builtinFound {
		t.Error("Builtin skill not found")
	}

	if !userFound {
		t.Error("User skill not found")
	}
}

// TestIntegrationLazyLoadingWithRealFiles 测试真实文件的懒加载
func TestIntegrationLazyLoadingWithRealFiles(t *testing.T) {
	tmpDir := t.TempDir()

	skillDir := filepath.Join(tmpDir, "test-skill")
	if err := os.Mkdir(skillDir, 0755); err != nil {
		t.Fatalf("Failed to create skill dir: %v", err)
	}

	// 创建大文件内容来测试懒加载的效果
	largeContent := `---
name: large-skill
description: Large skill for testing lazy loading
version: 1.0.0
priority: 5
---

# Large Content

`

	// 添加大量内容
	for i := 0; i < 100; i++ {
		largeContent += "This is line " + string(rune('0'+i%10)) + " of the large skill content.\n"
	}

	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(largeContent), 0644); err != nil {
		t.Fatalf("Failed to write large skill: %v", err)
	}

	// 创建 skill 但不加载内容
	skill := &Skill{
		Name:   "large-skill",
		path:   filepath.Join(skillDir, "SKILL.md"),
		Source: SkillSourceBuiltin,
	}

	// 验证初始状态
	if skill.loaded {
		t.Error("Skill should not be loaded initially")
	}

	// 获取内容
	content, err := skill.GetContent()
	if err != nil {
		t.Fatalf("GetContent failed: %v", err)
	}

	// 验证内容被加载
	if !skill.loaded {
		t.Error("Skill should be loaded after GetContent")
	}

	// 验证内容不为空
	if content == "" {
		t.Error("Content should not be empty")
	}

	// 验证内容包含预期文本
	if !containsString(content, "Large Content") {
		t.Error("Content should contain 'Large Content'")
	}
}

// 辅助函数
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
