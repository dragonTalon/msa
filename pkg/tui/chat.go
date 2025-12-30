package tui

import (
	"context"
	"strings"

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

	return &Chat{
		textInput: ti,
		pendingMsgs: []Message{
			{Role: "logo", Content: GetStyledLogo()},
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
			input := strings.TrimSpace(c.textInput.Value())
			if input == "" {
				return c, nil
			}

			// 处理特殊命令
			switch strings.ToLower(input) {
			case "clear":
				c.textInput.Reset()
				c.addMessage("system", "对话已清空，重新开始吧！")
				return c, c.Flush()

			case "help", "?":
				c.textInput.Reset()
				c.addMessage("system", "📋 可用命令:\n  • clear - 清空对话\n  • help/? - 显示帮助\n  • quit/exit - 退出程序")
				return c, c.Flush()

			case "quit", "exit":
				return c, tea.Quit
			}

			// 添加用户消息
			c.addMessage("user", input)
			log.Debugf("用户输入: %s", input)

			// 模拟 AI 回复（后续可接入真正的 AI）
			c.addMessage("assistant", "📊 已收到您的问题: \""+input+"\"\n正在分析中...")

			// 清空输入框
			c.textInput.Reset()

			// flush 消息到终端
			return c, c.Flush()

		case tea.KeyCtrlK:
			c.textInput.Reset()
			c.addMessage("system", "对话已清空，重新开始吧！")
			return c, c.Flush()
		}
	}

	// 更新输入组件
	c.textInput, tiCmd = c.textInput.Update(msg)

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

	// 帮助提示
	help := ChatHelpStyle.Render(
		"ESC/Ctrl+C: 退出 | Ctrl+K: 清空 | Enter: 发送",
	)
	sb.WriteString("\n")
	sb.WriteString(help)

	return sb.String()
}
