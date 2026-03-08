package agent

import (
	"context"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	log "github.com/sirupsen/logrus"
)

// BuildQueryMessages 使用 eino 的 prompt.FromMessages + schema.GoTemplate 构建完整的查询消息
// 流程：SystemMessage(BasePromptTemplate) -> chat_history -> UserMessage(BaseUserPrompt + skill + question)
func BuildQueryMessages(ctx context.Context, question string, history []*schema.Message, skillPrompt string, vars map[string]any) ([]*schema.Message, error) {
	hasHistory := len(history) > 0

	template := prompt.FromMessages(schema.GoTemplate,
		// 系统消息模板 - 角色定义 + 核心规范
		schema.SystemMessage(BasePromptTemplate),

		// 对话历史占位符（兼容多轮对话）
		schema.MessagesPlaceholder("chat_history", hasHistory),

		// 用户消息模板 - 技能模块 + 任务要求 + 用户问题
		schema.UserMessage(BaseUserPrompt),
	)

	// 合并模板变量
	templateVars := map[string]any{
		"question":     question,
		"skill_prompt": skillPrompt,
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
