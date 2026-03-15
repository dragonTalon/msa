package knowledge

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	log "github.com/sirupsen/logrus"
	"msa/pkg/logic/memory"
	"msa/pkg/logic/message"
	"msa/pkg/model"
)

// QuerySessionsByDateParam 按日期查询 Session 参数
type QuerySessionsByDateParam struct {
	Date string `json:"date" jsonschema:"description=日期，支持 YYYY-MM-DD 格式或 today/yesterday，required"`
	Tag  string `json:"tag" jsonschema:"description=可选的标签过滤，如 morning-session, afternoon-session, close-session"`
}

// QuerySessionsByDateTool 按日期查询 Session 工具
type QuerySessionsByDateTool struct{}

func (t *QuerySessionsByDateTool) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(t.GetName(), t.GetDescription(), QuerySessionsByDate)
}

func (t *QuerySessionsByDateTool) GetName() string {
	return "query_sessions_by_date"
}

func (t *QuerySessionsByDateTool) GetDescription() string {
	return "按日期查询 Session 记录 | Query sessions by date (supports today/yesterday)"
}

func (t *QuerySessionsByDateTool) GetToolGroup() model.ToolGroup {
	return model.KnowledgeToolGroup
}

// QuerySessionsByDate 按日期查询 Session
func QuerySessionsByDate(ctx context.Context, param *QuerySessionsByDateParam) (string, error) {
	message.BroadcastToolStart("query_sessions_by_date", fmt.Sprintf("日期: %s, 标签: %s", param.Date, param.Tag))

	// 解析日期
	targetDate, err := parseDate(param.Date)
	if err != nil {
		message.BroadcastToolEnd("query_sessions_by_date", "", err)
		return "", err
	}

	// 获取 Memory Manager
	manager := memory.GetManager()
	store := manager.GetStore()
	if store == nil {
		err := fmt.Errorf("记忆系统未初始化")
		message.BroadcastToolEnd("query_sessions_by_date", "", err)
		return "", err
	}

	// 列出所有会话（获取足够多的会话来筛选）
	sessions, err := store.ListSessions(0, 100)
	if err != nil {
		message.BroadcastToolEnd("query_sessions_by_date", "", err)
		return "", fmt.Errorf("查询会话失败: %w", err)
	}

	// 按日期和标签过滤
	var filtered []model.SessionIndex
	for _, s := range sessions {
		// 检查日期
		if !isSameDate(s.StartTime, targetDate) {
			continue
		}

		// 检查标签（如果指定）
		if param.Tag != "" {
			if !containsTag(s.Tags, param.Tag) {
				continue
			}
		}

		filtered = append(filtered, s)
	}

	// 格式化输出
	result := formatSessionsResult(targetDate, filtered, param.Tag)
	message.BroadcastToolEnd("query_sessions_by_date", result, nil)
	return result, nil
}

// parseDate 解析日期参数
func parseDate(dateStr string) (time.Time, error) {
	now := time.Now()

	switch strings.ToLower(dateStr) {
	case "today":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()), nil
	case "yesterday":
		yesterday := now.AddDate(0, 0, -1)
		return time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location()), nil
	default:
		// 尝试解析 YYYY-MM-DD 格式
		return time.Parse("2006-01-02", dateStr)
	}
}

// isSameDate 检查两个时间是否同一天
func isSameDate(t time.Time, date time.Time) bool {
	return t.Year() == date.Year() && t.Month() == date.Month() && t.Day() == date.Day()
}

// containsTag 检查标签列表是否包含指定标签
func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}
	return false
}

