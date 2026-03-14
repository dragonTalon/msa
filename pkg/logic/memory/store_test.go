package memory

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"msa/pkg/model"
)

// TestNewFileStore 测试存储初始化
func TestNewFileStore(t *testing.T) {
	store, err := NewFileStore()
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}

	// 验证基础目录设置
	if store.baseDir == "" {
		t.Error("基础目录未设置")
	}

	// 验证目录结构存在（至少应该有基础目录）
	if _, err := os.Stat(store.baseDir); os.IsNotExist(err) {
		t.Errorf("基础目录不存在: %s", store.baseDir)
	}

	_ = store
}

// TestSaveAndGetSession 测试会话保存和获取
func TestSaveAndGetSession(t *testing.T) {
	store, err := NewFileStore()
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}

	// 创建测试会话
	startTime := time.Now().Add(-1 * time.Hour)
	sessionID := uuid.New().String()
	session := &model.Session{
		ID:        sessionID,
		StartTime: startTime,
		EndTime:   time.Now(),
		Messages: []model.MemoryMessage{
			{
				ID:        uuid.New().String(),
				Type:      model.MemoryMessageTypeUser,
				Content:   "测试消息",
				Timestamp: time.Now(),
			},
		},
		MessageCount: 1,
		Context:      &model.SessionContext{}, // 初始化为空指针
	}

	// 保存会话
	err = store.SaveSession(session)
	if err != nil {
		t.Fatalf("保存会话失败: %v", err)
	}

	// 获取会话（需要传入 startTime）
	retrieved, err := store.GetSession(sessionID, startTime)
	if err != nil {
		t.Fatalf("获取会话失败: %v", err)
	}

	// 验证会话内容
	if retrieved.ID != session.ID {
		t.Errorf("会话 ID 不匹配: got %s, want %s", retrieved.ID, session.ID)
	}
	if len(retrieved.Messages) != len(session.Messages) {
		t.Errorf("消息数量不匹配: got %d, want %d", len(retrieved.Messages), len(session.Messages))
	}

	// 清理：删除测试会话
	_ = store.DeleteSession(sessionID, startTime)
}

// TestListSessions 测试会话列表
func TestListSessions(t *testing.T) {
	store, err := NewFileStore()
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}

	// 创建多个测试会话
	now := time.Now()
	sessionIDs := []string{}
	startTimes := []time.Time{}
	for i := 0; i < 5; i++ {
		sessionID := uuid.New().String()
		sessionIDs = append(sessionIDs, sessionID)
		startTime := now.Add(time.Duration(-i) * time.Hour)
		startTimes = append(startTimes, startTime)
		session := &model.Session{
			ID:        sessionID,
			StartTime: startTime,
			EndTime:   startTime.Add(30 * time.Minute),
			Messages: []model.MemoryMessage{
				{
					ID:        uuid.New().String(),
					Type:      model.MemoryMessageTypeUser,
					Content:   "测试消息",
					Timestamp: now,
				},
			},
			MessageCount: 1,
			Context:      &model.SessionContext{},
		}
		if err := store.SaveSession(session); err != nil {
			t.Fatalf("保存会话 %d 失败: %v", i, err)
		}
	}

	// 列出会话
	sessions, err := store.ListSessions(0, 10)
	if err != nil {
		t.Fatalf("列出会话失败: %v", err)
	}

	// 验证数量（可能包含之前的会话，所以只检查至少有我们创建的会话）
	if len(sessions) < 5 {
		t.Errorf("会话数量不足: got %d, want at least 5", len(sessions))
	}

	// 验证排序（应该按时间倒序）
	for i := 0; i < len(sessions)-1; i++ {
		if sessions[i].StartTime.Before(sessions[i+1].StartTime) {
			t.Errorf("会话未按时间倒序排列")
			break
		}
	}

	// 清理
	for i, id := range sessionIDs {
		_ = store.DeleteSession(id, startTimes[i])
	}
}

// TestSaveAndListKnowledge 测试知识保存和列表
func TestSaveAndListKnowledge(t *testing.T) {
	store, err := NewFileStore()
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}

	// 创建测试知识
	concept := model.ConceptLearned{
		ID:          uuid.New().String(),
		Name:        "测试概念",
		Description: "这是一个测试概念",
		Category:    "test",
		Confidence:  0.8,
		LearnedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	knowledge1 := &model.Knowledge{
		ID:        uuid.New().String(),
		Type:      model.KnowledgeTypeConcept,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Source:    "test-session",
		Content:   concept,
	}

	strategy := model.TradingStrategy{
		ID:          uuid.New().String(),
		Name:        "测试策略",
		Description: "这是一个测试策略",
		Rules:       []string{"规则1", "规则2"},
		Conditions:  []string{"条件1"},
		Version:     1,
		CreatedAt:   time.Now(),
		ValidFrom:   time.Now(),
	}

	knowledge2 := &model.Knowledge{
		ID:        uuid.New().String(),
		Type:      model.KnowledgeTypeStrategy,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Source:    "test-session",
		Content:   strategy,
	}

	// 保存知识
	for _, k := range []*model.Knowledge{knowledge1, knowledge2} {
		if err := store.SaveKnowledge(k); err != nil {
			t.Fatalf("保存知识失败: %v", err)
		}
	}

	// 获取所有知识
	all, err := store.GetAllKnowledge()
	if err != nil {
		t.Fatalf("获取所有知识失败: %v", err)
	}

	// 验证类型和数量
	concepts, ok := all[model.KnowledgeTypeConcept]
	if !ok {
		t.Error("没有找到概念类型的知识")
	} else if len(concepts) == 0 {
		t.Error("概念知识列表为空")
	}

	strategies, ok := all[model.KnowledgeTypeStrategy]
	if !ok {
		t.Error("没有找到策略类型的知识")
	} else if len(strategies) == 0 {
		t.Error("策略知识列表为空")
	}
}

