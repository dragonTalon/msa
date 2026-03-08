package config

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"msa/pkg/model"
)

// EditorMsg 编辑器完成消息
type EditorMsg struct {
	Value string
	Done  bool
}

// ValidateTickMsg 验证定时器消息
type ValidateTickMsg time.Time

// ValidationSeverity 验证严重程度
type ValidationSeverity int

const (
	SeverityError ValidationSeverity = iota
	SeverityWarning
)

// ValidationError 验证错误
type ValidationError struct {
	Field    string
	Message  string
	Severity ValidationSeverity
}

func (e *ValidationError) Error() string {
	return e.Message
}

// EditorModel 编辑模型
type EditorModel struct {
	title     string
	label     string
	value     string
	input     textinput.Model
	isSecret  bool
	validator func(string) []*ValidationError
	errors    []*ValidationError
	quitting  bool
	aborted   bool
	width     int
	height    int
}

// NewEditorModel 创建新的编辑模型
func NewEditorModel(title, label, value string, isSecret bool) *EditorModel {
	input := textinput.New()
	input.Placeholder = "输入新值..."
	input.SetValue(value)
	input.Focus()

	if isSecret {
		input.EchoMode = textinput.EchoPassword
	}

	return &EditorModel{
		title:    title,
		label:    label,
		value:    value,
		input:    input,
		isSecret: isSecret,
		quitting: false,
		aborted:  false,
	}
}

// SetValidator 设置验证函数
func (m *EditorModel) SetValidator(validator func(string) []*ValidationError) {
	m.validator = validator
}

// Init bubbletea.Model 接口实现
func (m *EditorModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update bubbletea.Model 接口实现
func (m *EditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.aborted = true
			m.quitting = true
			return m, tea.Quit

		case tea.KeyEnter:
			// 执行验证
			if m.validator != nil {
				m.errors = m.validator(m.input.Value())
				// 如果只有警告，允许继续
				hasError := false
				for _, e := range m.errors {
					if e.Severity == SeverityError {
						hasError = true
						break
					}
				}
				if hasError {
					// 有错误，不退出
					return m, nil
				}
			}
			m.quitting = true
			return m, tea.Quit

		case tea.KeyCtrlD:
			// 强制保存（跳过验证）
			m.quitting = true
			return m, tea.Quit
		}

	case ValidateTickMsg:
		// 延迟验证
		if m.validator != nil {
			m.errors = m.validator(m.input.Value())
		}
		return m, nil
	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View bubbletea.Model 接口实现
func (m *EditorModel) View() string {
	if m.quitting {
		return ""
	}

	var s string

	// 标题
	s += TitleStyle.Render(m.title) + "\n\n"

	// 标签
	s += SectionTitleStyle.Render(m.label) + "\n\n"

	// 输入框
	s += InputStyle.Width(m.width-8).Render(m.input.View()) + "\n\n"

	// 验证错误/警告
	if len(m.errors) > 0 {
		for _, e := range m.errors {
			if e.Severity == SeverityError {
				s += ErrorStyle.Render("✗ "+e.Error()) + "\n"
			} else {
				s += WarningStyle.Render("⚠ "+e.Error()) + "\n"
			}
		}
		s += "\n"
	}

	// 验证通过
	if m.validator != nil && len(m.errors) == 0 && m.input.Value() != "" {
		s += SuccessStyle.Render("✓ 有效") + "\n\n"
	}

	// 帮助信息
	s += HelpStyle.Render("操作: [Enter] 确认 | [Esc] 取消 | [Ctrl+D] 强制保存")

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(1, 2).
		Render(s)
}

// GetValue 获取编辑后的值
func (m *EditorModel) GetValue() string {
	return m.input.Value()
}

// IsAborted 是否被中止
func (m *EditorModel) IsAborted() bool {
	return m.aborted
}

// IsDone 是否已完成
func (m *EditorModel) IsDone() bool {
	return m.quitting
}

// StartValidateTimer 启动验证定时器（延迟验证）
func StartValidateTimer(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return ValidateTickMsg(t)
	})
}

// 本地验证函数

// validateAPIKey 验证 API Key
func validateAPIKey(key, provider string) []*ValidationError {
	var errors []*ValidationError

	// 非空验证
	if len(key) == 0 {
		errors = append(errors, &ValidationError{
			Field:    "API Key",
			Message:  "API Key 不能为空",
			Severity: SeverityError,
		})
		return errors
	}

	// 前缀验证
	p := model.LlmProvider(provider)
	if err := p.ValidateAPIKey(key); err != nil {
		errors = append(errors, &ValidationError{
			Field:    "API Key",
			Message:  err.Error(),
			Severity: SeverityError,
		})
	}

	// 长度验证（警告）
	if len(key) < 10 {
		errors = append(errors, &ValidationError{
			Field:    "API Key",
			Message:  "API Key 长度可能不正确（少于 10 个字符）",
			Severity: SeverityWarning,
		})
	}

	return errors
}

// validateBaseURL 验证 Base URL
func validateBaseURL(urlStr string) []*ValidationError {
	var errors []*ValidationError

	// 空字符串是有效的
	if len(urlStr) == 0 {
		return nil
	}

	// 简单检查是否有 http 或 https 前缀
	if len(urlStr) < 7 {
		errors = append(errors, &ValidationError{
			Field:    "Base URL",
			Message:  "URL 格式不正确",
			Severity: SeverityError,
		})
		return errors
	}

	hasPrefix := false
	if len(urlStr) >= 7 && urlStr[:7] == "http://" {
		hasPrefix = true
	}
	if len(urlStr) >= 8 && urlStr[:8] == "https://" {
		hasPrefix = true
	}

	if !hasPrefix {
		errors = append(errors, &ValidationError{
			Field:    "Base URL",
			Message:  "URL 必须以 http:// 或 https:// 开头",
			Severity: SeverityError,
		})
	}

	return errors
}

// validateLogLevel 验证日志级别
func validateLogLevel(level string) []*ValidationError {
	var errors []*ValidationError

	validLevels := map[string]bool{
		"debug":   true,
		"info":    true,
		"warn":    true,
		"warning": true,
		"error":   true,
		"fatal":   true,
		"panic":   true,
	}

	if !validLevels[level] {
		errors = append(errors, &ValidationError{
			Field:    "日志级别",
			Message:  "无效的日志级别，支持：debug, info, warn, error, fatal, panic",
			Severity: SeverityError,
		})
	}

	return errors
}

// validateLogPath 验证日志路径
func validateLogPath(path string) []*ValidationError {
	// 简化验证，只检查是否为空
	if len(path) == 0 {
		return []*ValidationError{
			{
				Field:    "日志路径",
				Message:  "日志路径不能为空",
				Severity: SeverityWarning,
			},
		}
	}
	return nil
}
