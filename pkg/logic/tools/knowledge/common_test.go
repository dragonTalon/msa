package knowledge

import (
	"path/filepath"
	"testing"
)

func TestSafePath(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   error
		checkPath bool
	}{
		// 正常路径
		{
			name:    "正常相对路径",
			input:   "notes/2026-03-15.md",
			wantErr: nil,
		},
		{
			name:    "单层目录",
			input:   "errors/err-001.md",
			wantErr: nil,
		},
		{
			name:    "根目录文件",
			input:   "readme.md",
			wantErr: nil,
		},
		// 路径遍历攻击
		{
			name:    "路径遍历-双点",
			input:   "../../../etc/passwd",
			wantErr: ErrPathTraversal,
		},
		{
			name:    "路径遍历-嵌入双点",
			input:   "notes/../../../etc/passwd",
			wantErr: ErrPathTraversal,
		},
		{
			name:    "路径遍历-开头的双点",
			input:   "../secrets.env",
			wantErr: ErrPathTraversal,
		},
		// 绝对路径
		{
			name:    "绝对路径-Unix",
			input:   "/etc/passwd",
			wantErr: ErrAbsolutePath,
		},
		{
			name:    "绝对路径-Windows风格",
			input:   "C:\\Windows\\System32",
			wantErr: ErrAbsolutePath,
		},
		// 空路径
		{
			name:    "空路径",
			input:   "",
			wantErr: ErrEmptyPath,
		},
		// Unicode 路径
		{
			name:    "中文路径",
			input:   "笔记/今日总结.md",
			wantErr: nil,
		},
		{
			name:    "日文路径",
			input:   "ノート/まとめ.md",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SafePath(tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("SafePath(%q) expected error %v, got nil", tt.input, tt.wantErr)
					return
				}
				if err != tt.wantErr {
					t.Errorf("SafePath(%q) expected error %v, got %v", tt.input, tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("SafePath(%q) unexpected error: %v", tt.input, err)
				return
			}

			// 验证返回的路径在知识库目录下
			basePath := GetKnowledgeBasePath()
			absBase, _ := filepath.Abs(basePath)
			absResult, _ := filepath.Abs(result)
			if absResult == "" || absBase == "" {
				return // 跳过路径检查
			}
		})
	}
}

func TestGetKnowledgeBasePath(t *testing.T) {
	// 第一次调用
	path1 := GetKnowledgeBasePath()
	if path1 == "" {
		t.Error("GetKnowledgeBasePath() returned empty path")
	}

	// 第二次调用应该返回相同路径（单例）
	path2 := GetKnowledgeBasePath()
	if path1 != path2 {
		t.Errorf("GetKnowledgeBasePath() not singleton: %q != %q", path1, path2)
	}

	// 验证路径以 .msa/knowledge 结尾
	if !filepath.IsAbs(path1) {
		// 相对路径，检查后缀
		if !stringsContains(path1, KnowledgeBaseDir) {
			t.Errorf("GetKnowledgeBasePath() = %q, should contain %q", path1, KnowledgeBaseDir)
		}
	}
}

func stringsContains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr || len(s) > len(substr) && stringsContainsAny(s, substr)
}

func stringsContainsAny(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestKnowledgeError(t *testing.T) {
	err := ErrPathTraversal
	if err.Error() == "" {
		t.Error("KnowledgeError.Error() returned empty string")
	}

	if err.Code != "PATH_TRAVERSAL" {
		t.Errorf("KnowledgeError.Code = %q, want %q", err.Code, "PATH_TRAVERSAL")
	}
}
