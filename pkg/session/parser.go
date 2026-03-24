package session

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"msa/pkg/model"
)

// ParseError 解析错误
type ParseError struct {
	Message string
}

func (e *ParseError) Error() string {
	return e.Message
}

// ParsedSession 解析后的会话
type ParsedSession struct {
	Session  *Session
	Messages []model.Message
}

// LoadSession 从文件加载会话
func (m *Manager) LoadSession(sessionID string) (*ParsedSession, error) {
	// 解析 sessionID 为文件路径
	filePath, err := m.parseSessionID(sessionID)
	if err != nil {
		return nil, err
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("会话不存在：%s", sessionID)
	}

	// 读取文件
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取会话文件失败：%w", err)
	}

	// 解析文件
	return parseSessionFile(content, filePath)
}

// parseSessionID 解析 sessionID 为文件路径
// 格式：YYYY-MM-DD_uuid前8位 或 完整 uuid
func (m *Manager) parseSessionID(sessionID string) (string, error) {
	parts := strings.Split(sessionID, "_")
	if len(parts) != 2 {
		return "", &ParseError{Message: "会话 ID 格式错误，正确格式：YYYY-MM-DD_uuid"}
	}

	dateStr := parts[0]

	// 验证日期格式
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return "", &ParseError{Message: "日期格式错误，正确格式：YYYY-MM-DD"}
	}

	// 在 memory 目录中查找匹配的文件
	// 文件名格式: {YYYY-MM-DD_uuid}.md
	entries, err := os.ReadDir(m.memoryDir)
	if err != nil {
		return "", fmt.Errorf("会话不存在：%s", sessionID)
	}

	// 查找匹配的文件：前缀为 sessionID（日期_uuid前8位）
	prefix := sessionID
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), prefix) && strings.HasSuffix(entry.Name(), ".md") {
			return filepath.Join(m.memoryDir, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("会话不存在：%s", sessionID)
}

// parseSessionFile 解析会话文件
func parseSessionFile(content []byte, filePath string) (*ParsedSession, error) {
	text := string(content)

	// 检查 frontmatter
	if !strings.HasPrefix(text, "---\n") {
		return nil, &ParseError{Message: "会话文件缺少元数据"}
	}

	// 解析 frontmatter
	endIndex := strings.Index(text[4:], "\n---\n")
	if endIndex == -1 {
		return nil, &ParseError{Message: "会话文件格式错误：frontmatter 未正确闭合"}
	}

	frontmatterText := text[4 : 4+endIndex]
	session, err := parseFrontmatter(frontmatterText, filePath)
	if err != nil {
		return nil, err
	}

	// 解析消息
	bodyText := text[4+endIndex+5:] // 跳过 frontmatter
	messages := parseMessages(bodyText)

	return &ParsedSession{
		Session:  session,
		Messages: messages,
	}, nil
}

// parseFrontmatter 解析 YAML frontmatter
func parseFrontmatter(text string, filePath string) (*Session, error) {
	session := &Session{
		FilePath: filePath,
	}

	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "uuid":
			session.UUID = value
		case "created_at":
			t, err := time.Parse(time.RFC3339, value)
			if err != nil {
				log.Warnf("解析 created_at 失败: %v", err)
			}
			session.CreatedAt = t
		case "updated_at":
			t, err := time.Parse(time.RFC3339, value)
			if err != nil {
				log.Warnf("解析 updated_at 失败: %v", err)
			}
			session.UpdatedAt = t
		case "mode":
			session.Mode = Mode(value)
		}
	}

	if session.UUID == "" {
		return nil, &ParseError{Message: "会话文件缺少 uuid"}
	}

	return session, nil
}

// parseMessages 解析消息
func parseMessages(text string) []model.Message {
	var messages []model.Message
	var currentRole model.MessageRole
	var currentContent strings.Builder

	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()

		// 检查消息标题
		if strings.HasPrefix(line, "## 👤 用户") {
			// 保存之前的消息
			if currentRole != "" && currentContent.Len() > 0 {
				messages = append(messages, model.Message{
					Role:    currentRole,
					Content: strings.TrimSpace(currentContent.String()),
				})
			}
			currentRole = model.RoleUser
			currentContent.Reset()
			continue
		}

		if strings.HasPrefix(line, "## 🤖 MSA") {
			// 保存之前的消息
			if currentRole != "" && currentContent.Len() > 0 {
				messages = append(messages, model.Message{
					Role:    currentRole,
					Content: strings.TrimSpace(currentContent.String()),
				})
			}
			currentRole = model.RoleAssistant
			currentContent.Reset()
			continue
		}

		// 追加到当前消息
		if currentRole != "" {
			if currentContent.Len() > 0 {
				currentContent.WriteString("\n")
			}
			currentContent.WriteString(line)
		}
	}

	// 保存最后一条消息
	if currentRole != "" && currentContent.Len() > 0 {
		messages = append(messages, model.Message{
			Role:    currentRole,
			Content: strings.TrimSpace(currentContent.String()),
		})
	}

	return messages
}
