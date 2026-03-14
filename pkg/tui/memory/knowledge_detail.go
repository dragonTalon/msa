package memory

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"msa/pkg/model"
)

// ==================== 知识详情视图 ====================

// KnowledgeDetailModel 知识详情模型
type KnowledgeDetailModel struct {
	knowledge    *model.Knowledge // 知识数据
	scrollOffset int              // 滚动偏移
	width        int              // 终端宽度
	height       int              // 终端高度
	quitting     bool             // 是否退出
	loading      bool             // 加载中
	err          error            // 错误
	parentModel  tea.Model        // 父模型（用于返回）
}

// NewKnowledgeDetailModel 创建知识详情模型
func NewKnowledgeDetailModel(knowledge *model.Knowledge) *KnowledgeDetailModel {
	return &KnowledgeDetailModel{
		knowledge:    knowledge,
		scrollOffset: 0,
		loading:      false,
	}
}

// SetParentModel 设置父模型
func (m *KnowledgeDetailModel) SetParentModel(parent tea.Model) {
	m.parentModel = parent
}

// Init 实现 tea.Model
func (m *KnowledgeDetailModel) Init() tea.Cmd {
	return nil
}

// Update 实现 tea.Model
func (m *KnowledgeDetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// handleKey 处理按键
func (m *KnowledgeDetailModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	visibleLines := m.height - 6
	if visibleLines < 5 {
		visibleLines = 5
	}
	totalLines := len(m.renderContentLines())

	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "esc", "q":
		// 返回父模型（知识列表）
		if m.parentModel != nil {
			return m.parentModel, nil
		}
		m.quitting = true
		return m, tea.Quit

	case "up", "k":
		if m.scrollOffset > 0 {
			m.scrollOffset--
		}

	case "down", "j":
		if m.scrollOffset < totalLines-visibleLines {
			m.scrollOffset++
		}

	case "pgup":
		m.scrollOffset -= visibleLines
		if m.scrollOffset < 0 {
			m.scrollOffset = 0
		}

	case "pgdown":
		m.scrollOffset += visibleLines
		if m.scrollOffset > totalLines-visibleLines {
			m.scrollOffset = totalLines - visibleLines
		}
		if m.scrollOffset < 0 {
			m.scrollOffset = 0
		}

	case "g":
		// 跳到开头
		m.scrollOffset = 0

	case "G":
		// 跳到结尾
		m.scrollOffset = totalLines - visibleLines
		if m.scrollOffset < 0 {
			m.scrollOffset = 0
		}
	}

	return m, nil
}

// View 实现 tea.Model
func (m *KnowledgeDetailModel) View() string {
	if m.quitting {
		return ""
	}

	if m.loading {
		return memoryTitleStyle.Render("🧠 知识详情") + "\n" +
			memoryEmptyStyle.Render("正在加载...")
	}

	if m.err != nil {
		return memoryTitleStyle.Render("🧠 知识详情") + "\n" +
			memoryEmptyStyle.Render(fmt.Sprintf("加载失败: %v", m.err))
	}

	if m.knowledge == nil {
		return memoryTitleStyle.Render("🧠 知识详情") + "\n" +
			memoryEmptyStyle.Render("知识不存在")
	}

	// 标题
	title := memoryTitleStyle.Render("🧠 知识详情")

	// 元数据
	metadata := m.renderMetadata()

	// 内容
	lines := m.renderContentLines()

	// 应用滚动偏移
	visibleLines := m.height - 6
	if visibleLines < 5 {
		visibleLines = 5
	}

	start := m.scrollOffset
	end := start + visibleLines
	if end > len(lines) {
		end = len(lines)
	}
	if start < 0 {
		start = 0
	}
	if start > len(lines) {
		start = len(lines)
	}

	var content string
	if start < end {
		content = strings.Join(lines[start:end], "\n")
	}

	// 位置信息
	posInfo := ""
	if len(lines) > 0 {
		pos := m.scrollOffset
		if pos >= len(lines) {
			pos = len(lines) - 1
		}
		posInfo = fmt.Sprintf(" 行 %d/%d", pos+1, len(lines))
	}

	// 帮助信息
	help := memoryHelpStyle.Render("ESC: 返回 | ↑/↓: 滚动 | PgUp/PgDn: 翻页 | g/G: 首/尾")

	// 组合
	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		metadata,
		"",
		content,
		posInfo,
		"",
		help,
	)
}

