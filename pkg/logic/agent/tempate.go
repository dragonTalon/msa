package agent

import (
	"context"
	"time"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	log "github.com/sirupsen/logrus"
)

func GetDefaultTemplate(ctx context.Context, question string, history []*schema.Message) ([]*schema.Message, error) {
	// 构建股票分析专用的消息模板
	historyTag := false
	if len(history) > 0 {
		historyTag = true
	}
	template := prompt.FromMessages(schema.GoTemplate,
		// 系统消息模板 - 定义股票分析助手的角色和行为准则
		schema.SystemMessage(BasePromptTemplate),

		// 对话历史占位符（兼容多轮对话）
		schema.MessagesPlaceholder("chat_history", historyTag),

		// 用户消息模板 - 接收用户的股票分析问题
		schema.UserMessage("问题: {{.question}}"),
	)

	// 格式化模板并填充参数
	now := time.Now()
	messages, err := template.Format(ctx, map[string]any{
		"role":         "专业股票分析助手",
		"style":        "理性、专业、客观且严谨",
		"time":         now.Format("2006-01-02 15:04:05"),
		"weekday":      now.Format("2006年01月02日 星期Monday"),
		"question":     question,
		"chat_history": history,
	})
	if err != nil {
		log.Errorf("格式化股票分析模板失败: %v", err)
		return nil, err
	}
	return messages, nil
}
