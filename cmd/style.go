package cmd

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

var logo = `
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
	// 主色调：沉稳蓝色（贴合股票/金融工具的视觉定位）
	primaryColor = lipgloss.Color("#2563eb")
	// 辅助色调：浅蓝灰（用于次要文本，提升层次感）
	secondaryColor = lipgloss.Color("#93c5fd")

	// 分隔线样式（替换单调等号，带装饰符号）
	separatorStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(false)

	// MSA LOGO 样式（加粗+主色调，突出品牌感）
	logoStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Align(lipgloss.Center) // 居中对齐

	// 原有 ASCII 艺术图样式（主色调+轻微加粗，保持质感）
	asciiArtStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(false).
			Align(lipgloss.Center) // 居中对齐，让排版更规整

	// 标题样式（辅助色调+居中，搭配整体布局）
	titleStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true).
			Align(lipgloss.Center).
			MarginTop(1) // 顶部留白，提升呼吸感
)

func PrintPrettyMSALogo() {
	// 1. 定义美观分隔线（替换等号，带前后装饰符，长度适配LOGO）
	separator := separatorStyle.Render("╭───────────────────────────────────────────────────╮")
	separatorBottom := separatorStyle.Render("╰───────────────────────────────────────────────────╯")
	// 修复：移除嵌套的反引号，仅用一对反引号包裹多行字符串，替换硬制表符为空格
	renderedAsciiArt := asciiArtStyle.Render(logo)
	renderedTitle := titleStyle.Render("📈 专业股票代理工具 | 高效管理你的投资")

	// 3. 拼接并打印最终布局（对称美观，层次分明）
	fmt.Println(separator)
	fmt.Println(renderedAsciiArt)
	fmt.Println(separatorBottom)
	fmt.Println(renderedTitle)
}
