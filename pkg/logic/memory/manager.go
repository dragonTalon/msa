package memory

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"msa/pkg/model"
)

// ==================== Manager 结构体 ====================

// Manager 记忆管理器（单例）
type Manager struct {
	store          Store                // 存储实例
	knowledgeCache *AggregatedKnowledge // 知识缓存
	currentSession *SessionManager      // 当前会话
	lastSessionID  string               // 最后结束的会话ID（用于恢复功能）
	mu             sync.RWMutex         // 读写锁
	initialized    bool                 // 是否已初始化
}

// AggregatedKnowledge 聚合知识
type AggregatedKnowledge struct {
	UserProfiles   []*model.UserProfile     // 用户画像（取最新的）
	Watchlist      []*model.Watchlist       // 关注列表（去重合并）
	Concepts       []*model.ConceptLearned  // 学过的概念
	Strategies     []*model.TradingStrategy // 交易策略
	QAndAs         []*model.QAndA           // 常见问答
	MarketInsights []*model.MarketInsight   // 市场洞察
	LastUpdated    time.Time                // 最后更新时间
}

// SessionManager 会话管理器
type SessionManager struct {
	ID        string                 // 会话 ID (UUID)
	StartTime time.Time              // 开始时间
	Messages  []model.MemoryMessage  // 消息列表
	Context   *model.SessionContext  // 会话上下文
	ShortTerm *model.ShortTermMemory // 短期记忆
}

// ==================== 单例模式 ====================

var (
	globalManager *Manager
	managerOnce   sync.Once
)

// GetManager 获取全局唯一的管理器实例
func GetManager() *Manager {
	managerOnce.Do(func() {
		store, err := NewFileStore()
		if err != nil {
			log.Errorf("创建存储失败: %v", err)
			globalManager = &Manager{}
			return
		}

		globalManager = &Manager{
			store:          store,
			knowledgeCache: &AggregatedKnowledge{},
			currentSession: nil,
			initialized:    false,
		}

		// 启动时加载知识
		globalManager.LoadKnowledge()

		log.Info("记忆管理器已初始化")
	})

	return globalManager
}

// ==================== 初始化方法 ====================

// Initialize 初始化新会话
func (m *Manager) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 生成 UUID 作为会话 ID
	sessionID := uuid.New().String()

	// 创建会话管理器
	m.currentSession = &SessionManager{
		ID:        sessionID,
		StartTime: time.Now(),
		Messages:  []model.MemoryMessage{},
		Context: &model.SessionContext{
			StockCodes: []string{},
			ToolsUsed:  []string{},
			Topics:     []string{},
			Actions:    []string{},
		},
		ShortTerm: &model.ShortTermMemory{
			SessionID: sessionID,
			Messages:  []model.MemoryMessage{},
			UpdatedAt: time.Now(),
		},
	}

	m.initialized = true

	log.Infof("新会话已初始化: %s", sessionID)
	return nil
}

// IsInitialized 检查是否已初始化
func (m *Manager) IsInitialized() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.initialized
}

// GetSessionID 获取当前会话 ID
func (m *Manager) GetSessionID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.currentSession != nil {
		return m.currentSession.ID
	}
	return ""
}

// ==================== 消息记录 ====================

// AddMessage 添加消息到当前会话
func (m *Manager) AddMessage(msg model.MemoryMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized || m.currentSession == nil {
		return fmt.Errorf("会话未初始化")
	}

	// 敏感信息过滤
	msg.Content = FilterMessage(msg.Content)

	// 添加到消息列表
	m.currentSession.Messages = append(m.currentSession.Messages, msg)

	// 更新会话上下文
	m.updateContext(msg)

	// 更新短期记忆（保持最近 10 条）
	m.updateShortTermMemory(msg)

	// 持久化短期记忆
	if err := m.store.SaveShortTermMemory(m.currentSession.ShortTerm); err != nil {
		log.Warnf("保存短期记忆失败: %v", err)
		// 不影响聊天继续进行
	}

	return nil
}

