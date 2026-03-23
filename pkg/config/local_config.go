package config

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"msa/pkg/db"
	"msa/pkg/model"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const CONFIG_FILE = "msa_config.json"
const LOG_FILE = "msa.log"

var (
	configCache     *LocalStoreConfig
	configCacheOnce sync.Once
	// cliConfig 从命令行参数加载的配置
	cliConfig *LocalStoreConfig
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
		log.Warnf("无法获取用户主目录: %v", err)
		// 继续使用默认配置
	}

	// 创建配置目录
	configDir := filepath.Join(homeDir, ".msa")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Warnf("无法创建配置目录: %v", err)
		// 继续执行，可能使用内存配置
	}

	// 使用 ConfigBuilder 构建配置
	builder := NewConfigBuilder()

	// 1. 加载配置文件
	var fileCfg *LocalStoreConfig
	configPath := filepath.Join(configDir, CONFIG_FILE)
	if data, err := os.ReadFile(configPath); err == nil {
		var cfg LocalStoreConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			// JSON 解析失败
			log.Warnf("配置文件格式错误，已使用默认配置: %v", err)
			log.Info("运行 'msa config' 重新配置")
		} else {
			fileCfg = &cfg
		}
	} else if !os.IsNotExist(err) {
		// 读取失败（不是文件不存在）
		log.Warnf("无法读取配置文件，使用默认配置: %v", err)
	}
	builder.WithFileConfig(fileCfg)

	// 2. 加载环境变量
	envCfg := LoadFromEnv()
	builder.WithEnvConfig(envCfg)

	// 3. 应用命令行参数
	builder.WithCLIConfig(cliConfig)

	// 4. 构建最终配置
	finalCfg := builder.Build()

	// 5. 验证配置
	if errors := ValidateConfig(finalCfg); len(errors) > 0 {
		// 只记录错误，不阻止启动
		for _, e := range errors {
			if e.Severity == SeverityError {
				log.Warnf("配置验证失败: %s", e.Error())
			}
		}
		log.Info("运行 'msa config' 修复配置问题")
	}

	// 6. 如果配置文件不存在，创建默认配置文件
	configPath = filepath.Join(configDir, CONFIG_FILE)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultCfg := GetDefaultConfig()
		data, err := json.MarshalIndent(defaultCfg, "", "  ")
		if err == nil {
			os.WriteFile(configPath, data, 0644)
		}
	}

	// 7. 更新缓存
	configCache = finalCfg

	// 8. 初始化日志系统
	if finalCfg.LogConfig != nil {
		// 确保日志路径设置
		if finalCfg.LogConfig.File == "" {
			finalCfg.LogConfig.File = filepath.Join(configDir, LOG_FILE)
		}
		if finalCfg.LogConfig.ErrFile == "" {
			finalCfg.LogConfig.ErrFile = filepath.Join(configDir, ERR_LOG_FILE)
		}
		InitLogger(finalCfg.LogConfig)
	} else {
		defaultLogConfig := DefaultConfig()
		defaultLogConfig.File = filepath.Join(configDir, LOG_FILE)
		defaultLogConfig.ErrFile = filepath.Join(configDir, ERR_LOG_FILE)
		InitLogger(defaultLogConfig)
	}

	// 8. 初始化 SQLite（非阻塞）
	db.InitGlobalDB()

	return nil
}

// SaveConfig 保存配置到本地文件
func SaveConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Errorf("SaveConfig: 获取用户主目录失败: %v", err)
		return err
	}

	// 创建配置目录
	configDir := filepath.Join(homeDir, ".msa")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Errorf("SaveConfig: 创建配置目录失败: %v", err)
		return err
	}

	// 写入配置
	configPath := filepath.Join(configDir, CONFIG_FILE)
	cfg := configCache

	log.Infof("SaveConfig: 准备保存配置到 %s", configPath)
	log.Infof("SaveConfig: 配置内容 Provider=%q APIKey=%q BaseURL=%q Model=%q",
		cfg.Provider, maskKey(cfg.APIKey), cfg.BaseURL, cfg.Model)

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		log.Errorf("SaveConfig: 序列化配置失败: %v", err)
		return err
	}

	log.Infof("SaveConfig: JSON 数据: %s", string(data))

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		log.Errorf("SaveConfig: 写入文件失败: %v", err)
		return err
	}

	log.Infof("SaveConfig: 文件写入成功，大小=%d 字节", len(data))

	// 验证文件内容
	verifyData, err := os.ReadFile(configPath)
	if err != nil {
		log.Errorf("SaveConfig: 验证读取失败: %v", err)
		return err
	}
	log.Infof("SaveConfig: 验证读取成功，大小=%d 字节", len(verifyData))

	// 重新加载配置以验证保存是否成功
	var reloaded LocalStoreConfig
	if err := json.Unmarshal(verifyData, &reloaded); err != nil {
		log.Errorf("SaveConfig: 重新加载解析失败: %v", err)
		return err
	}

	configCache = &reloaded
	log.Infof("SaveConfig: 配置已保存并重新加载成功")
	return nil
}

