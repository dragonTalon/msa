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

// ==================== 主页视图 ====================

// MemoryHomeModel 记忆浏览器主页模型
type MemoryHomeModel struct {
	choices     []*model.SelectorItem // 选项列表
	choiceIDs   []string              // 选项ID列表
	selected    int                   // 当前选中的索引
	width       int                   // 终端宽度
	height      int                   // 终端高度
	quitting    bool                  // 是否退出
	parentModel tea.Model             // 父模型（Chat）用于返回
}

// NewMemoryHomeModel 创建主页模型
func NewMemoryHomeModel(choices []*model.SelectorItem) *MemoryHomeModel {
	// ID映射 - 按顺序匹配
	ids := []string{"sessions", "knowledge", "stats"}

	return &MemoryHomeModel{
		choices:     choices,
		choiceIDs:   ids,
		selected:    0,
		parentModel: nil, // 稍后设置
	}
}

// SetParentModel 设置父模型
func (m *MemoryHomeModel) SetParentModel(parent tea.Model) {
	m.parentModel = parent
}

// Init 实现 tea.Model
func (m *MemoryHomeModel) Init() tea.Cmd {
	return nil
}

// Update 实现 tea.Model
func (m *MemoryHomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "esc", "q":
			// 返回父模型（聊天界面）
			if m.parentModel != nil {
				m.quitting = true
				return m.parentModel, nil
			}
			// 如果没有父模型，则退出
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.choices)-1 {
				m.selected++
			}
		case "enter", " ":
			// 直接返回目标视图
			return m.navigateToTarget()
		case "1", "2", "3", "4":
			// 数字快捷键
			idx := int(msg.String()[0] - '1')
			if idx >= 0 && idx < len(m.choices) {
				m.selected = idx
				return m.navigateToTarget()
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View 实现 tea.Model
func (m *MemoryHomeModel) View() string {
	if m.quitting {
		return ""
	}

	// 标题
	title := memoryTitleStyle.Render("📚 记忆浏览器")

	// 选项列表
	var choices string
	for i, choice := range m.choices {
		cursor := " "
		if i == m.selected {
			cursor = "▶"
		}
		// 快捷键数字
		hotkey := i + 1
		if hotkey <= 9 {
			choices += cursor + memoryOptionStyle.Render(fmt.Sprintf(" [%d] %s", hotkey, choice.Name)) + "\n"
			choices += memoryDescStyle.Render("     "+choice.Description) + "\n\n"
		} else {
			choices += cursor + memoryOptionStyle.Render(fmt.Sprintf(" [ ] %s", choice.Name)) + "\n"
			choices += memoryDescStyle.Render("      "+choice.Description) + "\n\n"
		}
	}

	// 帮助信息
	help := memoryHelpStyle.Render("ESC/q: 返回聊天 | ↑/↓: 导航 | Enter: 选择")

	// 组合
	content := lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		"",
		choices,
		"",
		help,
	)

	// 居中
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// navigateToSelected 导航到选中的视图（返回命令）
func (m *MemoryHomeModel) navigateToSelected() tea.Cmd {
	target := m.choiceIDs[m.selected]
	return func() tea.Msg {
		return NavigationMsg{
			Target: target,
			Source: "home",
		}
	}
}

// navigateToTarget 导航到选中的视图（直接返回Model）
func (m *MemoryHomeModel) navigateToTarget() (tea.Model, tea.Cmd) {
	target := m.choiceIDs[m.selected]

	switch target {
	case "sessions":
		// 返回会话列表视图
		sessionList := GetSessionListModel()
		sessionList.SetParentModel(m) // 设置父模型
		return sessionList, sessionList.Init()
	case "knowledge":
		// 返回知识库列表视图
		knowledgeList := GetKnowledgeListModel()
		knowledgeList.SetParentModel(m)
		return knowledgeList, knowledgeList.Init()
	case "stats":
		// 返回统计视图
		statsView := GetStatsView()
		statsView.SetParentModel(m)
		return statsView, statsView.Init()
	default:
		return m, nil
	}
}

// ==================== 工厂函数 ====================

// GetMemoryHomeModel 获取记忆浏览器主页模型
func GetMemoryHomeModel() *MemoryHomeModel {
	// 检查记忆系统是否初始化
	manager := memory.GetManager()
	if manager == nil || !manager.IsInitialized() {
		fmt.Println("记忆系统未初始化")
	}

	// 获取统计数据
	sessionCount := 0
	knowledgeCount := 0

	if manager != nil {
		sessionCount = manager.GetSessionCount()
		knowledgeCount = manager.GetKnowledgeCount()
	}

	// 创建记忆浏览器的主页选项
	options := []*model.SelectorItem{
		{
			Name:        "📝 历史会话",
			Description: fmt.Sprintf("查看和管理历史聊天记录 (%d 个会话)", sessionCount),
		},
		{
			Name:        "🧠 知识库",
			Description: fmt.Sprintf("查看从对话中提取的知识 (%d 条知识)", knowledgeCount),
		},
		{
			Name:        "📊 统计信息",
			Description: "查看记忆系统的使用统计",
		},
	}

	return NewMemoryHomeModel(options)
}

// GetSessionListModel 获取会话列表模型
func GetSessionListModel() *SessionListModel {
	manager := memory.GetManager()
	if manager == nil {
		return NewSessionListModel()
	}

	// 从 Manager 加载会话列表
	sessions, err := manager.ListSessions(0, 50)
	if err != nil {
		fmt.Printf("加载会话列表失败: %v\n", err)
		return NewSessionListModel()
	}

	model := NewSessionListModel()
	model.SetSessions(sessions)
	return model
}

// ==================== 导航消息 ====================

// NavigationMsg 导航消息
type NavigationMsg struct {
	Target string      // 目标视图
	Source string      // 来源视图
	Data   interface{} // 附加数据
}

// ==================== 会话列表视图 ====================

// SessionListModel 会话列表模型
type SessionListModel struct {
	sessions         []model.SessionIndex // 会话列表（原始）
	filteredSessions []model.SessionIndex // 过滤后的会话列表
	selected         int                  // 当前选中的索引
	offset           int                  // 分页偏移
	limit            int                  // 每页数量
	width            int                  // 终端宽度
	height           int                  // 终端高度
	quitting         bool                 // 是否退出
	searchMode       bool                 // 搜索模式
	searchText       string               // 搜索文本
	parentModel      tea.Model            // 父模型（主页）用于返回
}

// NewSessionListModel 创建会话列表模型
func NewSessionListModel() *SessionListModel {
	return &SessionListModel{
		sessions: []model.SessionIndex{},
		selected: 0,
		offset:   0,
		limit:    10,
	}
}

// SetSessions 设置会话列表
func (m *SessionListModel) SetSessions(sessions []model.SessionIndex) {
	m.sessions = sessions
	m.filteredSessions = sessions // 初始时过滤列表等于原始列表
	if len(sessions) == 0 {
		m.selected = 0
	} else if m.selected >= len(sessions) {
		m.selected = len(sessions) - 1
	}
}

// SetParentModel 设置父模型
func (m *SessionListModel) SetParentModel(parent tea.Model) {
	m.parentModel = parent
}

// filterSessions 过滤会话列表
func (m *SessionListModel) filterSessions() {
	if m.searchText == "" {
		m.filteredSessions = m.sessions
	} else {
		var filtered []model.SessionIndex
		query := strings.ToLower(m.searchText)
		for _, session := range m.sessions {
			// 搜索摘要和 ID
			if strings.Contains(strings.ToLower(session.Summary), query) ||
				strings.Contains(strings.ToLower(session.ID), query) {
				filtered = append(filtered, session)
			}
		}
		m.filteredSessions = filtered
	}
	// 重置选中索引
	if len(m.filteredSessions) == 0 {
		m.selected = 0
	} else if m.selected >= len(m.filteredSessions) {
		m.selected = len(m.filteredSessions) - 1
	}
}

// getDisplaySessions 获取当前显示的会话列表
func (m *SessionListModel) getDisplaySessions() []model.SessionIndex {
	return m.filteredSessions
}

// Init 实现 tea.Model
func (m *SessionListModel) Init() tea.Cmd {
	return func() tea.Msg {
		// 返回当前会话列表（已在 SetSessions 中设置）
		return SessionsLoadedMsg{Sessions: m.sessions}
	}
}

// Update 实现 tea.Model
func (m *SessionListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.searchMode {
			return m.handleSearchKey(msg)
		}
		return m.handleNormalKey(msg)
	case SessionsLoadedMsg:
		// 会话加载完成
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// handleNormalKey 处理普通模式按键
func (m *SessionListModel) handleNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	displaySessions := m.getDisplaySessions()

	switch msg.String() {
	case "ctrl+c":
		// Ctrl+C 直接退出程序
		return m, tea.Quit
	case "esc", "q":
		// ESC/q 返回主页
		if m.parentModel != nil {
			return m.parentModel, nil
		}
		// 如果没有父模型，创建新的主页
		homeModel := GetMemoryHomeModel()
		return homeModel, homeModel.Init()
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}
	case "down", "j":
		if m.selected < len(displaySessions)-1 {
			m.selected++
		}
	case "pgup":
		m.selected -= 10
		if m.selected < 0 {
			m.selected = 0
		}
	case "pgdown":
		m.selected += 10
		if m.selected >= len(displaySessions) {
			m.selected = len(displaySessions) - 1
		}
	case "d":
		// 删除选中的会话
		if len(displaySessions) > 0 && m.selected < len(displaySessions) {
			sessionID := displaySessions[m.selected].ID
			// 从原始列表中移除
			var newSessions []model.SessionIndex
			for _, s := range m.sessions {
				if s.ID != sessionID {
					newSessions = append(newSessions, s)
				}
			}
			m.sessions = newSessions
			m.filterSessions() // 重新过滤
			return m, nil
		}
	case "enter", " ":
		// 查看会话详情
		if len(displaySessions) > 0 {
			sessionID := displaySessions[m.selected].ID
			sessionDetail := NewSessionDetailModel(sessionID)
			sessionDetail.SetParentModel(m)
			return sessionDetail, sessionDetail.Init()
		}
	case "/":
		m.searchMode = true
		m.searchText = ""
	}

	return m, nil
}

