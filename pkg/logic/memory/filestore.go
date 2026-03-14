package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"msa/pkg/model"
)

// ==================== FileStore 结构体 ====================

// FileStore 基于文件系统的存储实现
type FileStore struct {
	baseDir      string       // 基础目录: ~/.msa/remember/
	mu           sync.RWMutex // 读写锁
	indexes      *Indexes     // 内存索引
	indexVersion int          // 索引版本号（用于原子更新）
}

// ==================== 构造函数 ====================

// NewFileStore 创建新的文件存储实例
func NewFileStore() (*FileStore, error) {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("获取用户主目录失败: %w", err)
	}

	// 创建基础目录
	baseDir := filepath.Join(homeDir, ".msa", "remember")

	store := &FileStore{
		baseDir:      baseDir,
		indexes:      &Indexes{},
		indexVersion: 0,
	}

	// 创建目录结构
	if err := store.createDirectoryStructure(); err != nil {
		return nil, fmt.Errorf("创建目录结构失败: %w", err)
	}

	// 加载索引
	if err := store.loadIndexes(); err != nil {
		log.Warnf("加载索引失败: %v，将重建索引", err)
		if err := store.rebuildIndexes(); err != nil {
			log.Warnf("重建索引失败: %v", err)
		}
	}

	return store, nil
}

// createDirectoryStructure 创建目录结构
func (s *FileStore) createDirectoryStructure() error {
	dirs := []string{
		s.baseDir,
		filepath.Join(s.baseDir, DirNameSessions),
		filepath.Join(s.baseDir, DirNameKnowledge),
		filepath.Join(s.baseDir, DirNameShortTerm),
		filepath.Join(s.baseDir, DirNameIndexes),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return NewStoreError("创建目录", dir, err)
		}
	}

	log.Debugf("记忆系统目录结构已创建: %s", s.baseDir)
	return nil
}

// ==================== 索引操作 ====================

// loadIndexes 加载索引
// 启动时调用，失败时自动重建
func (s *FileStore) loadIndexes() error {
	indexFile := filepath.Join(s.baseDir, DirNameIndexes, IndexFileNameAll)

	data, err := os.ReadFile(indexFile)
	if err != nil {
		if os.IsNotExist(err) {
			// 索引文件不存在，初始化空索引
			s.indexes = &Indexes{
				Sessions:   []model.SessionIndex{},
				Knowledge:  []model.KnowledgeIndex{},
				LastUpdate: time.Now(),
				Version:    0,
			}
			return nil
		}
		return NewStoreError("读取索引", indexFile, err)
	}

	var indexes Indexes
	if err := json.Unmarshal(data, &indexes); err != nil {
		// 索引文件损坏，需要重建
		return fmt.Errorf("索引文件损坏: %w", err)
	}

	s.mu.Lock()
	s.indexes = &indexes
	s.indexVersion = indexes.Version
	s.mu.Unlock()

	log.Debugf("索引已加载: %d 个会话, %d 个知识",
		len(s.indexes.Sessions), len(s.indexes.Knowledge))
	return nil
}

// saveIndexes 保存索引
// 会话或知识更新后调用
// 注意：调用此方法前必须持有 s.mu 锁，或者使用 saveIndexesUnsafe
func (s *FileStore) saveIndexes() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveIndexesUnsafe()
}

// saveIndexesUnsafe 保存索引（不加锁版本）
// 调用者必须确保已持有 s.mu 锁
func (s *FileStore) saveIndexesUnsafe() error {
	s.indexes.LastUpdate = time.Now()
	s.indexes.Version++
	s.indexVersion = s.indexes.Version
	data, err := json.MarshalIndent(s.indexes, "", "  ")

	if err != nil {
		return fmt.Errorf("序列化索引失败: %w", err)
	}

	indexFile := filepath.Join(s.baseDir, DirNameIndexes, IndexFileNameAll)

	// 原子写入：先写临时文件，再重命名
	tmpFile := indexFile + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return NewStoreError("写入索引", tmpFile, err)
	}

	if err := os.Rename(tmpFile, indexFile); err != nil {
		return NewStoreError("重命名索引", indexFile, err)
	}

	log.Debugf("索引已保存，版本: %d", s.indexVersion)
	return nil
}

