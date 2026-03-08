package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
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

var (
	agentCache     *react.Agent
	chatModelCache *openai.ChatModel
)

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

	config := &openai.ChatModelConfig{
		Model:   cfg.Model,
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
	}
	if strings.HasPrefix(cfg.Model, "deepseek-") {
		config.ExtraFields = map[string]any{
			"enable_thinking": true,
		}
	} else {
		config.ReasoningEffort = openai.ReasoningEffortLevelMedium
	}

	// 创建并缓存 ChatModel
	chatModel, err := openai.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}
	chatModelCache = chatModel

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

	// 初始化所需的 tools
	tools := compose.ToolsNodeConfig{
		Tools: allTools,
	}

	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel:      chatModel,
		ToolsConfig:           tools,
		MaxStep:               40,
		StreamToolCallChecker: toolCallChecker,
	})
	if err != nil {
		return nil, err
	}
	agentCache = agent
	return agent, nil
}

// GetChatModelOnly 只返回 ChatModel（用于 Selector 等前置 sub-agent）
func GetChatModelOnly(ctx context.Context) (*openai.ChatModel, error) {
	if chatModelCache != nil {
		return chatModelCache, nil
	}
	// 触发 GetChatModel 来初始化
	_, err := GetChatModel(ctx)
	if err != nil {
		return nil, err
	}
	return chatModelCache, nil
}

func toolCallChecker(ctx context.Context, sr *schema.StreamReader[*schema.Message]) (bool, error) {
	defer sr.Close()

	streamManager := msamessage.GetStreamManager()

	for {
		msg, err := sr.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				// 发送结束标记
				streamManager.Broadcast(&msamodel.StreamChunk{
					IsDone: true,
				})
				break
			}
			// 发送错误
			streamManager.Broadcast(&msamodel.StreamChunk{
				Err:    err,
				IsDone: true,
			})
			return false, err
		}

		// 检查是否有 tool call
		isToolCall := len(msg.ToolCalls) > 0
		if isToolCall {
			log.Infof("tool call: %v", msg.ToolCalls)
			return true, nil
		}
		if msg.ReasoningContent != "" {
			// 广播给所有订阅者
			streamManager.Broadcast(&msamodel.StreamChunk{
				Content: msg.ReasoningContent,
				MsgType: model.StreamMsgTypeReason,
			})
			return false, nil
		}
		// 广播给所有订阅者
		streamManager.Broadcast(&msamodel.StreamChunk{
			Content: msg.Content,
			MsgType: model.StreamMsgTypeText,
		})
	}

	return false, nil
}

// Ask 聊天（异步处理流式数据，通过 StreamOutputManager 广播）
// 流程：
//  1. 同步调用 sub-agent（Selector）获取需要使用的 skills
//  2. 根据 skills 构建完整的 system prompt，通过 prompt.FromMessages 拼接消息模板
//  3. 异步调用 agent.Stream 获取流式结果
func Ask(ctx context.Context, messages string, history []model.Message, manualSkills []string) error {
	// ===== 第一步：通过 sub-agent 同步获取需要使用的 skills =====
	skillManager := skills.GetManager()
	selectedSkills, err := selectSkillsViaSubAgent(ctx, messages, history, manualSkills, skillManager)
	if err != nil {
		return err
	}

	// ===== 第二步：根据 skills 构建 system prompt，并通过 eino 模板拼接消息 =====
	// 构建 skill 信息文本（注入到 system prompt 中）
	skillPromptText := buildSkillPromptText(skillManager, selectedSkills)

	// 过滤并截取历史消息，转为 eino schema.Message
	filteredHistory := filterAndTrimHistory(history, maxHistoryRounds)
	historyMessages := convertToSchemaMessages(filteredHistory)

	// 使用 prompt.FromMessages 构建消息模板
	now := time.Now()
	queryMsg, err := BuildQueryMessages(ctx, messages, historyMessages, skillPromptText, map[string]any{
		"role":    "专业股票分析助手",
		"style":   "理性、专业、客观且严谨",
		"time":    now.Format("2006-01-02 15:04:05"),
		"weekday": now.Format("2006年01月02日 星期Monday"),
	})
	if err != nil {
		log.Errorf("Failed to build query messages: %v", err)
		return err
	}

	log.Infof("queryMsg: %v", queryMsg)

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

// selectSkillsViaSubAgent 通过前置 sub-agent 同步获取需要使用的 skills
// 如果 manualSkills 不为空，则直接使用手动指定的 skills（跳过 LLM 选择）
func selectSkillsViaSubAgent(ctx context.Context, userInput string, history []model.Message, manualSkills []string, skillManager *skills.Manager) ([]string, error) {
	// 初始化 Skills 系统
	if err := skillManager.Initialize(); err != nil {
		log.Errorf("Failed to initialize skills: %v", err)
		return nil, err
	}

	// 手动指定 skills：验证后直接返回
	if len(manualSkills) > 0 {
		validSkills := validateAndEnsureBase(manualSkills, skillManager)
		log.Infof("[Sub-Agent] Using manual skills: %v", validSkills)
		return validSkills, nil
	}

	// 通过 sub-agent (Selector) 调用 LLM Generate 自动选择 skills
	chatModel, err := GetChatModelOnly(ctx)
	if err != nil {
		return nil, err
	}

	selector := skills.NewSelector(chatModel, skillManager.GetRegistry())
	result, err := selector.SelectSkills(ctx, userInput, history)
	if err != nil {
		log.Warnf("[Sub-Agent] Skill selection failed: %v, fallback to base skill", err)
		return []string{}, nil
	}

	log.Infof("[Sub-Agent] Auto-selected skills: %v, reasoning: %s", result.SelectedSkills, result.Reasoning)
	return result.SelectedSkills, nil
}

// validateAndEnsureBase 验证手动指定的 skills 并确保包含 base
func validateAndEnsureBase(manualSkills []string, skillManager *skills.Manager) []string {
	validSkills := make([]string, 0, len(manualSkills))
	for _, skillName := range manualSkills {
		if _, err := skillManager.GetSkill(skillName); err != nil {
			log.Warnf("Skill '%s' not found, skipping", skillName)
		} else {
			validSkills = append(validSkills, skillName)
		}
	}
	return validSkills
}

// buildSkillPromptText 根据选中的 skills 构建 skill 描述文本，注入到 system prompt
func buildSkillPromptText(skillManager *skills.Manager, selectedSkills []string) string {
	registry := skillManager.GetRegistry()
	var sb strings.Builder
	sb.WriteString("\n## Skill List")
	for _, skillName := range selectedSkills {
		skill, err := registry.Get(skillName)
		if err != nil || skill == nil {
			log.Warnf("Skill not found when building prompt: %s", skillName)
			continue
		}
		content, err := skill.GetContent()
		if err != nil {
			return ""
		}
		sb.WriteString(fmt.Sprintf("\n- name: %s\n  describe: %s\n  content: %s", skill.Name, skill.Description, content))
	}
	return sb.String()
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
	// 先过滤：只保留 User 和 Assistant
	filtered := make([]model.Message, 0, len(history))
	for _, msg := range history {
		if msg.Role == model.RoleUser || msg.Role == model.RoleAssistant {
			filtered = append(filtered, msg)
		}
	}

	// 再修剪：限制最大轮数（一轮约 2 条消息）
	maxMessages := maxRounds * 2
	if len(filtered) > maxMessages {
		filtered = filtered[len(filtered)-maxMessages:]
	}

	return filtered
}
