package tui

import (
	"context"
	"fmt"
	"msa/pkg/config"
	command "msa/pkg/logic/command"
	"msa/pkg/model"
	"msa/pkg/tui/cmd"
	"strings"

	listStyle "github.com/charmbracelet/lipgloss/list"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	log "github.com/sirupsen/logrus"
)

// Message 聊天消息结构
type Message struct {
	Role    string // "user", "system", "assistant"
	Content string
}

// Chat TUI聊天模型
type Chat struct {
	textInput   textinput.Model // 文本输入组件
	pendingMsgs []Message       // 待 flush 的消息
	ctx         context.Context // 上下文
	width       int             // 终端宽度
	height      int             // 终端高度
	cmdFlag     bool            // 是否处于命令模式
	cmdList     []string
}

// maskAPIKey 隐藏 APIKey，只显示前4个和后4个字符
func maskAPIKey(apiKey string) string {
	if apiKey == "" {
		return "未设置"
	}
	if len(apiKey) <= 8 {
		return strings.Repeat("*", len(apiKey))
	}
	return apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
}

// NewChat 创建新的聊天模型
func NewChat(ctx context.Context) *Chat {
	// 初始化文本输入组件
	ti := textinput.New()
	ti.Placeholder = "输入你的理财问题..."
	ti.Focus()
	ti.CharLimit = 0
	ti.Width = 50
	ti.PromptStyle = ChatInputPromptStyle
	ti.Prompt = "MSA > "
	ti.TextStyle = ChatInputTextStyle
	cfg := config.GetLocalStoreConfig()
	m := cfg.Model
	if m == "" {
		m = "未设置"
	}
	return &Chat{
		textInput: ti,
		pendingMsgs: []Message{
			{Role: "logo", Content: GetStyledLogo()},
			{Role: "system", Content: fmt.Sprintf("模型供应商: %s", cfg.Provider)},
			{Role: "system", Content: fmt.Sprintf("模型 : %s", m)},
			{Role: "system", Content: fmt.Sprintf("APIKey : %s", maskAPIKey(cfg.APIKey))},
			{Role: "system", Content: "欢迎使用 MSA！输入你的理财问题吧..."},
		},
		ctx: ctx,
	}
}

// Init 实现 tea.Model 接口
func (c *Chat) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, c.Flush())
}

// Flush 将待输出消息 flush 到终端
func (c *Chat) Flush() tea.Cmd {
	if len(c.pendingMsgs) == 0 {
		return nil
	}

	// 渲染所有待输出消息
	content := c.renderPendingMessages()
	// 清空待输出队列
	c.pendingMsgs = nil

	return tea.Println(content)
}

// renderPendingMessages 渲染待输出的消息
func (c *Chat) renderPendingMessages() string {
	var sb strings.Builder

	for i, msg := range c.pendingMsgs {
		switch msg.Role {
		case "logo":
			sb.WriteString(msg.Content)
		case "user":
			sb.WriteString(ChatUserMsgStyle.Render("👤 你: "))
			sb.WriteString(ChatNormalMsgStyle.Render(msg.Content))
		case "system":
			sb.WriteString(ChatSystemMsgStyle.Render("🔧 系统: "))
			sb.WriteString(ChatNormalMsgStyle.Render(msg.Content))
		case "assistant":
			sb.WriteString(ChatSystemMsgStyle.Render("🤖 MSA: "))
			sb.WriteString(ChatNormalMsgStyle.Render(msg.Content))
		}
		if i < len(c.pendingMsgs)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// addMessage 添加消息到待 flush 队列
func (c *Chat) addMessage(role, content string) {
	c.pendingMsgs = append(c.pendingMsgs, Message{
		Role:    role,
		Content: content,
	})
}

// Update 实现 tea.Model 接口，处理消息更新
func (c *Chat) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var tiCmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height
		c.textInput.Width = msg.Width - 10

	case tea.KeyMsg:
		log.Debugf("捕获按键: %s, Type: %v", msg.String(), msg.Type)
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return c, tea.Quit

		case tea.KeyEnter:
			c.cmdFlag = false
			input := strings.TrimSpace(c.textInput.Value())
			if input == "" {
				return c, nil
			}
			// 添加用户消息
			c.addMessage("user", input)
			log.Debugf("用户输入: %s", input)
			if strings.HasPrefix(input, "/") {
				return c.commandHandler(input)
			}
			// 处理特殊命令
			switch strings.ToLower(input) {
			case "clear":
				c.textInput.Reset()
				c.addMessage("system", "对话已清空，重新开始吧！")
			case "help", "?":
				c.textInput.Reset()
				c.addMessage("system", "📋 可用命令:\n  • clear - 清空对话\n  • help/? - 显示帮助\n  • quit/exit - 退出程序")
			case "quit", "exit":
				return c, tea.Quit
			}
			// 清空输入框
			c.textInput.Reset()

			// flush 消息到终端
			return c, c.Flush()

		case tea.KeyCtrlK:
			c.textInput.Reset()
			c.addMessage("system", "对话已清空，重新开始吧！")
			return c, c.Flush()
		default:
			c.textInput, tiCmd = c.textInput.Update(msg)
			if strings.HasPrefix(c.textInput.Value(), "/") {
				c.cmdFlag = true
				log.Infof("进入命令模式 %s\n", c.textInput.Value())
				c.cmdList = command.GetLikeCommand(c.textInput.Value())
				log.Infof("命令列表: %v", c.cmdList)
			}
		}
	}
	return c, tiCmd
}

