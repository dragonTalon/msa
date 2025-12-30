package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Logo MSA ASCII艺术LOGO
var Logo = `
$$\      $$\        $$$$$$\         $$$$$$\  
$$$\    $$$ |      $$  __$$\       $$  __$$\ 
$$$$\  $$$$ |      $$ /  \__|      $$ /  $$ |
$$\$$\$$ $$ |      \$$$$$$\        $$$$$$$$ |
$$ \$$$  $$ |       \____$$\       $$  __$$ |
$$ |\$  /$$ |      $$\   $$ |      $$ |  $$ |
$$ | \_/ $$ |      \$$$$$$  |      $$ |  $$ |
\__|     \__|       \______/       \__|  \__|
(My Stock Agent CLI)`

var (
	// PrimaryColor 主色调：沉稳蓝色（贴合股票/金融工具的视觉定位）
	PrimaryColor = lipgloss.Color("#2563eb")
	// SecondaryColor 辅助色调：浅蓝灰（用于次要文本，提升层次感）
	SecondaryColor = lipgloss.Color("#93c5fd")

	// separatorStyle 分隔线样式（替换单调等号，带装饰符号）
	separatorStyle = lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Bold(false)

	// logoStyle MSA LOGO 样式（加粗+主色调，突出品牌感）
	logoStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			Align(lipgloss.Center)

	// asciiArtStyle 原有 ASCII 艺术图样式（主色调+轻微加粗，保持质感）
	asciiArtStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(false).
			Align(lipgloss.Center)

	// titleStyle 标题样式（辅助色调+居中，搭配整体布局）
	titleStyle = lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Bold(true).
			Align(lipgloss.Center).
			MarginTop(1)
)

// PrintPrettyMSALogo 打印美化的 MSA LOGO
func PrintPrettyMSALogo() {
	// 定义美观分隔线（替换等号，带前后装饰符，长度适配LOGO）
	separator := separatorStyle.Render("╭───────────────────────────────────────────────────╮")
	separatorBottom := separatorStyle.Render("╰───────────────────────────────────────────────────╯")
	renderedAsciiArt := asciiArtStyle.Render(Logo)
	renderedTitle := titleStyle.Render("📈 专业股票代理工具 | 高效管理你的投资")

	// 拼接并打印最终布局（对称美观，层次分明）
	fmt.Println(separator)
	fmt.Println(renderedAsciiArt)
	fmt.Println(separatorBottom)
	fmt.Println(renderedTitle)
}
