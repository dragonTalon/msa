package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"msa/pkg/config"
	tools2 "msa/pkg/logic/tools"
	"msa/pkg/model"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	log "github.com/sirupsen/logrus"
)

var agentCache *react.Agent

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
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   cfg.Model,
		APIKey:  cfg.APIKey,  // OpenAI API 密钥
		BaseURL: cfg.BaseURL, // OpenAI 基础 URL
	})
	allTools := tools2.GetAllTools()
	log.Infof("tools: %v", len(allTools))

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

func toolCallChecker(ctx context.Context, sr *schema.StreamReader[*schema.Message]) (bool, error) {
	defer sr.Close()

	streamManager := GetStreamManager()

	for {
		msg, err := sr.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				// 发送结束标记
				streamManager.Broadcast(&StreamChunk{
					IsDone: true,
				})
				break
			}
			// 发送错误
			streamManager.Broadcast(&StreamChunk{
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
		// 广播给所有订阅者
		streamManager.Broadcast(&StreamChunk{
			Content: msg.Content,
		})
	}

	return false, nil
}

// Ask 聊天（异步处理流式数据，通过 StreamOutputManager 广播）
func Ask(ctx context.Context, messages string, history []model.Message) error {
	chatModel, err := GetChatModel(ctx)
	if err != nil {
		return err
	}
	historyMsg := make([]*schema.Message, 0, len(history)) // ✅ 使用 0 长度，但预分配容量
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

	queryMsg, err := GetDefaultTemplate(ctx, messages, historyMsg)
	log.Infof("queryMsg: %v", queryMsg)
	if err != nil {
		log.Errorf("获取模板失败: %v", err)
		return err
	}

	// 启动异步流式处理
	go func() {
		streamResult, err := chatModel.Stream(ctx, queryMsg)
		if err != nil {
			log.Errorf("流式请求失败: %v", err)
			// 广播错误
			GetStreamManager().Broadcast(&StreamChunk{
				Err:    err,
				IsDone: true,
			})
			return
		}
		log.Infof("streamResult: %v", streamResult)
	}()

	return nil
}
