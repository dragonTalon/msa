package style

import (
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var (
	// markdownRenderer Markdown 渲染器实例
	markdownRenderer *glamour.TermRenderer
)

// InitMarkdownRenderer 初始化 Markdown 渲染器
func InitMarkdownRenderer(width int) error {
	var err error
	markdownRenderer, err = glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	return err
}

// RenderMarkdown 渲染 Markdown 文本
// 如果渲染失败，返回原始文本
func RenderMarkdown(content string) string {
	if markdownRenderer == nil {
		// 如果渲染器未初始化，尝试初始化
		if err := InitMarkdownRenderer(80); err != nil {
			return content
		}
	}

	rendered, err := markdownRenderer.Render(content)
	if err != nil {
		// 渲染失败，返回原始文本
		return content
	}

	// 去除末尾多余的换行
	return strings.TrimRight(rendered, "\n")
}

// RenderMarkdownWithStyle 渲染 Markdown 并应用样式
func RenderMarkdownWithStyle(content string, style lipgloss.Style) string {
	rendered := RenderMarkdown(content)
	return style.Render(rendered)
}
