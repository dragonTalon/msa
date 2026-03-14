package model

import (
	"time"
)

// ==================== 消息模型 ====================

// MemoryMessageType 消息类型
type MemoryMessageType string

const (
	MemoryMessageTypeUser      MemoryMessageType = "user"      // 用户消息
	MemoryMessageTypeAssistant MemoryMessageType = "assistant" // AI 回复
	MemoryMessageTypeSystem    MemoryMessageType = "system"    // 系统消息
)

// MemoryMessage 记忆系统消息结构（与 chat.Message 区分）
type MemoryMessage struct {
	ID        string            `json:"id"`                 // 消息唯一 ID
	Type      MemoryMessageType `json:"type"`               // 消息类型
	Content   string            `json:"content"`            // 消息内容
	Timestamp time.Time         `json:"timestamp"`          // 时间戳
	ToolCall  *MemoryToolCall   `json:"toolCall,omitempty"` // 工具调用（可选）
}

// MemoryToolCall 工具调用记录
type MemoryToolCall struct {
	Name      string                 `json:"name"`   // 工具名称
	Arguments map[string]interface{} `json:"args"`   // 调用参数
	Result    string                 `json:"result"` // 调用结果
}

// ==================== 会话模型 ====================

// Session 完整会话数据
type Session struct {
	ID           string          `json:"id"`           // 会话唯一 ID (UUID)
	StartTime    time.Time       `json:"startTime"`    // 会话开始时间
	EndTime      time.Time       `json:"endTime"`      // 会话结束时间
	Messages     []MemoryMessage `json:"messages"`     // 完整消息列表
	MessageCount int             `json:"messageCount"` // 消息总数
	Context      *SessionContext `json:"context"`      // 会话上下文
	Summary      string          `json:"summary"`      // 会话摘要
	Tags         []string        `json:"tags"`         // 标签
	Model        string          `json:"model"`        // 使用的模型
}

// SessionContext 会话上下文（从对话中提取的信息）
type SessionContext struct {
	StockCodes []string `json:"stockCodes"` // 提及的股票代码
	ToolsUsed  []string `json:"toolsUsed"`  // 使用的工具
	Topics     []string `json:"topics"`     // 讨论的主题
	Actions    []string `json:"actions"`    // 执行的操作
}

// ==================== 知识模型 ====================

// KnowledgeType 知识类型
type KnowledgeType string

const (
	KnowledgeTypeUserProfile   KnowledgeType = "user-profile"   // 用户画像
	KnowledgeTypeWatchlist     KnowledgeType = "watchlist"      // 关注列表
	KnowledgeTypeConcept       KnowledgeType = "concept"        // 学过的概念
	KnowledgeTypeStrategy      KnowledgeType = "strategy"       // 交易策略
	KnowledgeTypeQAndA         KnowledgeType = "qanda"          // 常见问答
	KnowledgeTypeMarketInsight KnowledgeType = "market-insight" // 市场洞察
)

// Knowledge 知识结构（基础类型）
type Knowledge struct {
	ID         string        `json:"id"`         // 知识唯一 ID
	Type       KnowledgeType `json:"type"`       // 知识类型
	CreatedAt  time.Time     `json:"createdAt"`  // 创建时间
	UpdatedAt  time.Time     `json:"updatedAt"`  // 更新时间
	Source     string        `json:"source"`     // 来源会话 ID
	Confidence float64       `json:"confidence"` // AI 提取的置信度 (0-1)
	Content    interface{}   `json:"content"`    // 具体内容（根据类型不同）
}

// UserProfile 用户画像
type UserProfile struct {
	ID              string    `json:"id"`
	RiskPreference  string    `json:"riskPreference"`            // 风险偏好: conservative/moderate/aggressive
	InvestmentStyle []string  `json:"investmentStyle"`           // 投资风格: value/growth/momentum
	FocusAreas      []string  `json:"focusAreas"`                // 关注领域: tech/finance/energy
	ExperienceLevel string    `json:"experienceLevel"`           // 经验水平: beginner/intermediate/advanced
	ValidFrom       time.Time `json:"validFrom"`                 // 生效时间
	ValidTo         time.Time `json:"validTo"`                   // 失效时间（零值表示当前生效）
	Source          string    `json:"source"`                    // 来源: "ai-extracted" | "user-edited"
	IsMajorChange   bool      `json:"isMajorChange"`             // 是否重大变化
	PreviousProfile string    `json:"previousProfile,omitempty"` // 上一个版本 ID
}