// renderMetadata 渲染知识元数据
func (m *KnowledgeDetailModel) renderMetadata() string {
	k := m.knowledge
	var parts []string

	// 类型图标和名称
	typeIcon := m.getTypeIcon(k.Type)
	typeName := m.getTypeName(k.Type)
	parts = append(parts, fmt.Sprintf("%s 类型: %s", typeIcon, typeName))

	// 时间
	timeStr := k.CreatedAt.Format("2006-01-02 15:04")
	parts = append(parts, fmt.Sprintf("📅 创建: %s", timeStr))

	// 来源
	if k.Source != "" {
		sourceID := k.Source
		if len(sourceID) > 8 {
			sourceID = sourceID[:8] + "..."
		}
		parts = append(parts, fmt.Sprintf("📚 来源: %s", sourceID))
	}

	// 置信度
	if k.Confidence > 0 {
		parts = append(parts, fmt.Sprintf("🎯 置信度: %.0f%%", k.Confidence*100))
	}

	return memoryDescStyle.Render(strings.Join(parts, " | "))
}

// renderContentLines 渲染内容为文本行
func (m *KnowledgeDetailModel) renderContentLines() []string {
	var lines []string
	k := m.knowledge

	// 根据类型渲染内容
	switch k.Type {
	case model.KnowledgeTypeUserProfile:
		lines = m.renderUserProfileContent()
	case model.KnowledgeTypeWatchlist:
		lines = m.renderWatchlistContent()
	case model.KnowledgeTypeConcept:
		lines = m.renderConceptContent()
	case model.KnowledgeTypeStrategy:
		lines = m.renderStrategyContent()
	case model.KnowledgeTypeQAndA:
		lines = m.renderQAndAContent()
	case model.KnowledgeTypeMarketInsight:
		lines = m.renderMarketInsightContent()
	default:
		lines = append(lines, memoryDescStyle.Render("未知类型"))
	}

	return lines
}

// renderUserProfileContent 渲染用户画像内容
func (m *KnowledgeDetailModel) renderUserProfileContent() []string {
	var lines []string

	lines = append(lines, "")
	lines = append(lines, memoryOptionStyle.Render("👤 用户画像"))

	if content, ok := m.knowledge.Content.(map[string]interface{}); ok {
		if risk := getStringFromMap(content, "riskPreference"); risk != "" {
			lines = append(lines, "")
			lines = append(lines, memoryDescStyle.Render("风险偏好"))
			lines = append(lines, fmt.Sprintf("  %s", risk))
		}

		if styles := getStringSliceFromMap(content, "investmentStyle"); len(styles) > 0 {
			lines = append(lines, "")
			lines = append(lines, memoryDescStyle.Render("投资风格"))
			for _, style := range styles {
				lines = append(lines, fmt.Sprintf("  • %s", style))
			}
		}

		if areas := getStringSliceFromMap(content, "focusAreas"); len(areas) > 0 {
			lines = append(lines, "")
			lines = append(lines, memoryDescStyle.Render("关注领域"))
			for _, area := range areas {
				lines = append(lines, fmt.Sprintf("  • %s", area))
			}
		}

		if level := getStringFromMap(content, "experienceLevel"); level != "" {
			lines = append(lines, "")
			lines = append(lines, memoryDescStyle.Render("经验水平"))
			lines = append(lines, fmt.Sprintf("  %s", level))
		}
	} else if profile, ok := m.knowledge.Content.(*model.UserProfile); ok {
		lines = append(lines, "")
		if profile.RiskPreference != "" {
			lines = append(lines, fmt.Sprintf("  风险偏好: %s", profile.RiskPreference))
		}
		if len(profile.InvestmentStyle) > 0 {
			lines = append(lines, fmt.Sprintf("  投资风格: %s", strings.Join(profile.InvestmentStyle, "、")))
		}
		if len(profile.FocusAreas) > 0 {
			lines = append(lines, fmt.Sprintf("  关注领域: %s", strings.Join(profile.FocusAreas, "、")))
		}
		if profile.ExperienceLevel != "" {
			lines = append(lines, fmt.Sprintf("  经验水平: %s", profile.ExperienceLevel))
		}
	}

	return lines
}

