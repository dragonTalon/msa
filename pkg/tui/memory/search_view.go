package memory

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"msa/pkg/logic/memory"
	"msa/pkg/model"
)

// ==================== 搜索视图 ====================

// SearchViewModel 搜索视图模型
type SearchViewModel struct {
	searchInput  string             // 搜索输入
	results      []SearchResultItem // 搜索结果
	selected     int                // 当前选中的索引
	scrollOffset int                // 滚动偏移
	width        int                // 终端宽度
	height       int                // 终端高度
	quitting     bool               // 是否退出
	searchMode   bool               // 是否在搜索模式
	filterType   string             // 过滤类型 (all/session/knowledge)
	loading      bool               // 加载中
	err          error              // 错误
	history      []string           // 搜索历史
	historyIndex int                // 历史索引
	parentModel  tea.Model          // 父模型（用于返回主页）
}

// SearchResultItem 搜索结果项
type SearchResultItem struct {
	Type      string // "session" 或 "knowledge"
	ID        string
	Title     string
	Snippet   string
	Date      time.Time
	Relevance int // 相关性分数
}

// NewSearchViewModel 创建搜索视图模型
func NewSearchViewModel() *SearchViewModel {
	return &SearchViewModel{
		searchInput:  "",
		results:      []SearchResultItem{},
		selected:     0,
		scrollOffset: 0,
		searchMode:   true,
		filterType:   "all",
		loading:      false,
		history:      []string{},
		historyIndex: -1,
	}
}

// Init 实现 tea.Model
func (m *SearchViewModel) Init() tea.Cmd {
	// 初始化时自动执行空搜索，显示所有结果
	return func() tea.Msg {
		return SearchExecMsg{Query: "", FilterType: "all"}
	}
}

// SetParentModel 设置父模型
func (m *SearchViewModel) SetParentModel(parent tea.Model) {
	m.parentModel = parent
}

// Update 实现 tea.Model
func (m *SearchViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.searchMode {
			return m.handleSearchKey(msg)
		}
		return m.handleResultsKey(msg)

	case SearchExecMsg:
		// 执行搜索
		return m, m.performSearch(msg.Query, msg.FilterType)

	case SearchResultsMsg:
		m.results = msg.Results
		m.loading = false
		m.selected = 0
		m.scrollOffset = 0
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// handleSearchKey 处理搜索模式按键
func (m *SearchViewModel) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "esc":
		if m.searchInput == "" {
			// 返回父模型（主页）
			if m.parentModel != nil {
				return m.parentModel, nil
			}
			m.quitting = true
			return m, tea.Quit
		}
		m.searchInput = ""
		return m, nil

	case "enter":
		// 执行搜索（包括空搜索）
		m.searchMode = false
		m.loading = true
		// 只有非空搜索才添加到历史
		if m.searchInput != "" {
			m.history = append(m.history, m.searchInput)
			m.historyIndex = len(m.history)
		}
		return m, func() tea.Msg {
			return SearchExecMsg{Query: m.searchInput, FilterType: m.filterType}
		}

	case "backspace":
		if len(m.searchInput) > 0 {
			m.searchInput = m.searchInput[:len(m.searchInput)-1]
		}

	case "up":
		// 浏览历史
		if m.historyIndex > 0 {
			m.historyIndex--
			if m.historyIndex < len(m.history) {
				m.searchInput = m.history[m.historyIndex]
			}
		}

	case "down":
		// 浏览历史
		if m.historyIndex < len(m.history)-1 {
			m.historyIndex++
			m.searchInput = m.history[m.historyIndex]
		} else if m.historyIndex == len(m.history)-1 {
			m.historyIndex = len(m.history)
			m.searchInput = ""
		}

	case "tab":
		// 切换过滤类型
		types := []string{"all", "session", "knowledge"}
		for i, t := range types {
			if t == m.filterType {
				m.filterType = types[(i+1)%len(types)]
				break
			}
		}

	default:
		if len(msg.String()) == 1 {
			m.searchInput += msg.String()
		}
	}

	return m, nil
}

