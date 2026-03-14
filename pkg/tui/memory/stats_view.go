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

// ==================== 统计视图 ====================

// StatsModel 统计视图模型
type StatsModel struct {
	stats       *model.MemoryStats // 统计数据
	width       int                // 终端宽度
	height      int                // 终端高度
	quitting    bool               // 是否退出
	loading     bool               // 加载中
	err         error              // 错误
	parentModel tea.Model          // 父模型（用于返回主页）
}

// NewStatsModel 创建统计视图模型
func NewStatsModel() *StatsModel {
	return &StatsModel{
		loading: true,
	}
}

// Init 实现 tea.Model
func (m *StatsModel) Init() tea.Cmd {
	return func() tea.Msg {
		return StatsLoadMsg{}
	}
}

// SetParentModel 设置父模型
func (m *StatsModel) SetParentModel(parent tea.Model) {
	m.parentModel = parent
}

// Update 实现 tea.Model
func (m *StatsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
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
		}

	case StatsLoadMsg:
		// 加载统计数据
		return m, m.loadStats()

	case StatsLoadedMsg:
		m.stats = msg.Stats
		m.loading = false
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View 实现 tea.Model
func (m *StatsModel) View() string {
	if m.quitting {
		return ""
	}

	title := memoryTitleStyle.Render("📊 统计信息")

	var content string
	if m.loading {
		content = memoryEmptyStyle.Render("正在加载...")
	} else if m.err != nil {
		content = memoryEmptyStyle.Render(fmt.Sprintf("加载失败: %v", m.err))
	} else if m.stats == nil {
		content = memoryEmptyStyle.Render("暂无统计数据")
	} else {
		content = m.renderStats()
	}

	help := memoryHelpStyle.Render("ESC/q: 返回聊天")

	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		"",
		content,
		help,
	)
}

// renderStats 渲染统计数据
func (m *StatsModel) renderStats() string {
	var sections []string

	// 概览统计
	sections = append(sections, m.renderOverview())

	// 会话统计
	sections = append(sections, m.renderSessionStats())

	// 知识统计
	sections = append(sections, m.renderKnowledgeStats())

	// TOP 5 股票
	sections = append(sections, m.renderTopStocks())

	// TOP 5 概念
	sections = append(sections, m.renderTopConcepts())

	// 时间统计
	sections = append(sections, m.renderTimeStats())

	return strings.Join(sections, "\n\n")
}

// renderOverview 渲染概览统计
func (m *StatsModel) renderOverview() string {
	var lines []string
	lines = append(lines, memoryOptionStyle.Render("📈 概览"))
	lines = append(lines, "")

	lines = append(lines, fmt.Sprintf("  总会话数: %d", m.stats.TotalSessions))
	lines = append(lines, fmt.Sprintf("  总消息数: %d", m.stats.TotalMessages))
	lines = append(lines, fmt.Sprintf("  总知识数: %d", m.stats.TotalKnowledge))

	return strings.Join(lines, "\n")
}

// renderSessionStats 渲染会话统计
func (m *StatsModel) renderSessionStats() string {
	var lines []string
	lines = append(lines, memoryOptionStyle.Render("💬 会话统计"))
	lines = append(lines, "")

	if !m.stats.EarliestSession.IsZero() {
		lines = append(lines, fmt.Sprintf("  最早会话: %s", m.stats.EarliestSession.Format("2006-01-02 15:04")))
	} else {
		lines = append(lines, "  最早会话: 无")
	}

	if !m.stats.LatestSession.IsZero() {
		lines = append(lines, fmt.Sprintf("  最新会话: %s", m.stats.LatestSession.Format("2006-01-02 15:04")))
	} else {
		lines = append(lines, "  最新会话: 无")
	}

	// 计算平均消息数
	avgMessages := 0.0
	if m.stats.TotalSessions > 0 {
		avgMessages = float64(m.stats.TotalMessages) / float64(m.stats.TotalSessions)
	}
	lines = append(lines, fmt.Sprintf("  平均消息/会话: %.1f", avgMessages))

	return strings.Join(lines, "\n")
}

