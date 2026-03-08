package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

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

// GetChatModelOnly 只返回 ChatModel（用于 Selector）
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
// manualSkills: 手动指定的 skills 列表（优先级高于 LLM 自动选择），为空时使用自动选择
func Ask(ctx context.Context, messages string, history []model.Message, manualSkills []string) error {
	// 初始化 Skills 系统
	skillManager := skills.GetManager()
	if err := skillManager.Initialize(); err != nil {
		log.Warnf("Failed to initialize skills: %v", err)
		// 继续执行，使用基础 Prompt
	}

	var selectedSkills []string

	// 检查是否有手动指定的 skills
	if len(manualSkills) > 0 {
		// 使用手动指定的 skills
		selectedSkills = manualSkills
		log.Infof("Using manual skills: %v", selectedSkills)

		// 验证 skills 是否存在
		validSkills := make([]string, 0, len(manualSkills))
		for _, skillName := range manualSkills {
			if _, err := skillManager.GetSkill(skillName); err != nil {
				log.Warnf("Skill '%s' not found, skipping", skillName)
			} else {
				validSkills = append(validSkills, skillName)
			}
		}

		// 确保 base skill 总是被包含
		hasBase := false
		for _, skillName := range validSkills {
			if skillName == "base" {
				hasBase = true
				break
			}
		}
		if !hasBase {
			validSkills = append([]string{"base"}, validSkills...)
		}

		selectedSkills = validSkills
	} else {
		// 使用 LLM 自动选择
		chatModel, err := GetChatModelOnly(ctx)
		if err != nil {
			return err
		}

		selector := skills.NewSelector(chatModel, skillManager.GetRegistry())
		result, err := selector.SelectSkills(ctx, messages)
		if err != nil {
			log.Warnf("Skill selection failed: %v, using base skill", err)
			selectedSkills = []string{"base"}
		} else {
			selectedSkills = result.SelectedSkills
		}

		log.Infof("Auto-selected skills: %v", selectedSkills)
	}

	// 构建 System Prompt
	builder := skills.NewBuilder(skillManager.GetRegistry())
	systemPrompt, err := builder.BuildSystemPrompt(selectedSkills)
	if err != nil {
		log.Errorf("Failed to build system prompt: %v", err)
		return err
	}

	// 准备历史消息
	historyMsg := make([]*schema.Message, 0, len(history))
	log.Infof("history: %v", len(history))
	for _, msg := range history {
		role := schema.System
		switch msg.Role {
		case model.RoleSystem:
			role = schema.System
		case model.RoleUser:
			role = schema.User
		case model.RoleAssistant:
			role = schema.Assistant
		default:
			log.Warnf("unknown role: %s", msg.Role)
		}
		historyMsg = append(historyMsg, &schema.Message{
			Role:    role,
			Content: msg.Content,
		})
	}

	// 构建查询消息
	queryMsg := []*schema.Message{
		{
			Role:    schema.System,
			Content: systemPrompt,
		},
	}
	queryMsg = append(queryMsg, historyMsg...)
	queryMsg = append(queryMsg, &schema.Message{
		Role:    schema.User,
		Content: fmt.Sprintf("问题: %s", messages),
	})

	log.Infof("queryMsg: %v", queryMsg)

	// 获取 Agent
	agent, err := GetChatModel(ctx)
	if err != nil {
		return err
	}

	// 启动异步流式处理
	go func() {
		// 发送流式请求
		_, err := agent.Stream(ctx, queryMsg)
		if err != nil {
			log.Errorf("流式请求失败: %v", err)
			// 广播错误
			msamessage.GetStreamManager().Broadcast(&msamodel.StreamChunk{
				Err:    err,
				IsDone: true,
			})
			return
		}
	}()

	return nil
}
