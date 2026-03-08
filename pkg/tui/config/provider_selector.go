package config

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"msa/pkg/logic/provider"
	"msa/pkg/model"
)

// ProviderSelectorMsg Provider 选择器消息
type ProviderSelectorMsg struct {
	Provider model.LlmProvider
	Aborted  bool
}

// ProviderSelectorModel Provider 选择器模型
type ProviderSelectorModel struct {
	providers []model.LlmProvider
	index     int
	quitting  bool
	aborted   bool
	width     int
	height    int
}

// NewProviderSelectorModel 创建新的 Provider 选择器模型
func NewProviderSelectorModel(currentProvider model.LlmProvider) *ProviderSelectorModel {
	providers := provider.GetRegisteredProviders()

	// 找到当前 provider 的索引
	index := 0
	for i, p := range providers {
		if p == currentProvider {
			index = i
			break
		}
	}

	return &ProviderSelectorModel{
		providers: providers,
		index:     index,
		quitting:  false,
		aborted:   false,
	}
}

// Init bubbletea.Model 接口实现
func (m *ProviderSelectorModel) Init() tea.Cmd {
	return nil
}

// Update bubbletea.Model 接口实现
func (m *ProviderSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.aborted = true
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.index > 0 {
				m.index--
			}
			return m, nil

		case "down", "j":
			if m.index < len(m.providers)-1 {
				m.index++
			}
			return m, nil

		case "enter":
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// View bubbletea.Model 接口实现
func (m *ProviderSelectorModel) View() string {
	if m.quitting {
		return ""
	}

	var s string

	// 标题
	s += TitleStyle.Render("选择 Provider") + "\n\n"

	// Provider 列表
	for i, p := range m.providers {
		info := model.ProviderRegistry[p]
		cursor := " "
		if i == m.index {
			cursor = CursorStyle.Render("▸")
		}

		// 当前选中的标记
		mark := ""
		if i == m.index {
			mark = " (当前选择)"
		}

		line := cursor + " " + info.DisplayName + mark + "\n"
		s += line

		// 描述
		s += DimStyle.Render("      "+info.Description) + "\n"

		// 默认 URL
		s += DimStyle.Render("      默认 URL: "+info.DefaultBaseURL) + "\n\n"
	}

	// 帮助信息
	s += "\n" + HelpStyle.Render("操作: [↑↓] 选择 | [Enter] 确认 | [Esc] 取消")

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(1, 2).
		Render(s)
}

// GetSelectedProvider 获取选中的 Provider
func (m *ProviderSelectorModel) GetSelectedProvider() model.LlmProvider {
	if m.index >= 0 && m.index < len(m.providers) {
		return m.providers[m.index]
	}
	return model.Siliconflow // 默认
}

// IsAborted 是否被中止
func (m *ProviderSelectorModel) IsAborted() bool {
	return m.aborted
}

// IsDone 是否已完成
func (m *ProviderSelectorModel) IsDone() bool {
	return m.quitting
}
