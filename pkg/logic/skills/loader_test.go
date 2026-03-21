package skills

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoaderParseMetadata 测试 YAML frontmatter 解析
func TestLoaderParseMetadata(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
		expected    *skillMetadataYAML
	}{
		{
			name: "Valid frontmatter",
			content: `---
name: test-skill
description: A test skill
version: 1.0.0
priority: 7
---

Content here`,
			expectError: false,
			expected: &skillMetadataYAML{
				Name:        "test-skill",
				Description: "A test skill",
				Version:     "1.0.0",
				Priority:    7,
			},
		},
		{
			name: "Missing priority (should use default)",
			content: `---
name: test-skill
description: A test skill
version: 1.0.0
---

Content here`,
			expectError: false,
			expected: &skillMetadataYAML{
				Name:        "test-skill",
				Description: "A test skill",
				Version:     "1.0.0",
				Priority:    5, // 默认值
			},
		},
		{
			name: "Missing name (should error)",
			content: `---
description: A test skill
version: 1.0.0
---

Content here`,
			expectError: true,
			expected:    nil,
		},
		{
			name: "Missing description (should error)",
			content: `---
name: test-skill
version: 1.0.0
---

Content here`,
			expectError: true,
			expected:    nil,
		},
		{
			name: "Invalid YAML",
			content: `---
name: test-skill
description: A test skill
version: 1.0.0
priority: invalid
---

Content here`,
			expectError: true,
			expected:    nil,
		},
		{
			name:        "No frontmatter",
			content:     `Just content without frontmatter`,
			expectError: true,
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			skillPath := filepath.Join(tmpDir, "SKILL.md")

			if err := os.WriteFile(skillPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			loader := &Loader{}
			metadata, err := loader.parseSkillMetadata(skillPath)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if metadata.Name != tt.expected.Name {
				t.Errorf("Expected name %q, got %q", tt.expected.Name, metadata.Name)
			}

			if metadata.Description != tt.expected.Description {
				t.Errorf("Expected description %q, got %q", tt.expected.Description, metadata.Description)
			}

			if metadata.Version != tt.expected.Version {
				t.Errorf("Expected version %q, got %q", tt.expected.Version, metadata.Version)
			}

			if metadata.Priority != tt.expected.Priority {
				t.Errorf("Expected priority %d, got %d", tt.expected.Priority, metadata.Priority)
			}
		})
	}
}

// TestLoaderExtractBody 测试 frontmatter 提取
func TestLoaderExtractBody(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Valid frontmatter with content",
			input: `---
name: test
---

This is the content.`,
			expected: "This is the content.",
		},
		{
			name: "Frontmatter with leading newlines",
			input: `---

name: test
---

This is the content.`,
			expected: "This is the content.",
		},
		{
			name:     "No frontmatter",
			input:    `This is just content`,
			expected: "This is just content",
		},
		{
			name: "Empty content",
			input: `---
name: test
---

`,
			expected: "",
		},
		{
			name: "Multiline content",
			input: `---
name: test
---

Line 1
Line 2
Line 3`,
			expected: "Line 1\nLine 2\nLine 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractBody([]byte(tt.input))
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestLoaderErrorHandling 测试错误处理
func TestLoaderErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建子目录来存放 SKILL.md 文件
	skillDir := filepath.Join(tmpDir, "skill1")
	if err := os.Mkdir(skillDir, 0755); err != nil {
		t.Fatalf("Failed to create skill directory: %v", err)
	}

	// 创建一个有效的 skill
	validSkill := `---
name: valid-skill
description: A valid skill
version: 1.0.0
priority: 5
---

Valid content`

	validPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(validPath, []byte(validSkill), 0644); err != nil {
		t.Fatalf("Failed to create valid skill: %v", err)
	}

	// 创建另一个子目录来存放无效的 skill
	invalidDir := filepath.Join(tmpDir, "skill2")
	if err := os.Mkdir(invalidDir, 0755); err != nil {
		t.Fatalf("Failed to create invalid skill directory: %v", err)
	}

	// 创建一个无效的 skill（损坏的 YAML）
	invalidSkill := `---
name: invalid-skill
description: A skill with bad frontmatter
version: 1.0.0
priority: not-a-number
---

Content`

	invalidPath := filepath.Join(invalidDir, "SKILL.md")
	if err := os.WriteFile(invalidPath, []byte(invalidSkill), 0644); err != nil {
		t.Fatalf("Failed to create invalid skill: %v", err)
	}

	// 创建一个非 SKILL.md 文件（应该被忽略）
	otherFile := filepath.Join(tmpDir, "not-a-skill.txt")
	if err := os.WriteFile(otherFile, []byte("ignored"), 0644); err != nil {
		t.Fatalf("Failed to create other file: %v", err)
	}

	// 创建 registry 并加载
	registry := NewRegistry()
	loader := NewLoader(tmpDir, "", registry)

	if err := loader.LoadAll(); err != nil {
		// LoadAll 不应该返回错误，即使有无效的文件
		t.Errorf("LoadAll should not return error, got: %v", err)
	}

	// 验证只有有效的 skill 被加载
	allSkills := registry.ListAll()

	if len(allSkills) != 1 {
		t.Errorf("Expected 1 valid skill, got %d", len(allSkills))
	}

	if len(allSkills) > 0 && allSkills[0].Name != "valid-skill" {
		t.Errorf("Expected skill name 'valid-skill', got '%s'", allSkills[0].Name)
	}
}

// TestLoaderNonExistentDirectory 测试目录不存在的情况
func TestLoaderNonExistentDirectory(t *testing.T) {
	registry := NewRegistry()
	loader := NewLoader("/non/existent/directory", "", registry)

	// 不应该报错，只是没有加载任何 skill
	if err := loader.LoadAll(); err != nil {
		t.Errorf("LoadAll should not error for non-existent directory, got: %v", err)
	}

	allSkills := registry.ListAll()
	if len(allSkills) != 0 {
		t.Errorf("Expected 0 skills, got %d", len(allSkills))
	}
}

// TestLoaderPriorityDefault 测试优先级默认值
func TestLoaderPriorityDefault(t *testing.T) {
	content := `---
name: test-default-priority
description: Test default priority
version: 1.0.0
---

Content`

	tmpDir := t.TempDir()
	skillPath := filepath.Join(tmpDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := &Loader{}
	metadata, err := loader.parseSkillMetadata(skillPath)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if metadata.Priority != 5 {
		t.Errorf("Expected default priority 5, got %d", metadata.Priority)
	}
}
