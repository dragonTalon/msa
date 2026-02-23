package siliconflow

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"msa/pkg/config"
	"msa/pkg/model"
	"msa/pkg/utils"
)

var SILICONFLOW_BASE_URL = "https://api.siliconflow.cn/v1"

// SiliconflowProvider 表示 SiliconFlow API 客户端
type SiliconflowProvider struct {
}

// GetProvider 获取 provider 类型
func (c SiliconflowProvider) GetProvider(ctx context.Context) model.LlmProvider {
	return model.Siliconflow
}

// ListModels 获取可用的模型列表
func (c SiliconflowProvider) ListModels(ctx context.Context) ([]*model.LLMModel, error) {
	var result ModelsResponse

	client := utils.GetRestyClient()

	cfg := config.GetLocalStoreConfig()
	baseURL := SILICONFLOW_BASE_URL
	if cfg.BaseURL != "" {
		baseURL = cfg.BaseURL
	}

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", cfg.APIKey)).
		SetResult(&result).
		Get(fmt.Sprintf("%s/models?type=text&sub_type=chat", baseURL))

	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("API 请求失败 (状态码 %d): %s", resp.StatusCode(), resp.String())
	}
	log.Infof("ListModels: %s", utils.ToJSONString(result))
	// 转换为 LlModel 格式
	models := make([]*model.LLMModel, len(result.Data))
	for i, m := range result.Data {
		models[i] = &model.LLMModel{
			Name:        m.ID,
			Description: m.Object,
		}
	}

	return models, nil
}