// renderWatchlistContent 渲染关注列表内容
func (m *KnowledgeDetailModel) renderWatchlistContent() []string {
	var lines []string

	lines = append(lines, "")
	lines = append(lines, memoryOptionStyle.Render("📊 关注列表"))

	if content, ok := m.knowledge.Content.(map[string]interface{}); ok {
		if stockCodes, ok := content["stockCodes"].([]interface{}); ok {
			lines = append(lines, "")
			for _, item := range stockCodes {
				if stockMap, ok := item.(map[string]interface{}); ok {
					code := getStringFromMap(stockMap, "code")
					reason := getStringFromMap(stockMap, "reason")
					if reason != "" {
						lines = append(lines, fmt.Sprintf("  • %s - %s", code, reason))
					} else {
						lines = append(lines, fmt.Sprintf("  • %s", code))
					}
				}
			}
		}
	} else if list, ok := m.knowledge.Content.(*model.Watchlist); ok {
		for _, item := range list.StockCodes {
			if item.Reason != "" {
				lines = append(lines, fmt.Sprintf("  • %s - %s", item.Code, item.Reason))
			} else {
				lines = append(lines, fmt.Sprintf("  • %s", item.Code))
			}
		}
	}

	return lines
}

// renderConceptContent 渲染概念内容
func (m *KnowledgeDetailModel) renderConceptContent() []string {
	var lines []string

	lines = append(lines, "")
	lines = append(lines, memoryOptionStyle.Render("📚 概念详情"))

	if content, ok := m.knowledge.Content.(map[string]interface{}); ok {
		if name := getStringFromMap(content, "name"); name != "" {
			lines = append(lines, "")
			lines = append(lines, memoryDescStyle.Render("名称"))
			lines = append(lines, fmt.Sprintf("  %s", name))
		}

		if desc := getStringFromMap(content, "description"); desc != "" {
			lines = append(lines, "")
			lines = append(lines, memoryDescStyle.Render("描述"))
			// 将描述按行分割
			for _, line := range strings.Split(desc, "\n") {
				lines = append(lines, fmt.Sprintf("  %s", line))
			}
		}

		if category := getStringFromMap(content, "category"); category != "" {
			lines = append(lines, "")
			lines = append(lines, memoryDescStyle.Render("分类"))
			lines = append(lines, fmt.Sprintf("  %s", category))
		}

		if confidence := getFloatFromMap(content, "confidence"); confidence > 0 {
			lines = append(lines, "")
			lines = append(lines, memoryDescStyle.Render("理解程度"))
			lines = append(lines, fmt.Sprintf("  %.0f%%", confidence*100))
		}
	} else if concept, ok := m.knowledge.Content.(*model.ConceptLearned); ok {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("  名称: %s", concept.Name))
		lines = append(lines, "")
		if concept.Description != "" {
			for _, line := range strings.Split(concept.Description, "\n") {
				lines = append(lines, fmt.Sprintf("  %s", line))
			}
		}
		if concept.Confidence > 0 {
			lines = append(lines, "")
			lines = append(lines, fmt.Sprintf("  理解程度: %.0f%%", concept.Confidence*100))
		}
	}

	return lines
}

// renderStrategyContent 渲染策略内容
func (m *KnowledgeDetailModel) renderStrategyContent() []string {
	var lines []string

	lines = append(lines, "")
	lines = append(lines, memoryOptionStyle.Render("📋 策略详情"))

	if content, ok := m.knowledge.Content.(map[string]interface{}); ok {
		if name := getStringFromMap(content, "name"); name != "" {
			lines = append(lines, "")
			lines = append(lines, memoryDescStyle.Render("名称"))
			lines = append(lines, fmt.Sprintf("  %s", name))
		}

		if desc := getStringFromMap(content, "description"); desc != "" {
			lines = append(lines, "")
			lines = append(lines, memoryDescStyle.Render("描述"))
			for _, line := range strings.Split(desc, "\n") {
				lines = append(lines, fmt.Sprintf("  %s", line))
			}
		}
	} else if strategy, ok := m.knowledge.Content.(*model.TradingStrategy); ok {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("  名称: %s", strategy.Name))
		lines = append(lines, "")
		if strategy.Description != "" {
			for _, line := range strings.Split(strategy.Description, "\n") {
				lines = append(lines, fmt.Sprintf("  %s", line))
			}
		}
	}

	return lines
}