// updateContext 更新会话上下文
func (m *Manager) updateContext(msg model.MemoryMessage) {
	content := msg.Content

	// TODO: 提取股票代码
	// TODO: 检测使用的工具
	// TODO: 检测交易操作
	// TODO: 生成标签候选

	// 这里简化处理，实际应该使用正则表达式或 AI 提取
	_ = content
}

// updateShortTermMemory 更新短期记忆
func (m *Manager) updateShortTermMemory(msg model.MemoryMessage) {
	// 添加到短期记忆
	m.currentSession.ShortTerm.Messages = append(m.currentSession.ShortTerm.Messages, msg)

	// 保持最近 10 条
	if len(m.currentSession.ShortTerm.Messages) > MaxShortTermMessages {
		// 只保留最新的 10 条
		start := len(m.currentSession.ShortTerm.Messages) - MaxShortTermMessages
		m.currentSession.ShortTerm.Messages = m.currentSession.ShortTerm.Messages[start:]
	}

	m.currentSession.ShortTerm.UpdatedAt = time.Now()
}

// ==================== 会话结束处理 ====================

// EndSession 结束当前会话
func (m *Manager) EndSession() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized || m.currentSession == nil {
		return fmt.Errorf("会话未初始化")
	}

	// 构建完整会话
	session := &model.Session{
		ID:           m.currentSession.ID,
		StartTime:    m.currentSession.StartTime,
		EndTime:      time.Now(),
		Messages:     m.currentSession.Messages,
		MessageCount: len(m.currentSession.Messages),
		Context:      m.currentSession.Context,
		Summary:      m.generateSummary(),
		Tags:         m.generateTags(),
		Model:        "", // TODO: 从配置获取
	}

	// 同步保存完整会话
	if err := m.store.SaveSession(session); err != nil {
		log.Errorf("保存会话失败: %v", err)
		return err
	}

	// 保存会话ID用于恢复功能
	m.lastSessionID = session.ID

	// 异步触发知识提取（不阻塞）
	go m.extractKnowledgeAsync(session)

	// 清空短期记忆文件
	m.currentSession.ShortTerm = &model.ShortTermMemory{
		Messages:  []model.MemoryMessage{},
		UpdatedAt: time.Now(),
	}
	_ = m.store.SaveShortTermMemory(m.currentSession.ShortTerm)

	// 清空当前会话
	m.currentSession = nil
	m.initialized = false

	log.Infof("会话已结束: %s", session.ID)
	return nil
}

// GetLastSessionID 获取最后结束的会话ID（用于会话恢复）
func (m *Manager) GetLastSessionID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastSessionID
}

// LoadSession 加载指定会话的历史消息（用于会话恢复）
func (m *Manager) LoadSession(sessionID string) (*model.Session, error) {
	// 从存储中获取会话
	session, err := m.store.GetSession(sessionID, time.Time{}) // time.Time{} 表示不指定时间，让存储自动查找
	if err != nil {
		return nil, fmt.Errorf("加载会话失败: %w", err)
	}

	// 初始化新会话，但使用历史会话的ID
	m.mu.Lock()
	defer m.mu.Unlock()

	m.currentSession = &SessionManager{
		ID:        session.ID,
		StartTime: time.Now(),                                           // 恢复时创建新会话开始时间
		Messages:  append([]model.MemoryMessage{}, session.Messages...), // 复制历史消息
		Context:   session.Context,
		ShortTerm: &model.ShortTermMemory{},
	}
	m.initialized = true

	log.Infof("会话已恢复: %s (%d 条历史消息)", session.ID, len(session.Messages))
	return session, nil
}

// generateSummary 生成会话摘要
func (m *Manager) generateSummary() string {
	// 取第一条用户消息作为摘要
	for _, msg := range m.currentSession.Messages {
		if msg.Type == model.MemoryMessageTypeUser {
			if len(msg.Content) > 50 {
				return msg.Content[:50] + "..."
			}
			return msg.Content
		}
	}
	return "空会话"
}