// rebuildIndexes 重建索引
// 索引损坏或不存在时调用
func (s *FileStore) rebuildIndexes() error {
	log.Info("开始重建索引...")

	newIndexes := &Indexes{
		Sessions:   []model.SessionIndex{},
		Knowledge:  []model.KnowledgeIndex{},
		LastUpdate: time.Now(),
		Version:    s.indexVersion + 1,
	}

	// 扫描会话目录
	sessionsDir := filepath.Join(s.baseDir, DirNameSessions)
	if err := filepath.Walk(sessionsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".json") {
			return nil
		}

		// 读取会话文件
		data, err := os.ReadFile(path)
		if err != nil {
			log.Warnf("读取会话文件失败: %s, %v", path, err)
			return nil // 继续处理其他文件
		}

		var session model.Session
		if err := json.Unmarshal(data, &session); err != nil {
			log.Warnf("解析会话文件失败: %s, %v", path, err)
			return nil
		}

		// 添加到索引
		newIndexes.Sessions = append(newIndexes.Sessions, model.SessionIndex{
			ID:           session.ID,
			StartTime:    session.StartTime,
			EndTime:      session.EndTime,
			Summary:      session.Summary,
			MessageCount: session.MessageCount,
			Tags:         session.Tags,
			StockCodes:   session.Context.StockCodes,
		})

		return nil
	}); err != nil {
		return fmt.Errorf("扫描会话目录失败: %w", err)
	}

	// 扫描知识目录
	knowledgeDir := filepath.Join(s.baseDir, DirNameKnowledge)
	if err := filepath.Walk(knowledgeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".json") {
			return nil
		}

		// 读取知识文件
		data, err := os.ReadFile(path)
		if err != nil {
			log.Warnf("读取知识文件失败: %s, %v", path, err)
			return nil
		}

		var knowledge model.Knowledge
		if err := json.Unmarshal(data, &knowledge); err != nil {
			log.Warnf("解析知识文件失败: %s, %v", path, err)
			return nil
		}

		// 生成摘要
		summary := s.knowledgeSummary(&knowledge)

		// 添加到索引
		newIndexes.Knowledge = append(newIndexes.Knowledge, model.KnowledgeIndex{
			ID:        knowledge.ID,
			Type:      knowledge.Type,
			CreatedAt: knowledge.CreatedAt,
			UpdatedAt: knowledge.UpdatedAt,
			Summary:   summary,
		})

		return nil
	}); err != nil {
		return fmt.Errorf("扫描知识目录失败: %w", err)
	}

	// 更新索引
	s.mu.Lock()
	s.indexes = newIndexes
	s.indexVersion = newIndexes.Version
	s.mu.Unlock()

	// 保存索引
	if err := s.saveIndexes(); err != nil {
		return fmt.Errorf("保存重建的索引失败: %w", err)
	}

	log.Infof("索引重建完成: %d 个会话, %d 个知识",
		len(newIndexes.Sessions), len(newIndexes.Knowledge))
	return nil
}

// knowledgeSummary 生成知识摘要
func (s *FileStore) knowledgeSummary(k *model.Knowledge) string {
	switch k.Type {
	case model.KnowledgeTypeUserProfile:
		if profile, ok := k.Content.(*model.UserProfile); ok {
			return fmt.Sprintf("用户画像 - 风险偏好: %s, 投资风格: %v",
				profile.RiskPreference, profile.InvestmentStyle)
		}
	case model.KnowledgeTypeWatchlist:
		if list, ok := k.Content.(*model.Watchlist); ok {
			codes := make([]string, len(list.StockCodes))
			for i, item := range list.StockCodes {
				codes[i] = item.Code
			}
			return fmt.Sprintf("关注列表 - %d 只股票: %v", len(list.StockCodes), codes)
		}
	case model.KnowledgeTypeConcept:
		if concept, ok := k.Content.(*model.ConceptLearned); ok {
			return fmt.Sprintf("概念: %s - %s", concept.Name, concept.Description)
		}
	case model.KnowledgeTypeStrategy:
		if strategy, ok := k.Content.(*model.TradingStrategy); ok {
			return fmt.Sprintf("策略: %s - %s", strategy.Name, strategy.Description)
		}
	case model.KnowledgeTypeQAndA:
		if qa, ok := k.Content.(*model.QAndA); ok {
			return fmt.Sprintf("问答: %s -> %s", qa.Question, qa.Answer)
		}
	case model.KnowledgeTypeMarketInsight:
		if insight, ok := k.Content.(*model.MarketInsight); ok {
			return fmt.Sprintf("市场洞察: %s", insight.Topic)
		}
	}
	return "未知知识类型"
}

