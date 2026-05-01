package deepseek

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	"msa/pkg/config"
	"msa/pkg/model"
	"msa/pkg/utils"
)

var DEEPSEEK_BASE_URL = "https://api.deepseek.com"

// DeepseekProvider DeepSeek API 提供商
type DeepseekProvider struct{}

// GetProvider 获取 provider 类型
func (p DeepseekProvider) GetProvider(ctx context.Context) model.LlmProvider {
	return model.Deepseek
}

// ListModels 获取可用的模型列表
func (p DeepseekProvider) ListModels(ctx context.Context) ([]*model.LLMModel, error) {
	var result ModelsResponse

	client := utils.GetRestyClient()

	cfg := config.GetLocalStoreConfig()
	baseURL := DEEPSEEK_BASE_URL
	if cfg.BaseURL != "" {
		baseURL = cfg.BaseURL
	}

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", cfg.APIKey)).
		SetResult(&result).
		Get(fmt.Sprintf("%s/models", baseURL))

	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("API 请求失败 (状态码 %d): %s", resp.StatusCode(), resp.String())
	}

	log.Infof("DeepSeek ListModels: %s", utils.ToJSONString(result))

	models := make([]*model.LLMModel, len(result.Data))
	for i, m := range result.Data {
		models[i] = &model.LLMModel{
			Name:        m.ID,
			Description: m.OwnedBy,
		}
	}

	return models, nil
}
