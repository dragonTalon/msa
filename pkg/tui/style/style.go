package style

import (
	"github.com/charmbracelet/lipgloss"
)

// Logo MSA ASCII艺术LOGO
var Logo = `
╭───────────────────────────────────────────────────────╮
	$$\      $$\        $$$$$$\         $$$$$$\  
	$$$\    $$$ |      $$  __$$\       $$  __$$\ 
	$$$$\  $$$$ |      $$ /  \__|      $$ /  $$ |
	$$\$$\$$ $$ |      \$$$$$$\        $$$$$$$$ |
	$$ \$$$  $$ |       \____$$\       $$  __$$ |
	$$ |\$  /$$ |      $$\   $$ |      $$ |  $$ |
	$$ | \_/ $$ |      \$$$$$$  |      $$ |  $$ |
	\__|     \__|       \______/       \__|  \__|
╰───────────────────────────────────────────────────────╯
(My Stock Agent CLI)
📈 专业股票代理工具 | 高效管理你的投资
`

// ==================== 通用颜色定义 ====================
var (
	// PrimaryColor 主色调：沉稳蓝色（贴合股票/金融工具的视觉定位）
	PrimaryColor = lipgloss.Color("#2563eb")
	// ToolColor 工具色调：黄色（强调工具属性，提升用户参与感）
	ToolColor = lipgloss.Color("#FFED4E")
	// SecondaryColor 辅助色调：浅蓝灰（用于次要文本，提升层次感）
	SecondaryColor = lipgloss.Color("#93c5fd")
	// TextColor 文本颜色
	TextColor = lipgloss.Color("#e2e8f0")
	// WhiteColor 白色
	WhiteColor = lipgloss.Color("#ffffff")
	// logoStyle MSA LOGO 样式（加粗+主色调，突出品牌感）
	logoStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			Align(lipgloss.Center)
)

// ==================== 聊天界面样式 ====================
var (
	// ChatUserMsgStyle 用户消息样式
	ChatUserMsgStyle = lipgloss.NewStyle().
				Foreground(SecondaryColor).
				Bold(true)

	// ChatSystemMsgStyle 系统消息样式
	ChatSystemMsgStyle = lipgloss.NewStyle().
				Foreground(PrimaryColor).
				Bold(true)

	// ChatNormalMsgStyle 普通消息样式（正文-白色）
	ChatNormalMsgStyle = lipgloss.NewStyle().
				Foreground(WhiteColor)

	// ChatToolMsgStyle 工具消息样式（工具-黄色）
	ChatToolMsgStyle = lipgloss.NewStyle().
				Foreground(ToolColor)

	// ChatReasonMsgStyle 思考消息样式（思考-灰色）
	ChatReasonMsgStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9ca3af")).
				Italic(true)

	// ChatInputPromptStyle 输入框提示样式
	ChatInputPromptStyle = lipgloss.NewStyle().
				Foreground(SecondaryColor).
				Bold(true)

	// ChatInputTextStyle 输入框文本样式
	ChatInputTextStyle = lipgloss.NewStyle().
				Foreground(WhiteColor)

	// ChatHelpStyle 帮助文本样式
	ChatHelpStyle = lipgloss.NewStyle().
			Faint(true).
			MarginTop(1)

	// ==================== 命令建议样式 ====================
	// CommandSuggestionSelectedStyle 选中项样式
	CommandSuggestionSelectedStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#FFD700")).
					Bold(true)

	// CommandSuggestionNormalStyle 普通项样式
	CommandSuggestionNormalStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#AAAAAA"))

	// CommandSuggestionDescStyle 描述文本样式
	CommandSuggestionDescStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#666666")).
					Italic(true)

	// CommandSuggestionHintStyle 快捷键提示样式
	CommandSuggestionHintStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#555555")).
					Faint(true)
)

// ==================== 表格样式 ====================
var (
	// TableHeaderStyle 表格表头样式
	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FAFAFA")).
				Background(lipgloss.Color("#7D56F4")).
				Padding(0, 1).
				Width(30)

	// TableCellStyle 表格单元格基础样式
	TableCellStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Width(30)

	// TableEvenRowStyle 表格偶数行样式
	TableEvenRowStyle = TableCellStyle.
				Background(lipgloss.Color("#2E2E2E"))

	// TableOddRowStyle 表格奇数行样式
	TableOddRowStyle = TableCellStyle.
				Background(lipgloss.Color("#1E1E1E"))
)

// GetStyledLogo 返回带样式的 Logo 字符串
func GetStyledLogo() string {
	return logoStyle.Render(Logo)
}

// ==================== 流式输出样式 ====================
var (
	// ChatDividerStyle 分割线样式
	ChatDividerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#555555"))

	// 流式输出 emoji 前缀
	ChatThinkingPrefix   = "💭 思考: "
	ChatToolCallPrefix   = "⚙️ 工具: "
	ChatToolResultPrefix = "✅ 工具: "
	ChatToolErrorPrefix  = "❌ 工具: "
	ChatTextPrefix       = "💬 正文: "
)

// DividerLine 分割线内容
const DividerLine = "────────────────────────────"

// ==================== Markdown 渲染样式 ====================
var (
	// MDHeadingStyle 标题样式 (H1-H6) — bold + 主色调
	MDHeadingStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

	// MDBoldStyle 加粗样式
	MDBoldStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(WhiteColor)

	// MDItalicStyle 斜体样式
	MDItalicStyle = lipgloss.NewStyle().
			Italic(true).
			Foreground(TextColor)

	// MDCodeStyle 行内代码样式 — 灰底
	MDCodeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#2d2d2d")).
			Foreground(lipgloss.Color("#e5c07b"))

	// MDCodeBlockStyle 代码块样式 — 缩进 + 灰底边框
	MDCodeBlockStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Background(lipgloss.Color("#1a1a1a"))

	// MDBlockquoteStyle 引用块样式 — 竖线前缀 + 灰色
	MDBlockquoteStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9ca3af")).
				Italic(true).
				PaddingLeft(2)

	// MDListBullet 列表子弹
	MDListBullet = "•"

	// MDTableHeaderStyle 表格表头样式
	MDTableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(WhiteColor).
				Padding(0, 1)

	// MDTableCellStyle 表格数据单元格样式
	MDTableCellStyle = lipgloss.NewStyle().
				Foreground(TextColor).
				Padding(0, 1)

	// MDTableSeparator 表格列分隔符
	MDTableSeparator = "│"
)
