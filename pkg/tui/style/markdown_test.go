package style

import (
	"strings"
	"testing"
)

func TestRenderMarkdown_Heading(t *testing.T) {
	result := RenderMarkdown("#### 标题测试")

	if strings.Contains(result, "####") {
		t.Errorf("output should not contain '####' markdown syntax, got: %s", result)
	}
	if !strings.Contains(result, "标题测试") {
		t.Errorf("output should contain heading text '标题测试', got: %s", result)
	}
}

func TestRenderMarkdown_Bold(t *testing.T) {
	result := RenderMarkdown("这是**加粗**文字")

	if strings.Contains(result, "**") {
		t.Errorf("output should not contain '**' markdown syntax, got: %s", result)
	}
	if !strings.Contains(result, "加粗") {
		t.Errorf("output should contain bold text '加粗', got: %s", result)
	}
	if !strings.Contains(result, "文字") {
		t.Errorf("output should contain surrounding text '文字', got: %s", result)
	}
}

func TestRenderMarkdown_Italic(t *testing.T) {
	result := RenderMarkdown("这是*斜体*文字")

	if strings.Contains(result, "*斜体") {
		t.Errorf("output should not contain '*斜体' markdown syntax, got: %s", result)
	}
	if !strings.Contains(result, "斜体") {
		t.Errorf("output should contain italic text '斜体', got: %s", result)
	}
}

func TestRenderMarkdown_CodeSpan(t *testing.T) {
	result := RenderMarkdown("股票代码 `AAPL` 已更新")

	if strings.Contains(result, "`AAPL`") {
		t.Errorf("output should not contain backtick syntax, got: %s", result)
	}
	if !strings.Contains(result, "AAPL") {
		t.Errorf("output should contain code text 'AAPL', got: %s", result)
	}
}

func TestRenderMarkdown_UnorderedList(t *testing.T) {
	result := RenderMarkdown("- 项目1\n- 项目2")

	// 不应保留 "- " 标记
	if strings.Contains(result, "- 项目") {
		t.Errorf("output should not contain '- ' markdown syntax, got: %s", result)
	}
	if !strings.Contains(result, "项目1") {
		t.Errorf("output should contain list item '项目1', got: %s", result)
	}
	if !strings.Contains(result, "项目2") {
		t.Errorf("output should contain list item '项目2', got: %s", result)
	}
}

func TestRenderMarkdown_FencedCodeBlock(t *testing.T) {
	input := "```go\nfmt.Println(\"hello\")\n```"
	result := RenderMarkdown(input)

	if strings.Contains(result, "```") {
		t.Errorf("output should not contain fence ``` syntax, got: %s", result)
	}
	// chroma 语法高亮在 token 之间插入 ANSI 码，分别检查关键子串
	if !strings.Contains(result, "fmt") || !strings.Contains(result, "Println") {
		t.Errorf("output should contain code content, got: %s", result)
	}
}

func TestRenderMarkdown_Blockquote(t *testing.T) {
	result := RenderMarkdown("> 重要提示内容")

	if strings.Contains(result, "> ") {
		t.Errorf("output should not contain '> ' markdown syntax, got: %s", result)
	}
	if !strings.Contains(result, "重要提示内容") {
		t.Errorf("output should contain blockquote text, got: %s", result)
	}
}

func TestRenderMarkdown_EmptyInput(t *testing.T) {
	result := RenderMarkdown("")
	if result != "" {
		t.Errorf("empty input should produce empty output, got: %q", result)
	}
}

func TestRenderMarkdown_PlainText(t *testing.T) {
	input := "这是普通文本，没有任何格式"
	result := RenderMarkdown(input)

	if !strings.Contains(result, input) {
		t.Errorf("plain text should be preserved verbatim, got: %s", result)
	}
}

