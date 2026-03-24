package session

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// 文件权限
const (
	dirPerm  = 0755
	filePerm = 0644
)

// NewSession 创建新会话
func (m *Manager) NewSession(mode Mode) *Session {
	now := time.Now()
	sessionUUID := uuid.New().String()

	// 生成文件路径: ~/.msa/memory/{YYYY-MM-DD_uuid}.md
	filePath := m.generateFilePath(now, sessionUUID)

	return &Session{
		UUID:      sessionUUID,
		CreatedAt: now,
		UpdatedAt: now,
		Mode:      mode,
		FilePath:  filePath,
	}
}

// generateFilePath 生成文件路径
func (m *Manager) generateFilePath(t time.Time, sessionUUID string) string {
	// 文件名格式: {YYYY-MM-DD_uuid}.md
	filename := fmt.Sprintf("%s_%s.md", t.Format("2006-01-02"), sessionUUID)
	return filepath.Join(m.memoryDir, filename)
}

// CreateSessionFile 创建会话文件
func (m *Manager) CreateSessionFile(session *Session) error {
	// 确保目录存在
	dir := filepath.Dir(session.FilePath)
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		log.Warnf("创建会话目录失败: %v", err)
		return err
	}

	// 创建文件并写入 frontmatter
	f, err := os.OpenFile(session.FilePath, os.O_CREATE|os.O_WRONLY, filePerm)
	if err != nil {
		log.Warnf("创建会话文件失败: %v", err)
		return err
	}
	defer f.Close()

	frontmatter := formatFrontmatter(session)
	if _, err := f.WriteString(frontmatter); err != nil {
		log.Warnf("写入 frontmatter 失败: %v", err)
		return err
	}

	return nil
}

// formatFrontmatter 格式化 frontmatter
func formatFrontmatter(session *Session) string {
	return fmt.Sprintf(`---
uuid: %s
created_at: %s
updated_at: %s
mode: %s
---

`,
		session.UUID,
		session.CreatedAt.Format(time.RFC3339),
		session.UpdatedAt.Format(time.RFC3339),
		session.Mode,
	)
}

// AppendMessage 追加消息到会话文件
func (m *Manager) AppendMessage(session *Session, role, content string) error {
	if session == nil || session.FilePath == "" {
		return nil
	}

	// 打开文件追加
	f, err := os.OpenFile(session.FilePath, os.O_APPEND|os.O_WRONLY, filePerm)
	if err != nil {
		log.Warnf("打开会话文件失败: %v", err)
		return err
	}
	defer f.Close()

	// 格式化消息
	var msg string
	switch role {
	case "user":
		msg = fmt.Sprintf("## 👤 用户\n%s\n\n", content)
	case "assistant":
		msg = fmt.Sprintf("## 🤖 MSA\n%s\n\n", content)
	default:
		return nil
	}

	if _, err := f.WriteString(msg); err != nil {
		log.Warnf("追加消息失败: %v", err)
		return err
	}

	// 更新时间
	session.UpdatedAt = time.Now()

	return nil
}

// ShortID 返回 UUID 前 8 位
func (s *Session) ShortID() string {
	if len(s.UUID) >= 8 {
		return s.UUID[:8]
	}
	return s.UUID
}

// SessionID 返回完整的会话 ID（日期+UUID前8位）
func (s *Session) SessionID() string {
	return fmt.Sprintf("%s_%s", s.CreatedAt.Format("2006-01-02"), s.ShortID())
}
