package agent

import (
	"context"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	log "github.com/sirupsen/logrus"

	"msa/pkg/logic/skills"
)

// SkillMeta 注入模板的轻量级 Skill 元数据
type SkillMeta struct {
	Name        string
	Description string
	Priority    int
}

// BuildQueryMessages 使用 eino 的 prompt.FromMessages + schema.GoTemplate 构建完整的查询消息
// 流程：SystemMessage(BasePromptTemplate + skills) -> chat_history -> UserMessage(question)
func BuildQueryMessages(ctx context.Context, question string, history []*schema.Message, vars map[string]any) ([]*schema.Message, error) {
	hasHistory := len(history) > 0

	template := prompt.FromMessages(schema.GoTemplate,
		// 系统消息模板 - 角色定义 + 核心规范 + skill 列表
		schema.SystemMessage(BasePromptTemplate),

		// 对话历史占位符（兼容多轮对话）
		schema.MessagesPlaceholder("chat_history", hasHistory),

		// 用户消息模板 - 用户问题
		schema.UserMessage(BaseUserPrompt),
	)

	// 从全局 Manager 获取可用 skill 元数据列表
	skillMetas := buildSkillMetas()

	// 合并模板变量
	templateVars := map[string]any{
		"question":     question,
		"skills":       skillMetas,
		"chat_history": history,
	}
	// 注入外部传入的变量（role, style, time, weekday 等）
	for k, v := range vars {
		templateVars[k] = v
	}

	messages, err := template.Format(ctx, templateVars)
	if err != nil {
		log.Errorf("格式化查询消息模板失败: %v", err)
		return nil, err
	}
	return messages, nil
}

// buildSkillMetas 从全局 Manager 获取可用 skill 的轻量级元数据列表
func buildSkillMetas() []SkillMeta {
	manager := skills.GetManager()
	skillList := manager.ListSkills()
	metas := make([]SkillMeta, 0, len(skillList))
	for _, sk := range skillList {
		metas = append(metas, SkillMeta{
			Name:        sk.Name,
			Description: sk.Description,
			Priority:    sk.Priority,
		})
	}
	return metas
}
