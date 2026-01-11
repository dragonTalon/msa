package agent

import (
	"context"
	"fmt"
	"msa/pkg/config"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	log "github.com/sirupsen/logrus"
)

var chatCache *openai.ChatModel

func GetChatModel(ctx context.Context) (*openai.ChatModel, error) {
	if chatCache != nil {
		return chatCache, nil
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
	chatCache = chatModel
	return chatCache, err
}

// Chat 聊天
func Ask(ctx context.Context, messages string) (*schema.StreamReader[*schema.Message], error) {
	chatModel, err := GetChatModel(ctx)
	log.Infof("chatModel: %v, err: %v", chatModel, err)
	if err != nil {
		return nil, err
	}
	queryMsg, err := GetDefaultTemplate(ctx, messages)
	log.Infof("queryMsg: %v, err: %v", queryMsg, err)
	if err != nil {
		log.Errorf("获取模板失败: %v", err)
		return nil, err
	}
	streamResult, err := chatModel.Stream(ctx, queryMsg)
	if err != nil {
		log.Errorf("流式请求失败: %v", err)
		return nil, err
	}
	return streamResult, nil
}
