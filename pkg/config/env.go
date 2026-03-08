package config

import (
	"os"
	"strings"

	"msa/pkg/model"
)

// LoadFromEnv 从环境变量加载配置
// 环境变量：MSA_PROVIDER, MSA_API_KEY, MSA_BASE_URL, MSA_LOG_LEVEL, MSA_LOG_FILE
func LoadFromEnv() *LocalStoreConfig {
	cfg := &LocalStoreConfig{}

	// MSA_PROVIDER
	if provider := os.Getenv("MSA_PROVIDER"); provider != "" {
		cfg.Provider = model.LlmProvider(provider)
	}

	// MSA_API_KEY
	if apiKey := os.Getenv("MSA_API_KEY"); apiKey != "" {
		cfg.APIKey = apiKey
	}

	// MSA_BASE_URL
	if baseURL := os.Getenv("MSA_BASE_URL"); baseURL != "" {
		cfg.BaseURL = baseURL
	}

	// MSA_LOG_LEVEL
	if logLevel := os.Getenv("MSA_LOG_LEVEL"); logLevel != "" {
		// 转换为小写
		logLevel = strings.ToLower(logLevel)
		// 验证是否为有效日志级别
		if isValidLogLevel(logLevel) {
			if cfg.LogConfig == nil {
				cfg.LogConfig = &LogConfig{}
			}
			cfg.LogConfig.Level = logLevel
		}
	}

	// MSA_LOG_FILE
	if logFile := os.Getenv("MSA_LOG_FILE"); logFile != "" {
		if cfg.LogConfig == nil {
			cfg.LogConfig = &LogConfig{}
		}
		cfg.LogConfig.File = logFile
	}

	// 如果没有任何环境变量被设置，返回 nil
	if cfg.Provider == "" && cfg.APIKey == "" && cfg.BaseURL == "" && cfg.LogConfig == nil {
		return nil
	}

	return cfg
}

// isValidLogLevel 验证日志级别是否有效
func isValidLogLevel(level string) bool {
	validLevels := map[string]bool{
		"debug":   true,
		"info":    true,
		"warn":    true,
		"warning": true,
		"error":   true,
		"fatal":   true,
		"panic":   true,
	}
	return validLevels[level]
}
