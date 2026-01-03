package config

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"msa/pkg/model"
	"os"
	"path/filepath"
	"sync"
)

const CONFIG_FILE = "msa_config.json"
const LOG_FILE = "msa.log"

var (
	configCache     *LocalStoreConfig
	configCacheOnce sync.Once
)

// LocalStoreConfig 本地存储配置结构
type LocalStoreConfig struct {
	APIKey    string            `json:"apiKey"`
	BaseURL   string            `json:"baseURL"`
	Provider  model.LlmProvider `json:"provider"`
	Model     string            `json:"model"`
	LogConfig *LogConfig        `json:"logConfig,omitempty"`
}

// GetLocalStoreConfig 获取本地存储配置（带缓存）
func GetLocalStoreConfig() *LocalStoreConfig {
	configCacheOnce.Do(func() {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			configCache = &LocalStoreConfig{}
			return
		}

		configPath := filepath.Join(homeDir, ".msa", CONFIG_FILE)
		data, err := os.ReadFile(configPath)
		if err != nil {
			configCache = &LocalStoreConfig{}
			return
		}

		var cfg LocalStoreConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			configCache = &LocalStoreConfig{}
			return
		}

		configCache = &cfg
	})

	return configCache
}

// ReloadConfig 重新加载配置
func ReloadConfig() *LocalStoreConfig {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		configCache = &LocalStoreConfig{}
		return configCache
	}

	configPath := filepath.Join(homeDir, ".msa", CONFIG_FILE)
	data, err := os.ReadFile(configPath)
	if err != nil {
		configCache = &LocalStoreConfig{}
		return configCache
	}

	var cfg LocalStoreConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		configCache = &LocalStoreConfig{}
		return configCache
	}

	configCache = &cfg
	return configCache
}

// InitConfig 初始化配置文件和日志系统
func InitConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// 创建配置目录
	configDir := filepath.Join(homeDir, ".msa")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// 检查配置文件是否存在
	configPath := filepath.Join(configDir, CONFIG_FILE)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 配置文件不存在，创建默认配置
		defaultConfig := &LocalStoreConfig{
			APIKey:   "",
			BaseURL:  "",
			Provider: model.Siliconflow,
			LogConfig: &LogConfig{
				Level:      "info",
				Format:     "text",
				Output:     "file",
				File:       filepath.Join(configDir, LOG_FILE),
				ShowColor:  false,
				TimeFormat: "2006-01-02 15:04:05",
			},
		}

		// 写入默认配置
		data, err := json.MarshalIndent(defaultConfig, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return err
		}
	}

	// 加载配置
	cfg := ReloadConfig()

	// 初始化日志系统
	if cfg.LogConfig != nil {
		InitLogger(cfg.LogConfig)
	} else {
		cfg := DefaultConfig()
		cfg.File = filepath.Join(configDir, LOG_FILE)
		InitLogger(cfg)
	}

	return nil
}

// SaveConfig 保存配置到本地文件
func SaveConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// 创建配置目录
	configDir := filepath.Join(homeDir, ".msa")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// 写入配置
	configPath := filepath.Join(configDir, CONFIG_FILE)
	cfg := configCache
	log.Info("SaveConfig: %v", cfg)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return err
	}

	// 更新缓存
	configCache = cfg

	return nil
}