// handleResultsKey 处理结果模式按键
func (m *SearchViewModel) handleResultsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	visibleLines := m.height - 8

	switch msg.String() {
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
		// 返回搜索模式
		m.searchMode = true
		return m, nil

	case "/":
		// 返回搜索模式
		m.searchMode = true
		return m, nil

	case "up", "k":
		if m.selected > 0 {
			m.selected--
			if m.scrollOffset > m.selected {
				m.scrollOffset = m.selected
			}
		}

	case "down", "j":
		if m.selected < len(m.results)-1 {
			m.selected++
			if m.scrollOffset+visibleLines-1 < m.selected {
				m.scrollOffset = m.selected - visibleLines + 1
			}
		}

	case "pgup":
		m.scrollOffset -= visibleLines
		if m.scrollOffset < 0 {
			m.scrollOffset = 0
		}
		m.selected = m.scrollOffset

	case "pgdown":
		m.scrollOffset += visibleLines
		if m.scrollOffset > len(m.results)-visibleLines {
			m.scrollOffset = len(m.results) - visibleLines
		}
		if m.scrollOffset < 0 {
			m.scrollOffset = 0
		}
		m.selected = m.scrollOffset

	case "enter":
		// 打开结果
		if len(m.results) > 0 && m.selected < len(m.results) {
			result := m.results[m.selected]
			if result.Type == "session" {
				// 返回会话详情视图
				return m, func() tea.Msg {
					return NavigationMsg{
						Target: "session-detail",
						Source: "search",
						Data:   result.ID,
					}
				}
			}
		}
	}

	return m, nil
}

// View 实现 tea.Model
func (m *SearchViewModel) View() string {
	if m.quitting {
		return ""
	}

	title := memoryTitleStyle.Render("🔍 搜索记忆")

	// 搜索框
	var searchBox string
	if m.searchMode {
		cursor := "_"
		if time.Now().Unix()/500%2 == 0 {
			cursor = " "
		}
		searchBox = memorySearchBoxStyle.Render("/" + m.searchInput + cursor)

		// 过滤器提示
		filterHint := m.getFilterHint()
		searchBox += "\n" + memoryDescStyle.Render(filterHint)

		// 历史提示
		if len(m.history) > 0 && m.searchInput == "" {
			searchBox += "\n" + memoryDescStyle.Render("↑/↓: 浏览历史")
		}
	} else {
		searchBox = memorySearchBoxStyle.Render(fmt.Sprintf("/ %s", m.searchInput))
		if m.filterType != "all" {
			searchBox += memoryDescStyle.Render(fmt.Sprintf(" (过滤: %s)", m.filterType))
		}
	}

	// 搜索结果
	var results string
	if m.loading {
		results = memoryEmptyStyle.Render("正在搜索...")
	} else if m.searchMode {
		results = memoryDescStyle.Render("输入搜索词并按 Enter 搜索")
	} else if len(m.results) == 0 {
		results = memoryEmptyStyle.Render("未找到匹配结果")
	} else {
		// 显示结果
		visibleLines := m.height - 8
		start := m.scrollOffset
		end := start + visibleLines
		if end > len(m.results) {
			end = len(m.results)
		}

		for i := start; i < end; i++ {
			if i >= 0 && i < len(m.results) {
				result := m.results[i]
				cursor := " "
				if i == m.selected {
					cursor = "▶"
				}
				results += m.renderResultItem(result, cursor) + "\n\n"
			}
		}
	}

	// 位置信息
	posInfo := ""
	if len(m.results) > 0 && !m.searchMode {
		posInfo = fmt.Sprintf(" %d/%d", m.selected+1, len(m.results))
	}

	// 帮助信息
	var help string
	if m.searchMode {
		help = memoryHelpStyle.Render("直接输入搜索词 | Enter: 搜索 | Tab: 切换类型 | ESC/q: 返回")
	} else {
		help = memoryHelpStyle.Render("/: 重新搜索 | ↑/↓: 导航 | Enter: 查看 | ESC: 返回主页")
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		"",
		searchBox,
		"",
		results,
		posInfo,
		help,
	)

	return content
}

// renderResultItem 渲染搜索结果项
func (m *SearchViewModel) renderResultItem(result SearchResultItem, cursor string) string {
	var lines []string

	// 类型图标
	var icon string
	switch result.Type {
	case "session":
		icon = "💬"
	case "knowledge":
		icon = "🧠"
	default:
		icon = "📄"
	}

	// 标题行
	dateStr := result.Date.Format("2006-01-02 15:04")
	header := fmt.Sprintf("%s [%s] %s %s", cursor, dateStr, icon, result.Title)
	lines = append(lines, memoryOptionStyle.Render(header))

	// 摘要（高亮搜索词）
	snippet := result.Snippet
	if m.searchInput != "" {
		snippet = m.highlightSearchTerm(result.Snippet, m.searchInput)
	}

	// 截断过长的摘要
	if len(snippet) > 80 {
		snippet = snippet[:80] + "..."
	}
	lines = append(lines, "    "+memoryDescStyle.Render(snippet))

	return strings.Join(lines, "\n")
}

