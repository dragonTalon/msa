package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"

	"msa/pkg/config"
	"msa/pkg/core/event"
	corelogger "msa/pkg/core/logger"
	"msa/pkg/logic/tools"
	"msa/pkg/model"

	log "github.com/sirupsen/logrus"
)

// Agent wraps the Eino ReAct agent, exposing only the Run() interface.
type Agent struct {
	einoAgent *react.Agent
	adapter   *StreamAdapter
}

// New creates an Agent from the current config.
// Creates a fresh instance each time (no global cache — just call New() again to switch models).
func New(ctx context.Context) (*Agent, error) {
	cfg := config.GetLocalStoreConfig()
	if cfg == nil {
		return nil, fmt.Errorf("config is nil, please run `msa config` first")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("openai api key is empty, please run `msa config` first")
	}
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("openai base url is empty, please run `msa config` first")
	}
	if cfg.Model == "" {
		return nil, fmt.Errorf("openai model is empty, please run `/models choose model` first")
	}

	modelConfig := &openai.ChatModelConfig{
		Model:   cfg.Model,
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
	}
	if strings.HasPrefix(cfg.Model, "deepseek-") {
		modelConfig.ExtraFields = map[string]any{"enable_thinking": true}
	} else {
		modelConfig.ReasoningEffort = openai.ReasoningEffortLevelMedium
	}

	chatModel, err := openai.NewChatModel(ctx, modelConfig)
	if err != nil {
		return nil, fmt.Errorf("创建 chat model 失败: %w", err)
	}

	allTools := tools.GetAllTools()
	log.Infof("[Agent] 注册工具数: %d", len(allTools))

	einoAgent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig:      compose.ToolsNodeConfig{Tools: allTools},
		MaxStep:          100,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 Eino agent 失败: %w", err)
	}

	return &Agent{
		einoAgent: einoAgent,
		adapter:   &StreamAdapter{},
	}, nil
}

// Run starts a conversation round (ReAct loop) and returns a read-only event channel.
// When ctx is cancelled, the channel is closed automatically and all goroutines exit cleanly.
// Usage: for e := range agent.Run(ctx, msgs) { ... }
func (a *Agent) Run(ctx context.Context, messages []*schema.Message) <-chan event.Event {
	ch := make(chan event.Event, 32)
	go a.run(ctx, messages, ch)
	return ch
}

func (a *Agent) run(ctx context.Context, messages []*schema.Message, ch chan<- event.Event) {
	defer close(ch) // channel close = termination signal; range loop in consumer exits automatically
	logger := corelogger.FromCtx(ctx)
	logger.Infof("[Agent] 开始对话, 历史消息数=%d", len(messages))

	round := 0
	for {
		round++
		logger.Infof("[Agent] 第 %d 轮 LLM 调用开始", round)

		sr, err := a.einoAgent.Stream(ctx, messages)
		if err != nil {
			logger.Errorf("[Agent] 第 %d 轮 LLM 调用失败: %v", round, err)
			sendEvent(ctx, ch, event.Event{Type: event.EventError, Err: err})
			return
		}
		logger.Infof("[Agent] 第 %d 轮 LLM stream 已建立", round)

		result, err := a.adapter.Process(ctx, sr, ch)
		if err != nil {
			if err == context.Canceled {
				logger.Warnf("[Agent] 第 %d 轮 context canceled，干净退出", round)
				return
			}
			logger.Errorf("[Agent] 第 %d 轮 stream 处理失败: %v", round, err)
			sendEvent(ctx, ch, event.Event{Type: event.EventError, Err: err})
			return
		}

		if len(result.ToolCalls) == 0 {
			logger.Infof("[Agent] 第 %d 轮无工具调用，对话结束", round)
			sendEvent(ctx, ch, event.Event{Type: event.EventRoundDone})
			return
		}

		logger.Infof("[Agent] 第 %d 轮需要调用工具: %v", round, toolCallNames(result.ToolCalls))

		toolMsgs := a.executeTools(ctx, result.ToolCalls, ch)
		messages = append(messages, result.AssistantMsg)
		messages = append(messages, toolMsgs...)
	}
}