// Watchlist 关注列表
type Watchlist struct {
	ID         string      `json:"id"`
	StockCodes []StockItem `json:"stockCodes"` // 股票列表
	Category   string      `json:"category"`   // 分类: long-term/short-term/watching
	Notes      string      `json:"notes"`      // 备注
	ValidFrom  time.Time   `json:"validFrom"`  // 生效时间
	ValidTo    time.Time   `json:"validTo"`    // 失效时间（零值表示当前生效）
}

// StockItem 股票项
type StockItem struct {
	Code    string    `json:"code"`    // 股票代码
	Name    string    `json:"name"`    // 股票名称
	AddedAt time.Time `json:"addedAt"` // 添加时间
	Reason  string    `json:"reason"`  // 添加原因
}

// ConceptLearned 学过的概念
type ConceptLearned struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`        // 概念名称
	Description string              `json:"description"` // 概念描述
	Category    string              `json:"category"`    // 分类: indicator/strategy/term
	Confidence  float64             `json:"confidence"`  // 当前理解程度 (0-1)
	LearnedAt   time.Time           `json:"learnedAt"`   // 首次学习时间
	UpdatedAt   time.Time           `json:"updatedAt"`   // 更新时间
	History     []ConfidenceHistory `json:"history"`     // 理解程度变化历史
}

// ConfidenceHistory 理解程度历史记录
type ConfidenceHistory struct {
	Confidence float64   `json:"confidence"` // 理解程度
	ChangedAt  time.Time `json:"changedAt"`  // 变化时间
	Reason     string    `json:"reason"`     // 原因: "首次询问" "成功应用" "复习"
	Source     string    `json:"source"`     // 来源: "ai-assessed" | "user-confirmed"
}

// TradingStrategy 交易策略
type TradingStrategy struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`                       // 策略名称
	Description      string    `json:"description"`                // 策略描述
	Rules            []string  `json:"rules"`                      // 策略规则
	Conditions       []string  `json:"conditions"`                 // 适用条件
	Version          int       `json:"version"`                    // 版本号
	CreatedAt        time.Time `json:"createdAt"`                  // 创建时间
	ValidFrom        time.Time `json:"validFrom"`                  // 生效时间
	ValidTo          time.Time `json:"validTo"`                    // 失效时间（零值表示当前生效）
	PreviousStrategy string    `json:"previousStrategy,omitempty"` // 上一个版本 ID
}

// QAndA 常见问答
type QAndA struct {
	ID        string    `json:"id"`
	Question  string    `json:"question"`  // 问题
	Answer    string    `json:"answer"`    // 答案
	Category  string    `json:"category"`  // 分类
	AskCount  int       `json:"askCount"`  // 询问次数
	LastAsked time.Time `json:"lastAsked"` // 最后询问时间
	CreatedAt time.Time `json:"createdAt"` // 创建时间
	UpdatedAt time.Time `json:"updatedAt"` // 更新时间
}

// MarketInsight 市场洞察
type MarketInsight struct {
	ID        string    `json:"id"`
	Topic     string    `json:"topic"`     // 主题
	Insight   string    `json:"insight"`   // 洞察内容
	Relevant  []string  `json:"relevant"`  // 相关股票/板块
	CreatedAt time.Time `json:"createdAt"` // 创建时间
	ValidFrom time.Time `json:"validFrom"` // 生效时间
	ValidTo   time.Time `json:"validTo"`   // 失效时间（零值表示当前有效）
}

// ==================== 短期记忆 ====================

// ShortTermMemory 短期记忆（最多 10 条消息）
type ShortTermMemory struct {
	SessionID string          `json:"sessionId"` // 所属会话 ID
	Messages  []MemoryMessage `json:"messages"`  // 最近的消息（最多 10 条）
	UpdatedAt time.Time       `json:"updatedAt"` // 更新时间
}