// formatSessionsResult 格式化会话查询结果
func formatSessionsResult(date time.Time, sessions []model.SessionIndex, tag string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("日期: %s\n", date.Format("2006-01-02")))
	if tag != "" {
		sb.WriteString(fmt.Sprintf("标签过滤: %s\n", tag))
	}
	sb.WriteString("\n")

	if len(sessions) == 0 {
		sb.WriteString("未找到符合条件的会话")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("找到 %d 个会话：\n\n", len(sessions)))

	for i, s := range sessions {
		sb.WriteString(fmt.Sprintf("### 会话 %d\n", i+1))
		sb.WriteString(fmt.Sprintf("- ID: %s\n", s.ID))
		sb.WriteString(fmt.Sprintf("- 开始时间: %s\n", s.StartTime.Format("15:04:05")))
		sb.WriteString(fmt.Sprintf("- 消息数: %d\n", s.MessageCount))
		if len(s.Tags) > 0 {
			sb.WriteString(fmt.Sprintf("- 标签: %s\n", strings.Join(s.Tags, ", ")))
		}
		if s.Summary != "" {
			sb.WriteString(fmt.Sprintf("- 摘要: %s\n", s.Summary))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// AddSessionTagParam 添加 Session 标签参数
type AddSessionTagParam struct {
	SessionID string `json:"sessionId" jsonschema:"description=会话 ID，为空则使用当前会话"`
	Tag       string `json:"tag" jsonschema:"description=要添加的标签，如 morning-session, afternoon-session, close-session，required"`
}

// AddSessionTagTool 添加 Session 标签工具
type AddSessionTagTool struct{}

func (t *AddSessionTagTool) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(t.GetName(), t.GetDescription(), AddSessionTag)
}

func (t *AddSessionTagTool) GetName() string {
	return "add_session_tag"
}

func (t *AddSessionTagTool) GetDescription() string {
	return "为 Session 添加标签 | Add a tag to a session (morning-session/afternoon-session/close-session)"
}

func (t *AddSessionTagTool) GetToolGroup() model.ToolGroup {
	return model.KnowledgeToolGroup
}

// AddSessionTag 添加 Session 标签
func AddSessionTag(ctx context.Context, param *AddSessionTagParam) (string, error) {
	message.BroadcastToolStart("add_session_tag", fmt.Sprintf("会话: %s, 标签: %s", param.SessionID, param.Tag))

	// 验证标签格式
	if err := validateTag(param.Tag); err != nil {
		message.BroadcastToolEnd("add_session_tag", "", err)
		return "", err
	}

	// 标准化标签（转小写）
	tag := strings.ToLower(param.Tag)

	// 获取 Memory Manager
	manager := memory.GetManager()

	// 确定目标会话 ID
	sessionID := param.SessionID
	if sessionID == "" {
		// 使用当前会话
		sessionID = manager.GetSessionID()
		if sessionID == "" {
			err := fmt.Errorf("当前没有活动会话")
			message.BroadcastToolEnd("add_session_tag", "", err)
			return "", err
		}
	}

	// 添加标签到会话
	// 注意：当前实现中，会话在结束时才保存
	// 这里需要更新会话的标签，然后持久化
	err := addTagToSession(manager, sessionID, tag)
	if err != nil {
		message.BroadcastToolEnd("add_session_tag", "", err)
		return "", err
	}

	result := fmt.Sprintf("标签 '%s' 已添加到会话 %s", tag, sessionID)
	message.BroadcastToolEnd("add_session_tag", result, nil)
	return result, nil
}

// validateTag 验证标签格式
func validateTag(tag string) error {
	if tag == "" {
		return ErrInvalidTag
	}

	// 标签只允许字母、数字和连字符
	matched, err := regexp.MatchString("^[a-zA-Z0-9-]+$", tag)
	if err != nil {
		return err
	}
	if !matched {
		return ErrInvalidTag
	}

	return nil
}

// addTagToSession 添加标签到会话
func addTagToSession(manager *memory.Manager, sessionID string, tag string) error {
	// 获取当前会话 ID
	currentSessionID := manager.GetSessionID()

	// 如果是当前会话，使用 Manager 的新方法添加标签
	if sessionID == currentSessionID {
		return manager.AddSessionTag(tag)
	}

	// 对于历史会话，需要加载、修改、保存
	store := manager.GetStore()
	if store == nil {
		return fmt.Errorf("存储未初始化")
	}

	// 加载会话
	session, err := store.GetSession(sessionID, time.Time{})
	if err != nil {
		return fmt.Errorf("加载会话失败: %w", err)
	}

	// 检查标签是否已存在
	for _, t := range session.Tags {
		if strings.EqualFold(t, tag) {
			// 标签已存在，幂等操作
			return nil
		}
	}

	// 添加标签
	session.Tags = append(session.Tags, tag)

	// 保存会话
	if err := store.SaveSession(session); err != nil {
		return fmt.Errorf("保存会话失败: %w", err)
	}

	log.Infof("为历史会话 %s 添加标签: %s", sessionID, tag)
	return nil
}
