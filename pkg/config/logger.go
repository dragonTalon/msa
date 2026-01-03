package config

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// LogConfig 日志配置结构
type LogConfig struct {
	Level      string `json:"level"`      // 日志级别: debug, info, warn, error, fatal, panic
	Format     string `json:"format"`     // 日志格式: json, text
	Output     string `json:"output"`     // 输出目标: stdout, stderr, file
	File       string `json:"file"`       // 日志文件路径 (当output为file时)
	ShowColor  bool   `json:"showColor"`  // 是否显示颜色
	TimeFormat string `json:"timeFormat"` // 时间格式
}

// DefaultConfig 返回默认配置
func DefaultConfig() *LogConfig {
	return &LogConfig{
		Level:      "info",
		Format:     "text",
		Output:     "file",
		File:       "",
		ShowColor:  true,
		TimeFormat: "2006-01-02 15:04:05",
	}
}

// InitLogger 初始化日志系统
func InitLogger(config *LogConfig) {
	if config == nil {
		config = DefaultConfig()
	}

	// 设置日志级别
	level, err := logrus.ParseLevel(strings.ToLower(config.Level))
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	// 设置日志格式
	switch strings.ToLower(config.Format) {
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: config.TimeFormat,
		})
	case "text":
		fallthrough
	default:
		logrus.SetFormatter(&logrus.TextFormatter{
			ForceColors:      config.ShowColor,
			DisableTimestamp: false,
			FullTimestamp:    true,
			TimestampFormat:  config.TimeFormat,
		})
	}

	// 设置输出目标
	switch strings.ToLower(config.Output) {
	case "stderr":
		logrus.SetOutput(os.Stderr)
	case "file":
		if config.File != "" {
			file, err := os.OpenFile(config.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err == nil {
				logrus.SetOutput(file)
			} else {
				logrus.SetOutput(os.Stdout)
				logrus.Warnf("无法打开日志文件 %s，使用标准输出", config.File)
			}
		} else {
			logrus.SetOutput(os.Stdout)
		}
	case "stdout":
		fallthrough
	default:
		logrus.SetOutput(os.Stdout)
	}
}

// ShowLogConfig 显示指定的日志配置
func ShowLogConfig(cfg *LogConfig) {
	if cfg == nil {
		logrus.Info("No logging configuration found | 未找到日志配置")
		return
	}

	logrus.Info("Current Logging Configuration | 当前日志配置:")
	logrus.Infof("  Level | 级别:     %s", cfg.Level)
	logrus.Infof("  Format | 格式:    %s", cfg.Format)
	logrus.Infof("  Output | 输出:    %s", cfg.Output)
	if cfg.Output == "file" {
		if cfg.File != "" {
			logrus.Infof("  File | 文件:      %s", cfg.File)
		} else {
			logrus.Info("  File | 文件:      Not specified | 未指定")
		}
	}
	logrus.Infof("  Show Color | 显示颜色: %t", cfg.ShowColor)
	logrus.Infof("  Time Format | 时间格式: %s", cfg.TimeFormat)
}

// ShowCurrentConfig 显示当前日志配置
func ShowCurrentConfig() {
	config := GetConfig()
	ShowLogConfig(config)
}

// GetConfig 获取当前日志配置
func GetConfig() *LogConfig {
	formatter := logrus.StandardLogger().Formatter
	level := logrus.StandardLogger().GetLevel()

	config := DefaultConfig()
	config.Level = level.String()

	switch f := formatter.(type) {
	case *logrus.JSONFormatter:
		config.Format = "json"
		if f.TimestampFormat != "" {
			config.TimeFormat = f.TimestampFormat
		}
	case *logrus.TextFormatter:
		config.Format = "text"
		config.ShowColor = f.ForceColors
		config.TimeFormat = f.TimestampFormat
	}

	return config
}
