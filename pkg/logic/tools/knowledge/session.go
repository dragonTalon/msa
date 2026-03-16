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

// SessionItem 会话项
type SessionItem struct {
	ID           string   `json:"id"`
	StartTime    string   `json:"start_time"`
	MessageCount int      `json:"message_count"`
	Tags         []string `json:"tags,omitempty"`
	Summary      string   `json:"summary,omitempty"`
}

// QuerySessionsData 查询会话数据
type QuerySessionsData struct {
	Date     string        `json:"date"`
	Tag      string        `json:"tag,omitempty"`
	Total    int           `json:"total"`
	Sessions []SessionItem `json:"sessions"`
}

// QuerySessionsByDate 按日期查询 Session
func QuerySessionsByDate(ctx context.Context, param *QuerySessionsByDateParam) (string, error) {
	message.BroadcastToolStart("query_sessions_by_date", fmt.Sprintf("日期: %s, 标签: %s", param.Date, param.Tag))

	// 解析日期
	targetDate, err := parseDate(param.Date)
	if err != nil {
		message.BroadcastToolEnd("query_sessions_by_date", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	// 获取 Memory Manager
	manager := memory.GetManager()
	store := manager.GetStore()
	if store == nil {
		err := fmt.Errorf("记忆系统未初始化")
		message.BroadcastToolEnd("query_sessions_by_date", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	// 列出所有会话（获取足够多的会话来筛选）
	sessions, err := store.ListSessions(0, 100)
	if err != nil {
		message.BroadcastToolEnd("query_sessions_by_date", "", err)
		return model.NewErrorResult(fmt.Sprintf("查询会话失败: %v", err)), nil
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

	// 构建返回数据
	items := make([]SessionItem, 0, len(filtered))
	for _, s := range filtered {
		items = append(items, SessionItem{
			ID:           s.ID,
			StartTime:    s.StartTime.Format("15:04:05"),
			MessageCount: s.MessageCount,
			Tags:         s.Tags,
			Summary:      s.Summary,
		})
	}

	data := &QuerySessionsData{
		Date:     targetDate.Format("2006-01-02"),
		Tag:      param.Tag,
		Total:    len(filtered),
		Sessions: items,
	}

	message.BroadcastToolEnd("query_sessions_by_date", fmt.Sprintf("找到 %d 个会话", len(filtered)), nil)
	return model.NewSuccessResult(data, fmt.Sprintf("找到 %d 个会话", len(filtered))), nil
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

// AddSessionTagData 添加会话标签数据
type AddSessionTagData struct {
	SessionID string `json:"session_id"`
	Tag       string `json:"tag"`
}

// AddSessionTag 添加 Session 标签
func AddSessionTag(ctx context.Context, param *AddSessionTagParam) (string, error) {
	message.BroadcastToolStart("add_session_tag", fmt.Sprintf("会话: %s, 标签: %s", param.SessionID, param.Tag))

	// 验证标签格式
	if err := validateTag(param.Tag); err != nil {
		message.BroadcastToolEnd("add_session_tag", "", err)
		return model.NewErrorResult(err.Error()), nil
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
			return model.NewErrorResult(err.Error()), nil
		}
	}

	// 添加标签到会话
	err := addTagToSession(manager, sessionID, tag)
	if err != nil {
		message.BroadcastToolEnd("add_session_tag", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	data := &AddSessionTagData{
		SessionID: sessionID,
		Tag:       tag,
	}

	message.BroadcastToolEnd("add_session_tag", fmt.Sprintf("标签 '%s' 已添加", tag), nil)
	return model.NewSuccessResult(data, fmt.Sprintf("标签 '%s' 已添加到会话", tag)), nil
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
