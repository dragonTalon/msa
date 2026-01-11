package cmd

import (
	"fmt"
	"msa/pkg/model"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	log "github.com/sirupsen/logrus"
)

// SelectorView 选择器视图，包装 BaseSelector 并实现 tea.Model 接口
type SelectorView struct {
	selector *model.BaseSelector
	styles   *SelectorStyles
}

// NewSelectorView 创建新的选择器视图
func NewSelectorView(selector *model.BaseSelector) *SelectorView {
	return &SelectorView{
		selector: selector,
		styles:   NewSelectorStyles(),
	}
}

// Init 实现 tea.Model 接口
func (v *SelectorView) Init() tea.Cmd {
	return nil
}

// Update 实现 tea.Model 接口
func (v *SelectorView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m := v.selector

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		// 根据终端高度调整视口大小，预留标题和提示行
		m.ViewportSize = msg.Height - 8
		if m.ViewportSize < 5 {
			m.ViewportSize = 5
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return v, tea.Quit

		case "esc":
			// ESC: 如果有搜索内容，清空搜索；否则退出
			if m.SearchQuery != "" {
				m.SearchQuery = ""
				v.filterItems()
				m.Cursor = 0
				m.ViewportTop = 0
			} else {
				return v, tea.Quit
			}

		case "q":
			// q: 只有在没有搜索内容时才退出
			if m.SearchQuery == "" {
				return v, tea.Quit
			} else {
				// 否则作为普通字符输入
				m.SearchQuery += "q"
				v.filterItems()
				m.Cursor = 0
				m.ViewportTop = 0
			}

		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
				// 向上滚动视口
				if m.Cursor < m.ViewportTop {
					m.ViewportTop = m.Cursor
				}
			}

		case "backspace", "delete":
			// 删除搜索字符
			if len(m.SearchQuery) > 0 {
				m.SearchQuery = m.SearchQuery[:len(m.SearchQuery)-1]
				v.filterItems()
				m.Cursor = 0
				m.ViewportTop = 0
			}

		case "down", "j":
			if m.Cursor < len(m.FilteredItems)-1 {
				m.Cursor++
				// 向下滚动视口
				if m.Cursor >= m.ViewportTop+m.ViewportSize {
					m.ViewportTop = m.Cursor - m.ViewportSize + 1
				}
			}

		case "pgup": // Page Up
			m.Cursor -= m.ViewportSize
			if m.Cursor < 0 {
				m.Cursor = 0
			}
			m.ViewportTop = m.Cursor

		case "pgdown": // Page Down
			m.Cursor += m.ViewportSize
			if m.Cursor >= len(m.Items) {
				m.Cursor = len(m.Items) - 1
			}
			if m.Cursor >= m.ViewportTop+m.ViewportSize {
				m.ViewportTop = m.Cursor - m.ViewportSize + 1
			}

		case "home": // Home - 跳到第一个
			m.Cursor = 0
			m.ViewportTop = 0

		case "end": // End - 跳到最后一个
			m.Cursor = len(m.FilteredItems) - 1
			m.ViewportTop = m.Cursor - m.ViewportSize + 1
			if m.ViewportTop < 0 {
				m.ViewportTop = 0
			}

		case "enter":
			// 确认选择
			if len(m.FilteredItems) > 0 {
				m.Selected = m.FilteredItems[m.Cursor].Name
				m.Confirmed = true

				// 调用确认回调
				if m.OnConfirm != nil {
					err := m.OnConfirm(m.Selected)
					if err != nil {
						m.Err = err
						log.Errorf("确认回调失败: %v", err)
					}
				}
			}
			return v, tea.Quit

		default:
			// 其他字符作为搜索输入
			if len(msg.String()) == 1 {
				m.SearchQuery += msg.String()
				v.filterItems()
				m.Cursor = 0
				m.ViewportTop = 0
			}
		}
	}

	return v, nil
}

// filterItems 根据搜索关键字过滤项目
func (v *SelectorView) filterItems() {
	m := v.selector
	if m.SearchQuery == "" {
		// 没有搜索关键字，显示所有项
		m.FilteredItems = m.Items
		return
	}

	// 前缀匹配过滤
	query := strings.ToLower(m.SearchQuery)
	var filtered []*model.SelectorItem
	for _, item := range m.Items {
		// 检查名称或描述是否包含搜索关键字（前缀匹配）
		if strings.HasPrefix(strings.ToLower(item.Name), query) ||
			strings.HasPrefix(strings.ToLower(item.Description), query) {
			filtered = append(filtered, item)
		}
	}
	m.FilteredItems = filtered
}

// View 实现 tea.Model 接口
func (v *SelectorView) View() string {
	m := v.selector

	if m.Confirmed {
		return v.renderConfirmation()
	}

	var s string

	// 渲染标题
	s += v.renderHeader()

	// 渲染搜索框
	s += v.renderSearchBox()

	// 渲染位置信息
	s += v.renderPositionInfo()

	// 如果没有匹配结果
	if len(m.FilteredItems) == 0 {
		return s + v.renderEmptyResult()
	}

	// 渲染选择项列表
	s += v.renderItemList()

	// 渲染底部帮助信息
	s += v.renderFooter()

	return s
}