// generateTags 生成标签
func (m *Manager) generateTags() []string {
	var tags []string

	// 从上下文生成标签
	if len(m.currentSession.Context.StockCodes) > 0 {
		tags = append(tags, "股票查询")
	}
	if len(m.currentSession.Context.ToolsUsed) > 0 {
		tags = append(tags, "工具使用")
	}

	return tags
}

// ==================== 知识提取 ====================

// extractKnowledgeAsync 异步提取知识
func (m *Manager) extractKnowledgeAsync(session *model.Session) {
	log.Infof("开始提取知识: %s", session.ID)

	// 尝试 AI 提取
	knowledges, err := ExtractKnowledgeWithAI(session)
	if err != nil {
		log.Warnf("AI 提取知识失败: %v，使用降级策略", err)
		// 使用降级策略
		knowledges, err = ExtractKnowledgeWithFallback(session)
		if err != nil {
			log.Errorf("降级策略也失败: %v", err)
			return
		}
	}

	// 设置来源并保存知识
	for _, knowledge := range knowledges {
		knowledge.Source = session.ID

		// 设置 ID（如果还没有）
		if knowledge.ID == "" {
			knowledge.ID = uuid.New().String()
		}

		if err := m.store.SaveKnowledge(knowledge); err != nil {
			log.Errorf("保存知识失败: %v", err)
		} else {
			log.Debugf("知识已保存: %s (类型: %s)", knowledge.ID, knowledge.Type)
		}
	}

	// 刷新知识缓存
	m.LoadKnowledge()

	log.Infof("知识提取完成: %d 条", len(knowledges))
}

// ==================== 知识管理 ====================

// LoadKnowledge 加载并聚合知识
func (m *Manager) LoadKnowledge() error {
	allKnowledge, err := m.store.GetAllKnowledge()
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 聚合用户画像（取最新的）
	m.knowledgeCache.UserProfiles = aggregateUserProfiles(allKnowledge[model.KnowledgeTypeUserProfile])

	// 合并关注列表（去重）
	m.knowledgeCache.Watchlist = aggregateWatchlist(allKnowledge[model.KnowledgeTypeWatchlist])

	// 收集所有概念（从 Knowledge 转换）
	m.knowledgeCache.Concepts = convertToConcepts(allKnowledge[model.KnowledgeTypeConcept])

	// 收集策略
	m.knowledgeCache.Strategies = convertToStrategies(allKnowledge[model.KnowledgeTypeStrategy])

	// 收集问答
	m.knowledgeCache.QAndAs = convertToQAndAs(allKnowledge[model.KnowledgeTypeQAndA])

	// 收集市场洞察
	m.knowledgeCache.MarketInsights = convertToMarketInsights(allKnowledge[model.KnowledgeTypeMarketInsight])

	m.knowledgeCache.LastUpdated = time.Now()

	log.Infof("知识已加载: %d 个用户画像, %d 个概念, %d 个策略",
		len(m.knowledgeCache.UserProfiles),
		len(m.knowledgeCache.Concepts),
		len(m.knowledgeCache.Strategies))

	return nil
}

// aggregateUserProfiles 聚合用户画像（取最新的）
func aggregateUserProfiles(profiles []*model.Knowledge) []*model.UserProfile {
	var result []*model.UserProfile
	now := time.Now()

	for _, k := range profiles {
		if profile, ok := k.Content.(*model.UserProfile); ok {
			// 检查是否当前生效
			if profile.ValidTo.IsZero() || profile.ValidTo.After(now) {
				result = append(result, profile)
			}
		}
	}

	// 按生效时间排序，取最新的
	// 简化处理，实际可能需要更复杂的合并逻辑

	return result
}

