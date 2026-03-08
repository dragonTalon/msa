package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"msa/pkg/model"
)

// ValidationSeverity 验证严重程度
type ValidationSeverity int

const (
	SeverityError ValidationSeverity = iota
	SeverityWarning
)

// ValidationError 验证错误
type ValidationError struct {
	Field    string
	Message  string
	Severity ValidationSeverity
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", severityToString(e.Severity), e.Field, e.Message)
}

func severityToString(s ValidationSeverity) string {
	switch s {
	case SeverityError:
		return "错误"
	case SeverityWarning:
		return "警告"
	default:
		return "未知"
	}
}

// ValidateAPIKey 验证 API Key
func ValidateAPIKey(key, provider string) []*ValidationError {
	var errors []*ValidationError

	// 非空验证
	if strings.TrimSpace(key) == "" {
		errors = append(errors, &ValidationError{
			Field:    "API Key",
			Message:  "API Key 不能为空",
			Severity: SeverityError,
		})
		return errors
	}

	// 前缀验证
	p := model.LlmProvider(provider)
	if err := p.ValidateAPIKey(key); err != nil {
		errors = append(errors, &ValidationError{
			Field:    "API Key",
			Message:  err.Error(),
			Severity: SeverityError,
		})
	}

	// 长度验证（警告）
	if len(key) < 10 {
		errors = append(errors, &ValidationError{
			Field:    "API Key",
			Message:  "API Key 长度可能不正确（少于 10 个字符）",
			Severity: SeverityWarning,
		})
	}

	return errors
}

// ValidateBaseURL 验证 Base URL
func ValidateBaseURL(urlStr string) []*ValidationError {
	var errors []*ValidationError

	// 空字符串是有效的（使用默认值）
	if strings.TrimSpace(urlStr) == "" {
		return nil
	}

	// URL 格式验证
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		errors = append(errors, &ValidationError{
			Field:    "Base URL",
			Message:  fmt.Sprintf("无效的 URL 格式: %v", err),
			Severity: SeverityError,
		})
		return errors
	}

	// 协议验证
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		errors = append(errors, &ValidationError{
			Field:    "Base URL",
			Message:  "URL 必须以 http:// 或 https:// 开头",
			Severity: SeverityError,
		})
	}

	return errors
}

// ValidateLogLevel 验证日志级别
func ValidateLogLevel(level string) []*ValidationError {
	var errors []*ValidationError

	validLevels := map[string]bool{
		"debug":   true,
		"info":    true,
		"warn":    true,
		"warning": true,
		"error":   true,
		"fatal":   true,
		"panic":   true,
	}

	normalizedLevel := strings.ToLower(strings.TrimSpace(level))
	if !validLevels[normalizedLevel] {
		errors = append(errors, &ValidationError{
			Field:    "日志级别",
			Message:  fmt.Sprintf("无效的日志级别，支持：debug, info, warn, error, fatal, panic（当前值: %s）", level),
			Severity: SeverityError,
		})
	}

	return errors
}

// ValidateLogPath 验证日志文件路径
func ValidateLogPath(path string) []*ValidationError {
	var errors []*ValidationError

	// 空字符串是有效的
	if strings.TrimSpace(path) == "" {
		return nil
	}

	// 展开路径
	expandedPath := path
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			errors = append(errors, &ValidationError{
				Field:    "日志路径",
				Message:  fmt.Sprintf("无法展开 ~ 路径: %v", err),
				Severity: SeverityError,
			})
			return errors
		}
		expandedPath = filepath.Join(homeDir, path[1:])
	}

	// 获取目录
	dir := filepath.Dir(expandedPath)

	// 检查目录是否存在
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// 尝试创建目录
		if err := os.MkdirAll(dir, 0755); err != nil {
			errors = append(errors, &ValidationError{
				Field:    "日志路径",
				Message:  fmt.Sprintf("无法创建目录: %v", err),
				Severity: SeverityError,
			})
			return errors
		}
	}

	// 检查可写性
	testFile := filepath.Join(dir, ".msa_write_test")
	f, err := os.Create(testFile)
	if err != nil {
		errors = append(errors, &ValidationError{
			Field:    "日志路径",
			Message:  "无法写入日志文件，请检查路径权限",
			Severity: SeverityError,
		})
	} else {
		f.Close()
		os.Remove(testFile)
	}

	return errors
}

// ValidateProvider 验证 Provider
func ValidateProvider(provider string) []*ValidationError {
	var errors []*ValidationError

	p := model.LlmProvider(provider)

	// 检查是否在注册表中
	if _, ok := model.ProviderRegistry[p]; !ok {
		errors = append(errors, &ValidationError{
			Field:    "Provider",
			Message:  fmt.Sprintf("未知的 Provider: %s", provider),
			Severity: SeverityError,
		})
	}

	return errors
}

// ValidateConfig 验证完整配置
func ValidateConfig(cfg *LocalStoreConfig) []*ValidationError {
	var allErrors []*ValidationError

	// 验证 Provider
	if cfg.Provider != "" {
		allErrors = append(allErrors, ValidateProvider(string(cfg.Provider))...)
	}

	// 验证 API Key
	allErrors = append(allErrors, ValidateAPIKey(cfg.APIKey, string(cfg.Provider))...)

	// 验证 Base URL
	allErrors = append(allErrors, ValidateBaseURL(cfg.BaseURL)...)

	// 验证日志配置
	if cfg.LogConfig != nil {
		allErrors = append(allErrors, ValidateLogLevel(cfg.LogConfig.Level)...)
		allErrors = append(allErrors, ValidateLogPath(cfg.LogConfig.File)...)
	}

	// 按严重程度排序（错误在前，警告在后）
	sortErrors(allErrors)

	return allErrors
}

// sortErrors 按严重程度排序错误
func sortErrors(errors []*ValidationError) {
	// 使用简单的冒泡排序
	n := len(errors)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if errors[j].Severity > errors[j+1].Severity {
				errors[j], errors[j+1] = errors[j+1], errors[j]
			}
		}
	}
}
