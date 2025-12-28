package cmd

import (
	"msa/pkg/config"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var LoggerCmd = &cobra.Command{
	Use:   "logger",
	Short: "Manage logging configuration | 管理日志配置",
	Long: `Configure logging settings for MSA CLI | 配置MSA CLI的日志设置

Usage | 使用方法:
  msa-cli logger init                          # Initialize default logging configuration | 初始化默认日志配置
  msa-cli logger set --level=debug            # Set log level | 设置日志级别
  msa-cli logger set --format=json            # Set log format | 设置日志格式
  msa-cli logger set --output=file --file=/path/to/log  # Set file output | 设置文件输出
  msa-cli logger show                          # Show current logging configuration | 显示当前日志配置
  msa-cli logger reset                         # Reset to default configuration | 重置为默认配置`,
}

var loggerSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set logging configuration ",
	Long: `Set logging configuration parameters | 设置日志配置参数

Available parameters | 可用参数:
  --level      Log level (debug|info|warn|error|fatal|panic)
  --format     Log format (text|json)
  --output     Output target (stdout|stderr|file)
  --file       Log file path (required when output=file)
  --show-color Enable/disable colors (true|false)
  --time-format Time format string (e.g. "2006-01-02 15:04:05")`,
	Run: func(cmd *cobra.Command, args []string) {
		logConfig, err := loadConfig()
		if err != nil {
			logrus.Error("Configuration file does not exist or format error | 配置文件不存在或格式错误")
			logrus.Info("Please run 'msa-cli config init' first to initialize configuration file | 请先运行 'msa-cli config init' 初始化配置文件")
			return
		}

		// 初始化日志配置（如果不存在）
		if logConfig.LogConfig == nil {
			logConfig.LogConfig = config.DefaultConfig()
		}

		updated := false

		if cmd.Flags().Changed("level") {
			level, _ := cmd.Flags().GetString("level")
			if isValidLogLevel(level) {
				logConfig.LogConfig.Level = level
				logrus.Infof("✓ Log level set | 日志级别已设置: %s", level)
				updated = true
			} else {
				logrus.Errorf("Invalid log level | 无效的日志级别: %s", level)
				logrus.Info("Valid levels: debug, info, warn, error, fatal, panic")
				return
			}
		}

		if cmd.Flags().Changed("format") {
			format, _ := cmd.Flags().GetString("format")
			if isValidLogFormat(format) {
				logConfig.LogConfig.Format = format
				logrus.Infof("✓ Log format set | 日志格式已设置: %s", format)
				updated = true
			} else {
				logrus.Errorf("Invalid log format | 无效的日志格式: %s", format)
				logrus.Info("Valid formats: text, json")
				return
			}
		}

		if cmd.Flags().Changed("output") {
			output, _ := cmd.Flags().GetString("output")
			if isValidLogOutput(output) {
				logConfig.LogConfig.Output = output
				logrus.Infof("✓ Log output set | 日志输出已设置: %s", output)
				updated = true

				// 如果设置为文件输出，检查是否指定了文件路径
				if output == "file" && logConfig.LogConfig.File == "" {
					logrus.Warn("File output selected but no file path specified | 选择了文件输出但未指定文件路径")
					logrus.Info("Use --file parameter to specify log file path | 使用 --file 参数指定日志文件路径")
				}
			} else {
				logrus.Errorf("Invalid log output | 无效的日志输出: %s", output)
				logrus.Info("Valid outputs: stdout, stderr, file")
				return
			}
		}

		if cmd.Flags().Changed("file") {
			file, _ := cmd.Flags().GetString("file")
			logConfig.LogConfig.File = file
			logrus.Infof("✓ Log file set | 日志文件已设置: %s", file)

			// 如果设置了文件路径但输出不是file，自动设置为file
			if logConfig.LogConfig.Output != "file" {
				logConfig.LogConfig.Output = "file"
				logrus.Info("Output automatically set to 'file' | 输出已自动设置为 'file'")
			}
			updated = true
		}

		if cmd.Flags().Changed("show-color") {
			showColor, _ := cmd.Flags().GetBool("show-color")
			logConfig.LogConfig.ShowColor = showColor
			logrus.Infof("✓ Color display set | 颜色显示已设置: %t", showColor)
			updated = true
		}

		if cmd.Flags().Changed("time-format") {
			timeFormat, _ := cmd.Flags().GetString("time-format")
			logConfig.LogConfig.TimeFormat = timeFormat
			logrus.Infof("✓ Time format set | 时间格式已设置: %s", timeFormat)
			updated = true
		}

		if !updated {
			logrus.Error("Error | 错误: Please specify at least one configuration item | 请至少指定一个配置项")
			return
		}

		if err := saveConfig(logConfig); err != nil {
			logrus.Errorf("Failed to save logging configuration | 保存日志配置失败: %v", err)
			return
		}

		// 重新初始化日志系统
		config.InitLogger(logConfig.LogConfig)
		logrus.Info("✓ Logging configuration applied successfully | 日志配置应用成功")
	},
}

var loggerShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current logging configuration | 显示当前日志配置",
	Long:  `Display all current logging configuration items | 显示当前所有日志配置项`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := loadConfig()
		if err != nil {
			logrus.Error("Configuration file does not exist or format error | 配置文件不存在或格式错误")
			logrus.Info("Please run 'msa-cli config init' first to initialize configuration file | 请先运行 'msa-cli config init' 初始化配置文件")
			return
		}

		if config.LogConfig == nil {
			logrus.Info("No logging configuration found | 未找到日志配置")
			logrus.Info("Use 'msa-cli logger init' to create default logging configuration | 使用 'msa-cli logger init' 创建默认日志配置")
			return
		}

		logrus.Info("Current Logging Configuration | 当前日志配置:")
		logrus.Infof("  Level | 级别:     %s", config.LogConfig.Level)
		logrus.Infof("  Format | 格式:    %s", config.LogConfig.Format)
		logrus.Infof("  Output | 输出:    %s", config.LogConfig.Output)
		if config.LogConfig.Output == "file" {
			if config.LogConfig.File != "" {
				logrus.Infof("  File | 文件:      %s", config.LogConfig.File)
			} else {
				logrus.Info("  File | 文件:      Not specified | 未指定")
			}
		}
		logrus.Infof("  Show Color | 显示颜色: %t", config.LogConfig.ShowColor)
		logrus.Infof("  Time Format | 时间格式: %s", config.LogConfig.TimeFormat)
	},
}

var loggerResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset logging configuration to defaults | 重置日志配置为默认值",
	Long:  `Reset all logging configuration to default values | 将所有日志配置重置为默认值`,
	Run: func(cmd *cobra.Command, args []string) {
		logConfig, err := loadConfig()
		if err != nil {
			logrus.Error("Configuration file does not exist or format error | 配置文件不存在或格式错误")
			logrus.Info("Please run 'msa-cli config init' first to initialize configuration file | 请先运行 'msa-cli config init' 初始化配置文件")
			return
		}

		logConfig.LogConfig = config.DefaultConfig()

		if err := saveConfig(logConfig); err != nil {
			logrus.Errorf("Failed to reset logging configuration | 重置日志配置失败: %v", err)
			return
		}

		// 重新初始化日志系统
		config.InitLogger(logConfig.LogConfig)

		logrus.Info("✓ Logging configuration reset to defaults | 日志配置已重置为默认值")
		config.ShowCurrentConfig()
	},
}

func isValidLogLevel(level string) bool {
	validLevels := []string{"debug", "info", "warn", "error", "fatal", "panic"}
	for _, l := range validLevels {
		if strings.ToLower(level) == l {
			return true
		}
	}
	return false
}

func isValidLogFormat(format string) bool {
	validFormats := []string{"text", "json"}
	for _, f := range validFormats {
		if strings.ToLower(format) == f {
			return true
		}
	}
	return false
}

func isValidLogOutput(output string) bool {
	validOutputs := []string{"stdout", "stderr", "file"}
	for _, o := range validOutputs {
		if strings.ToLower(output) == o {
			return true
		}
	}
	return false
}

func init() {
	loggerSetCmd.Flags().StringP("level", "l", "", "Log level (debug|info|warn|error|fatal|panic)")
	loggerSetCmd.Flags().StringP("format", "f", "", "Log format (text|json)")
	loggerSetCmd.Flags().StringP("output", "o", "", "Output target (stdout|stderr|file)")
	loggerSetCmd.Flags().StringP("file", "F", "", "Log file path (when output=file)")
	loggerSetCmd.Flags().BoolP("show-color", "c", false, "Enable/disable colors (true|false)")
	loggerSetCmd.Flags().StringP("time-format", "t", "", "Time format (e.g. '2006-01-02 15:04:05')")

	LoggerCmd.AddCommand(loggerSetCmd)
	LoggerCmd.AddCommand(loggerShowCmd)
	LoggerCmd.AddCommand(loggerResetCmd)
}
