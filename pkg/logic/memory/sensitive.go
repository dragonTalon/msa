package memory

import (
	"regexp"
	"strings"
)

// ==================== 敏感信息过滤 ====================

// SensitivePattern 敏感信息模式
type SensitivePattern struct {
	Pattern string // 正则表达式
	Context string // 描述
}

// 敏感信息检测模式
// 使用捕获组：第一组是前缀（保留），第二组是敏感值（替换）
var sensitivePatterns = []SensitivePattern{
	// API Key 模式
	{`(sk-[a-zA-Z0-9]{5,})`, "OpenAI/SiliconFlow API Key"},
	{`(Bearer)\s+(\S+)`, "Bearer Token"},
	{`(api[_-]?key["']?\s*[:=]\s*)["']?([a-zA-Z0-9\-._~+/]{20,})["']?`, "API Key 配置"},
	{`(apikey["']?\s*[:=]\s*)["']?([a-zA-Z0-9\-._~+/]{20,})["']?`, "ApiKey 配置"},
	{`(access[_-]?token["']?\s*[:=]\s*)["']?([a-zA-Z0-9\-._~+/]{20,})["']?`, "Access Token"},

	// 密码模式 - 支持中文和英文
	{`(password["']?\s*[:：]\s*)(\S+)`, "密码"},
	{`(passwd["']?\s*[:：]\s*)(\S+)`, "密码"},
	{`(pwd["']?\s*[:：]\s*)(\S+)`, "密码"},
	// 中文密码关键词（支持"密码是:"格式）
	// 使用非捕获组 (?:...) 来避免增加捕获组数量
	{`((?:密码|密匙)(?:是)?\s*[:：]\s*)(\S+)`, "密码"},

	// Token 模式
	{`(token["']?\s*[:=]\s*)["']?([a-zA-Z0-9\-._~+/]{20,})["']?`, "Token"},
	{`(secret["']?\s*[:=]\s*)["']?([a-zA-Z0-9\-._~+/]{20,})["']?`, "Secret"},
	{`(private[_-]?key["']?\s*[:=]\s*)["']?([a-zA-Z0-9\-._~+/]{20,})["']?`, "Private Key"},
}

// 编译后的正则表达式缓存
var compiledPatterns = func() []*regexp.Regexp {
	patterns := make([]*regexp.Regexp, len(sensitivePatterns))
	for i, p := range sensitivePatterns {
		patterns[i] = regexp.MustCompile(`(?i)` + p.Pattern) // (?i) 表示不区分大小写
	}
	return patterns
}()

// RedactSensitive 脱敏处理
// 检测并替换敏感信息为 ***REDACTED***
// 使用捕获组保留上下文信息（如 "Bearer "、"password:" 等）
func RedactSensitive(content string) string {
	result := content

	for i, pattern := range compiledPatterns {
		// 使用 ReplaceAllStringFunc 进行更精细的替换控制
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			// 记录检测到的敏感信息类型
			logDetection(sensitivePatterns[i].Context)

			// 检查是否有子组（捕获组）
			submatches := pattern.FindStringSubmatch(match)
			if len(submatches) > 2 {
				// 有多个捕获组：保留第一个组（前缀），替换第二个组（敏感值）
				// 例如："(Bearer)\s+(xxx)" -> "Bearer ***REDACTED***"
				prefix := submatches[1]
				// 检查原匹配在前缀后是否有空格
				hasSpace := len(match) > len(prefix) && match[len(prefix)] == ' '
				if hasSpace {
					return prefix + " ***REDACTED***"
				}
				return prefix + "***REDACTED***"
			}
			// 没有子组，直接替换整个匹配
			// 例如："sk-xxx" -> "***REDACTED***"
			return "***REDACTED***"
		})
	}

	return result
}

// logDetection 记录敏感信息检测（可选）
func logDetection(context string) {
	// 这里可以记录日志，但要注意不要记录敏感信息本身
	// 为了避免日志过多，可以只在调试模式下记录
	// log.Debugf("检测到敏感信息: %s", context)
}

// HasSensitiveInfo 检查内容是否包含敏感信息
func HasSensitiveInfo(content string) bool {
	for _, pattern := range compiledPatterns {
		if pattern.MatchString(content) {
			return true
		}
	}
	return false
}

// ==================== 消息过滤 ====================

// FilterMessage 过滤消息内容（用于保存前）
func FilterMessage(content string) string {
	// 1. 脱敏敏感信息
	content = RedactSensitive(content)

	// 2. 可以添加其他过滤规则
	// 例如：过滤系统命令的输出

	return content
}

// ==================== 系统命令过滤 ====================

// IsSensitiveCommand 检查是否是敏感命令
// 敏感命令的输出不应该被记录
func IsSensitiveCommand(command string) bool {
	sensitiveCommands := []string{
		"/config",
		"/set",
		"/env",
		"/password",
		"/token",
		"/apikey",
		"/api",
	}

	command = strings.TrimSpace(strings.ToLower(command))
	for _, sensitive := range sensitiveCommands {
		if strings.HasPrefix(command, sensitive) {
			return true
		}
	}

	return false
}

// ShouldRecordCommand 判断命令输出是否应该被记录
func ShouldRecordCommand(command string) bool {
	// 如果是敏感命令，不记录
	if IsSensitiveCommand(command) {
		return false
	}
	return true
}

// ==================== 批量处理 ====================

// RedactMessages 批量脱敏消息
func RedactMessages(messages []string) []string {
	result := make([]string, len(messages))
	for i, msg := range messages {
		result[i] = RedactSensitive(msg)
	}
	return result
}
