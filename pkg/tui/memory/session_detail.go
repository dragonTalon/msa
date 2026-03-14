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

// ==================== 会话详情视图 ====================

// SessionDetailModel 会话详情模型
type SessionDetailModel struct {
	sessionID     string                // 会话ID
	session       *model.Session        // 完整会话数据
	messages      []model.MemoryMessage // 消息列表
	scrollOffset  int                   // 滚动偏移
	width         int                   // 终端宽度
	height        int                   // 终端高度
	quitting      bool                  // 是否退出
	searchMode    bool                  // 搜索模式
	searchQuery   string                // 搜索文本
	searchResults []int                 // 匹配的消息索引
	searchIndex   int                   // 当前搜索结果索引
	loading       bool                  // 加载中
	err           error                 // 错误
	parentModel   tea.Model             // 父模型（用于返回）
}

// NewSessionDetailModel 创建会话详情模型
func NewSessionDetailModel(sessionID string) *SessionDetailModel {
	return &SessionDetailModel{
		sessionID:     sessionID,
		session:       nil,
		scrollOffset:  0,
		searchMode:    false,
		searchQuery:   "",
		searchResults: []int{},
		searchIndex:   -1,
		loading:       true,
	}
}

// SetParentModel 设置父模型
func (m *SessionDetailModel) SetParentModel(parent tea.Model) {
	m.parentModel = parent
}

// Init 实现 tea.Model
func (m *SessionDetailModel) Init() tea.Cmd {
	// 异步加载会话数据
	return func() tea.Msg {
		manager := memory.GetManager()
		if manager == nil {
			return SessionLoadedMsg{Session: nil, Err: fmt.Errorf("记忆管理器未初始化")}
		}

		// 使用 ListSessions 获取会话索引，然后加载完整会话
		sessions, err := manager.ListSessions(0, 1000)
		if err != nil {
			return SessionLoadedMsg{Session: nil, Err: err}
		}

		// 找到匹配的会话
		for _, sessionIdx := range sessions {
			if sessionIdx.ID == m.sessionID {
				// 加载完整会话
				session, err := manager.GetStore().GetSession(m.sessionID, sessionIdx.StartTime)
				if err != nil {
					return SessionLoadedMsg{Session: nil, Err: err}
				}
				return SessionLoadedMsg{Session: session, Err: nil}
			}
		}

		return SessionLoadedMsg{Session: nil, Err: fmt.Errorf("会话不存在: %s", m.sessionID)}
	}
}

// Update 实现 tea.Model
func (m *SessionDetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.searchMode {
			return m.handleSearchKey(msg)
		}
		return m.handleNormalKey(msg)

	case SessionLoadedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.loading = false
		} else if msg.Session != nil {
			m.session = msg.Session
			m.messages = msg.Session.Messages
			m.loading = false
		} else {
			m.err = fmt.Errorf("会话不存在")
			m.loading = false
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// handleNormalKey 处理普通模式按键
func (m *SessionDetailModel) handleNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "esc", "q":
		// 返回父模型（会话列表）
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
		visibleLines := m.height - 6 // 减去标题和帮助的行数
		totalLines := len(m.renderMessageLines())
		if m.scrollOffset < totalLines-visibleLines {
			m.scrollOffset++
		}

	case "pgup":
		visibleLines := m.height - 6
		m.scrollOffset -= visibleLines
		if m.scrollOffset < 0 {
			m.scrollOffset = 0
		}

	case "pgdown":
		visibleLines := m.height - 6
		totalLines := len(m.renderMessageLines())
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
		visibleLines := m.height - 6
		totalLines := len(m.renderMessageLines())
		m.scrollOffset = totalLines - visibleLines
		if m.scrollOffset < 0 {
			m.scrollOffset = 0
		}

	case "/":
		// 进入搜索模式
		m.searchMode = true
		m.searchQuery = ""
		m.searchResults = []int{}
		m.searchIndex = -1

	case "n":
		// 下一个搜索结果
		if len(m.searchResults) > 0 {
			m.searchIndex = (m.searchIndex + 1) % len(m.searchResults)
			m.scrollToMessage(m.searchResults[m.searchIndex])
		}

	case "N":
		// 上一个搜索结果
		if len(m.searchResults) > 0 {
			m.searchIndex = (m.searchIndex - 1 + len(m.searchResults)) % len(m.searchResults)
			m.scrollToMessage(m.searchResults[m.searchIndex])
		}

	case "e":
		// 导出会话
		return m, nil
	}

	return m, nil
}