// aggregateWatchlist 聚合关注列表（去重）
func aggregateWatchlist(lists []*model.Knowledge) []*model.Watchlist {
	stockMap := make(map[string]*model.StockItem)
	var result []*model.Watchlist
	now := time.Now()

	for _, k := range lists {
		if list, ok := k.Content.(*model.Watchlist); ok {
			// 检查是否当前生效
			if list.ValidTo.IsZero() || list.ValidTo.After(now) {
				// 去重合并股票
				for _, item := range list.StockCodes {
					if existing, found := stockMap[item.Code]; found {
						// 保留最新的
						if item.AddedAt.After(existing.AddedAt) {
							stockMap[item.Code] = &item
						}
					} else {
						stockMap[item.Code] = &item
					}
				}
				result = append(result, list)
			}
		}
	}

	return result
}

// GetKnowledgeForPrompt 获取用于注入 AI Prompt 的知识
func (m *Manager) GetKnowledgeForPrompt() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sections []string
	totalLength := 0
	maxLength := 8000 // 约 2000 tokens，粗略估计 1 token ≈ 4 字符

	// 1. 用户画像
	if len(m.knowledgeCache.UserProfiles) > 0 {
		section := buildUserProfileSection(m.knowledgeCache.UserProfiles)
		if totalLength+len(section) < maxLength {
			sections = append(sections, section)
			totalLength += len(section)
		}
	}

	// 2. 关注列表
	if len(m.knowledgeCache.Watchlist) > 0 {
		section := buildWatchlistSection(m.knowledgeCache.Watchlist)
		if totalLength+len(section) < maxLength {
			sections = append(sections, section)
			totalLength += len(section)
		}
	}

	// 3. 学过的概念（按理解程度排序）
	if len(m.knowledgeCache.Concepts) > 0 {
		section := buildConceptsSection(m.knowledgeCache.Concepts)
		if totalLength+len(section) < maxLength {
			sections = append(sections, section)
			totalLength += len(section)
		}
	}

	// 4. 常见问答
	if len(m.knowledgeCache.QAndAs) > 0 {
		section := buildQAndASection(m.knowledgeCache.QAndAs)
		if totalLength+len(section) < maxLength {
			sections = append(sections, section)
			totalLength += len(section)
		}
	}

	if len(sections) == 0 {
		return ""
	}

	return `# 【记忆与上下文】

以下是从历史对话中提取的用户知识和偏好，请在回复时参考：

` + strings.Join(sections, "\n\n")
}

// buildUserProfileSection 构建用户画像部分
func buildUserProfileSection(profiles []*model.UserProfile) string {
	if len(profiles) == 0 {
		return ""
	}

	profile := profiles[len(profiles)-1] // 取最新的

	var parts []string
	parts = append(parts, "## 用户画像")

	if profile.RiskPreference != "" {
		parts = append(parts, fmt.Sprintf("- **风险偏好**: %s", profile.RiskPreference))
	}
	if len(profile.InvestmentStyle) > 0 {
		parts = append(parts, fmt.Sprintf("- **投资风格**: %s", strings.Join(profile.InvestmentStyle, "、")))
	}
	if len(profile.FocusAreas) > 0 {
		parts = append(parts, fmt.Sprintf("- **关注领域**: %s", strings.Join(profile.FocusAreas, "、")))
	}
	if profile.ExperienceLevel != "" {
		parts = append(parts, fmt.Sprintf("- **经验水平**: %s", profile.ExperienceLevel))
	}

	return strings.Join(parts, "\n")
}

// buildWatchlistSection 构建关注列表部分
func buildWatchlistSection(lists []*model.Watchlist) string {
	if len(lists) == 0 {
		return ""
	}

	var parts []string
	parts = append(parts, "## 关注列表")

	// 合并所有股票代码（去重）
	stockMap := make(map[string]string) // code -> reason
	for _, list := range lists {
		for _, item := range list.StockCodes {
			if item.Reason != "" {
				stockMap[item.Code] = item.Reason
			} else {
				stockMap[item.Code] = ""
			}
		}
	}

	if len(stockMap) > 0 {
		var codes []string
		for code, reason := range stockMap {
			if reason != "" {
				codes = append(codes, fmt.Sprintf("- %s (%s)", code, reason))
			} else {
				codes = append(codes, fmt.Sprintf("- %s", code))
			}
		}
		parts = append(parts, strings.Join(codes, "\n"))
	}

	return strings.Join(parts, "\n")
}

