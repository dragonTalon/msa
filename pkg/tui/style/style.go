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

	// ChatNormalMsgStyle 普通消息样式
	ChatNormalMsgStyle = lipgloss.NewStyle().
				Foreground(TextColor)

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

	// ChatBorderStyle 消息区域边框样式
	ChatBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			Padding(0, 1)

	// ChatTitleStyle 聊天标题样式
	ChatTitleStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			Padding(0, 1)
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