// handleSearchKey 处理搜索模式按键
func (m *SessionDetailModel) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.searchMode = false
		m.searchQuery = ""
		m.searchResults = []int{}
		m.searchIndex = -1

	case "enter":
		m.searchMode = false
		// 执行搜索
		if m.searchQuery != "" {
			m.searchResults = m.searchMessages(m.searchQuery)
			if len(m.searchResults) > 0 {
				m.searchIndex = 0
				m.scrollToMessage(m.searchResults[0])
			}
		}

	case "backspace":
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
		}

	default:
		// 处理可打印字符（包括中文）
		if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
			m.searchQuery += msg.String()
		}
	}

	return m, nil
}

// View 实现 tea.Model
func (m *SessionDetailModel) View() string {
	if m.quitting {
		return ""
	}

	if m.loading {
		return memoryTitleStyle.Render("📝 会话详情") + "\n" +
			memoryEmptyStyle.Render("正在加载...")
	}

	if m.err != nil {
		return memoryTitleStyle.Render("📝 会话详情") + "\n" +
			memoryEmptyStyle.Render(fmt.Sprintf("加载失败: %v", m.err))
	}

	if m.session == nil {
		return memoryTitleStyle.Render("📝 会话详情") + "\n" +
			memoryEmptyStyle.Render("会话不存在")
	}

	// 标题
	title := memoryTitleStyle.Render("📝 会话详情")

	// 会话元数据
	metadata := m.renderMetadata()

	// 搜索框
	var searchBox string
	if m.searchMode {
		searchBox = memorySearchBoxStyle.Render("/" + m.searchQuery + "_")
		if len(m.searchResults) > 0 {
			searchBox += memorySearchBoxStyle.Render(fmt.Sprintf(" [%d/%d]", m.searchIndex+1, len(m.searchResults)))
		}
	} else if len(m.searchResults) > 0 {
		searchBox = memorySearchBoxStyle.Render(fmt.Sprintf("搜索: %s [%d/%d] (n/N: 下/上, /: 新搜索)", m.searchQuery, m.searchIndex+1, len(m.searchResults)))
	}

	// 消息列表
	lines := m.renderMessageLines()

	// 应用滚动偏移
	visibleLines := m.height - 6 // 减去标题、元数据、搜索框、帮助
	if visibleLines < 5 {
		visibleLines = 5
	}

	// 裁剪可见区域
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

	var messages string
	if start < end {
		messages = strings.Join(lines[start:end], "\n")
	}

	// 位置信息
	posInfo := ""
	totalLines := len(lines)
	if totalLines > 0 {
		pos := m.scrollOffset
		if pos >= totalLines {
			pos = totalLines - 1
		}
		posInfo = fmt.Sprintf(" 行 %d/%d", pos+1, totalLines)
	}

	// 帮助信息
	help := memoryHelpStyle.Render("ESC: 返回 | ↑/↓: 滚动 | PgUp/PgDn: 翻页 | g/G: 首/尾 | /: 搜索 | e: 导出")

	// 组合
	content := lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		metadata,
		"",
		searchBox,
		"",
		messages,
		posInfo,
		"",
		help,
	)

	return content
}