// ==================== 辅助方法 ====================

// insertOrUpdateIndex 插入或更新会话索引
func (s *FileStore) insertOrUpdateSessionIndex(session *model.Session) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 查找是否已存在
	found := false
	for i, idx := range s.indexes.Sessions {
		if idx.ID == session.ID {
			// 更新现有索引
			s.indexes.Sessions[i] = model.SessionIndex{
				ID:           session.ID,
				StartTime:    session.StartTime,
				EndTime:      session.EndTime,
				Summary:      session.Summary,
				MessageCount: session.MessageCount,
				Tags:         session.Tags,
				StockCodes:   session.Context.StockCodes,
			}
			found = true
			break
		}
	}

	// 如果不存在，添加新索引
	if !found {
		s.indexes.Sessions = append(s.indexes.Sessions, model.SessionIndex{
			ID:           session.ID,
			StartTime:    session.StartTime,
			EndTime:      session.EndTime,
			Summary:      session.Summary,
			MessageCount: session.MessageCount,
			Tags:         session.Tags,
			StockCodes:   session.Context.StockCodes,
		})
	}
}

// insertOrUpdateKnowledgeIndex 插入或更新知识索引
func (s *FileStore) insertOrUpdateKnowledgeIndex(knowledge *model.Knowledge) {
	s.mu.Lock()
	defer s.mu.Unlock()

	summary := s.knowledgeSummary(knowledge)

	// 查找是否已存在
	found := false
	for i, idx := range s.indexes.Knowledge {
		if idx.ID == knowledge.ID {
			// 更新现有索引
			s.indexes.Knowledge[i] = model.KnowledgeIndex{
				ID:        knowledge.ID,
				Type:      knowledge.Type,
				CreatedAt: knowledge.CreatedAt,
				UpdatedAt: knowledge.UpdatedAt,
				Summary:   summary,
			}
			found = true
			break
		}
	}

	// 如果不存在，添加新索引
	if !found {
		s.indexes.Knowledge = append(s.indexes.Knowledge, model.KnowledgeIndex{
			ID:        knowledge.ID,
			Type:      knowledge.Type,
			CreatedAt: knowledge.CreatedAt,
			UpdatedAt: knowledge.UpdatedAt,
			Summary:   summary,
		})
	}
}

// getMonthDir 获取年月目录
func (s *FileStore) getMonthDir(t time.Time) string {
	return t.Format("2006-01")
}

// getKnowledgeTypeDir 获取知识类型目录
func (s *FileStore) getKnowledgeTypeDir(knowledgeType model.KnowledgeType) string {
	return string(knowledgeType)
}

// ==================== 会话操作实现 ====================

// SaveSession 保存完整会话
// 按年月组织目录（sessions/2024-03/），JSON 格式保存
func (s *FileStore) SaveSession(session *model.Session) error {
	const maxRetries = 2
	var lastErr error

	// 重试机制：尝试保存，失败时重试
	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := s.writeSessionFile(session)
		if err == nil {
			// 保存成功，更新索引
			s.insertOrUpdateSessionIndex(session)
			if err := s.saveIndexes(); err != nil {
				log.Warnf("保存索引失败: %v", err)
			}
			return nil
		}

		lastErr = err

		// 判断是否可重试
		if attempt < maxRetries && isTemporaryError(err) {
			log.Debugf("保存会话失败，重试 %d/%d: %v", attempt+1, maxRetries, err)
			time.Sleep(time.Duration(RetryDelayMs) * time.Duration(attempt+1) * time.Millisecond)
			continue
		}
	}

	return NewStoreError("保存会话", session.ID, lastErr)
}

