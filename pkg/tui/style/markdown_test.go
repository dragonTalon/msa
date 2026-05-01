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