// highlightSearchTerm 高亮搜索词
func (m *SearchViewModel) highlightSearchTerm(text, term string) string {
	if term == "" {
		return text
	}

	lowerText := strings.ToLower(text)
	lowerTerm := strings.ToLower(term)

	start := 0
	for {
		idx := strings.Index(lowerText[start:], lowerTerm)
		if idx == -1 {
			break
		}

		idx += start
		before := text[:idx]
		match := text[idx : idx+len(term)]
		after := text[idx+len(term):]

		highlighted := lipgloss.NewStyle().
			Background(lipgloss.Color("226")).
			Foreground(lipgloss.Color("0")).
			Render(match)

		text = before + highlighted + after

		start = idx + len(match) + len(highlighted) - len(match)
		lowerText = strings.ToLower(text)
	}

	return text
}

// getFilterHint 获取过滤器提示
func (m *SearchViewModel) getFilterHint() string {
	hints := map[string]string{
		"all":       "过滤: 全部 (Tab 切换)",
		"session":   "过滤: 会话 (Tab 切换)",
		"knowledge": "过滤: 知识 (Tab 切换)",
	}
	return hints[m.filterType]
}

// ==================== 消息类型 ====================

// SearchExecMsg 执行搜索消息
type SearchExecMsg struct {
	Query      string
	FilterType string
}

// SearchResultsMsg 搜索结果消息
type SearchResultsMsg struct {
	Results []SearchResultItem
}

// ==================== 搜索执行 ====================

// performSearch 执行搜索
func (m *SearchViewModel) performSearch(query, filterType string) tea.Cmd {
	return func() tea.Msg {
		manager := memory.GetManager()
		if manager == nil {
			return SearchResultsMsg{
				Results: []SearchResultItem{},
			}
		}

		var results []SearchResultItem

		// 根据过滤类型搜索
		switch filterType {
		case "all", "":
			// 搜索全部
			sessions, _ := manager.ListSessions(0, 100)
			knowledges, _ := manager.GetAllKnowledge()

			// 搜索会话
			for _, session := range sessions {
				// 检查摘要、ID、标签和股票代码
				match := m.matches(session.Summary, query) || m.matches(session.ID, query)
				// 检查标签
				for _, tag := range session.Tags {
					if m.matches(tag, query) {
						match = true
						break
					}
				}
				// 检查股票代码
				for _, code := range session.StockCodes {
					if m.matches(code, query) {
						match = true
						break
					}
				}

				// 如果查询为空或匹配，则添加结果
				if query == "" || match {
					// 构建标题和摘要
					title := session.Summary
					if title == "" {
						title = fmt.Sprintf("会话 (%d 条消息)", session.MessageCount)
					}
					snippet := session.Summary
					if snippet == "" && len(session.StockCodes) > 0 {
						snippet = fmt.Sprintf("股票代码: %s", strings.Join(session.StockCodes, ", "))
					}
					if len(snippet) > 80 {
						snippet = snippet[:80] + "..."
					}

					results = append(results, SearchResultItem{
						Type:      "session",
						ID:        session.ID,
						Title:     title,
						Snippet:   snippet,
						Date:      session.StartTime,
						Relevance: 80,
					})
				}
			}

			// 搜索知识（空查询显示所有知识）
			for _, k := range knowledges {
				// 如果查询为空，或匹配内容/ID，则添加
				if query == "" || m.matches(m.knowledgeToString(k), query) || m.matches(k.ID, query) {
					content := m.knowledgeToString(k)
					snippet := content
					if len(snippet) > 80 {
						snippet = snippet[:80] + "..."
					}

					results = append(results, SearchResultItem{
						Type:      "knowledge",
						ID:        k.ID,
						Title:     m.getKnowledgeTitle(k),
						Snippet:   snippet,
						Date:      k.CreatedAt,
						Relevance: 80,
					})
				}
			}

		case "session":
			// 只搜索会话
			sessions, _ := manager.ListSessions(0, 100)
			for _, session := range sessions {
				// 检查摘要、ID、标签和股票代码
				match := m.matches(session.Summary, query) || m.matches(session.ID, query)
				// 检查标签
				for _, tag := range session.Tags {
					if m.matches(tag, query) {
						match = true
						break
					}
				}
				// 检查股票代码
				for _, code := range session.StockCodes {
					if m.matches(code, query) {
						match = true
						break
					}
				}

				// 如果查询为空或匹配，则添加结果
				if query == "" || match {
					// 构建标题和摘要
					title := session.Summary
					if title == "" {
						title = fmt.Sprintf("会话 (%d 条消息)", session.MessageCount)
					}
					snippet := session.Summary
					if snippet == "" && len(session.StockCodes) > 0 {
						snippet = fmt.Sprintf("股票代码: %s", strings.Join(session.StockCodes, ", "))
					}
					if len(snippet) > 80 {
						snippet = snippet[:80] + "..."
					}

					results = append(results, SearchResultItem{
						Type:      "session",
						ID:        session.ID,
						Title:     title,
						Snippet:   snippet,
						Date:      session.StartTime,
						Relevance: 80,
					})
				}
			}

		case "knowledge":
			// 只搜索知识（空查询显示所有知识）
			knowledges, _ := manager.GetAllKnowledge()
			for _, k := range knowledges {
				// 如果查询为空，或匹配内容/ID，则添加
				if query == "" || m.matches(m.knowledgeToString(k), query) || m.matches(k.ID, query) {
					content := m.knowledgeToString(k)
					snippet := content
					if len(snippet) > 80 {
						snippet = snippet[:80] + "..."
					}

					results = append(results, SearchResultItem{
						Type:      "knowledge",
						ID:        k.ID,
						Title:     m.getKnowledgeTitle(k),
						Snippet:   snippet,
						Date:      k.CreatedAt,
						Relevance: 80,
					})
				}
			}
		}

		return SearchResultsMsg{
			Results: results,
		}
	}
}

