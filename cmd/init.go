package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"msa/pkg/config"
	"msa/pkg/utils"
	"os"
)

var InitCmd = &cobra.Command{
	Use:   "init [apiKey] [baseURL]",
	Short: "Initialize Configuration File ",
	Long: `Create configuration file ~/.msa/config.json in user home directory | 在用户主目录创建配置文件 ~/.msa/config.json

Parameters | 参数:
  apiKey   Optional, API key | 可选，API密钥
  baseURL  Optional, base URL | 可选，基础URL

If no parameters are provided, an empty configuration file will be created | 如果不提供参数，将创建空的配置文件`,
	Args: cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		logConfig := config.DefaultConfig()
		config.InitLogger(logConfig)

		logrus.Info("MSA CLI starting up...")

		configPath, err := utils.GetConfigPath()
		if err != nil {
			logrus.Errorf("Failed to get configuration path | 获取配置路径失败: %v", err)
			return
		}

		// Check if configuration file already exists | 检查配置文件是否已存在
		if _, err := os.Stat(configPath); err == nil {
			logrus.Warnf("Configuration file already exists | 配置文件已存在: %s", configPath)
			logrus.Info("To reconfigure, please delete the existing configuration file or use 'config set' command to modify | 如需重新配置，请删除现有配置文件或使用 'config set' 命令修改")
			return
		}

		// Create configuration | 创建配置
		config := &config.LocalStoreConfig{
			APIKey:  "",
			BaseURL: "",
		}

		// Set configuration based on parameters | 根据参数设置配置
		if len(args) >= 1 {
			config.APIKey = args[0]
		}
		if len(args) >= 2 {
			config.BaseURL = args[1]
		}

		if err := saveConfig(config); err != nil {
			logrus.Errorf("Failed to create configuration file | 创建配置文件失败: %v", err)
			return
		}

		logrus.Infof("✓ Configuration file created | 配置文件已创建: %s", configPath)
		if config.APIKey != "" {
			logrus.Infof("✓ API key set | API密钥已设置: %s", maskAPIKey(config.APIKey))
		} else {
			logrus.Info("• API key not set, you can use 'msa-cli config set apiKey <your-api-key>' to set | API密钥未设置，可使用 'msa-cli config set apiKey <your-api-key>' 设置")
		}
		if config.BaseURL != "" {
			logrus.Infof("✓ Base URL set | 基础URL已设置: %s", config.BaseURL)
		} else {
			logrus.Info("• Base URL not set, you can use 'msa-cli config set baseURL <your-base-url>' to set | 基础URL未设置，可使用 'msa-cli config set baseURL <your-base-url>' 设置")
		}
	},
}

func init() {
	InitCmd.Flags().StringP("apiKey", "k", "", "API key")
	InitCmd.Flags().StringP("baseURL", "u", "", "Base URL")

}
