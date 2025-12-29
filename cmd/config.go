/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"msa/pkg/config"
	"msa/pkg/utils"
)

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage MSA CLI Configuration",
	Long: `Configure API keys and base URLs | 配置API密钥和基础URL

Usage | 使用方法:
  msa-cli config init                           # Initialize empty configuration file | 初始化空配置文件
  msa-cli config init <apiKey> <baseURL>        # Initialize and directly set API key and base URL | 初始化并直接设置API密钥和基础URL
  msa-cli config set --apiKey=xxx               # Set API key | 设置API密钥
  msa-cli config set --baseURL=xxx              # Set base URL | 设置基础URL
  msa-cli config set --apiKey=xxx --baseURL=xxx # Set both API key and base URL | 同时设置API密钥和基础URL
  msa-cli config list all                       # View all configurations | 查看所有配置
  msa-cli config path                           # Show configuration file path | 显示配置文件路径`,
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set Configuration Items ",
	Long: `Set API key and base URL configuration items ( 设置API密钥和基础URL配置项)

Use --flag=value format to set configuration items ( 使用 --flag=value 的方式设置配置项)

Examples | 示例:
  msa-cli config set --apiKey=your-api-key
  msa-cli config set --baseURL=https://api.example.com
  msa-cli config set --apiKey=xxx --baseURL=xxx`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := loadConfig()
		if err != nil {
			cfg = &config.LocalStoreConfig{}
		}

		updated := false

		if cmd.Flags().Changed("apiKey") {
			apiKey, _ := cmd.Flags().GetString("apiKey")
			cfg.APIKey = apiKey
			logrus.Infof("✓ API key set | API密钥已设置: %s", maskAPIKey(apiKey))
			updated = true
		}

		if cmd.Flags().Changed("baseURL") {
			baseURL, _ := cmd.Flags().GetString("baseURL")
			cfg.BaseURL = baseURL
			logrus.Infof("✓ Base URL set | 基础URL已设置: %s", baseURL)
			updated = true
		}

		if !updated {
			logrus.Error("Error | 错误: Please specify at least one configuration item | 请至少指定一个配置项")
			logrus.Info("Use --apiKey and/or --baseURL parameters | 使用 --apiKey 和/或 --baseURL 参数")
			return
		}

		if err := saveConfig(cfg); err != nil {
			logrus.Errorf("Failed to save configuration | 保存配置失败: %v", err)
			return
		}
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List All Configurations | 列出所有配置",
	Long:  `Display all current configuration items | 显示当前所有的配置项`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := loadConfig()
		if err != nil {
			logrus.Error("Configuration file does not exist or format error | 配置文件不存在或格式错误")
			logrus.Info("Please run 'msa-cli config init' first to initialize configuration file | 请先运行 'msa-cli config init' 初始化配置文件")
			return
		}

		logrus.Info("Current Configuration | 当前配置:")
		if cfg.APIKey == "" {
			logrus.Info("  apiKey: Not set | 未设置")
		} else {
			logrus.Infof("  apiKey: %s", maskAPIKey(cfg.APIKey))
		}
		if cfg.BaseURL == "" {
			logrus.Info("  baseURL: Not set | 未设置")
		} else {
			logrus.Infof("  baseURL: %s", cfg.BaseURL)
		}
	},
}

func loadConfig() (*config.LocalStoreConfig, error) {
	configPath, err := utils.GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("配置文件不存在")
		}
		return nil, err
	}

	var cfg config.LocalStoreConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("配置文件格式错误: %v", err)
	}

	return &cfg, nil
}

func saveConfig(cfg *config.LocalStoreConfig) error {
	configPath, err := utils.GetConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "***"
	}
	return apiKey[:4] + "***" + apiKey[len(apiKey)-4:]
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show Configuration File Path | 显示配置文件路径",
	Long:  `Display the full path of the current configuration file | 显示当前配置文件的完整路径`,
	Run: func(cmd *cobra.Command, args []string) {
		configPath, err := utils.GetConfigPath()
		if err != nil {
			logrus.Errorf("Failed to get configuration path | 获取配置路径失败: %v", err)
			return
		}
		logrus.Infof("Configuration file path | 配置文件路径: %s", configPath)
	},
}

func init() {
	configSetCmd.Flags().StringP("apiKey", "k", "", "Set API key | 设置API密钥")
	configSetCmd.Flags().StringP("baseURL", "u", "", "Set base URL | 设置基础URL")

	ConfigCmd.AddCommand(configSetCmd)

	ConfigCmd.AddCommand(configListCmd)
	ConfigCmd.AddCommand(configPathCmd)
}