// View 实现 tea.Model 接口，渲染界面（只渲染输入框和帮助信息）
func (c *Chat) View() string {
	var sb strings.Builder

	// 输入区域
	inputBox := lipgloss.NewStyle().
		Padding(0, 1).
		Render(c.textInput.View())
	sb.WriteString(inputBox)
	if c.cmdFlag {
		styles := listStyle.New()
		for _, cmd := range c.cmdList {
			styles.Item("/" + cmd)
		}
		log.Infof("view styles %s", styles)
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("%s", styles))
	}
	// 帮助提示
	help := ChatHelpStyle.Render(
		"ESC/Ctrl+C: 退出 | Ctrl+K: 清空 | Enter: 发送",
	)
	sb.WriteString("\n")
	sb.WriteString(help)

	return sb.String()
}

// analyzeResult 分析结果
func analyzeResult(result *model.CmdResult) string {
	if result == nil {
		return "结果为空"
	}
	sb := strings.Builder{}
	sb.WriteString("\n")
	switch result.Type {
	case "list":
		list, ok := result.Data.([]string)
		if !ok {
			return "结果类型错误"
		}
		styles := listStyle.New()
		for _, v := range list {
			styles = styles.Item(v)
		}
		log.Infof("list styles %s", styles)
		sb.WriteString(fmt.Sprintf("%s", styles))

	case "table":
		table, ok := result.Data.(map[string]string)
		if !ok {
			return "结果类型错误"
		}
		// 渲染表格
		sb.WriteString(renderTable(table))

	case "boolean":
		b, ok := result.Data.(bool)
		if !ok {
			return "结果类型错误"
		}
		if b {
			sb.WriteString(ChatNormalMsgStyle.Render(result.Msg))
		} else {
			sb.WriteString(ChatNormalMsgStyle.Render(result.Error.Error()))
		}

	}
	return sb.String()
}

// renderTable 渲染表格，展示 key-value 数据
func renderTable(data map[string]string) string {
	if len(data) == 0 {
		return "无数据"
	}

	// 定义表格样式
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(30)

	cellStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Width(30)

	evenRowStyle := cellStyle.Copy().
		Background(lipgloss.Color("#2E2E2E"))

	oddRowStyle := cellStyle.Copy().
		Background(lipgloss.Color("#1E1E1E"))

	var sb strings.Builder

	// 表头
	sb.WriteString(headerStyle.Render("模型名称"))
	sb.WriteString(headerStyle.Render("描述"))
	sb.WriteString("\n")

	// 表格内容
	rowIndex := 0
	for key, value := range data {
		var rowStyle lipgloss.Style
		if rowIndex%2 == 0 {
			rowStyle = evenRowStyle
		} else {
			rowStyle = oddRowStyle
		}

		sb.WriteString(rowStyle.Render(key))
		sb.WriteString(rowStyle.Render(value))
		sb.WriteString("\n")
		rowIndex++
	}

	return sb.String()
}

// commandHandler 命令处理器
func (c *Chat) commandHandler(input string) (tea.Model, tea.Cmd) {
	input = strings.TrimPrefix(input, "/")
	split := strings.Split(input, " ")
	cmdName := split[0]

	msaCmd := command.GetCommand(cmdName)
	if msaCmd == nil {
		c.addMessage("system", "未找到命令: "+input)
		c.addMessage("system", fmt.Sprintf("可用命令: %v", command.GetLikeCommand("/")))
		return c, c.Flush()
	}

	var args []string
	if len(split) > 1 {
		args = split[1:]
	}

	// 执行命令
	runResult, err := msaCmd.Run(c.ctx, args)
	if err != nil {
		c.addMessage("system", "执行命令失败: "+err.Error())
		log.Errorf("执行命令失败: %v", err)
		return c, c.Flush()
	}

	log.Infof("执行命令成功: %v", runResult)

	// 检查是否需要启动交互式选择器
	// 如果命令返回的是 selector 类型，则启动选择器
	if runResult.Type == "selector" {
		items, ok := runResult.Data.([]*model.SelectorItem)
		if !ok {
			c.addMessage("system", "选择器数据类型错误")
			log.Errorf("选择器数据类型错误")
			return c, c.Flush()
		}

		// 调用命令的 ToSelect 方法创建选择器
		selector, err := msaCmd.ToSelect(items)
		if err != nil {
			c.addMessage("system", "创建选择器失败: "+err.Error())
			log.Errorf("创建选择器失败: %v", err)
			return c, c.Flush()
		}

		// 设置上下文
		selector.Ctx = c.ctx
		c.textInput.Reset()

		// 使用 SelectorView 包装 BaseSelector
		selectorView := cmd.NewSelectorView(selector)

		// 启动交互式选择器
		return selectorView, nil
	}

	// 普通命令结果，直接显示
	c.addMessage("system", analyzeResult(runResult))
	c.textInput.Reset()
	return c, c.Flush()
}
