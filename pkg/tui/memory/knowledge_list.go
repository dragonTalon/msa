package memory

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"msa/pkg/logic/memory"
	"msa/pkg/model"
)

// ==================== 知识列表视图 ====================

// KnowledgeListModel 知识列表模型
type KnowledgeListModel struct {
	state             string                      // "types" 或 "list"
	knowledgeType     model.KnowledgeType         // 当前知识类型
	knowledgeTypes    []*model.Knowledge          // 知识列表（原始）
	filteredKnowledge []*model.Knowledge          // 过滤后的知识列表
	typeCounts        map[model.KnowledgeType]int // 各类型数量
	selected          int                         // 当前选中的索引
	scrollOffset      int                         // 滚动偏移
	width             int                         // 终端宽度
	height            int                         // 终端高度
	quitting          bool                        // 是否退出
	loading           bool                        // 加载中
	err               error                       // 错误
	parentModel       tea.Model                   // 父模型（用于返回主页）
	searchMode        bool                        // 搜索模式
	searchText        string                      // 搜索文本
}

// NewKnowledgeListModel 创建知识列表模型
func NewKnowledgeListModel() *KnowledgeListModel {
	return &KnowledgeListModel{
		state:             "types",
		knowledgeTypes:    []*model.Knowledge{},
		filteredKnowledge: []*model.Knowledge{},
		typeCounts:        make(map[model.KnowledgeType]int),
		selected:          0,
		scrollOffset:      0,
		loading:           true,
		searchMode:        false,
		searchText:        "",
	}
}

// Init 实现 tea.Model
func (m *KnowledgeListModel) Init() tea.Cmd {
	return func() tea.Msg {
		return KnowledgeLoadTypesMsg{}
	}
}

// SetParentModel 设置父模型
func (m *KnowledgeListModel) SetParentModel(parent tea.Model) {
	m.parentModel = parent
}

// Update 实现 tea.Model
func (m *KnowledgeListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.state == "types" {
			return m.handleTypesKey(msg)
		}
		return m.handleListKey(msg)

	case KnowledgeLoadTypesMsg:
		// 加载知识类型统计
		return m, m.loadKnowledgeTypes()

	case KnowledgeTypesLoadedMsg:
		m.typeCounts = msg.Counts
		m.loading = false
		return m, nil

	case KnowledgeLoadMsg:
		// 加载指定类型的知识
		return m, m.loadKnowledge(msg.Type)

	case KnowledgeLoadedMsg:
		m.knowledgeTypes = msg.Knowledge
		m.filteredKnowledge = msg.Knowledge
		m.state = "list"
		m.selected = 0
		m.scrollOffset = 0
		m.loading = false
		m.searchMode = false // 确保不在搜索模式
		m.searchText = ""    // 清空搜索文本
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// handleTypesKey 处理类型选择按键
func (m *KnowledgeListModel) handleTypesKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	types := m.getAllTypes()

	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "esc", "q":
		// 返回父模型（主页）
		if m.parentModel != nil {
			return m.parentModel, nil
		}
		m.quitting = true
		return m, tea.Quit

	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}

	case "down", "j":
		if m.selected < len(types)-1 {
			m.selected++
		}

	case "enter", " ":
		// 选择类型
		if m.selected < len(types) {
			m.knowledgeType = types[m.selected]
			m.loading = true
			return m, func() tea.Msg {
				return KnowledgeLoadMsg{Type: types[m.selected]}
			}
		}

	case "1", "2", "3", "4", "5", "6":
		// 数字快捷键
		idx := int(msg.String()[0] - '1')
		if idx >= 0 && idx < len(types) {
			m.selected = idx
			m.knowledgeType = types[idx]
			m.loading = true
			return m, func() tea.Msg {
				return KnowledgeLoadMsg{Type: types[idx]}
			}
		}
	}

	return m, nil
}