// maskKey 隐藏密钥用于日志
func maskKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

// SetCLIConfig 设置命令行参数配置
func SetCLIConfig(cfg *LocalStoreConfig) {
	cliConfig = cfg
}

// GetDefaultConfig 获取默认配置
func GetDefaultConfig() *LocalStoreConfig {
	return &LocalStoreConfig{
		Provider: model.Siliconflow,
		LogConfig: &LogConfig{
			Level:      "info",
			Format:     "text",
			Output:     "file",
			ShowColor:  false,
			TimeFormat: "2006-01-02 15:04:05",
		},
	}
}

// ConfigBuilder 配置构建器
type ConfigBuilder struct {
	defaultCfg *LocalStoreConfig
	fileCfg    *LocalStoreConfig
	envCfg     *LocalStoreConfig
	cliCfg     *LocalStoreConfig
}

// NewConfigBuilder 创建配置构建器
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		defaultCfg: GetDefaultConfig(),
	}
}

// WithFileConfig 设置配置文件
func (b *ConfigBuilder) WithFileConfig(cfg *LocalStoreConfig) *ConfigBuilder {
	b.fileCfg = cfg
	return b
}

// WithEnvConfig 设置环境变量配置
func (b *ConfigBuilder) WithEnvConfig(cfg *LocalStoreConfig) *ConfigBuilder {
	b.envCfg = cfg
	return b
}

// WithCLIConfig 设置命令行配置
func (b *ConfigBuilder) WithCLIConfig(cfg *LocalStoreConfig) *ConfigBuilder {
	b.cliCfg = cfg
	return b
}

// Build 构建最终配置
func (b *ConfigBuilder) Build() *LocalStoreConfig {
	result := b.defaultCfg

	// 按优先级从低到高合并
	result = b.merge(result, b.fileCfg)
	result = b.merge(result, b.envCfg)
	result = b.merge(result, b.cliCfg)

	return result
}

// merge 合并配置，只覆盖非零值
func (b *ConfigBuilder) merge(base, override *LocalStoreConfig) *LocalStoreConfig {
	if override == nil {
		return base
	}

	result := base

	// 合并简单字段
	if override.APIKey != "" {
		result.APIKey = override.APIKey
	}
	if override.BaseURL != "" {
		result.BaseURL = override.BaseURL
	}
	if override.Provider != "" {
		result.Provider = override.Provider
	}
	if override.Model != "" {
		result.Model = override.Model
	}

	// 合并 LogConfig
	if override.LogConfig != nil {
		if result.LogConfig == nil {
			result.LogConfig = &LogConfig{}
		}
		if override.LogConfig.Level != "" {
			result.LogConfig.Level = override.LogConfig.Level
		}
		if override.LogConfig.Format != "" {
			result.LogConfig.Format = override.LogConfig.Format
		}
		if override.LogConfig.Output != "" {
			result.LogConfig.Output = override.LogConfig.Output
		}
		if override.LogConfig.File != "" {
			result.LogConfig.File = override.LogConfig.File
		}
		if override.LogConfig.TimeFormat != "" {
			result.LogConfig.TimeFormat = override.LogConfig.TimeFormat
		}
	}

	return result
}

// ParseConfigArg 解析 --config 参数
func ParseConfigArg(arg string) (*LocalStoreConfig, error) {
	cfg := &LocalStoreConfig{}

	// 检查是否包含 = 符号（key=value 格式）
	if strings.Contains(arg, "=") {
		parts := strings.SplitN(arg, "=", 2)
		key := strings.TrimSpace(strings.ToLower(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch key {
		case "provider":
			cfg.Provider = model.LlmProvider(value)
		case "apikey":
			cfg.APIKey = value
		case "baseurl":
			cfg.BaseURL = value
		case "loglevel":
			if cfg.LogConfig == nil {
				cfg.LogConfig = &LogConfig{}
			}
			cfg.LogConfig.Level = value
		case "logfile":
			if cfg.LogConfig == nil {
				cfg.LogConfig = &LogConfig{}
			}
			cfg.LogConfig.File = value
		default:
			return nil, fmt.Errorf("未知的配置项: %s", key)
		}

		return cfg, nil
	}

	// 否则视为文件路径
	return LoadConfigFile(arg)
}

// LoadConfigFile 从指定路径加载配置文件
func LoadConfigFile(path string) (*LocalStoreConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("无法读取配置文件: %w", err)
	}

	var cfg LocalStoreConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("配置文件格式错误: %w", err)
	}

	return &cfg, nil
}