// renderQAndAContent 渲染问答内容
func (m *KnowledgeDetailModel) renderQAndAContent() []string {
	var lines []string

	lines = append(lines, "")
	lines = append(lines, memoryOptionStyle.Render("❓ 常见问答"))

	if content, ok := m.knowledge.Content.(map[string]interface{}); ok {
		if question := getStringFromMap(content, "question"); question != "" {
			lines = append(lines, "")
			lines = append(lines, memoryDescStyle.Render("问题"))
			for _, line := range strings.Split(question, "\n") {
				lines = append(lines, fmt.Sprintf("  %s", line))
			}
		}

		if answer := getStringFromMap(content, "answer"); answer != "" {
			lines = append(lines, "")
			lines = append(lines, memoryDescStyle.Render("答案"))
			for _, line := range strings.Split(answer, "\n") {
				lines = append(lines, fmt.Sprintf("  %s", line))
			}
		}
	} else if qa, ok := m.knowledge.Content.(*model.QAndA); ok {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("  Q: %s", qa.Question))
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("  A: %s", qa.Answer))
	}

	return lines
}

// renderMarketInsightContent 渲染市场洞察内容
func (m *KnowledgeDetailModel) renderMarketInsightContent() []string {
	var lines []string

	lines = append(lines, "")
	lines = append(lines, memoryOptionStyle.Render("🔍 市场洞察"))

	if content, ok := m.knowledge.Content.(map[string]interface{}); ok {
		if topic := getStringFromMap(content, "topic"); topic != "" {
			lines = append(lines, "")
			lines = append(lines, memoryDescStyle.Render("主题"))
			lines = append(lines, fmt.Sprintf("  %s", topic))
		}

		if insight := getStringFromMap(content, "insight"); insight != "" {
			lines = append(lines, "")
			lines = append(lines, memoryDescStyle.Render("洞察"))
			for _, line := range strings.Split(insight, "\n") {
				lines = append(lines, fmt.Sprintf("  %s", line))
			}
		}
	} else if insightData, ok := m.knowledge.Content.(*model.MarketInsight); ok {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("  主题: %s", insightData.Topic))
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("  洞察: %s", insightData.Insight))
	}

	return lines
}

// 辅助函数
func (m *KnowledgeDetailModel) getTypeIcon(kType model.KnowledgeType) string {
	switch kType {
	case model.KnowledgeTypeUserProfile:
		return "👤"
	case model.KnowledgeTypeWatchlist:
		return "📊"
	case model.KnowledgeTypeConcept:
		return "📚"
	case model.KnowledgeTypeStrategy:
		return "📋"
	case model.KnowledgeTypeQAndA:
		return "❓"
	case model.KnowledgeTypeMarketInsight:
		return "🔍"
	default:
		return "📄"
	}
}

func (m *KnowledgeDetailModel) getTypeName(kType model.KnowledgeType) string {
	switch kType {
	case model.KnowledgeTypeUserProfile:
		return "用户画像"
	case model.KnowledgeTypeWatchlist:
		return "关注列表"
	case model.KnowledgeTypeConcept:
		return "学过的概念"
	case model.KnowledgeTypeStrategy:
		return "交易策略"
	case model.KnowledgeTypeQAndA:
		return "常见问答"
	case model.KnowledgeTypeMarketInsight:
		return "市场洞察"
	default:
		return "其他"
	}
}

// 从 map 获取值的辅助函数
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getStringSliceFromMap(m map[string]interface{}, key string) []string {
	if val, ok := m[key]; ok {
		if slice, ok := val.([]interface{}); ok {
			var result []string
			for _, item := range slice {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
	}
	return []string{}
}

func getFloatFromMap(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		}
	}
	return 0
}