// writeSessionFile 实际写入会话文件
func (s *FileStore) writeSessionFile(session *model.Session) error {
	// 按年月组织目录
	monthDir := s.getMonthDir(session.StartTime)
	sessionDir := filepath.Join(s.baseDir, DirNameSessions, monthDir)

	// 创建目录
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return err
	}

	// 文件名: session-id.json
	fileName := session.ID + ".json"
	filePath := filepath.Join(sessionDir, fileName)

	// 序列化
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化会话失败: %w", err)
	}

	// 原子写入
	tmpFile := filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}

	if err := os.Rename(tmpFile, filePath); err != nil {
		return err
	}

	log.Debugf("会话已保存: %s", filePath)
	return nil
}

// isTemporaryError 判断是否为临时错误
func isTemporaryError(err error) bool {
	// 检查是否为系统调用中的临时错误
	// EINTR, EAGAIN 等通常可以重试
	if err == nil {
		return false
	}

	// 检查错误字符串中是否包含临时错误标识
	errStr := err.Error()
	temporaryErrors := []string{
		"interrupted",
		"temporary",
		"timeout",
		"resource temporarily unavailable",
	}

	for _, tempErr := range temporaryErrors {
		if strings.Contains(strings.ToLower(errStr), tempErr) {
			return true
		}
	}

	return false
}

// GetSession 根据 ID 和时间获取完整会话
func (s *FileStore) GetSession(id string, startTime time.Time) (*model.Session, error) {
	// 根据时间确定目录
	monthDir := s.getMonthDir(startTime)
	sessionDir := filepath.Join(s.baseDir, DirNameSessions, monthDir)
	filePath := filepath.Join(sessionDir, id+".json")

	// 读取文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("会话不存在: %s", id)
		}
		return nil, NewStoreError("读取会话", filePath, err)
	}

	// 解析
	var session model.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, NewStoreError("解析会话", filePath, err)
	}

	return &session, nil
}

// ListSessions 列出会话（支持分页）
// 返回按时间倒序排列的会话索引列表
func (s *FileStore) ListSessions(offset, limit int) ([]model.SessionIndex, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 从索引中获取
	sessions := make([]model.SessionIndex, len(s.indexes.Sessions))
	copy(sessions, s.indexes.Sessions)

	// 按时间倒序排序
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartTime.After(sessions[j].StartTime)
	})

	// 计算分页
	total := len(sessions)
	if offset < 0 {
		offset = 0
	}
	if offset >= total {
		return []model.SessionIndex{}, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	// 返回分页结果
	return sessions[offset:end], nil
}

// DeleteSession 删除指定会话
func (s *FileStore) DeleteSession(id string, startTime time.Time) error {
	// 根据时间确定目录
	monthDir := s.getMonthDir(startTime)
	sessionDir := filepath.Join(s.baseDir, DirNameSessions, monthDir)
	filePath := filepath.Join(sessionDir, id+".json")

	// 删除文件
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("会话不存在: %s", id)
		}
		return NewStoreError("删除会话", filePath, err)
	}

	// 从索引中移除
	s.mu.Lock()
	defer s.mu.Unlock()

	newSessions := make([]model.SessionIndex, 0, len(s.indexes.Sessions))
	for _, session := range s.indexes.Sessions {
		if session.ID != id {
			newSessions = append(newSessions, session)
		}
	}
	s.indexes.Sessions = newSessions

	// 保存索引（使用不加锁版本，因为已经持有锁）
	if err := s.saveIndexesUnsafe(); err != nil {
		log.Warnf("保存索引失败: %v", err)
	}

	log.Infof("会话已删除: %s", id)
	return nil
}

