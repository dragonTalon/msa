package memory

import (
	"testing"

	"msa/pkg/model"
)

// TestFilterMessage 测试敏感信息过滤
func TestFilterMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "API Key with sk- prefix",
			input:    "我的 API Key 是 sk-abc123def4567890abcdef123456789",
			expected: "我的 API Key 是 ***REDACTED***",
		},
		{
			name:     "Bearer token",
			input:    "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expected: "Authorization: Bearer ***REDACTED***",
		},
		{
			name:     "password with colon",
			input:    "密码是: mypassword123",
			expected: "密码是: ***REDACTED***",
		},
		{
			name:     "passwd with colon",
			input:    "passwd: secret123",
			expected: "passwd: ***REDACTED***",
		},
		{
			name:     "normal message without sensitive info",
			input:    "今天腾讯股价不错",
			expected: "今天腾讯股价不错",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterMessage(tt.input)
			if result != tt.expected {
				t.Errorf("FilterMessage() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestFilterMessageInContent 测试在消息内容中的过滤
func TestFilterMessageInContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "message with API Key",
			input:    "请使用 sk-test123 这个 key 来访问",
			expected: "请使用 ***REDACTED*** 这个 key 来访问",
		},
		{
			name:     "message with multiple sensitive info",
			input:    "API Key: sk-abc123，密码: pass456",
			expected: "API Key: ***REDACTED***，密码: ***REDACTED***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterMessage(tt.input)
			if result != tt.expected {
				t.Errorf("FilterMessage() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestIsMemoryEnabled 测试记忆系统启用状态
func TestIsMemoryEnabled(t *testing.T) {
	// 默认应该启用
	if !IsMemoryEnabled() {
		t.Error("IsMemoryEnabled() should return true by default")
	}
}

// TestMemoryMessageTypeString 测试消息类型字符串
func TestMemoryMessageTypeString(t *testing.T) {
	tests := []struct {
		input    model.MemoryMessageType
		expected string
	}{
		{model.MemoryMessageTypeUser, "user"},
		{model.MemoryMessageTypeAssistant, "assistant"},
		{model.MemoryMessageTypeSystem, "system"},
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			if string(tt.input) != tt.expected {
				t.Errorf("MemoryMessageType.String() = %v, want %v", tt.input, tt.expected)
			}
		})
	}
}
