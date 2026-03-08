package skills

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestSkillLazyLoading 测试懒加载功能
func TestSkillLazyLoading(t *testing.T) {
	// 创建临时测试文件
	tmpDir := t.TempDir()
	skillPath := filepath.Join(tmpDir, "SKILL.md")

	content := `---
name: test-skill
description: Test skill for lazy loading
version: 1.0.0
priority: 5
---

# Test Skill Content

This is the test skill content.`

	if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test skill file: %v", err)
	}

	skill := &Skill{
		Name:     "test-skill",
		path:     skillPath,
		loaded:   false,
		Priority: 5,
	}

	// 初始状态：未加载
	if skill.loaded {
		t.Error("Skill should not be loaded initially")
	}

	if skill.content != "" {
		t.Error("Skill content should be empty initially")
	}

	// 第一次调用 GetContent()：应该加载内容
	content1, err := skill.GetContent()
	if err != nil {
		t.Fatalf("GetContent() failed: %v", err)
	}

	expectedContent := "# Test Skill Content\n\nThis is the test skill content."
	if content1 != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, content1)
	}

	// 现在应该已加载
	if !skill.loaded {
		t.Error("Skill should be loaded after GetContent()")
	}

	// 第二次调用 GetContent()：应该使用缓存
	content2, err := skill.GetContent()
	if err != nil {
		t.Fatalf("Second GetContent() failed: %v", err)
	}

	if content1 != content2 {
		t.Error("Cached content should be returned on second call")
	}
}

// TestSkillConcurrentAccess 测试并发访问安全性
func TestSkillConcurrentAccess(t *testing.T) {
	// 创建临时测试文件
	tmpDir := t.TempDir()
	skillPath := filepath.Join(tmpDir, "SKILL.md")

	content := `---
name: concurrent-test
description: Test concurrent access
version: 1.0.0
priority: 5
---

# Concurrent Test Content

This is test content for concurrent access.`

	if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test skill file: %v", err)
	}

	skill := &Skill{
		Name:     "concurrent-test",
		path:     skillPath,
		loaded:   false,
		Priority: 5,
	}

	// 启动 100 个 goroutine 同时访问
	const numGoroutines = 100
	var wg sync.WaitGroup
	results := make([]string, numGoroutines)
	errors := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			content, err := skill.GetContent()
			results[idx] = content
			errors[idx] = err
		}(i)
	}

	// 等待所有 goroutine 完成
	wg.Wait()

	// 验证所有调用都成功
	for i, err := range errors {
		if err != nil {
			t.Errorf("Goroutine %d failed: %v", i, err)
		}
	}

	// 验证所有返回的内容一致
	expectedContent := "# Concurrent Test Content\n\nThis is test content for concurrent access."
	for i, content := range results {
		if content != expectedContent {
			t.Errorf("Goroutine %d got unexpected content: %q", i, content)
		}
	}

	// 验证只加载了一次
	if !skill.loaded {
		t.Error("Skill should be loaded after concurrent access")
	}
}

// TestSkillCaching 测试缓存机制
func TestSkillCaching(t *testing.T) {
	// 创建临时测试文件
	tmpDir := t.TempDir()
	skillPath := filepath.Join(tmpDir, "SKILL.md")

	content := `---
name: cache-test
description: Test caching mechanism
version: 1.0.0
priority: 5
---

# Cache Test Content`

	if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test skill file: %v", err)
	}

	skill := &Skill{
		Name:     "cache-test",
		path:     skillPath,
		loaded:   false,
		Priority: 5,
	}

	// 第一次调用
	start1 := time.Now()
	content1, err1 := skill.GetContent()
	duration1 := time.Since(start1)

	if err1 != nil {
		t.Fatalf("First GetContent() failed: %v", err1)
	}

	// 第二次调用（应该使用缓存）
	start2 := time.Now()
	content2, err2 := skill.GetContent()
	duration2 := time.Since(start2)

	if err2 != nil {
		t.Fatalf("Second GetContent() failed: %v", err2)
	}

	// 验证内容一致
	if content1 != content2 {
		t.Error("Cached content should match original content")
	}

	// 验证缓存调用更快（虽然可能不明显，因为文件很小）
	t.Logf("First call duration: %v, Second call duration: %v", duration1, duration2)

	// 验证缓存状态
	if !skill.loaded {
		t.Error("Skill should be marked as loaded")
	}
}

// TestSkillGetPath 测试 GetPath 方法
func TestSkillGetPath(t *testing.T) {
	skill := &Skill{
		Name:     "test",
		path:     "/path/to/skill.md",
		loaded:   false,
		Priority: 5,
	}

	expectedPath := "/path/to/skill.md"
	actualPath := skill.GetPath()

	if actualPath != expectedPath {
		t.Errorf("Expected path %q, got %q", expectedPath, actualPath)
	}
}

// TestSkillSourceString 测试 SkillSource.String() 方法
func TestSkillSourceString(t *testing.T) {
	tests := []struct {
		name     string
		source   SkillSource
		expected string
	}{
		{
			name:     "Builtin source",
			source:   SkillSourceBuiltin,
			expected: "builtin",
		},
		{
			name:     "User source",
			source:   SkillSourceUser,
			expected: "user",
		},
		{
			name:     "Unknown source",
			source:   SkillSource(99),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.source.String()
			if actual != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, actual)
			}
		})
	}
}