func TestRenderMarkdown_MultipleHeadings(t *testing.T) {
	input := "## 标题2\n正文段落\n### 标题3\n另一段落"
	result := RenderMarkdown(input)

	if strings.Contains(result, "##") || strings.Contains(result, "###") {
		t.Errorf("output should not contain heading '#' syntax, got: %s", result)
	}
	if !strings.Contains(result, "标题2") || !strings.Contains(result, "标题3") {
		t.Errorf("output should contain all heading texts, got: %s", result)
	}
}

func TestRenderMarkdown_NewlineBehavior(t *testing.T) {
	input := "第一行\n\n第二行"
	result := RenderMarkdown(input)

	if !strings.Contains(result, "第一行") || !strings.Contains(result, "第二行") {
		t.Errorf("output should contain both lines, got: %s", result)
	}
}

func TestRenderMarkdown_Table(t *testing.T) {
	input := `| 项目 | 数值 |
|------|------|
| 初始资金 | **500,000.00 元** |
| 当前总资产 | 493,233.00 元 |
`
	result := RenderMarkdown(input)

	if strings.Contains(result, "|---") || strings.Contains(result, "------") {
		t.Errorf("output should not contain alignment separator line, got: %s", result)
	}
	if !strings.Contains(result, "初始资金") {
		t.Errorf("output should contain '初始资金', got: %s", result)
	}
	if !strings.Contains(result, "500,000.00") {
		t.Errorf("output should contain '500,000.00', got: %s", result)
	}
	if !strings.Contains(result, "当前总资产") {
		t.Errorf("output should contain '当前总资产', got: %s", result)
	}
	if !strings.Contains(result, "493,233.00") {
		t.Errorf("output should contain '493,233.00', got: %s", result)
	}
	if strings.Contains(result, "**500,000.00") || strings.Contains(result, "500,000.00**") {
		t.Errorf("output should not contain '**' bold syntax inside table cells, got: %s", result)
	}
	if !strings.Contains(result, MDTableSeparator) {
		t.Errorf("output should contain column separator '%s', got: %s", MDTableSeparator, result)
	}
}

func TestRenderMarkdown_TableHeaderBold(t *testing.T) {
	input := `| 表头 |
|------|
| 数据 |
`
	result := RenderMarkdown(input)

	if !strings.Contains(result, "表头") {
		t.Errorf("output should contain header text '表头', got: %s", result)
	}
	if !strings.Contains(result, "数据") {
		t.Errorf("output should contain cell text '数据', got: %s", result)
	}
}

func TestRenderMarkdown_EmptyTable(t *testing.T) {
	input := `| 项目 | 数值 |
|------|------|
`
	result := RenderMarkdown(input)

	_ = result // 不应 panic
}

func TestRenderMarkdown_LinkShowsURL(t *testing.T) {
	result := RenderMarkdown("请[点击这里](https://example.com)查看")

	if !strings.Contains(result, "点击这里") {
		t.Errorf("output should contain link text '点击这里', got: %s", result)
	}
	if !strings.Contains(result, "https://example.com") {
		t.Errorf("output should contain link URL, got: %s", result)
	}
	if strings.Contains(result, "[点击这里]") {
		t.Errorf("output should not contain markdown link syntax, got: %s", result)
	}
}

func TestRenderMarkdown_AutoLink(t *testing.T) {
	result := RenderMarkdown("访问 https://example.com 了解更多")

	if !strings.Contains(result, "https://example.com") {
		t.Errorf("output should contain autolink URL, got: %s", result)
	}
}

func TestRenderMarkdown_AutoLinkEmail(t *testing.T) {
	result := RenderMarkdown("联系 user@example.com 获取帮助")

	if !strings.Contains(result, "user@example.com") {
		t.Errorf("output should contain email autolink, got: %s", result)
	}
}

func TestRenderMarkdown_HeadingLevels(t *testing.T) {
	result := RenderMarkdown("# H1\n## H2\n### H3\n#### H4\n##### H5\n###### H6")

	if strings.Contains(result, "#") {
		t.Errorf("output should not contain '#' syntax, got: %s", result)
	}
	for _, h := range []string{"H1", "H2", "H3", "H4", "H5", "H6"} {
		if !strings.Contains(result, h) {
			t.Errorf("output should contain heading text '%s', got: %s", h, result)
		}
	}
}