// ==================== 知识操作实现 ====================

// SaveKnowledge 保存知识
// 根据类型分目录保存，自动更新索引
func (s *FileStore) SaveKnowledge(knowledge *model.Knowledge) error {
	const maxRetries = 2
	var lastErr error

	// 重试机制
	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := s.writeKnowledgeFile(knowledge)
		if err == nil {
			// 保存成功，更新索引
			s.insertOrUpdateKnowledgeIndex(knowledge)
			if err := s.saveIndexes(); err != nil {
				log.Warnf("保存索引失败: %v", err)
			}
			return nil
		}

		lastErr = err

		// 判断是否可重试
		if attempt < maxRetries && isTemporaryError(err) {
			log.Debugf("保存知识失败，重试 %d/%d: %v", attempt+1, maxRetries, err)
			time.Sleep(time.Duration(RetryDelayMs) * time.Duration(attempt+1) * time.Millisecond)
			continue
		}
	}

	return NewStoreError("保存知识", knowledge.ID, lastErr)
}

// writeKnowledgeFile 实际写入知识文件
func (s *FileStore) writeKnowledgeFile(knowledge *model.Knowledge) error {
	// 根据类型分目录
	typeDir := s.getKnowledgeTypeDir(knowledge.Type)
	knowledgeDir := filepath.Join(s.baseDir, DirNameKnowledge, typeDir)

	// 创建目录
	if err := os.MkdirAll(knowledgeDir, 0755); err != nil {
		return err
	}

	// 文件名: knowledge-id.json
	fileName := knowledge.ID + ".json"
	filePath := filepath.Join(knowledgeDir, fileName)

	// 序列化
	data, err := json.MarshalIndent(knowledge, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化知识失败: %w", err)
	}

	// 原子写入
	tmpFile := filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}

	if err := os.Rename(tmpFile, filePath); err != nil {
		return err
	}

	log.Debugf("知识已保存: %s (类型: %s)", knowledge.ID, knowledge.Type)
	return nil
}

// GetKnowledge 根据 ID 获取知识
func (s *FileStore) GetKnowledge(id string) (*model.Knowledge, error) {
	// 需要遍历所有类型目录查找
	for _, knowledgeType := range []model.KnowledgeType{
		model.KnowledgeTypeUserProfile,
		model.KnowledgeTypeWatchlist,
		model.KnowledgeTypeConcept,
		model.KnowledgeTypeStrategy,
		model.KnowledgeTypeQAndA,
		model.KnowledgeTypeMarketInsight,
	} {
		typeDir := s.getKnowledgeTypeDir(knowledgeType)
		filePath := filepath.Join(s.baseDir, DirNameKnowledge, typeDir, id+".json")

		data, err := os.ReadFile(filePath)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, NewStoreError("读取知识", filePath, err)
			}
			// 文件不存在，继续查找下一个类型
			continue
		}

		var knowledge model.Knowledge
		if err := json.Unmarshal(data, &knowledge); err != nil {
			return nil, NewStoreError("解析知识", filePath, err)
		}

		return &knowledge, nil
	}

	return nil, fmt.Errorf("知识不存在: %s", id)
}

// ListKnowledgeByType 按类型列出知识
func (s *FileStore) ListKnowledgeByType(knowledgeType model.KnowledgeType) ([]model.KnowledgeIndex, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []model.KnowledgeIndex

	for _, idx := range s.indexes.Knowledge {
		if idx.Type == knowledgeType {
			result = append(result, idx)
		}
	}

	return result, nil
}