// renderKnowledgeStats 渲染知识统计
func (m *StatsModel) renderKnowledgeStats() string {
	var lines []string
	lines = append(lines, memoryOptionStyle.Render("🧠 知识统计"))
	lines = append(lines, "")

	// 按类型统计
	if m.stats.KnowledgeByType != nil {
		typeIcons := map[model.KnowledgeType]string{
			model.KnowledgeTypeUserProfile:   "👤",
			model.KnowledgeTypeWatchlist:     "📊",
			model.KnowledgeTypeConcept:       "📚",
			model.KnowledgeTypeStrategy:      "📋",
			model.KnowledgeTypeQAndA:         "❓",
			model.KnowledgeTypeMarketInsight: "🔍",
		}

		typeNames := map[model.KnowledgeType]string{
			model.KnowledgeTypeUserProfile:   "用户画像",
			model.KnowledgeTypeWatchlist:     "关注列表",
			model.KnowledgeTypeConcept:       "学过的概念",
			model.KnowledgeTypeStrategy:      "交易策略",
			model.KnowledgeTypeQAndA:         "常见问答",
			model.KnowledgeTypeMarketInsight: "市场洞察",
		}

		for kType, count := range m.stats.KnowledgeByType {
			if count > 0 {
				icon := typeIcons[kType]
				name := typeNames[kType]
				lines = append(lines, fmt.Sprintf("  %s %s: %d 条", icon, name, count))
			}
		}
	}

	// 如果没有任何知识
	if len(lines) == 1 {
		lines = append(lines, "  暂无知识")
	}

	return strings.Join(lines, "\n")
}

// renderTopStocks 渲染最常查询的股票 TOP 5
func (m *StatsModel) renderTopStocks() string {
	var lines []string
	lines = append(lines, memoryOptionStyle.Render("📈 最常查询的股票 TOP 5"))
	lines = append(lines, "")

	if len(m.stats.TopStocks) == 0 {
		lines = append(lines, "  暂无数据")
		return strings.Join(lines, "\n")
	}

	for i, stock := range m.stats.TopStocks {
		medal := ""
		switch i {
		case 0:
			medal = "🥇"
		case 1:
			medal = "🥈"
		case 2:
			medal = "🥉"
		default:
			medal = fmt.Sprintf("%d.", i+1)
		}
		lines = append(lines, fmt.Sprintf("  %s %s - 查询 %d 次", medal, stock.Code, stock.QueryCount))
	}

	return strings.Join(lines, "\n")
}

// renderTopConcepts 渲染学得最多的概念 TOP 5
func (m *StatsModel) renderTopConcepts() string {
	var lines []string
	lines = append(lines, memoryOptionStyle.Render("📚 学得最多的概念 TOP 5"))
	lines = append(lines, "")

	if len(m.stats.TopConcepts) == 0 {
		lines = append(lines, "  暂无数据")
		return strings.Join(lines, "\n")
	}

	for i, concept := range m.stats.TopConcepts {
		medal := ""
		switch i {
		case 0:
			medal = "🥇"
		case 1:
			medal = "🥈"
		case 2:
			medal = "🥉"
		default:
			medal = fmt.Sprintf("%d.", i+1)
		}

		category := ""
		if concept.Category != "" {
			category = fmt.Sprintf(" [%s]", concept.Category)
		}
		lines = append(lines, fmt.Sprintf("  %s %s%s - 学习 %d 次", medal, concept.Name, category, concept.LearnCount))
	}

	return strings.Join(lines, "\n")
}

// renderTimeStats 渲染时间统计
func (m *StatsModel) renderTimeStats() string {
	var lines []string
	lines = append(lines, memoryOptionStyle.Render("⏰ 使用时长"))
	lines = append(lines, "")

	// 显示总时长
	duration := time.Since(m.stats.EarliestSession)
	if !m.stats.EarliestSession.IsZero() && m.stats.TotalSessions > 1 {
		days := int(duration.Hours() / 24)
		hours := int(duration.Hours()) % 24
		lines = append(lines, fmt.Sprintf("  记忆跨度: 约 %d 天 %d 小时", days, hours))
	} else {
		lines = append(lines, "  记忆跨度: 从首次使用开始")
	}

	return strings.Join(lines, "\n")
}

// loadStats 加载统计数据
func (m *StatsModel) loadStats() tea.Cmd {
	return func() tea.Msg {
		stats, err := memory.GetMemoryStats()
		if err != nil {
			m.err = err
			return StatsLoadedMsg{Stats: nil}
		}

		return StatsLoadedMsg{Stats: stats}
	}
}

// ==================== 消息类型 ====================

// StatsLoadMsg 加载统计消息
type StatsLoadMsg struct{}

// StatsLoadedMsg 统计加载完成消息
type StatsLoadedMsg struct {
	Stats *model.MemoryStats
}
