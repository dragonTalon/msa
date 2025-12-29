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
	"os"
	"os/signal"
	"syscall"
)

var rootCmd = &cobra.Command{
	Use:   "msa",
	Short: "My Stock Agent CLI ",
	Long:  logo,
	Run:   entry,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	ctx, cancel := notify(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Errorf("MSA execute failed: %v", err)
		log.Fatal(err)
	}
}

func entry(cmd *cobra.Command, args []string) {
	// 步骤 1：打印样式化的 MSA LOGO + ASCII 图（核心优化点）
	PrintPrettyMSALogo()
	ctx := cmd.Context()
	// 步骤 2：（可选）添加功能提示，优化用户体验
	p := tea.NewProgram(model{
		messages: []string{"欢迎使用 MSA！输入你的理财梦想吧..."},
		ctx:      ctx,
	})
	if _, err := p.Run(); err != nil {
		log.Fatalf("运行 TUI 失败: %v", err)
	}
}

type model struct {
	input    string
	messages []string
	cursor   int
	ctx      context.Context
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			if m.input != "" {
				m.messages = append(m.messages, "你: "+m.input)
				log.Infof("你输入了: %s", m.input)
				// 处理输入逻辑
				//response := handleCommand(m.input)
				//m.messages = append(m.messages, "MSA: "+response)
				m.messages = append(m.messages, "MSA: recovery -->"+m.input)
				m.input = ""
			}
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		default:
			m.input += msg.String()
		}
	}
	return m, nil
}

func (m model) View() string {
	s := lipgloss.NewStyle().Foreground(primaryColor).Render("\n")

	// 显示历史消息
	for _, msg := range m.messages {
		s += msg + "\n"
	}

	// 显示输入框
	s += "\n" + lipgloss.NewStyle().Foreground(secondaryColor).Render("MSA> ") + m.input
	s += "\n\n" + lipgloss.NewStyle().Faint(true).Render("(按 ESC 或 Ctrl+C 退出)")

	return s
}

func notify(parent context.Context, signals ...os.Signal) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)

	// 绑定信号通知
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)

	if ctx.Err() == nil {
		// 监听信号
		go func() {
			// 第一次收到信号取消上下文
			select {
			case <-ctx.Done():
				return
			case <-ch:
				cancel()
			}
			// 第二次直接退出
			select {
			case s, ok := <-ch:
				if !ok || s == nil {
					os.Exit(1)
				}
				if syscallSignal, isSyscallSignal := s.(syscall.Signal); isSyscallSignal {
					os.Exit(128 + int(syscallSignal)) // 128+n 被信号终止的退出码
				}
				os.Exit(1)
			}
		}()
	}

	return ctx, cancel
}