// ==================== 索引模型 ====================

// SessionIndex 会话索引（用于快速查询）
type SessionIndex struct {
	ID           string    `json:"id"`           // 会话 ID
	StartTime    time.Time `json:"startTime"`    // 开始时间
	EndTime      time.Time `json:"endTime"`      // 结束时间
	Summary      string    `json:"summary"`      // 会话摘要
	MessageCount int       `json:"messageCount"` // 消息数量
	Tags         []string  `json:"tags"`         // 标签
	StockCodes   []string  `json:"stockCodes"`   // 股票代码
}

// KnowledgeIndex 知识索引（用于快速查询）
type KnowledgeIndex struct {
	ID        string        `json:"id"`        // 知识 ID
	Type      KnowledgeType `json:"type"`      // 知识类型
	CreatedAt time.Time     `json:"createdAt"` // 创建时间
	UpdatedAt time.Time     `json:"updatedAt"` // 更新时间
	Summary   string        `json:"summary"`   // 摘要（用于搜索）
}

// ==================== 统计模型 ====================

// MemoryStats 记忆系统统计信息
type MemoryStats struct {
	// 会话统计
	TotalSessions   int       `json:"totalSessions"`   // 总会话数
	EarliestSession time.Time `json:"earliestSession"` // 最早会话
	LatestSession   time.Time `json:"latestSession"`   // 最新会话
	TotalMessages   int       `json:"totalMessages"`   // 总消息数

	// 知识统计
	TotalKnowledge  int                   `json:"totalKnowledge"`  // 总知识数
	KnowledgeByType map[KnowledgeType]int `json:"knowledgeByType"` // 按类型统计

	// 股票统计
	TopStocks []StockStat `json:"topStocks"` // 最常查询的股票 TOP 5

	// 概念统计
	TopConcepts []ConceptStat `json:"topConcepts"` // 学得最多的概念 TOP 5

	// 系统信息
	IndexSize   int64     `json:"indexSize"`   // 索引大小（字节）
	CacheSize   int64     `json:"cacheSize"`   // 缓存大小（字节）
	LastUpdated time.Time `json:"lastUpdated"` // 最后更新时间
}

// StockStat 股票统计
type StockStat struct {
	Code       string `json:"code"`       // 股票代码
	Name       string `json:"name"`       // 股票名称
	QueryCount int    `json:"queryCount"` // 查询次数
}

// ConceptStat 概念统计
type ConceptStat struct {
	Name       string `json:"name"`       // 概念名称
	Category   string `json:"category"`   // 分类
	LearnCount int    `json:"learnCount"` // 学习次数
}

// ==================== 搜索模型 ====================

// SearchFilters 搜索过滤器
type SearchFilters struct {
	Query      string          `json:"query"`      // 搜索关键词
	Types      []KnowledgeType `json:"types"`      // 知识类型过滤（空表示全部）
	StartDate  *time.Time      `json:"startDate"`  // 开始日期（可选）
	EndDate    *time.Time      `json:"endDate"`    // 结束日期（可选）
	StockCodes []string        `json:"stockCodes"` // 股票代码过滤
	Limit      int             `json:"limit"`      // 结果数量限制
	Offset     int             `json:"offset"`     // 分页偏移
}

// SearchResult 搜索结果
type SearchResult struct {
	Type      SearchResultType `json:"type"`      // 结果类型
	ID        string           `json:"id"`        // ID
	Title     string           `json:"title"`     // 标题
	Summary   string           `json:"summary"`   // 摘要
	Score     float64          `json:"score"`     // 相关性评分
	Timestamp time.Time        `json:"timestamp"` // 时间戳
	Highlight string           `json:"highlight"` // 匹配片段（高亮显示）
}

// SearchResultType 搜索结果类型
type SearchResultType string

const (
	SearchResultTypeSession   SearchResultType = "session"   // 会话
	SearchResultTypeKnowledge SearchResultType = "knowledge" // 知识
)
