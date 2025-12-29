package config

// LocalStoreConfig 本地存储配置结构
type LocalStoreConfig struct {
	APIKey    string     `json:"apiKey"`
	BaseURL   string     `json:"baseURL"`
	LogConfig *LogConfig `json:"logConfig,omitempty"`
}
