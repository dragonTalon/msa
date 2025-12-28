package config

type LocalStoreConfig struct {
	APIKey    string     `json:"apiKey"`
	BaseURL   string     `json:"baseURL"`
	LogConfig *LogConfig `json:"logConfig,omitempty"`
}