// handleSearchKey 处理搜索模式按键
func (m *SessionListModel) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.searchMode = false
		m.searchText = ""
		m.filterSessions() // 清空搜索时恢复原列表
	case "enter":
		m.searchMode = false
		// 搜索已经在输入时实时执行
		return m, nil
	default:
		// 处理可打印字符（包括中文）
		if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
			m.searchText += msg.String()
			m.filterSessions() // 实时过滤
		}
	case "backspace":
		if len(m.searchText) > 0 {
			m.searchText = m.searchText[:len(m.searchText)-1]
			m.filterSessions() // 实时过滤
		}
	}

	return m, nil
}

// View 实现 tea.Model
func (m *SessionListModel) View() string {
	if m.quitting {
		return ""
	}

	// 标题
	title := memoryTitleStyle.Render("📝 历史会话")

	// 搜索框
	searchBox := ""
	if m.searchMode {
		searchBox = memorySearchBoxStyle.Render("/" + m.searchText + "_")
	} else {
		searchBox = memorySearchBoxStyle.Render("/ 搜索...")
	}

	// 会话列表
	var list string
	displaySessions := m.getDisplaySessions()
	if len(displaySessions) == 0 {
		if m.searchText != "" {
			list = memoryEmptyStyle.Render("未找到匹配的会话")
		} else {
			list = memoryEmptyStyle.Render("暂无会话记录")
		}
	} else {
		for i, session := range displaySessions {
			cursor := " "
			if i == m.selected {
				cursor = "▶"
			}

			date := session.StartTime.Format("2006-01-02 15:04")
			item := fmt.Sprintf("%s %s | %d 条消息 | %s",
				cursor,
				date,
				session.MessageCount,
				session.Summary,
			)

			if i == m.selected {
				list += memorySelectedStyle.Render(item) + "\n"
			} else {
				list += item + "\n"
			}
		}
	}

	// 位置信息
	posInfo := ""
	if len(displaySessions) > 0 {
		posInfo = fmt.Sprintf(" %d/%d", m.selected+1, len(displaySessions))
	}

	// 帮助信息
	help := memoryHelpStyle.Render("ESC: 返回 | ↑/↓: 导航 | Enter: 查看 | /: 搜索")

	// 组合
	content := lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		searchBox,
		"",
		list,
		posInfo,
		"",
		help,
	)

	return content
}

