package session

import (
	"time"
)

// Mode 会话模式
type Mode string

const (
	// ModeTUI TUI 模式
	ModeTUI Mode = "tui"
	// ModeCLI CLI 模式
	ModeCLI Mode = "cli"
)

// Session 会话信息
type Session struct {
	UUID      string    // 完整 UUID
	CreatedAt time.Time // 创建时间
	UpdatedAt time.Time // 最后更新时间
	Mode      Mode      // 会话模式
	FilePath  string    // 文件路径
}