// GetAllKnowledge 获取所有知识（按类型分组）
func (s *FileStore) GetAllKnowledge() (map[model.KnowledgeType][]*model.Knowledge, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[model.KnowledgeType][]*model.Knowledge)

	// 初始化所有类型
	for _, kType := range []model.KnowledgeType{
		model.KnowledgeTypeUserProfile,
		model.KnowledgeTypeWatchlist,
		model.KnowledgeTypeConcept,
		model.KnowledgeTypeStrategy,
		model.KnowledgeTypeQAndA,
		model.KnowledgeTypeMarketInsight,
	} {
		result[kType] = []*model.Knowledge{}
	}

	// 遍历所有知识目录
	for _, kType := range []model.KnowledgeType{
		model.KnowledgeTypeUserProfile,
		model.KnowledgeTypeWatchlist,
		model.KnowledgeTypeConcept,
		model.KnowledgeTypeStrategy,
		model.KnowledgeTypeQAndA,
		model.KnowledgeTypeMarketInsight,
	} {
		typeDir := s.getKnowledgeTypeDir(kType)
		knowledgeDir := filepath.Join(s.baseDir, DirNameKnowledge, typeDir)

		files, err := os.ReadDir(knowledgeDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, NewStoreError("读取知识目录", knowledgeDir, err)
		}

		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
				continue
			}

			filePath := filepath.Join(knowledgeDir, file.Name())
			data, err := os.ReadFile(filePath)
			if err != nil {
				log.Warnf("读取知识文件失败: %s, %v", filePath, err)
				continue
			}

			var knowledge model.Knowledge
			if err := json.Unmarshal(data, &knowledge); err != nil {
				log.Warnf("解析知识文件失败: %s, %v", filePath, err)
				continue
			}

			result[kType] = append(result[kType], &knowledge)
		}
	}

	return result, nil
}

// ==================== 短期记忆操作实现 ====================

// SaveShortTermMemory 保存短期记忆
// 保存到 short-term/current.json，最多保存 10 条消息
func (s *FileStore) SaveShortTermMemory(memory *model.ShortTermMemory) error {
	shortTermDir := filepath.Join(s.baseDir, DirNameShortTerm)

	// 创建目录
	if err := os.MkdirAll(shortTermDir, 0755); err != nil {
		return NewStoreError("创建短期记忆目录", shortTermDir, err)
	}

	// 限制消息数量
	if len(memory.Messages) > MaxShortTermMessages {
		// 只保留最新的 N 条消息
		start := len(memory.Messages) - MaxShortTermMessages
		memory.Messages = memory.Messages[start:]
	}

	// 更新时间戳
	memory.UpdatedAt = time.Now()

	// 序列化
	data, err := json.MarshalIndent(memory, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化短期记忆失败: %w", err)
	}

	// 写入文件
	filePath := filepath.Join(shortTermDir, ShortTermMemoryFileName)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return NewStoreError("写入短期记忆", filePath, err)
	}

	log.Debugf("短期记忆已保存: %d 条消息", len(memory.Messages))
	return nil
}

// GetShortTermMemory 获取短期记忆
func (s *FileStore) GetShortTermMemory() (*model.ShortTermMemory, error) {
	filePath := filepath.Join(s.baseDir, DirNameShortTerm, ShortTermMemoryFileName)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 不存在时返回空的短期记忆
			return &model.ShortTermMemory{
				Messages:  []model.MemoryMessage{},
				UpdatedAt: time.Now(),
			}, nil
		}
		return nil, NewStoreError("读取短期记忆", filePath, err)
	}

	var memory model.ShortTermMemory
	if err := json.Unmarshal(data, &memory); err != nil {
		return nil, NewStoreError("解析短期记忆", filePath, err)
	}

	return &memory, nil
}

// ==================== 搜索操作实现 ====================

// Search 搜索会话和知识
func (s *FileStore) Search(filters model.SearchFilters) ([]model.SearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []model.SearchResult

	// 搜索会话
	sessionResults := s.searchSessions(filters)
	results = append(results, sessionResults...)

	// 搜索知识
	knowledgeResults := s.searchKnowledge(filters)
	results = append(results, knowledgeResults...)

	// 按相关性评分排序
	sortSearchResults(results)

	// 应用分页
	if filters.Limit <= 0 {
		filters.Limit = MaxSearchResults
	}

	start := filters.Offset
	if start < 0 {
		start = 0
	}
	if start > len(results) {
		return []model.SearchResult{}, nil
	}

	end := start + filters.Limit
	if end > len(results) {
		end = len(results)
	}

	return results[start:end], nil
}

