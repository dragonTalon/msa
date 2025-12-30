package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	log "github.com/sirupsen/logrus"
)

// Chat TUI聊天模型
type Chat struct {
	input       string
	messages    []string
	cursor      int
	ctx         context.Context
	cursorBlink bool // 光标闪烁状态
}

// NewChat 创建新的聊天模型
func NewChat(ctx context.Context) *Chat {
	return &Chat{
		messages: []string{"欢迎使用 MSA！输入你的理财梦想吧..."},
		ctx:      ctx,
	}
}

type tickMsg time.Time

// Init 实现 tea.Model 接口
func (c *Chat) Init() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update 实现 tea.Model 接口，处理消息更新
func (c *Chat) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		// 切换光标显示状态
		c.cursorBlink = !c.cursorBlink
		// 返回下一次定时器
		return c, tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	case tea.KeyMsg:
		log.Debugf("捕获按键: %s, Type: %v", msg.String(), msg.Type)
		switch msg.String() {
		case "ctrl+c", "esc":
			return c, tea.Quit
		case "enter":
			if c.input == "" {
				return c, nil
			}
			if c.input == "clear" {
				c.messages = []string{"欢迎使用 MSA！输入你的理财梦想吧..."}
				c.input = ""
				return c, nil
			}
			c.messages = append(c.messages, "你: "+c.input)
			log.Debugf("你输入了: %s", c.input)
			// 处理输入逻辑
			c.messages = append(c.messages, "MSA: recovery -->"+c.input)
			c.input = ""
		case "backspace":
			if len(c.input) > 0 {
				c.input = c.input[:len(c.input)-1]
			}
		case "ctrl+k":
			// 执行清空
			c.messages = []string{"欢迎使用 MSA！输入你的理财梦想吧..."}
			c.input = ""
		default:
			key := msg.String()
			if len(key) == 1 {
				c.input += key
			} else if key == "space" {
				c.input += " "
			}
		}
	}
	return c, nil
}

// View 实现 tea.Model 接口，渲染界面
func (c *Chat) View() string {
	if len(c.messages) == 0 {
		PrintPrettyMSALogo()
	}

	s := lipgloss.NewStyle().Foreground(PrimaryColor).Render("")
	for _, msg := range c.messages {
		s += msg + "\n"
	}

	cursor := ""
	if c.cursorBlink {
		cursor = "▊"
	} else {
		cursor = " "
	}
	// 显示输入框
	s += "\n" + lipgloss.NewStyle().Foreground(SecondaryColor).Render("MSA :  ") + c.input + cursor
	s += "\n\n" + lipgloss.NewStyle().Faint(true).Render(
		"(ESC/Ctrl+C: 退出 | Ctrl+K/clear: 清空对话)",
	)
	return s
}