func TestRenderMarkdown_TaskList(t *testing.T) {
	input := "- [ ] 待办事项\n- [x] 已完成\n- [ ] 另一待办"
	result := RenderMarkdown(input)

	if !strings.Contains(result, "[ ]") {
		t.Errorf("output should contain '[ ]' checkbox, got: %s", result)
	}
	if !strings.Contains(result, "[x]") {
		t.Errorf("output should contain '[x]' checkbox, got: %s", result)
	}
	if strings.Contains(result, "- [ ]") || strings.Contains(result, "- [x]") {
		t.Errorf("output should not contain '- [ ]' raw syntax, got: %s", result)
	}
	if !strings.Contains(result, "待办事项") || !strings.Contains(result, "已完成") {
		t.Errorf("output should contain task text, got: %s", result)
	}
}

func TestRenderMarkdown_Strikethrough(t *testing.T) {
	result := RenderMarkdown("这是~~删除的~~文字")

	if strings.Contains(result, "~~删除的~~") || strings.Contains(result, "~~") {
		t.Errorf("output should not contain '~~' markdown syntax, got: %s", result)
	}
	if !strings.Contains(result, "删除的") {
		t.Errorf("output should contain strikethrough text '删除的', got: %s", result)
	}
	if !strings.Contains(result, "文字") {
		t.Errorf("output should contain surrounding text '文字', got: %s", result)
	}
}

func TestRenderMarkdown_StrikethroughInTable(t *testing.T) {
	input := `| 项目 | 状态 |
|------|------|
| 旧功能 | ~~已移除~~ |
`
	result := RenderMarkdown(input)

	if strings.Contains(result, "~~") {
		t.Errorf("output should not contain '~~' syntax in table, got: %s", result)
	}
	if !strings.Contains(result, "已移除") {
		t.Errorf("output should contain strikethrough text in table cell, got: %s", result)
	}
}

func TestRenderMarkdown_TableColumnAlignment(t *testing.T) {
	input := `| ID | 描述 |
|----|------|
| 1 | 短 |
| 2 | 这是一个很长的描述 |
`
	result := RenderMarkdown(input)

	if !strings.Contains(result, "ID") || !strings.Contains(result, "描述") {
		t.Errorf("output should contain headers, got: %s", result)
	}
	if !strings.Contains(result, "1") || !strings.Contains(result, "2") {
		t.Errorf("output should contain row data, got: %s", result)
	}
	if strings.Contains(result, "|---") || strings.Contains(result, "----") {
		t.Errorf("output should not contain separator line, got: %s", result)
	}
}

func TestRenderMarkdown_TableUniformWidth(t *testing.T) {
	input := `| 项目 | 数值 |
|------|------|
| 初始资金 | 500,000.00 元 |
| 当前总资产 | 493,233.00 元 |
`
	result := RenderMarkdown(input)

	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines, got %d", len(lines))
	}

	// 验证每行数据都包含列分隔符
	for _, line := range lines {
		if strings.Contains(line, "初始资金") || strings.Contains(line, "当前总资产") {
			sepIdx := strings.Index(line, MDTableSeparator)
			if sepIdx < 0 {
				t.Errorf("line should contain separator '%s': %s", MDTableSeparator, line)
			}
		}
	}
}

func TestRenderMarkdown_TableWithInlineFormatting(t *testing.T) {
	input := `| 股票 | 涨跌幅 |
|------|--------|
| AAPL | +3.5% |
| GOOGL | *待更新* |
`
	result := RenderMarkdown(input)

	if !strings.Contains(result, "AAPL") || !strings.Contains(result, "GOOGL") {
		t.Errorf("output should contain stock codes, got: %s", result)
	}
	if strings.Contains(result, "*待更新*") {
		t.Errorf("output should not contain '*' italic syntax, got: %s", result)
	}
}