// buildConceptsSection 构建概念部分
func buildConceptsSection(concepts []*model.ConceptLearned) string {
	if len(concepts) == 0 {
		return ""
	}

	var parts []string
	parts = append(parts, "## 学过的概念")

	// 只显示理解程度 > 0.5 的概念
	count := 0
	for _, concept := range concepts {
		if concept.Confidence > 0.5 {
			parts = append(parts, fmt.Sprintf("- **%s**: %s (理解程度: %.0f%%)",
				concept.Name, concept.Description, concept.Confidence*100))
			count++
			if count >= 5 { // 最多显示 5 个
				break
			}
		}
	}

	return strings.Join(parts, "\n")
}

// buildQAndASection 构建问答部分
func buildQAndASection(qas []*model.QAndA) string {
	if len(qas) == 0 {
		return ""
	}

	var parts []string
	parts = append(parts, "## 常见问答")

	count := 0
	for _, qa := range qas {
		parts = append(parts, fmt.Sprintf("- **Q**: %s\n  **A**: %s", qa.Question, qa.Answer))
		count++
		if count >= 3 { // 最多显示 3 个
			break
		}
	}

	return strings.Join(parts, "\n")
}

// convertToConcepts 转换 Knowledge 列表为 ConceptLearned 列表
func convertToConcepts(knowledges []*model.Knowledge) []*model.ConceptLearned {
	var result []*model.ConceptLearned
	for _, k := range knowledges {
		if concept, ok := k.Content.(*model.ConceptLearned); ok {
			result = append(result, concept)
		}
	}
	return result
}

// convertToStrategies 转换 Knowledge 列表为 TradingStrategy 列表
func convertToStrategies(knowledges []*model.Knowledge) []*model.TradingStrategy {
	var result []*model.TradingStrategy
	for _, k := range knowledges {
		if strategy, ok := k.Content.(*model.TradingStrategy); ok {
			result = append(result, strategy)
		}
	}
	return result
}

// convertToQAndAs 转换 Knowledge 列表为 QAndA 列表
func convertToQAndAs(knowledges []*model.Knowledge) []*model.QAndA {
	var result []*model.QAndA
	for _, k := range knowledges {
		if qa, ok := k.Content.(*model.QAndA); ok {
			result = append(result, qa)
		}
	}
	return result
}

// convertToMarketInsights 转换 Knowledge 列表为 MarketInsight 列表
func convertToMarketInsights(knowledges []*model.Knowledge) []*model.MarketInsight {
	var result []*model.MarketInsight
	for _, k := range knowledges {
		if insight, ok := k.Content.(*model.MarketInsight); ok {
			result = append(result, insight)
		}
	}
	return result
}

// ==================== 公共方法：获取数据 ====================

// ListSessions 获取会话列表（支持分页）
func (m *Manager) ListSessions(offset, limit int) ([]model.SessionIndex, error) {
	return m.store.ListSessions(offset, limit)
}

// GetSessionCount 获取会话总数
func (m *Manager) GetSessionCount() int {
	if stats, err := m.store.GetStats(); err == nil {
		return stats.TotalSessions
	}
	return 0
}

// GetKnowledgeCount 获取知识总数
func (m *Manager) GetKnowledgeCount() int {
	if stats, err := m.store.GetStats(); err == nil {
		return stats.TotalKnowledge
	}
	return 0
}

// GetAllKnowledge 获取所有知识
func (m *Manager) GetAllKnowledge() ([]*model.Knowledge, error) {
	allKnowledge, err := m.store.GetAllKnowledge()
	if err != nil {
		return nil, err
	}

	// 将 map 转换为切片
	var result []*model.Knowledge
	for _, kList := range allKnowledge {
		result = append(result, kList...)
	}

	return result, nil
}

// GetStore 获取存储实例（供会话详情视图使用）
func (m *Manager) GetStore() Store {
	return m.store
}