// searchSessions 搜索会话
func (s *FileStore) searchSessions(filters model.SearchFilters) []model.SearchResult {
	var results []model.SearchResult
	query := strings.ToLower(filters.Query)

	for _, session := range s.indexes.Sessions {
		// 日期过滤（先检查指针是否为 nil）
		if filters.StartDate != nil && session.StartTime.Before(*filters.StartDate) {
			continue
		}
		if filters.EndDate != nil && session.EndTime.After(*filters.EndDate) {
			continue
		}

		// 股票代码过滤
		if len(filters.StockCodes) > 0 {
			found := false
			for _, code := range filters.StockCodes {
				if contains(session.StockCodes, code) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// 关键词搜索
		score := 0.0
		if query != "" {
			summary := strings.ToLower(session.Summary)
			if strings.Contains(summary, query) {
				score += 1.0
			}
			for _, tag := range session.Tags {
				if strings.Contains(strings.ToLower(tag), query) {
					score += 0.5
				}
			}
			for _, code := range session.StockCodes {
				if strings.Contains(strings.ToLower(code), query) {
					score += 0.8
				}
			}
		} else {
			score = float64(session.StartTime.Unix())
		}

		if score > 0 || query == "" {
			highlight := generateHighlight(session.Summary, query)
			results = append(results, model.SearchResult{
				Type:      model.SearchResultTypeSession,
				ID:        session.ID,
				Title:     session.StartTime.Format("2006-01-02 15:04"),
				Summary:   session.Summary,
				Score:     score,
				Timestamp: session.StartTime,
				Highlight: highlight,
			})
		}
	}

	return results
}

// searchKnowledge 搜索知识
func (s *FileStore) searchKnowledge(filters model.SearchFilters) []model.SearchResult {
	var results []model.SearchResult
	query := strings.ToLower(filters.Query)

	for _, knowledge := range s.indexes.Knowledge {
		// 日期过滤（先检查指针是否为 nil）
		if filters.StartDate != nil && knowledge.CreatedAt.Before(*filters.StartDate) {
			continue
		}
		if filters.EndDate != nil && knowledge.UpdatedAt.After(*filters.EndDate) {
			continue
		}

		score := 0.0
		if query != "" {
			summary := strings.ToLower(knowledge.Summary)
			if strings.Contains(summary, query) {
				score += 1.0
			}
		} else {
			score = float64(knowledge.UpdatedAt.Unix())
		}

		if score > 0 || query == "" {
			highlight := generateHighlight(knowledge.Summary, query)
			results = append(results, model.SearchResult{
				Type:      model.SearchResultTypeKnowledge,
				ID:        knowledge.ID,
				Title:     string(knowledge.Type),
				Summary:   knowledge.Summary,
				Score:     score,
				Timestamp: knowledge.UpdatedAt,
				Highlight: highlight,
			})
		}
	}

	return results
}

// ==================== 统计操作实现 ====================

// GetStats 获取统计信息
func (s *FileStore) GetStats() (*model.MemoryStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &model.MemoryStats{
		KnowledgeByType: make(map[model.KnowledgeType]int),
	}

	stats.TotalSessions = len(s.indexes.Sessions)

	// 计算总消息数并找到最早/最新会话
	var earliest, latest time.Time
	totalMessages := 0

	// 统计股票查询次数
	stockCounts := make(map[string]int)
	for _, session := range s.indexes.Sessions {
		totalMessages += session.MessageCount

		if earliest.IsZero() || session.StartTime.Before(earliest) {
			earliest = session.StartTime
		}
		if latest.IsZero() || session.StartTime.After(latest) {
			latest = session.StartTime
		}

		// 统计股票代码
		for _, code := range session.StockCodes {
			stockCounts[code]++
		}
	}

	stats.TotalMessages = totalMessages
	stats.EarliestSession = earliest
	stats.LatestSession = latest

	// 统计知识数量
	stats.TotalKnowledge = len(s.indexes.Knowledge)

	// 统计概念学习次数
	conceptCounts := make(map[string]int)
	conceptCategories := make(map[string]string)

	for _, k := range s.indexes.Knowledge {
		stats.KnowledgeByType[k.Type]++

		// 统计概念
		if k.Type == model.KnowledgeTypeConcept {
			// 尝试从知识内容中提取概念名称
			if knowledge, err := s.GetKnowledge(k.ID); err == nil {
				if concept, ok := knowledge.Content.(*model.ConceptLearned); ok {
					conceptCounts[concept.Name]++
					conceptCategories[concept.Name] = concept.Category
				} else if contentMap, ok := knowledge.Content.(map[string]interface{}); ok {
					if name, ok := contentMap["name"].(string); ok && name != "" {
						conceptCounts[name]++
						if category, ok := contentMap["category"].(string); ok {
							conceptCategories[name] = category
						}
					}
				}
			}
		}
	}

	// 计算 TOP 5 股票
	stats.TopStocks = calculateTopStocks(stockCounts, 5)

	// 计算 TOP 5 概念
	stats.TopConcepts = calculateTopConcepts(conceptCounts, conceptCategories, 5)

	stats.LastUpdated = time.Now()

	return stats, nil
}

// calculateTopStocks 计算最常查询的股票 TOP N
func calculateTopStocks(stockCounts map[string]int, topN int) []model.StockStat {
	// 转换为切片并排序
	type stockCount struct {
		code  string
		count int
	}

	var stocks []stockCount
	for code, count := range stockCounts {
		stocks = append(stocks, stockCount{code: code, count: count})
	}

	// 按次数降序排序
	for i := 0; i < len(stocks); i++ {
		for j := i + 1; j < len(stocks); j++ {
			if stocks[j].count > stocks[i].count {
				stocks[i], stocks[j] = stocks[j], stocks[i]
			}
		}
	}

	// 取 TOP N
	result := make([]model.StockStat, 0, topN)
	for i := 0; i < len(stocks) && i < topN; i++ {
		result = append(result, model.StockStat{
			Code:       stocks[i].code,
			QueryCount: stocks[i].count,
		})
	}

	return result
}

// calculateTopConcepts 计算学得最多的概念 TOP N
func calculateTopConcepts(conceptCounts map[string]int, conceptCategories map[string]string, topN int) []model.ConceptStat {
	// 转换为切片并排序
	type conceptCount struct {
		name     string
		category string
		count    int
	}

	var concepts []conceptCount
	for name, count := range conceptCounts {
		concepts = append(concepts, conceptCount{
			name:     name,
			category: conceptCategories[name],
			count:    count,
		})
	}

	// 按次数降序排序
	for i := 0; i < len(concepts); i++ {
		for j := i + 1; j < len(concepts); j++ {
			if concepts[j].count > concepts[i].count {
				concepts[i], concepts[j] = concepts[j], concepts[i]
			}
		}
	}

	// 取 TOP N
	result := make([]model.ConceptStat, 0, topN)
	for i := 0; i < len(concepts) && i < topN; i++ {
		result = append(result, model.ConceptStat{
			Name:       concepts[i].name,
			Category:   concepts[i].category,
			LearnCount: concepts[i].count,
		})
	}

	return result
}

// ==================== 辅助函数 ====================

func generateHighlight(text, query string) string {
	if query == "" {
		if len(text) > 100 {
			return text[:100] + "..."
		}
		return text
	}

	lowerText := strings.ToLower(text)
	lowerQuery := strings.ToLower(query)
	idx := strings.Index(lowerText, lowerQuery)

	if idx == -1 {
		if len(text) > 100 {
			return text[:100] + "..."
		}
		return text
	}

	start := idx - 30
	if start < 0 {
		start = 0
	}
	end := idx + len(query) + 30
	if end > len(text) {
		end = len(text)
	}

	prefix := ""
	if start > 0 {
		prefix = "..."
	}
	suffix := ""
	if end < len(text) {
		suffix = "..."
	}

	return prefix + text[start:end] + suffix
}

func sortSearchResults(results []model.SearchResult) {
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Score < results[j].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
