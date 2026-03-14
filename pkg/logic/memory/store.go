package memory

import (
	"time"

	"msa/pkg/model"
)

// ==================== 存储接口 ====================

// Store 记忆存储接口
// 提供会话、知识、短期记忆的持久化和查询功能
type Store interface {
	// ==================== 会话操作 ====================

	// SaveSession 保存完整会话
	// 按年月组织目录（sessions/2024-03/），JSON 格式保存
	SaveSession(session *model.Session) error

	// GetSession 根据 ID 和时间获取完整会话
	GetSession(id string, startTime time.Time) (*model.Session, error)

	// ListSessions 列出会话（支持分页）
	// offset: 偏移量, limit: 数量限制
	// 返回按时间倒序排列的会话索引列表
	ListSessions(offset, limit int) ([]model.SessionIndex, error)

	// DeleteSession 删除指定会话
	DeleteSession(id string, startTime time.Time) error

	// ==================== 知识操作 ====================

	// SaveKnowledge 保存知识
	// 根据类型分目录保存，自动更新索引
	SaveKnowledge(knowledge *model.Knowledge) error

	// GetKnowledge 根据 ID 获取知识
	GetKnowledge(id string) (*model.Knowledge, error)

	// ListKnowledgeByType 按类型列出知识
	ListKnowledgeByType(knowledgeType model.KnowledgeType) ([]model.KnowledgeIndex, error)

	// GetAllKnowledge 获取所有知识（按类型分组）
	// 返回 map[KnowledgeType][]Knowledge
	GetAllKnowledge() (map[model.KnowledgeType][]*model.Knowledge, error)

	// ==================== 短期记忆操作 ====================

	// SaveShortTermMemory 保存短期记忆
	// 保存到 short-term/current.json，最多保存 10 条消息
	SaveShortTermMemory(memory *model.ShortTermMemory) error

	// GetShortTermMemory 获取短期记忆
	GetShortTermMemory() (*model.ShortTermMemory, error)

	// ==================== 搜索操作 ====================

	// Search 搜索会话和知识
	// 支持关键词搜索、股票代码搜索、日期范围过滤、类型过滤
	Search(filters model.SearchFilters) ([]model.SearchResult, error)

	// ==================== 统计操作 ====================

	// GetStats 获取统计信息
	GetStats() (*model.MemoryStats, error)

	// ==================== 索引操作 ====================

	// loadIndexes 加载索引
	// 启动时调用，失败时自动重建
	loadIndexes() error

	// saveIndexes 保存索引
	// 会话或知识更新后调用
	saveIndexes() error

	// rebuildIndexes 重建索引
	// 索引损坏或不存在时调用
	rebuildIndexes() error
}

// ==================== 错误类型 ====================

// StoreError 存储错误类型
type StoreError struct {
	Op   string // 操作名称
	Path string // 文件路径
	Err  error  // 底层错误
}

func (e *StoreError) Error() string {
	if e.Path != "" {
		return e.Op + " " + e.Path + ": " + e.Err.Error()
	}
	return e.Op + ": " + e.Err.Error()
}

func (e *StoreError) Unwrap() error {
	return e.Err
}

// NewStoreError 创建存储错误
func NewStoreError(op, path string, err error) *StoreError {
	return &StoreError{
		Op:   op,
		Path: path,
		Err:  err,
	}
}

// ==================== 常量定义 ====================

const (
	// 目录名称
	DirNameSessions  = "sessions"
	DirNameKnowledge = "knowledge"
	DirNameShortTerm = "short-term"
	DirNameIndexes   = "indexes"

	// 索引文件名
	IndexFileNameAll       = "all.json"
	IndexFileNameSessions  = "sessions.json"
	IndexFileNameKnowledge = "knowledge.json"

	// 短期记忆文件名
	ShortTermMemoryFileName = "current.json"

	// 限制
	MaxShortTermMessages = 10  // 短期记忆最多消息数
	MaxSearchResults     = 100 // 搜索结果最大数量

	// 重试配置
	MaxRetries   = 2   // 最大重试次数
	RetryDelayMs = 100 // 重试延迟（毫秒）
)

// ==================== 辅助类型 ====================

// Indexes 内存索引
type Indexes struct {
	Sessions   []model.SessionIndex   `json:"sessions"`
	Knowledge  []model.KnowledgeIndex `json:"knowledge"`
	LastUpdate time.Time              `json:"lastUpdate"`
	Version    int                    `json:"version"` // 索引版本号
}
