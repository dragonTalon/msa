package app

import (
	"context"
	log "github.com/sirupsen/logrus"

	tea "github.com/charmbracelet/bubbletea"

	"msa/pkg/tui"
)

// Run 启动应用主逻辑
func Run(ctx context.Context) error {
	// 启动 TUI 程序
	p := tea.NewProgram(tui.NewChat(ctx))
	if _, err := p.Run(); err != nil {
		log.Errorf("运行 TUI 失败: %v", err)
		return err
	}
	return nil
}
