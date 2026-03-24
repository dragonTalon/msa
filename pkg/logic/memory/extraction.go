package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"msa/pkg/config"
	"msa/pkg/model"
)

const (
	// 知识提取 Prompt 模板
	knowledgeExtractionPrompt = `你是一个专业的投资顾问助手。请分析以下对话内容，提取结构化的用户知识。

# 分析任务
从对话中提取以下类型的知识（如果存在）：

1. **用户画像** (user-profile)
   - 风险偏好: conservative(保守)/moderate(适中)/aggressive(激进)
   - 投资风格: 价值投资/成长投资/动量投资等
   - 关注领域: 科技/金融/能源等
   - 经验水平: beginner(初级)/intermediate(中级)/advanced(高级)

2. **关注列表** (watchlist)
   - 股票代码和名称
   - 关注原因
   - 分类: long-term(长期持有)/short-term(短期交易)/watching(观察中)

3. **学过的概念** (concept)
   - 概念名称
   - 概念描述
   - 分类: indicator(技术指标)/strategy(策略)/term(术语)
   - 理解程度: 0.0-1.0

4. **交易策略** (strategy)
   - 策略名称
   - 策略描述
   - 具体规则
   - 适用条件

5. **常见问答** (qanda)
   - 问题
   - 答案
   - 分类

6. **市场洞察** (market-insight)
   - 主题
   - 洞察内容
   - 相关股票/板块

# 输出格式
请以 JSON 格式输出，不要包含任何其他内容：

{
  "knowledges": [
    {
      "type": "知识类型",
      "confidence": 0.0-1.0,
      "content": {具体内容对象}
    }
  ]
}

# 对话内容
%s

# 重要提示
- 只提取明确表达的信息，不要推测
- 如果某类知识不存在，不要创建
- confidence 表示 AI 对提取结果的置信度
- 所有时间使用 RFC3339 格式`

	maxExtractionRetries = 2
	extractionTimeout    = 30 * time.Second
)

// ==================== AI 知识提取 ====================

