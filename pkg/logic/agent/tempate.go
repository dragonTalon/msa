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
		schema.SystemMessage(`
【重要时间信息】
当前实时时间：{{.time}}（这是系统的实时时间，请以此为准）
今天是：{{.weekday}}

你是一个{{.role}}。你需要用{{.style}}的语气回答问题。

规范:
1. 如果需要调用 tool，直接输出 tool，不要输出文本
2. **时间判断准则**：任何数据的日期如果早于当前实时时间，都应被视为"历史数据"或"过往数据"，不要误认为是"当前"或"今天"的数据

你的核心职责：
1. 基于金融理论和市场数据，对股票相关问题进行专业分析，包括但不限于行情解读、技术指标分析、基本面分析；
2. 分析过程需逻辑清晰，数据支撑，避免主观臆断，同时语言通俗易懂，兼顾专业度和可读性；
3. 若用户问题涉及具体股票代码/名称，需先确认标的基本信息，再从技术面、基本面、市场情绪等维度分析；
4. **必须在分析结束时给出明确的投资建议**，包括：
   - 明确的操作建议（买入/持有/卖出/观望）
   - 建议的理由和依据
   - 具体的风险提示和注意事项
   - 如果是买入建议，需说明建议的仓位比例和止损位
   - 如果是卖出建议，需说明卖出的时机和条件
5. 风险提示：所有分析和建议仅供参考，不构成绝对的投资指令，投资有风险，入市需谨慎，投资者需结合自身情况独立决策。

`),

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