// ==================== 消息类型 ====================

// LoadSessionsMsg 加载会话消息
type LoadSessionsMsg struct{}

// SessionsLoadedMsg 会话加载完成消息
type SessionsLoadedMsg struct {
	Sessions []model.SessionIndex
}

// DeleteSessionMsg 删除会话消息
type DeleteSessionMsg struct {
	SessionID string
	StartTime time.Time
}

// ==================== 样式 ====================

var (
	memoryTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				MarginBottom(1)

	memoryOptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("230")).
				Bold(true).
				MarginLeft(2)

	memoryDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginLeft(6).
			MarginBottom(1)

	memoryHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("242")).
			MarginTop(1)

	memorySearchBoxStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("226")).
				Bold(true)

	memorySelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("226")).
				Background(lipgloss.Color("235")).
				Padding(0, 1)

	memoryEmptyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Italic(true)
)

// ==================== 工厂函数 ====================

// GetKnowledgeListModel 获取知识库列表模型
func GetKnowledgeListModel() *KnowledgeListModel {
	return NewKnowledgeListModel()
}

// GetSearchView 获取搜索视图模型
func GetSearchView() *SearchViewModel {
	return NewSearchViewModel()
}

// GetStatsView 获取统计视图模型
func GetStatsView() *StatsModel {
	return NewStatsModel()
}

// GetSessionDetailModel 获取会话详情模型
func GetSessionDetailModel(sessionID string) *SessionDetailModel {
	return NewSessionDetailModel(sessionID)
}

// GetKnowledgeDetailModel 获取知识详情模型
func GetKnowledgeDetailModel(knowledge *model.Knowledge) *KnowledgeDetailModel {
	return NewKnowledgeDetailModel(knowledge)
}