// ExtractKnowledgeWithAI 使用 AI 从会话中提取知识
func ExtractKnowledgeWithAI(session *model.Session) ([]*model.Knowledge, error) {
	ctx, cancel := context.WithTimeout(context.Background(), extractionTimeout)
	defer cancel()

	// 1. 构建对话文本
	conversationText := buildConversationText(session)

	if len(strings.TrimSpace(conversationText)) == 0 {
		return nil, fmt.Errorf("对话内容为空")
	}

	// 2. 构建 Prompt
	prompt := fmt.Sprintf(knowledgeExtractionPrompt, conversationText)

	// 3. 调用 LLM API
	response, err := callLLMForExtraction(ctx, prompt)
	if err != nil {
		log.Errorf("调用 LLM 失败: %v", err)
		return nil, fmt.Errorf("调用 LLM 失败: %w", err)
	}

	// 4. 解析 JSON 响应
	knowledges, err := parseExtractionResponse(response)
	if err != nil {
		log.Errorf("解析响应失败: %v", err)
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	log.Infof("AI 提取了 %d 个知识", len(knowledges))
	return knowledges, nil
}

// buildConversationText 构建对话文本（过滤系统消息）
func buildConversationText(session *model.Session) string {
	var lines []string

	for _, msg := range session.Messages {
		// 跳过系统消息
		if msg.Type == model.MemoryMessageTypeSystem {
			continue
		}

		role := "用户"
		if msg.Type == model.MemoryMessageTypeAssistant {
			role = "助手"
		}

		lines = append(lines, fmt.Sprintf("[%s]: %s", role, msg.Content))
	}

	return strings.Join(lines, "\n\n")
}

// callLLMForExtraction 调用 LLM 进行知识提取
func callLLMForExtraction(ctx context.Context, prompt string) (string, error) {
	cfg := config.GetLocalStoreConfig()
	if cfg == nil {
		return "", fmt.Errorf("配置为空")
	}
	if cfg.APIKey == "" {
		return "", fmt.Errorf("API Key 为空")
	}
	if cfg.BaseURL == "" {
		return "", fmt.Errorf("Base URL 为空")
	}
	if cfg.Model == "" {
		return "", fmt.Errorf("Model 为空")
	}

	// 创建 ChatModel
	maxTokens := 4000
	temperature := float32(0.3)
	modelConfig := &openai.ChatModelConfig{
		Model:       cfg.Model,
		APIKey:      cfg.APIKey,
		BaseURL:     cfg.BaseURL,
		MaxTokens:   &maxTokens,
		Temperature: &temperature, // 低温度以获得更稳定的输出
	}

	chatModel, err := openai.NewChatModel(ctx, modelConfig)
	if err != nil {
		log.Errorf("创建 ChatModel 失败: %v", err)
		return "", fmt.Errorf("创建 ChatModel 失败: %w", err)
	}

	// 构建消息
	messages := []*schema.Message{
		{
			Role:    schema.System,
			Content: "你是一个专业的投资顾问助手，擅长从对话中提取结构化知识。",
		},
		{
			Role:    schema.User,
			Content: prompt,
		},
	}

	// 调用 LLM（使用 Generate 而不是 Stream，因为我们需要完整响应）
	resp, err := chatModel.Generate(ctx, messages)
	if err != nil {
		log.Errorf("LLM 生成失败: %v", err)
		return "", fmt.Errorf("LLM 生成失败: %w", err)
	}

	if resp == nil {
		return "", fmt.Errorf("LLM 返回空响应")
	}

	return resp.Content, nil
}

// parseExtractionResponse 解析 LLM 响应
func parseExtractionResponse(response string) ([]*model.Knowledge, error) {
	// 清理响应（移除 markdown 代码块标记）
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	// 解析 JSON
	var result struct {
		Knowledges []struct {
			Type       string                 `json:"type"`
			Confidence float64                `json:"confidence"`
			Content    map[string]interface{} `json:"content"`
		} `json:"knowledges"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("JSON 解析失败: %w, 响应内容: %s", err, response)
	}

	// 转换为 Knowledge 对象
	var knowledges []*model.Knowledge
	now := time.Now()

	for i, item := range result.Knowledges {
		knowledgeType := model.KnowledgeType(item.Type)
		if !isValidKnowledgeType(knowledgeType) {
			log.Warnf("无效的知识类型: %s，跳过", item.Type)
			continue
		}

		// 生成唯一 ID
		id := fmt.Sprintf("%s-%d", knowledgeType, i)

		// 根据类型解析内容
		content, err := parseKnowledgeContent(knowledgeType, item.Content)
		if err != nil {
			log.Warnf("解析知识内容失败: %v，跳过", err)
			continue
		}

		knowledge := &model.Knowledge{
			ID:         id,
			Type:       knowledgeType,
			CreatedAt:  now,
			UpdatedAt:  now,
			Source:     "", // 将由调用者设置
			Confidence: item.Confidence,
			Content:    content,
		}

		knowledges = append(knowledges, knowledge)
	}

	return knowledges, nil
}

// isValidKnowledgeType 检查是否是有效的知识类型
func isValidKnowledgeType(kType model.KnowledgeType) bool {
	switch kType {
	case model.KnowledgeTypeUserProfile,
		model.KnowledgeTypeWatchlist,
		model.KnowledgeTypeConcept,
		model.KnowledgeTypeStrategy,
		model.KnowledgeTypeQAndA,
		model.KnowledgeTypeMarketInsight:
		return true
	default:
		return false
	}
}

// parseKnowledgeContent 解析知识内容
func parseKnowledgeContent(kType model.KnowledgeType, content map[string]interface{}) (interface{}, error) {
	data, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}

	switch kType {
	case model.KnowledgeTypeUserProfile:
		var profile model.UserProfile
		if err := json.Unmarshal(data, &profile); err != nil {
			return nil, err
		}
		// 设置默认值
		if profile.ValidFrom.IsZero() {
			profile.ValidFrom = time.Now()
		}
		profile.Source = "ai-extracted"
		return &profile, nil

	case model.KnowledgeTypeWatchlist:
		var list model.Watchlist
		if err := json.Unmarshal(data, &list); err != nil {
			return nil, err
		}
		if list.ValidFrom.IsZero() {
			list.ValidFrom = time.Now()
		}
		return &list, nil

	case model.KnowledgeTypeConcept:
		var concept model.ConceptLearned
		if err := json.Unmarshal(data, &concept); err != nil {
			return nil, err
		}
		if concept.LearnedAt.IsZero() {
			concept.LearnedAt = time.Now()
		}
		if concept.UpdatedAt.IsZero() {
			concept.UpdatedAt = time.Now()
		}
		return &concept, nil

	case model.KnowledgeTypeStrategy:
		var strategy model.TradingStrategy
		if err := json.Unmarshal(data, &strategy); err != nil {
			return nil, err
		}
		if strategy.CreatedAt.IsZero() {
			strategy.CreatedAt = time.Now()
		}
		if strategy.ValidFrom.IsZero() {
			strategy.ValidFrom = time.Now()
		}
		return &strategy, nil

	case model.KnowledgeTypeQAndA:
		var qa model.QAndA
		if err := json.Unmarshal(data, &qa); err != nil {
			return nil, err
		}
		if qa.CreatedAt.IsZero() {
			qa.CreatedAt = time.Now()
		}
		if qa.UpdatedAt.IsZero() {
			qa.UpdatedAt = time.Now()
		}
		return &qa, nil

	case model.KnowledgeTypeMarketInsight:
		var insight model.MarketInsight
		if err := json.Unmarshal(data, &insight); err != nil {
			return nil, err
		}
		if insight.CreatedAt.IsZero() {
			insight.CreatedAt = time.Now()
		}
		if insight.ValidFrom.IsZero() {
			insight.ValidFrom = time.Now()
		}
		return &insight, nil

	default:
		return nil, fmt.Errorf("未知的知识类型: %s", kType)
	}
}

// ==================== 降级策略（规则提取）====================

// ExtractKnowledgeWithFallback 使用降级策略提取知识
// 当 AI 提取完全失败时使用规则提取
func ExtractKnowledgeWithFallback(session *model.Session) ([]*model.Knowledge, error) {
	log.Info("使用降级策略提取知识")

	var knowledges []*model.Knowledge
	now := time.Now()

	// 规则 1: 提取股票代码作为关注列表
	symbols := extractStockCodes(session)
	if len(symbols) > 0 {
		var stockItems []model.StockItem
		for _, code := range symbols {
			stockItems = append(stockItems, model.StockItem{
				Code:    code,
				Name:    "", // 名称需要后续查找
				AddedAt: now,
				Reason:  "对话中提及",
			})
		}

		watchlist := &model.Watchlist{
			ID:         uuid.New().String(),
			Category:   "watching",
			Notes:      "从对话中提取",
			ValidFrom:  now,
			StockCodes: stockItems,
		}

		knowledges = append(knowledges, &model.Knowledge{
			ID:         uuid.New().String(),
			Type:       model.KnowledgeTypeWatchlist,
			CreatedAt:  now,
			UpdatedAt:  now,
			Source:     session.ID,
			Confidence: 0.3, // 低置信度
			Content:    watchlist,
		})
	}

	log.Infof("降级策略提取了 %d 个知识", len(knowledges))
	return knowledges, nil
}

// extractStockCodes 从会话中提取股票代码
func extractStockCodes(session *model.Session) []string {
	// 简单的正则匹配
	// 港股: 5 位数字
	// A 股: 6 位数字
	// 这里简化处理，实际应该使用更精确的正则

	codes := make(map[string]bool)

	for _, msg := range session.Messages {
		content := msg.Content

		// 查找类似 00700.HK, 600519 等模式
		words := strings.Fields(content)
		for _, word := range words {
			// 清理标点
			word = strings.Trim(word, ",.!?;:，。！？；：")
			// 检查是否是股票代码
			if isStockCode(word) {
				codes[word] = true
			}
		}
	}

	result := make([]string, 0, len(codes))
	for code := range codes {
		result = append(result, code)
	}

	return result
}

// isStockCode 检查是否是股票代码
func isStockCode(s string) bool {
	// 港股: 5 位数字 + .HK
	if strings.HasSuffix(strings.ToUpper(s), ".HK") {
		num := strings.TrimSuffix(strings.ToUpper(s), ".HK")
		if len(num) == 5 && isDigit(num) {
			return true
		}
	}

	// A 股: 6 位纯数字
	if len(s) == 6 && isDigit(s) {
		return true
	}

	return false
}

// isDigit 检查字符串是否全是数字
func isDigit(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
