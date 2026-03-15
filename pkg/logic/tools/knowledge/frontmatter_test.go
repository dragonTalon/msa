package knowledge

import (
	"bufio"
	"strings"
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantMetaKeys  []string
		wantBody      string
		wantEmptyMeta bool
	}{
		{
			name: "完整 frontmatter",
			input: `---
id: err-001
date: 2026-03-10
severity: high
tags:
  - 追高
  - 止损
---
# 错误描述
这是正文内容。`,
			wantMetaKeys:  []string{"id", "date", "severity", "tags"},
			wantBody:      "# 错误描述\n这是正文内容。",
			wantEmptyMeta: false,
		},
		{
			name: "无 frontmatter",
			input: `# 纯 Markdown 文件
没有 frontmatter。`,
			wantMetaKeys:  []string{},
			wantBody:      "# 纯 Markdown 文件\n没有 frontmatter。",
			wantEmptyMeta: true,
		},
		{
			name:          "空文件",
			input:         "",
			wantMetaKeys:  []string{},
			wantBody:      "",
			wantEmptyMeta: true,
		},
		{
			name: "只有 frontmatter 无正文",
			input: `---
title: 测试
---`,
			wantMetaKeys:  []string{"title"},
			wantBody:      "",
			wantEmptyMeta: false,
		},
		{
			name: "格式错误的 YAML",
			input: `---
title: 测试
invalid yaml: [unclosed
---
正文内容`,
			wantMetaKeys:  []string{},
			wantBody:      "---\ntitle: 测试\ninvalid yaml: [unclosed\n---\n正文内容",
			wantEmptyMeta: true,
		},
		{
			name: "未闭合的 frontmatter",
			input: `---
title: 测试
正文内容`,
			wantMetaKeys:  []string{},
			wantBody:      "---\ntitle: 测试\n正文内容",
			wantEmptyMeta: true,
		},
		{
			name: "Unicode 内容",
			input: `---
标题: 中文测试
---
## 中文正文
日本語も対応。`,
			wantMetaKeys:  []string{"标题"},
			wantBody:      "## 中文正文\n日本語も対応。",
			wantEmptyMeta: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, body := ParseFrontmatter(tt.input)

			// 检查元数据
			if tt.wantEmptyMeta {
				if len(metadata) > 0 {
					t.Errorf("ParseFrontmatter() metadata = %v, want empty", metadata)
				}
			} else {
				for _, key := range tt.wantMetaKeys {
					if _, ok := metadata[key]; !ok {
						t.Errorf("ParseFrontmatter() metadata missing key %q", key)
					}
				}
			}

			// 检查正文
			if body != tt.wantBody {
				t.Errorf("ParseFrontmatter() body = %q, want %q", body, tt.wantBody)
			}
		})
	}
}

func TestGenerateFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]interface{}
		wantErr  bool
	}{
		{
			name: "简单元数据",
			metadata: map[string]interface{}{
				"id":    "err-001",
				"title": "测试错误",
			},
			wantErr: false,
		},
		{
			name: "嵌套元数据",
			metadata: map[string]interface{}{
				"id":   "test",
				"tags": []string{"tag1", "tag2"},
			},
			wantErr: false,
		},
		{
			name:     "空元数据",
			metadata: map[string]interface{}{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateFrontmatter(tt.metadata)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateFrontmatter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(tt.metadata) > 0 && result == "" {
				t.Error("GenerateFrontmatter() returned empty string for non-empty metadata")
			}

			if len(tt.metadata) == 0 && result != "" {
				t.Error("GenerateFrontmatter() returned non-empty string for empty metadata")
			}
		})
	}
}

func TestGenerateFileContent(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]interface{}
		body     string
		wantErr  bool
	}{
		{
			name: "完整内容",
			metadata: map[string]interface{}{
				"id":   "test",
				"date": "2026-03-15",
			},
			body:    "# 标题\n正文内容",
			wantErr: false,
		},
		{
			name:     "无元数据",
			metadata: map[string]interface{}{},
			body:     "纯正文",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateFileContent(tt.metadata, tt.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateFileContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 验证可以重新解析
			parsedMeta, parsedBody := ParseFrontmatter(result)
			if len(tt.metadata) > 0 && len(parsedMeta) == 0 {
				t.Error("GenerateFileContent() result cannot be parsed back")
			}
			if tt.body != "" && parsedBody != tt.body {
				t.Errorf("ParseFrontmatter() body = %q, want %q", parsedBody, tt.body)
			}
		})
	}
}

func TestReadFrontmatterLine(t *testing.T) {
	input := `---
id: test
title: 测试
---
正文内容`

	scanner := bufio.NewScanner(strings.NewReader(input))
	metadata, body, err := ReadFrontmatterLine(scanner)

	if err != nil {
		t.Errorf("ReadFrontmatterLine() error = %v", err)
	}

	if len(metadata) == 0 {
		t.Error("ReadFrontmatterLine() metadata is empty")
	}

	if metadata["id"] != "test" {
		t.Errorf("ReadFrontmatterLine() metadata['id'] = %v, want 'test'", metadata["id"])
	}

	if body != "正文内容" {
		t.Errorf("ReadFrontmatterLine() body = %q, want '正文内容'", body)
	}
}