// TestShortTermMemory 测试短期记忆
func TestShortTermMemory(t *testing.T) {
	store, err := NewFileStore()
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}

	// 创建短期记忆
	memory := &model.ShortTermMemory{
		SessionID: uuid.New().String(),
		Messages: []model.MemoryMessage{
			{
				ID:        uuid.New().String(),
				Type:      model.MemoryMessageTypeUser,
				Content:   "消息1",
				Timestamp: time.Now(),
			},
			{
				ID:        uuid.New().String(),
				Type:      model.MemoryMessageTypeAssistant,
				Content:   "回复1",
				Timestamp: time.Now(),
			},
		},
		UpdatedAt: time.Now(),
	}

	// 保存
	if err := store.SaveShortTermMemory(memory); err != nil {
		t.Fatalf("保存短期记忆失败: %v", err)
	}

	// 获取
	retrieved, err := store.GetShortTermMemory()
	if err != nil {
		t.Fatalf("获取短期记忆失败: %v", err)
	}

	if len(retrieved.Messages) != len(memory.Messages) {
		t.Errorf("消息数量不匹配: got %d, want %d", len(retrieved.Messages), len(memory.Messages))
	}
}

// TestSearch 测试搜索功能
func TestSearch(t *testing.T) {
	store, err := NewFileStore()
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}

	// 创建测试会话
	startTime := time.Now().Add(-1 * time.Hour)
	sessionID := uuid.New().String()
	session := &model.Session{
		ID:        sessionID,
		StartTime: startTime,
		EndTime:   time.Now(),
		Messages: []model.MemoryMessage{
			{
				ID:        uuid.New().String(),
				Type:      model.MemoryMessageTypeUser,
				Content:   "腾讯股价是多少",
				Timestamp: time.Now(),
			},
		},
		MessageCount: 1,
		Summary:      "腾讯股价查询", // 添加摘要以便搜索
		Context: &model.SessionContext{
			StockCodes: []string{"00700.HK"},
		},
	}

	if err := store.SaveSession(session); err != nil {
		t.Fatalf("保存会话失败: %v", err)
	}

	// 搜索腾讯
	filters := model.SearchFilters{
		Query:  "腾讯",
		Limit:  10,
		Offset: 0,
	}
	results, err := store.Search(filters)
	if err != nil {
		t.Fatalf("搜索失败: %v", err)
	}

	if len(results) == 0 {
		t.Error("搜索结果为空，期望找到包含'腾讯'的会话")
	}

	// 清理
	_ = store.DeleteSession(sessionID, startTime)
}

// TestDeleteSession 测试删除会话
func TestDeleteSession(t *testing.T) {
	store, err := NewFileStore()
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}

	// 创建测试会话
	startTime := time.Now()
	sessionID := uuid.New().String()
	session := &model.Session{
		ID:        sessionID,
		StartTime: startTime,
		EndTime:   time.Now(),
		Messages:  []model.MemoryMessage{},
		Context:   &model.SessionContext{},
	}

	if err := store.SaveSession(session); err != nil {
		t.Fatalf("保存会话失败: %v", err)
	}

	// 删除会话
	if err := store.DeleteSession(sessionID, startTime); err != nil {
		t.Fatalf("删除会话失败: %v", err)
	}

	// 验证删除
	_, err = store.GetSession(sessionID, startTime)
	if err == nil {
		t.Error("会话应该已被删除，但仍然可以获取")
	}
}

// TestGetStats 测试统计功能
func TestGetStats(t *testing.T) {
	store, err := NewFileStore()
	if err != nil {
		t.Fatalf("创建存储失败: %v", err)
	}

	// 创建一些测试数据
	startTime := time.Now()
	sessionID := uuid.New().String()
	session := &model.Session{
		ID:        sessionID,
		StartTime: startTime,
		EndTime:   time.Now(),
		Messages: []model.MemoryMessage{
			{
				ID:        uuid.New().String(),
				Type:      model.MemoryMessageTypeUser,
				Content:   "测试",
				Timestamp: time.Now(),
			},
		},
		MessageCount: 1,
		Context:      &model.SessionContext{},
	}

	if err := store.SaveSession(session); err != nil {
		t.Fatalf("保存会话失败: %v", err)
	}

	concept := model.ConceptLearned{
		ID:          uuid.New().String(),
		Name:        "测试",
		Description: "测试概念",
		Category:    "test",
		Confidence:  0.5,
		LearnedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	knowledge := &model.Knowledge{
		ID:        uuid.New().String(),
		Type:      model.KnowledgeTypeConcept,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Source:    sessionID,
		Content:   concept,
	}

	if err := store.SaveKnowledge(knowledge); err != nil {
		t.Fatalf("保存知识失败: %v", err)
	}

	// 获取统计
	stats, err := store.GetStats()
	if err != nil {
		t.Fatalf("获取统计失败: %v", err)
	}

	// 验证至少有我们创建的数据
	if stats.TotalSessions < 1 {
		t.Errorf("总会话数不足: got %d, want at least 1", stats.TotalSessions)
	}
	if stats.TotalKnowledge < 1 {
		t.Errorf("总知识数不足: got %d, want at least 1", stats.TotalKnowledge)
	}

	// 清理
	_ = store.DeleteSession(sessionID, startTime)
}