// renderMetadata 渲染会话元数据
func (m *SessionDetailModel) renderMetadata() string {
	duration := m.session.EndTime.Sub(m.session.StartTime)
	var parts []string

	parts = append(parts, fmt.Sprintf("📅 时间: %s - %s",
		m.session.StartTime.Format("2006-01-02 15:04:05"),
		m.session.EndTime.Format("15:04:05")))

	parts = append(parts, fmt.Sprintf("⏱️ 时长: %s", formatDuration(duration)))

	parts = append(parts, fmt.Sprintf("💬 消息: %d 条", len(m.session.Messages)))

	if m.session.Model != "" {
		parts = append(parts, fmt.Sprintf("🤖 模型: %s", m.session.Model))
	}

	if len(m.session.Context.StockCodes) > 0 {
		parts = append(parts, fmt.Sprintf("📈 股票: %s", strings.Join(m.session.Context.StockCodes, ", ")))
	}

	if len(m.session.Tags) > 0 {
		parts = append(parts, fmt.Sprintf("🏷️ 标签: %s", strings.Join(m.session.Tags, ", ")))
	}

	return memoryDescStyle.Render(strings.Join(parts, " | "))
}

// renderMessageLines 渲染消息列表为文本行
func (m *SessionDetailModel) renderMessageLines() []string {
	var lines []string

	for _, msg := range m.messages {
		// 消息头部
		var roleIcon, roleName string
		switch msg.Type {
		case model.MemoryMessageTypeUser:
			roleIcon = "👤"
			roleName = "你"
		case model.MemoryMessageTypeAssistant:
			roleIcon = "🤖"
			roleName = "MSA"
		case model.MemoryMessageTypeSystem:
			roleIcon = "🔧"
			roleName = "系统"
		default:
			roleIcon = "💬"
			roleName = "消息"
		}

		// 消息时间
		timeStr := msg.Timestamp.Format("15:04")

		// 消息头部行
		header := fmt.Sprintf("[%s] %s %s:", timeStr, roleIcon, roleName)
		lines = append(lines, memoryOptionStyle.Render(header))

		// 消息内容（可能多行）
		contentLines := strings.Split(msg.Content, "\n")
		for _, line := range contentLines {
			// 高亮搜索匹配
			if m.searchQuery != "" && strings.Contains(strings.ToLower(line), strings.ToLower(m.searchQuery)) {
				line = m.highlightSearchTerm(line, m.searchQuery)
			}
			lines = append(lines, "    "+line)
		}

		// 消息间间隔
		lines = append(lines, "")
	}

	return lines
}

// highlightSearchTerm 高亮搜索词
func (m *SessionDetailModel) highlightSearchTerm(text, term string) string {
	lowerText := strings.ToLower(text)
	lowerTerm := strings.ToLower(term)

	start := 0
	for {
		idx := strings.Index(lowerText[start:], lowerTerm)
		if idx == -1 {
			break
		}

		idx += start
		// 替换为高亮版本
		before := text[:idx]
		match := text[idx : idx+len(term)]
		after := text[idx+len(term):]

		// 使用颜色高亮
		highlighted := lipgloss.NewStyle().
			Background(lipgloss.Color("226")).
			Foreground(lipgloss.Color("0")).
			Render(match)

		text = before + highlighted + after

		// 更新索引和文本
		start = idx + len(match) + len(highlighted) - len(match)
		lowerText = strings.ToLower(text)
	}

	return text
}

// searchMessages 搜索消息
func (m *SessionDetailModel) searchMessages(query string) []int {
	var results []int
	lowerQuery := strings.ToLower(query)

	for i, msg := range m.messages {
		if strings.Contains(strings.ToLower(msg.Content), lowerQuery) {
			results = append(results, i)
		}
	}

	return results
}

// scrollToMessage 滚动到指定消息
func (m *SessionDetailModel) scrollToMessage(msgIndex int) {
	// 计算消息所在行数
	lineCount := 0
	for i := 0; i < msgIndex; i++ {
		lineCount += len(strings.Split(m.messages[i].Content, "\n")) + 2 // 消息内容行 + 头部 + 间隔
	}

	m.scrollOffset = lineCount
}

// formatDuration 格式化时长
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d秒", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%d分钟", int(d.Minutes()))
	} else {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		return fmt.Sprintf("%d小时%d分钟", hours, minutes)
	}
}

// ==================== 消息类型 ====================

// SessionLoadMsg 加载会话消息
type SessionLoadMsg struct {
	SessionID string
}

// SessionLoadedMsg 会话加载完成消息
type SessionLoadedMsg struct {
	Session *model.Session
	Err     error
}