// executeTools executes the tool calls, emitting EventToolStart/EventToolResult/EventToolError.
// Returns tool result messages to append to messages for the next ReAct round.
func (a *Agent) executeTools(
	ctx context.Context,
	calls []schema.ToolCall,
	ch chan<- event.Event,
) []*schema.Message {
	logger := corelogger.FromCtx(ctx)
	var toolMsgs []*schema.Message

	toolsMap := tools.GetToolsMap()

	for _, call := range calls {
		start := time.Now()
		toolName := call.Function.Name

		sendEvent(ctx, ch, event.Event{
			Type: event.EventToolStart,
			Tool: event.ToolCall{
				ID:    call.ID,
				Name:  toolName,
				Input: call.Function.Arguments,
			},
		})

		logger.Infof("[Tool] 开始执行: name=%s id=%s input=%s", toolName, call.ID, call.Function.Arguments)

		var output string
		var toolErr error

		msaTool, ok := toolsMap[toolName]
		if !ok {
			toolErr = fmt.Errorf("工具不存在: %s", toolName)
			output = fmt.Sprintf(`{"error": "工具不存在: %s"}`, toolName)
		} else {
			baseTool, err := msaTool.GetToolInfo()
			if err != nil {
				toolErr = err
				output = fmt.Sprintf(`{"error": "获取工具信息失败: %v"}`, err)
			} else {
				invokable, ok := baseTool.(tool.InvokableTool)
				if !ok {
					toolErr = fmt.Errorf("工具 %s 不支持 InvokableTool 接口", toolName)
					output = fmt.Sprintf(`{"error": "工具不支持调用: %s"}`, toolName)
				} else {
					output, toolErr = invokable.InvokableRun(ctx, call.Function.Arguments)
				}
			}
		}

		elapsed := time.Since(start)

		if toolErr != nil {
			logger.Errorf("[Tool] 执行失败: name=%s elapsed=%v err=%v", toolName, elapsed, toolErr)
			sendEvent(ctx, ch, event.Event{
				Type: event.EventToolError,
				Result: event.ToolResult{
					ToolCallID: call.ID,
					Name:       toolName,
					Output:     output,
					IsError:    true,
				},
				Err: toolErr,
			})
			output = fmt.Sprintf(`{"error": "%v"}`, toolErr)
		} else {
			logger.Infof("[Tool] 执行完成: name=%s elapsed=%v outputLen=%d", toolName, elapsed, len(output))
			sendEvent(ctx, ch, event.Event{
				Type: event.EventToolResult,
				Result: event.ToolResult{
					ToolCallID: call.ID,
					Name:       toolName,
					Output:     output,
					IsError:    false,
				},
			})
		}

		toolMsgs = append(toolMsgs, &schema.Message{
			Role:       schema.Tool,
			Content:    output,
			ToolCallID: call.ID,
		})
	}

	return toolMsgs
}

// BuildQueryMessages builds query messages (system prompt + history + user input).
// Delegates to the template.go implementation in this package.
func BuildQueryMessages(ctx context.Context, userInput string, history []*schema.Message, templateVars map[string]any) ([]*schema.Message, error) {
	return buildQueryMessages(ctx, userInput, history, templateVars)
}

// FilterAndTrimHistory filters and truncates history messages.
func FilterAndTrimHistory(history []model.Message, maxRounds int) []model.Message {
	filtered := make([]model.Message, 0, len(history))
	for _, msg := range history {
		if msg.Role == model.RoleUser || msg.Role == model.RoleAssistant {
			filtered = append(filtered, msg)
		}
	}
	maxMessages := maxRounds * 2
	if len(filtered) > maxMessages {
		filtered = filtered[len(filtered)-maxMessages:]
	}
	return filtered
}