// handleListKey 处理列表按键
func (m *KnowledgeListModel) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 获取当前显示的列表
	displayList := m.getDisplayKnowledge()
	visibleLines := m.height - 8
	keyStr := msg.String()

	switch keyStr {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "q":
		// q 返回主页
		if m.parentModel != nil {
			return m.parentModel, nil
		}
		m.quitting = true
		return m, tea.Quit

	case "esc":
		// ESC 返回类型选择
		m.state = "types"
		m.selected = 0
		m.scrollOffset = 0
		m.searchText = ""
		m.searchMode = false
		m.filteredKnowledge = m.knowledgeTypes // 重置过滤
		return m, nil

	case "/":
		// 进入搜索模式
		m.searchMode = true
		return m, nil

	case "up", "k":
		if !m.searchMode && m.selected > 0 {
			m.selected--
		}

	case "down", "j":
		if !m.searchMode {
			displayList := m.getDisplayKnowledge()
			if m.selected < len(displayList)-1 {
				m.selected++
			}
		}

	case "pgup":
		if !m.searchMode {
			m.scrollOffset -= visibleLines
			if m.scrollOffset < 0 {
				m.scrollOffset = 0
			}
			m.selected = m.scrollOffset
		}
		return m, nil

	case "pgdown":
		if !m.searchMode {
			m.scrollOffset += visibleLines
			if m.scrollOffset > len(displayList)-visibleLines {
				m.scrollOffset = len(displayList) - visibleLines
			}
			if m.scrollOffset < 0 {
				m.scrollOffset = 0
			}
			m.selected = m.scrollOffset
		}
		return m, nil

	case "enter":
		if m.searchMode {
			// 退出搜索模式
			m.searchMode = false
			return m, nil
		} else {
			// 查看详情
			if len(displayList) > 0 && m.selected < len(displayList) {
				knowledge := displayList[m.selected]
				detailView := NewKnowledgeDetailModel(knowledge)
				detailView.SetParentModel(m)
				return detailView, detailView.Init()
			}
		}

	case "backspace":
		if m.searchMode && len(m.searchText) > 0 {
			m.searchText = m.searchText[:len(m.searchText)-1]
			m.filterKnowledge()
			m.selected = 0
			m.scrollOffset = 0
		}

	default:
		// 搜索模式下处理字符输入（包括中文）
		if m.searchMode && msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
			m.searchText += msg.String()
			m.filterKnowledge()
			m.selected = 0
			m.scrollOffset = 0
		}
	}

	return m, nil
}

// handleListSearchKey 处理列表搜索模式按键
// filterKnowledge 过滤知识列表
func (m *KnowledgeListModel) filterKnowledge() {
	if m.searchText == "" {
		m.filteredKnowledge = m.knowledgeTypes
	} else {
		var filtered []*model.Knowledge
		query := strings.ToLower(m.searchText)
		for _, k := range m.knowledgeTypes {
			if m.matchesKnowledge(k, query) {
				filtered = append(filtered, k)
			}
		}
		m.filteredKnowledge = filtered
	}

	// 重置选中索引
	if len(m.filteredKnowledge) == 0 {
		m.selected = 0
	} else if m.selected >= len(m.filteredKnowledge) {
		m.selected = len(m.filteredKnowledge) - 1
	}
}

// matchesKnowledge 检查知识是否匹配搜索查询
func (m *KnowledgeListModel) matchesKnowledge(k *model.Knowledge, query string) bool {
	// 搜索 ID
	if strings.Contains(strings.ToLower(k.ID), query) {
		return true
	}

	// 搜索来源
	if strings.Contains(strings.ToLower(k.Source), query) {
		return true
	}

	// 搜索内容
	if content, ok := k.Content.(map[string]interface{}); ok {
		// 搜索名称字段
		if name := getString(content, "name"); name != "" {
			if strings.Contains(strings.ToLower(name), query) {
				return true
			}
		}
		// 搜索描述字段
		if desc := getString(content, "description"); desc != "" {
			if strings.Contains(strings.ToLower(desc), query) {
				return true
			}
		}
		// 搜索主题字段
		if topic := getString(content, "topic"); topic != "" {
			if strings.Contains(strings.ToLower(topic), query) {
				return true
			}
		}
		// 搜索问题字段
		if question := getString(content, "question"); question != "" {
			if strings.Contains(strings.ToLower(question), query) {
				return true
			}
		}
	}

	return false
}

// getDisplayKnowledge 获取当前显示的知识列表
func (m *KnowledgeListModel) getDisplayKnowledge() []*model.Knowledge {
	return m.filteredKnowledge
}

