// Package runner provides the unified entry point for MSA conversations.
// Both CLI and TUI use Runner.Ask() to start a conversation round.
package runner

import (
	"context"
	"fmt"
	"msa/pkg/utils"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"

	"msa/pkg/core/agent"
	"msa/pkg/core/event"
	corelogger "msa/pkg/core/logger"
	"msa/pkg/logic/skills"
	"msa/pkg/model"
	"msa/pkg/renderer"
	"msa/pkg/session"

	log "github.com/sirupsen/logrus"
)

// Runner is the core layer's external entry point.
// Both CLI and TUI start conversations through Runner.Ask().
type Runner struct {
	agent      *agent.Agent
	sessionMgr *session.Manager
	renderer   renderer.Renderer
}

// New creates a Runner.
func New(ag *agent.Agent, sessionMgr *session.Manager, r renderer.Renderer) *Runner {
	return &Runner{
		agent:      ag,
		sessionMgr: sessionMgr,
		renderer:   r,
	}
}

// Ask handles one user input round.
// CLI calls this once; TUI calls it on each Enter press.
// Blocks until the conversation round completes or ctx is cancelled.
func (r *Runner) Ask(ctx context.Context, input string, history []model.Message) error {
	ctx, reqID := corelogger.WithRequestID(ctx)
	logger := corelogger.FromCtx(ctx)
	logger.Infof("[Runner] 收到输入 reqID=%s len=%d", reqID, len(input))

	// Initialize skills
	skillManager := skills.GetManager()
	if err := skillManager.Initialize(); err != nil {
		log.Warnf("[Runner] skills initialize warning: %v", err)
	}

	// Filter and trim history
	filteredHistory := agent.FilterAndTrimHistory(history, 20)
	schemaHistory := convertToSchemaMessages(filteredHistory)

	// Build query messages (system prompt + history + user input)
	now := time.Now()
	log.Infof("[Runner] 构建消息开始 ，历史消息 %d 条", len(schemaHistory))
	messages, err := agent.BuildQueryMessages(ctx, input, schemaHistory, map[string]any{
		"role":    "专业股票分析助手",
		"style":   "理性、专业、客观且严谨",
		"time":    now.Format("2006-01-02 15:04:05"),
		"weekday": now.Format("2006年01月02日 星期Monday"),
	})
	if err != nil {
		return fmt.Errorf("构建消息失败: %w", err)
	}

	logger.Infof("[Runner] 构建消息完成, 共 %d 条", len(messages))

	// Start agent, get event channel
	eventCh := r.agent.Run(ctx, messages)

	// Consume events, hand to renderer; collect full reply for persistence
	var assistantReply strings.Builder
	for e := range eventCh {
		log.Infof("[Runner] 收到事件: %v", utils.ToJSONString(e))
		if err := r.renderer.Handle(ctx, e); err != nil {
			return err
		}
		if e.Type == event.EventTextChunk {
			assistantReply.WriteString(e.Text)
		}
	}

	logger.Infof("[Runner] 本轮对话完成 replyLen=%d", assistantReply.Len())
	return nil
}

// convertToSchemaMessages converts model.Message slice to Eino schema.Message slice.
func convertToSchemaMessages(history []model.Message) []*schema.Message {
	msgs := make([]*schema.Message, 0, len(history))
	for _, msg := range history {
		var role schema.RoleType
		switch msg.Role {
		case model.RoleUser:
			role = schema.User
		case model.RoleAssistant:
			role = schema.Assistant
		default:
			continue
		}
		msgs = append(msgs, &schema.Message{
			Role:    role,
			Content: msg.Content,
		})
	}
	return msgs
}
