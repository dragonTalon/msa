package model

type LlmProvider string

const (
	Siliconflow LlmProvider = "siliconflow"
)

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