// View 实现 tea.Model
func (m *KnowledgeListModel) View() string {
	if m.quitting {
		return ""
	}

	if m.loading {
		return memoryTitleStyle.Render("🧠 知识库") + "\n" +
			memoryEmptyStyle.Render("正在加载...")
	}

	if m.state == "types" {
		return m.renderTypes()
	}
	return m.renderList()
}

// renderTypes 渲染类型选择
func (m *KnowledgeListModel) renderTypes() string {
	title := memoryTitleStyle.Render("🧠 知识库")

	types := m.getAllTypes()
	var items string

	for i, kType := range types {
		cursor := " "
		if i == m.selected {
			cursor = "▶"
		}

		// 图标和名称
		icon := m.getTypeIcon(kType)
		name := m.getTypeName(kType)
		count := m.typeCounts[kType]

		// 快捷键数字
		hotkey := i + 1
		if hotkey <= 9 {
			items += cursor + memoryOptionStyle.Render(fmt.Sprintf(" [%d] %s %s", hotkey, icon, name))
		} else {
			items += cursor + memoryOptionStyle.Render(fmt.Sprintf(" [ ] %s %s", icon, name))
		}

		// 数量
		items += memoryDescStyle.Render(fmt.Sprintf(" (%d 条)", count))
		items += "\n\n"
	}

	help := memoryHelpStyle.Render("ESC/q: 返回聊天 | ↑/↓: 导航 | Enter: 选择")

	content := lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		"",
		items,
		help,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// renderList 渲染知识列表
func (m *KnowledgeListModel) renderList() string {
	title := memoryTitleStyle.Render(fmt.Sprintf("🧠 %s", m.getTypeName(m.knowledgeType)))

	// 返回提示
	backHint := memoryDescStyle.Render("ESC: 返回类型选择")

	// 搜索框
	var searchBox string
	if m.searchMode {
		cursor := "_"
		searchBox = memorySearchBoxStyle.Render("/" + m.searchText + cursor)
	} else {
		if m.searchText == "" {
			searchBox = memorySearchBoxStyle.Render("/ 搜索...")
		} else {
			searchBox = memorySearchBoxStyle.Render(fmt.Sprintf("/ %s", m.searchText))
		}
	}

	// 获取显示的列表
	displayList := m.getDisplayKnowledge()

	var items string
	if len(displayList) == 0 {
		if m.searchText != "" {
			items = memoryEmptyStyle.Render("未找到匹配的知识")
		} else {
			items = memoryEmptyStyle.Render("暂无知识")
		}
	} else {
		for i, k := range displayList {
			cursor := " "
			if i == m.selected {
				cursor = "▶"
			}

			// 渲染知识条目
			item := m.renderKnowledgeItem(k, cursor)
			items += item + "\n\n"
		}
	}

	// 位置信息
	posInfo := ""
	if len(displayList) > 0 {
		posInfo = fmt.Sprintf(" %d/%d", m.selected+1, len(displayList))
	}

	// 帮助信息
	var help string
	if m.searchMode {
		help = memoryHelpStyle.Render("ESC: 完成搜索 | Enter: 搜索 | 直接输入搜索词")
	} else {
		help = memoryHelpStyle.Render("/: 搜索 | ↑/↓: 导航 | PgUp/PgDn: 翻页 | Enter: 查看详情")
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		backHint,
		"",
		searchBox,
		"",
		items,
		posInfo,
		help,
	)

	return content
}

// renderKnowledgeItem 渲染单个知识项
func (m *KnowledgeListModel) renderKnowledgeItem(k *model.Knowledge, cursor string) string {
	var lines []string

	// 提取时间和来源
	timeStr := k.CreatedAt.Format("2006-01-02 15:04")
	sourceID := ""
	if len(k.Source) > 8 {
		sourceID = k.Source[:8] + "..."
	} else {
		sourceID = k.Source
	}

	header := fmt.Sprintf("%s [%s] 来源: %s", cursor, timeStr, sourceID)
	lines = append(lines, memoryOptionStyle.Render(header))

	// 根据类型渲染内容（处理 map[string]interface{} 类型）
	switch k.Type {
	case model.KnowledgeTypeUserProfile:
		if profile, ok := k.Content.(*model.UserProfile); ok {
			lines = append(lines, m.renderUserProfile(profile)...)
		} else if contentMap, ok := k.Content.(map[string]interface{}); ok {
			lines = append(lines, m.renderUserProfileFromMap(contentMap)...)
		}
	case model.KnowledgeTypeWatchlist:
		if list, ok := k.Content.(*model.Watchlist); ok {
			lines = append(lines, m.renderWatchlist(list)...)
		} else if contentMap, ok := k.Content.(map[string]interface{}); ok {
			lines = append(lines, m.renderWatchlistFromMap(contentMap)...)
		}
	case model.KnowledgeTypeConcept:
		if concept, ok := k.Content.(*model.ConceptLearned); ok {
			lines = append(lines, m.renderConcept(concept)...)
		} else if contentMap, ok := k.Content.(map[string]interface{}); ok {
			lines = append(lines, m.renderConceptFromMap(contentMap)...)
		}
	case model.KnowledgeTypeStrategy:
		if strategy, ok := k.Content.(*model.TradingStrategy); ok {
			lines = append(lines, m.renderStrategy(strategy)...)
		} else if contentMap, ok := k.Content.(map[string]interface{}); ok {
			lines = append(lines, m.renderStrategyFromMap(contentMap)...)
		}
	case model.KnowledgeTypeQAndA:
		if qa, ok := k.Content.(*model.QAndA); ok {
			lines = append(lines, m.renderQAndA(qa)...)
		} else if contentMap, ok := k.Content.(map[string]interface{}); ok {
			lines = append(lines, m.renderQAndAFromMap(contentMap)...)
		}
	case model.KnowledgeTypeMarketInsight:
		if insight, ok := k.Content.(*model.MarketInsight); ok {
			lines = append(lines, m.renderMarketInsight(insight)...)
		} else if contentMap, ok := k.Content.(map[string]interface{}); ok {
			lines = append(lines, m.renderMarketInsightFromMap(contentMap)...)
		}
	default:
		lines = append(lines, memoryDescStyle.Render("未知类型"))
	}

	return strings.Join(lines, "\n")
}

// renderUserProfile 渲染用户画像
func (m *KnowledgeListModel) renderUserProfile(profile *model.UserProfile) []string {
	var lines []string
	lines = append(lines, memoryDescStyle.Render("👤 用户画像"))

	if profile.RiskPreference != "" {
		lines = append(lines, fmt.Sprintf("   风险偏好: %s", profile.RiskPreference))
	}
	if len(profile.InvestmentStyle) > 0 {
		lines = append(lines, fmt.Sprintf("   投资风格: %s", strings.Join(profile.InvestmentStyle, "、")))
	}
	if len(profile.FocusAreas) > 0 {
		lines = append(lines, fmt.Sprintf("   关注领域: %s", strings.Join(profile.FocusAreas, "、")))
	}
	if profile.ExperienceLevel != "" {
		lines = append(lines, fmt.Sprintf("   经验水平: %s", profile.ExperienceLevel))
	}

	return lines
}

// renderWatchlist 渲染关注列表
func (m *KnowledgeListModel) renderWatchlist(list *model.Watchlist) []string {
	var lines []string
	lines = append(lines, memoryDescStyle.Render("📊 关注列表"))

	for _, item := range list.StockCodes {
		reason := ""
		if item.Reason != "" {
			reason = fmt.Sprintf(" (%s)", item.Reason)
		}
		lines = append(lines, fmt.Sprintf("   - %s%s", item.Code, reason))
	}

	return lines
}

// renderConcept 渲染概念
func (m *KnowledgeListModel) renderConcept(concept *model.ConceptLearned) []string {
	var lines []string
	lines = append(lines, memoryDescStyle.Render(fmt.Sprintf("📚 %s", concept.Name)))

	if concept.Description != "" {
		// 截断过长的描述
		desc := concept.Description
		if len(desc) > 100 {
			desc = desc[:100] + "..."
		}
		lines = append(lines, fmt.Sprintf("   %s", desc))
	}

	if concept.Confidence > 0 {
		lines = append(lines, fmt.Sprintf("   理解程度: %.0f%%", concept.Confidence*100))
	}

	return lines
}

// renderStrategy 渲染交易策略
func (m *KnowledgeListModel) renderStrategy(strategy *model.TradingStrategy) []string {
	var lines []string
	lines = append(lines, memoryDescStyle.Render(fmt.Sprintf("📋 %s", strategy.Name)))

	if strategy.Description != "" {
		desc := strategy.Description
		if len(desc) > 100 {
			desc = desc[:100] + "..."
		}
		lines = append(lines, fmt.Sprintf("   %s", desc))
	}

	return lines
}

// renderQAndA 渲染问答
func (m *KnowledgeListModel) renderQAndA(qa *model.QAndA) []string {
	var lines []string
	lines = append(lines, memoryDescStyle.Render("❓ 常见问答"))

	if qa.Question != "" {
		q := qa.Question
		if len(q) > 80 {
			q = q[:80] + "..."
		}
		lines = append(lines, fmt.Sprintf("   Q: %s", q))
	}

	if qa.Answer != "" {
		a := qa.Answer
		if len(a) > 80 {
			a = a[:80] + "..."
		}
		lines = append(lines, fmt.Sprintf("   A: %s", a))
	}

	return lines
}

// renderMarketInsight 渲染市场洞察
func (m *KnowledgeListModel) renderMarketInsight(insight *model.MarketInsight) []string {
	var lines []string

	// 使用 Topic 作为标题
	if insight.Topic != "" {
		lines = append(lines, memoryDescStyle.Render(fmt.Sprintf("🔍 %s", insight.Topic)))
	} else {
		lines = append(lines, memoryDescStyle.Render("🔍 市场洞察"))
	}

	// 显示洞察内容
	if insight.Insight != "" {
		content := insight.Insight
		if len(content) > 100 {
			content = content[:100] + "..."
		}
		lines = append(lines, fmt.Sprintf("   %s", content))
	}

	return lines
}

// getAllTypes 获取所有知识类型
func (m *KnowledgeListModel) getAllTypes() []model.KnowledgeType {
	return []model.KnowledgeType{
		model.KnowledgeTypeUserProfile,
		model.KnowledgeTypeWatchlist,
		model.KnowledgeTypeConcept,
		model.KnowledgeTypeStrategy,
		model.KnowledgeTypeQAndA,
		model.KnowledgeTypeMarketInsight,
	}
}

// getTypeIcon 获取类型图标
func (m *KnowledgeListModel) getTypeIcon(kType model.KnowledgeType) string {
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

// getTypeName 获取类型名称
func (m *KnowledgeListModel) getTypeName(kType model.KnowledgeType) string {
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

// loadKnowledgeTypes 加载知识类型统计
func (m *KnowledgeListModel) loadKnowledgeTypes() tea.Cmd {
	return func() tea.Msg {
		stats, err := memory.GetMemoryStats()
		if err != nil {
			return KnowledgeTypesLoadedMsg{
				Counts: map[model.KnowledgeType]int{},
			}
		}

		// 使用统计中的 KnowledgeByType
		return KnowledgeTypesLoadedMsg{
			Counts: stats.KnowledgeByType,
		}
	}
}

// loadKnowledge 加载指定类型的知识
func (m *KnowledgeListModel) loadKnowledge(kType model.KnowledgeType) tea.Cmd {
	return func() tea.Msg {
		manager := memory.GetManager()
		if manager == nil {
			return KnowledgeLoadedMsg{
				Knowledge: []*model.Knowledge{},
			}
		}

		// 获取所有知识并过滤类型
		allKnowledge, err := manager.GetAllKnowledge()
		if err != nil {
			return KnowledgeLoadedMsg{
				Knowledge: []*model.Knowledge{},
			}
		}

		// 过滤指定类型
		var filtered []*model.Knowledge
		for _, k := range allKnowledge {
			if k.Type == kType {
				filtered = append(filtered, k)
			}
		}

		return KnowledgeLoadedMsg{
			Knowledge: filtered,
		}
	}
}

// ==================== 消息类型 ====================

// KnowledgeLoadTypesMsg 加载知识类型消息
type KnowledgeLoadTypesMsg struct{}

// KnowledgeTypesLoadedMsg 知识类型加载完成消息
type KnowledgeTypesLoadedMsg struct {
	Counts map[model.KnowledgeType]int
}

// KnowledgeLoadMsg 加载知识消息
type KnowledgeLoadMsg struct {
	Type model.KnowledgeType
}

// KnowledgeLoadedMsg 知识加载完成消息
type KnowledgeLoadedMsg struct {
	Knowledge []*model.Knowledge
}

// ==================== 从 Map 渲染的辅助方法 ====================
// 用于处理 JSON 反序列化后的 map[string]interface{} 类型

func (m *KnowledgeListModel) renderConceptFromMap(content map[string]interface{}) []string {
	var lines []string
	name := getString(content, "name")
	if name == "" {
		name = "概念"
	}
	lines = append(lines, memoryDescStyle.Render(fmt.Sprintf("📚 %s", name)))

	if desc := getString(content, "description"); desc != "" {
		if len(desc) > 100 {
			desc = desc[:100] + "..."
		}
		lines = append(lines, fmt.Sprintf("   %s", desc))
	}

	if confidence := getFloat(content, "confidence"); confidence > 0 {
		lines = append(lines, fmt.Sprintf("   理解程度: %.0f%%", confidence*100))
	}

	return lines
}

func (m *KnowledgeListModel) renderStrategyFromMap(content map[string]interface{}) []string {
	var lines []string
	name := getString(content, "name")
	if name == "" {
		name = "策略"
	}
	lines = append(lines, memoryDescStyle.Render(fmt.Sprintf("📋 %s", name)))

	if desc := getString(content, "description"); desc != "" {
		if len(desc) > 100 {
			desc = desc[:100] + "..."
		}
		lines = append(lines, fmt.Sprintf("   %s", desc))
	}

	return lines
}

func (m *KnowledgeListModel) renderQAndAFromMap(content map[string]interface{}) []string {
	var lines []string
	lines = append(lines, memoryDescStyle.Render("❓ 常见问答"))

	if question := getString(content, "question"); question != "" {
		q := question
		if len(q) > 80 {
			q = q[:80] + "..."
		}
		lines = append(lines, fmt.Sprintf("   Q: %s", q))
	}

	if answer := getString(content, "answer"); answer != "" {
		a := answer
		if len(a) > 80 {
			a = a[:80] + "..."
		}
		lines = append(lines, fmt.Sprintf("   A: %s", a))
	}

	return lines
}

func (m *KnowledgeListModel) renderMarketInsightFromMap(content map[string]interface{}) []string {
	var lines []string

	topic := getString(content, "topic")
	if topic == "" {
		topic = "市场洞察"
	}
	lines = append(lines, memoryDescStyle.Render(fmt.Sprintf("🔍 %s", topic)))

	if insight := getString(content, "insight"); insight != "" {
		content := insight
		if len(content) > 100 {
			content = content[:100] + "..."
		}
		lines = append(lines, fmt.Sprintf("   %s", content))
	}

	return lines
}

func (m *KnowledgeListModel) renderUserProfileFromMap(content map[string]interface{}) []string {
	var lines []string
	lines = append(lines, memoryDescStyle.Render("👤 用户画像"))

	if risk := getString(content, "riskPreference"); risk != "" {
		lines = append(lines, fmt.Sprintf("   风险偏好: %s", risk))
	}
	if styles := getStringSlice(content, "investmentStyle"); len(styles) > 0 {
		lines = append(lines, fmt.Sprintf("   投资风格: %s", strings.Join(styles, "、")))
	}
	if areas := getStringSlice(content, "focusAreas"); len(areas) > 0 {
		lines = append(lines, fmt.Sprintf("   关注领域: %s", strings.Join(areas, "、")))
	}
	if level := getString(content, "experienceLevel"); level != "" {
		lines = append(lines, fmt.Sprintf("   经验水平: %s", level))
	}

	return lines
}

func (m *KnowledgeListModel) renderWatchlistFromMap(content map[string]interface{}) []string {
	var lines []string
	lines = append(lines, memoryDescStyle.Render("📊 关注列表"))

	if stockCodes, ok := content["stockCodes"].([]interface{}); ok {
		for _, item := range stockCodes {
			if stockMap, ok := item.(map[string]interface{}); ok {
				code := getString(stockMap, "code")
				reason := getString(stockMap, "reason")
				if reason != "" {
					lines = append(lines, fmt.Sprintf("   - %s (%s)", code, reason))
				} else {
					lines = append(lines, fmt.Sprintf("   - %s", code))
				}
			}
		}
	}

	return lines
}

// 辅助函数：从 map 中获取字符串值
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// 辅助函数：从 map 中获取字符串切片
func getStringSlice(m map[string]interface{}, key string) []string {
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

// 辅助函数：从 map 中获取浮点数值
func getFloat(m map[string]interface{}, key string) float64 {
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