// matches 检查文本是否匹配查询
func (m *SearchViewModel) matches(text, query string) bool {
	return strings.Contains(
		strings.ToLower(text),
		strings.ToLower(query),
	)
}

// knowledgeToString 将知识转换为字符串
func (m *SearchViewModel) knowledgeToString(k *model.Knowledge) string {
	switch k.Type {
	case model.KnowledgeTypeUserProfile:
		if p, ok := k.Content.(*model.UserProfile); ok {
			return fmt.Sprintf("用户画像: %s %s", p.RiskPreference, strings.Join(p.InvestmentStyle, " "))
		}
	case model.KnowledgeTypeWatchlist:
		if w, ok := k.Content.(*model.Watchlist); ok {
			var codes []string
			for _, item := range w.StockCodes {
				codes = append(codes, item.Code)
			}
			return fmt.Sprintf("关注列表: %s", strings.Join(codes, ", "))
		}
	case model.KnowledgeTypeConcept:
		if c, ok := k.Content.(*model.ConceptLearned); ok {
			return fmt.Sprintf("概念: %s - %s", c.Name, c.Description)
		}
	case model.KnowledgeTypeStrategy:
		if s, ok := k.Content.(*model.TradingStrategy); ok {
			return fmt.Sprintf("策略: %s - %s", s.Name, s.Description)
		}
	case model.KnowledgeTypeQAndA:
		if q, ok := k.Content.(*model.QAndA); ok {
			return fmt.Sprintf("问答: %s - %s", q.Question, q.Answer)
		}
	case model.KnowledgeTypeMarketInsight:
		if i, ok := k.Content.(*model.MarketInsight); ok {
			return fmt.Sprintf("洞察: %s - %s", i.Topic, i.Insight)
		}
	}
	return "未知知识类型"
}

// getKnowledgeTitle 获取知识标题
func (m *SearchViewModel) getKnowledgeTitle(k *model.Knowledge) string {
	switch k.Type {
	case model.KnowledgeTypeUserProfile:
		return "用户画像"
	case model.KnowledgeTypeWatchlist:
		return "关注列表"
	case model.KnowledgeTypeConcept:
		if c, ok := k.Content.(*model.ConceptLearned); ok {
			return c.Name
		}
		return "概念"
	case model.KnowledgeTypeStrategy:
		if s, ok := k.Content.(*model.TradingStrategy); ok {
			return s.Name
		}
		return "策略"
	case model.KnowledgeTypeQAndA:
		return "问答"
	case model.KnowledgeTypeMarketInsight:
		if i, ok := k.Content.(*model.MarketInsight); ok {
			return i.Topic
		}
		return "洞察"
	default:
		return "知识"
	}
}
