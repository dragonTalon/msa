package provider

import (
	"context"
	"github.com/sirupsen/logrus"
	"msa/pkg/config"
	"msa/pkg/model"
	"msa/pkg/utils"

	// 导入 provider 实现，触发 init 注册
	_ "msa/pkg/logic/provider/siliconflow"
)

// Provider provider 接口
type LLMProvider interface {
	// GetProvider 获取提供商
	GetProvider(ctx context.Context) model.LlmProvider

	// ListModels 获取模型列表
	ListModels(ctx context.Context) ([]*model.LLMModel, error)
}

var providerMap = map[model.LlmProvider]LLMProvider{}

// RegisterProvider 注册 provider
func RegisterProvider(provider model.LlmProvider, p LLMProvider) {
	providerMap[provider] = p
	logrus.Infof("provider registered: %v", provider)
}

// GetProvider 获取 provider
func GetProvider() LLMProvider {
	cfg := config.GetLocalStoreConfig()

	logrus.Infof("GetProvider: %v", utils.ToJSONString(cfg))
	if cfg.Provider == "" {
		cfg.Provider = model.Siliconflow
		logrus.Infof("provider init: %v", utils.ToJSONString(cfg))
	}
	provider, ok := providerMap[cfg.Provider]
	logrus.Infof("GetProvider result: %v", ok)
	if ok {
		return provider
	}
	return nil
}
