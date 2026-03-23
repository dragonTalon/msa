package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"msa/pkg/utils"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"

	"msa/pkg/config"
	msamessage "msa/pkg/logic/message"
	"msa/pkg/logic/skills"
	tools2 "msa/pkg/logic/tools"
	"msa/pkg/model"
	msamodel "msa/pkg/model"

	log "github.com/sirupsen/logrus"
)

var agentCache *react.Agent

// ResetCache 清除 agent 缓存，用于模型切换时重新初始化
func ResetCache() {
	agentCache = nil
}

// maxHistoryRounds 最大保留的对话轮数（一轮 = 一条 User + 一条 Assistant）
const maxHistoryRounds = 10

func GetChatModel(ctx context.Context) (*react.Agent, error) {
	if agentCache != nil {
		return agentCache, nil
	}
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
		modelConfig.ExtraFields = map[string]any{
			"enable_thinking": true,
		}
	} else {
		modelConfig.ReasoningEffort = openai.ReasoningEffortLevelMedium
	}

	chatModel, err := openai.NewChatModel(ctx, modelConfig)
	if err != nil {
		return nil, err
	}

	allTools := tools2.GetAllTools()
	log.Infof("tools: %v", len(allTools))
	for _, tool := range allTools {
		tInfo, err := tool.Info(ctx)
		if err != nil {
			log.Errorf("tool info error: %v", err)
			continue
		}
		log.Infof("tool: %v", tInfo.Name)
	}

	tools := compose.ToolsNodeConfig{
		Tools: allTools,
	}

	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel:      chatModel,
		ToolsConfig:           tools,
		MaxStep:               100,
		StreamToolCallChecker: toolCallChecker,
	})
	if err != nil {
		return nil, err
	}
	agentCache = agent
	return agent, nil
}

func toolCallChecker(ctx context.Context, sr *schema.StreamReader[*schema.Message]) (bool, error) {
	defer sr.Close()

	streamManager := msamessage.GetStreamManager()

	for {
		msg, err := sr.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				streamManager.Broadcast(&msamodel.StreamChunk{
					IsDone: true,
				})
				break
			}
			streamManager.Broadcast(&msamodel.StreamChunk{
				Err:    err,
				IsDone: true,
			})
			return false, err
		}

		isToolCall := len(msg.ToolCalls) > 0
		if isToolCall {
			log.Infof("tool call: %v", msg.ToolCalls)
			return true, nil
		}
		if msg.ReasoningContent != "" {
			streamManager.Broadcast(&msamodel.StreamChunk{
				Content: msg.ReasoningContent,
				MsgType: model.StreamMsgTypeReason,
			})
			continue
		}
		streamManager.Broadcast(&msamodel.StreamChunk{
			Content: msg.Content,
			MsgType: model.StreamMsgTypeText,
		})
	}

	return false, nil
}

// Ask 聊天（异步处理流式数据，通过 StreamOutputManager 广播）
// 流程：
//  1. 初始化 skills，确保 get_skill_content tool 可正常使用
//  2. 根据 skill 元数据构建 system prompt，通过 BuildQueryMessages 拼接消息
//  3. 异步调用 agent.Stream 获取流式结果
func Ask(ctx context.Context, messages string, history []model.Message) error {
	// ===== 第一步：初始化 skills =====
	skillManager := skills.GetManager()
	if err := skillManager.Initialize(); err != nil {
		log.Warnf("skills initialize warning: %v", err)
	}

	// ===== 第二步：构建消息 =====
	filteredHistory := filterAndTrimHistory(history, maxHistoryRounds)
	historyMessages := convertToSchemaMessages(filteredHistory)

	now := time.Now()
	queryMsg, err := BuildQueryMessages(ctx, messages, historyMessages, map[string]any{
		"role":    "专业股票分析助手",
		"style":   "理性、专业、客观且严谨",
		"time":    now.Format("2006-01-02 15:04:05"),
		"weekday": now.Format("2006年01月02日 星期Monday"),
	})
	if err != nil {
		log.Errorf("Failed to build query messages: %v", err)
		return err
	}

	log.Infof("queryMsg count: %d", len(queryMsg))
	log.Infof("query  msg\n %s", utils.ToJSONString(queryMsg))

	// ===== 第三步：获取 Agent 并异步调用 Stream =====
	agent, err := GetChatModel(ctx)
	if err != nil {
		return err
	}

	go func() {
		_, err := agent.Stream(ctx, queryMsg)
		if err != nil {
			log.Errorf("流式请求失败: %v", err)
			msamessage.GetStreamManager().Broadcast(&msamodel.StreamChunk{
				Err:    err,
				IsDone: true,
			})
			return
		}
	}()

	return nil
}

// convertToSchemaMessages 将 model.Message 切片转为 eino schema.Message 切片
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

// filterAndTrimHistory 过滤历史消息，只保留 User 和 Assistant 角色，并限制最大轮数
func filterAndTrimHistory(history []model.Message, maxRounds int) []model.Message {
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
