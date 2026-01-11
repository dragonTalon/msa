package style

import (
	"github.com/charmbracelet/lipgloss"
)

// SelectorStyles 选择器样式配置
type SelectorStyles struct {
	// 系统消息样式
	SystemMsg lipgloss.Style

	// 帮助信息样式
	Help lipgloss.Style

	// 标题样式
	Title lipgloss.Style

	// 选中项样式 - 整行高亮
	Selected lipgloss.Style

	// 普通项样式
	Normal lipgloss.Style

	// 描述样式
	Description lipgloss.Style

	// 光标样式 - 超级醒目的金色箭头
	Cursor lipgloss.Style

	// 行号样式
	LineNumber lipgloss.Style

	// 滚动指示器样式
	Scroll lipgloss.Style

	// 搜索框样式
	SearchBox lipgloss.Style

	// 搜索框提示样式
	SearchPlaceholder lipgloss.Style

	// 错误提示样式
	Error lipgloss.Style

	// 分隔线样式
	Separator lipgloss.Style
}

// NewSelectorStyles 创建默认的选择器样式
func NewSelectorStyles() *SelectorStyles {
	return &SelectorStyles{
		SystemMsg: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#2563eb")).
			Bold(true),

		Help: lipgloss.NewStyle().
			Faint(true).
			MarginTop(1),

		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			Padding(0, 0),

		Selected: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#FFD700")).
			Padding(0, 1),

		Normal: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC")),

		Description: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true),

		Cursor: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFD700")),

		LineNumber: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Width(4).
			Align(lipgloss.Right),

		Scroll: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")),

		SearchBox: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).
			Bold(true),

		SearchPlaceholder: lipgloss.NewStyle().
			Faint(true),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true),

		Separator: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")),
	}
}

// 样式常量
const (
	SeparatorLine = "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	CursorSymbol  = "►►► "
	CursorEmpty   = "    "
)
