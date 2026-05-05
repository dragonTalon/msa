package skills

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
	"unsafe"
)

// setDirPath 使用 unsafe 设置 Skill 的私有 dirPath 字段
func setDirPath(skill *Skill, dirPath string) {
	// 获取 Skill 结构体的反射值
	val := unsafe.Pointer(skill)
	// 获取 dirPath 字段的偏移量
	dirPathOffset := unsafe.Offsetof(skill.dirPath)
	// 获取 dirPath 字段的指针
	dirPathPtr := (*string)(unsafe.Pointer(uintptr(val) + dirPathOffset))
	// 设置值
	*dirPathPtr = dirPath
}

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
		Priority: 5,
	}
	setDirPath(skill, tmpDir)

	// 第一次调用 GetContent()：应该加载内容
	content1, err := skill.GetContent()
	if err != nil {
		t.Fatalf("GetContent() failed: %v", err)
	}

	expectedContent := "# Test Skill Content\n\nThis is the test skill content."
	if content1 != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, content1)
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
		Priority: 5,
	}
	setDirPath(skill, tmpDir)

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
		Priority: 5,
	}
	setDirPath(skill, tmpDir)

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
}

// TestSkillGetDirPath 测试 GetDirPath 方法
func TestSkillGetDirPath(t *testing.T) {
	skill := &Skill{
		Name:     "test",
		Priority: 5,
	}
	setDirPath(skill, "/path/to/skill")

	expectedPath := "/path/to/skill"
	actualPath := skill.GetDirPath()

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

// TestExtractBody 测试 extractFrontmatterAndBody 函数
func TestExtractBody(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Content with YAML frontmatter",
			input:    "---\nname: test\npriority: 5\n---\n# Title\nContent here",
			expected: "# Title\nContent here",
		},
		{
			name:     "Content without YAML frontmatter",
			input:    "# Title\nNo frontmatter here",
			expected: "# Title\nNo frontmatter here",
		},
		{
			name:     "Only YAML frontmatter, no body",
			input:    "---\nname: test\n---\n",
			expected: "",
		},
		{
			name:     "Empty content",
			input:    "",
			expected: "",
		},
		{
			name:     "Only frontmatter markers - no second marker",
			input:    "------\n",
			expected: "------\n", // 只有第一个 ---，找不到第二个，返回整个内容
		},
		{
			name:     "No closing frontmatter",
			input:    "---\nname: test\npriority: 5\n# Content without closing",
			expected: "---\nname: test\npriority: 5\n# Content without closing",
		},
		{
			name:     "Body with CRLF line endings",
			input:    "---\r\nname: test\r\n---\r\n# Title\r\nContent",
			expected: "# Title\r\nContent",
		},
		{
			name:     "Body with multiple newlines after frontmatter",
			input:    "---\nname: test\n---\n\n\n# Title\n\nContent",
			expected: "# Title\n\nContent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, result := extractFrontmatterAndBody([]byte(tt.input))
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
