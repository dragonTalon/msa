/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"time"
)

func run(cmd *cobra.Command, args []string) {
	// 步骤 1：打印样式化的 MSA LOGO + ASCII 图（核心优化点）
	PrintPrettyMSALogo()
	ctx := cmd.Context()
	// 步骤 2：（可选）添加功能提示，优化用户体验
	p := tea.NewProgram(&chat{
		messages: []string{"欢迎使用 MSA！输入你的理财梦想吧..."},
		ctx:      ctx,
	})
	if _, err := p.Run(); err != nil {
		log.Fatalf("运行 TUI 失败: %v", err)
	}
}

type chat struct {
	input       string
	messages    []string
	cursor      int
	ctx         context.Context
	cursorBlink bool // 新增：光标闪烁状态
}

type tickMsg time.Time

func (c *chat) Init() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (c *chat) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			//response := handleCommand(m.input)
			//m.messages = append(m.messages, "MSA: "+response)
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

func (c *chat) View() string {
	if len(c.messages) == 0 {
		PrintPrettyMSALogo()
	}

	s := lipgloss.NewStyle().Foreground(primaryColor).Render("")
	for _, msg := range c.messages {
		s += msg + "\n"
	}
	// 显示历史消息

	cursor := ""
	if c.cursorBlink {
		cursor = "▊" // 或者用 "_" 或 "|"
	} else {
		cursor = " "
	}
	// 显示输入框
	s += "\n" + lipgloss.NewStyle().Foreground(secondaryColor).Render("MSA :  ") + c.input + cursor
	s += "\n\n" + lipgloss.NewStyle().Faint(true).Render(
		"(ESC/Ctrl+C: 退出 | Ctrl+K/clear: 清空对话)",
	)
	return s
}
