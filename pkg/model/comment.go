package model

import (
	"fmt"
	"strings"
)

type LlmProvider string

const (
	Siliconflow LlmProvider = "siliconflow"
)

// ProviderInfo Provider 元信息
type ProviderInfo struct {
	ID             LlmProvider
	DisplayName    string
	Description    string
	DefaultBaseURL string
	KeyPrefix      string
}

// ProviderRegistry Provider 注册表
var ProviderRegistry = map[LlmProvider]ProviderInfo{
	Siliconflow: {
		ID:             Siliconflow,
		DisplayName:    "SiliconFlow (硅基流动)",
		Description:    "国内 LLM API 提供商，兼容 OpenAI 格式",
		DefaultBaseURL: "https://api.siliconflow.cn/v1",
		KeyPrefix:      "sk-",
	},
}

// GetDisplayName 获取 Provider 的友好显示名称
func (p LlmProvider) GetDisplayName() string {
	if info, ok := ProviderRegistry[p]; ok {
		return info.DisplayName
	}
	return string(p)
}

// GetDefaultBaseURL 获取 Provider 的默认 Base URL
func (p LlmProvider) GetDefaultBaseURL() string {
	if info, ok := ProviderRegistry[p]; ok {
		return info.DefaultBaseURL
	}
	return ""
}

// ValidateAPIKey 验证 API Key 格式
func (p LlmProvider) ValidateAPIKey(key string) error {
	if info, ok := ProviderRegistry[p]; ok && info.KeyPrefix != "" {
		if !strings.HasPrefix(key, info.KeyPrefix) {
			return fmt.Errorf("API Key 应该以 %s 开头", info.KeyPrefix)
		}
	}
	return nil
}

type ToolGroup string

const (
	// 金融工具组
	FinanceToolGroup ToolGroup = "finance"
	// 市场工具
	MarketToolGroup ToolGroup = "market"
	// 新闻工具组
	NewsToolGroup ToolGroup = "news"
	// 股票工具组
	StockToolGroup ToolGroup = "stock"
	// 搜索工具组
	SearchToolGroup ToolGroup = "search"
	// Skill 工具组
	SkillToolGroup ToolGroup = "skill"
)

type Pair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// MessageRole 消息角色类型
type MessageRole string

// 消息角色枚举常量
const (
	RoleLogo      MessageRole = "logo"      // Logo 显示
	RoleUser      MessageRole = "user"      // 用户消息
	RoleSystem    MessageRole = "system"    // 系统消息
	RoleAssistant MessageRole = "assistant" // AI 助手消息
)

// Message 聊天消息结构
type Message struct {
	Role    MessageRole   // 消息角色
	Content string        // 消息内容
	MsgType StreamMsgType // 消息类型（text/tool/reason）
}