// renderConfirmation 渲染确认信息
func (v *SelectorView) renderConfirmation() string {
	m := v.selector
	if m.Err != nil {
		return v.styles.SystemMsg.Render(fmt.Sprintf("❌ 设置失败: %v\n", m.Err))
	}
	return v.styles.SystemMsg.Render(fmt.Sprintf("✅ 已选择: %s\n", m.Selected))
}

// renderHeader 渲染标题
func (v *SelectorView) renderHeader() string {
	var s string
	s += v.styles.Title.Render("🎯 模型选择器") + "\n"
	s += v.styles.Separator.Render(SeparatorLine) + "\n"
	return s
}

// renderSearchBox 渲染搜索框
func (v *SelectorView) renderSearchBox() string {
	m := v.selector
	var s string
	if m.SearchQuery != "" {
		s += v.styles.SearchBox.Render(fmt.Sprintf("🔍 搜索: %s_", m.SearchQuery)) + "\n"
	} else {
		s += v.styles.SearchPlaceholder.Render("🔍 搜索: (输入关键字进行前缀匹配...)") + "\n"
	}
	s += "\n"
	return s
}

// renderPositionInfo 渲染位置信息
func (v *SelectorView) renderPositionInfo() string {
	m := v.selector
	posInfo := fmt.Sprintf("📍 位置: %d/%d", m.Cursor+1, len(m.FilteredItems))

	if len(m.Items) != len(m.FilteredItems) {
		posInfo += fmt.Sprintf("  |  已过滤: %d/%d", len(m.FilteredItems), len(m.Items))
	}

	if len(m.FilteredItems) > m.ViewportSize {
		posInfo += fmt.Sprintf("  |  显示: %d-%d",
			m.ViewportTop+1,
			min(m.ViewportTop+m.ViewportSize, len(m.FilteredItems)))
	}

	return v.styles.Title.Render(posInfo) + "\n\n"
}

// renderEmptyResult 渲染空结果提示
func (v *SelectorView) renderEmptyResult() string {
	var s string
	s += v.styles.Error.Render("❌ 没有找到匹配的模型") + "\n\n"
	s += v.styles.SearchPlaceholder.Render("提示: 按 ESC 清空搜索，按 Ctrl+C 退出") + "\n"
	return s
}

// renderItemList 渲染选择项列表
func (v *SelectorView) renderItemList() string {
	m := v.selector
	var s string

	// 计算可见范围
	viewportEnd := m.ViewportTop + m.ViewportSize
	if viewportEnd > len(m.FilteredItems) {
		viewportEnd = len(m.FilteredItems)
	}

	// 上方滚动指示器
	if m.ViewportTop > 0 {
		s += v.styles.Scroll.Render(fmt.Sprintf("     ▲▲▲ 上方还有 %d 项 ▲▲▲", m.ViewportTop)) + "\n"
	}

	// 显示可见的选择项
	for i := m.ViewportTop; i < viewportEnd; i++ {
		s += v.renderItem(i) + "\n"
	}

	// 下方滚动指示器
	if viewportEnd < len(m.FilteredItems) {
		remaining := len(m.FilteredItems) - viewportEnd
		s += v.styles.Scroll.Render(fmt.Sprintf("     ▼▼▼ 下方还有 %d 项 ▼▼▼", remaining)) + "\n"
	}

	return s
}

// renderItem 渲染单个选择项
func (v *SelectorView) renderItem(index int) string {
	m := v.selector
	item := m.FilteredItems[index]

	// 行号
	lineNum := v.styles.LineNumber.Render(fmt.Sprintf("%d.", index+1))

	// 光标指示器
	var cursor string
	if m.Cursor == index {
		cursor = v.styles.Cursor.Render(CursorSymbol)
	} else {
		cursor = CursorEmpty
	}

	// 根据是否选中使用不同样式
	if m.Cursor == index {
		// 选中行：金色背景 + 黑色文字
		content := fmt.Sprintf("%s %s%s", lineNum, cursor, item.Name)
		if item.Description != "" {
			content += fmt.Sprintf(" - %s", item.Description)
		}
		return v.styles.Selected.Render(content)
	}

	// 普通行
	line := lineNum + " " + cursor + v.styles.Normal.Render(item.Name)
	if item.Description != "" {
		line += " " + v.styles.Description.Render("- "+item.Description)
	}
	return line
}

// renderFooter 渲染底部帮助信息
func (v *SelectorView) renderFooter() string {
	var s string
	s += "\n" + v.styles.Separator.Render(SeparatorLine) + "\n"
	s += v.styles.Help.Render("⌨️  输入:搜索  ↑/↓:移动  PgUp/PgDn:翻页  Enter:确认  ESC:清空搜索  Ctrl+C:退出")
	return s
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
