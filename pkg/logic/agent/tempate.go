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
		schema.SystemMessage(`你是一个{{.role}}。你需要用{{.style}}的语气回答问题。
你的核心职责：
1. 基于金融理论和市场数据，对股票相关问题进行专业分析，包括但不限于行情解读、技术指标分析、基本面分析；
2. 严格遵守风险提示原则：明确告知所有分析仅为参考，不构成投资建议，投资有风险，入市需谨慎；
3. 分析过程需逻辑清晰，数据支撑，避免主观臆断，同时语言通俗易懂，兼顾专业度和可读性；
4. 若用户问题涉及具体股票代码/名称，需先确认标的基本信息，再从技术面、基本面、市场情绪等维度分析；
5. 拒绝提供明确的买入/卖出指令，仅提供分析维度和风险提示。
 {{.time}}`),

		// 对话历史占位符（兼容多轮对话）
		schema.MessagesPlaceholder("chat_history", historyTag),

		// 用户消息模板 - 接收用户的股票分析问题
		schema.UserMessage("问题: {{.question}}"),
	)

	// 格式化模板并填充参数
	messages, err := template.Format(ctx, map[string]any{
		"role":         "专业股票分析助手",
		"style":        "理性、专业、客观且严谨",
		"time":         time.Now().Format("2006-01-02 15:04:05"),
		"question":     question,
		"chat_history": history,
	})
	if err != nil {
		log.Errorf("格式化股票分析模板失败: %v", err)
		return nil, err
	}
	return messages, nil
}
