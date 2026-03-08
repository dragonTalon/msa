package config

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	log "github.com/sirupsen/logrus"
	"msa/pkg/config"
	"msa/pkg/model"
)

// ConfigItem 配置项
type ConfigItem struct {
	Key         string
	Label       string
	Value       string
	DisplayFunc func(string) string
	IsEditable  bool
	IsSecret    bool
}

// ConfigState 配置状态
type ConfigState struct {
	Items          []*ConfigItem
	SelectedIndex  int
	ErrorMessage   string
	SuccessMessage string
	EditMode       bool
	EditorModel    interface{} // 可以是 EditorModel 或 ProviderSelectorModel
}

// ConfigModel TUI 配置主模型
type ConfigModel struct {
	state    *ConfigState
	quitting bool
	width    int
	height   int
}

// NewConfigModel 创建新的配置模型
func NewConfigModel() *ConfigModel {
	cfg := config.GetLocalStoreConfig()

	state := &ConfigState{
		Items: []*ConfigItem{
			{
				Key:         "provider",
				Label:       "Provider",
				Value:       string(cfg.Provider),
				DisplayFunc: func(v string) string { return model.LlmProvider(v).GetDisplayName() },
				IsEditable:  true,
			},
			{
				Key:         "apikey",
				Label:       "API Key",
				Value:       cfg.APIKey,
				DisplayFunc: MaskAPIKey,
				IsEditable:  true,
				IsSecret:    true,
			},
			{
				Key:         "baseurl",
				Label:       "Base URL",
				Value:       cfg.BaseURL,
				DisplayFunc: func(v string) string { return v },
				IsEditable:  true,
			},
			{
				Key:         "model",
				Label:       "Model",
				Value:       cfg.Model,
				DisplayFunc: func(v string) string { return v },
				IsEditable:  true,
			},
			{
				Key:         "loglevel",
				Label:       "日志级别",
				Value:       getLogLevel(cfg),
				DisplayFunc: func(v string) string { return v },
				IsEditable:  true,
			},
			{
				Key:         "logfile",
				Label:       "日志文件",
				Value:       getLogFile(cfg),
				DisplayFunc: func(v string) string { return v },
				IsEditable:  true,
			},
		},
		SelectedIndex: 0,
	}

	return &ConfigModel{
		state:    state,
		quitting: false,
	}
}

// getLogLevel 获取日志级别
func getLogLevel(cfg *config.LocalStoreConfig) string {
	if cfg.LogConfig != nil {
		return cfg.LogConfig.Level
	}
	return "info"
}

// getLogFile 获取日志文件路径
func getLogFile(cfg *config.LocalStoreConfig) string {
	if cfg.LogConfig != nil {
		return cfg.LogConfig.File
	}
	return ""
}

// Init bubbletea.Model 接口实现
func (m *ConfigModel) Init() tea.Cmd {
	return nil
}

// Update bubbletea.Model 接口实现
func (m *ConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// 如果在编辑模式，委托给编辑器处理
	if m.state.EditMode {
		return m.updateEditMode(msg)
	}

	// 主界面模式
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "退出"))):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "退出"))):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "上移"))):
			if m.state.SelectedIndex > 0 {
				m.state.SelectedIndex--
			}
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "下移"))):
			if m.state.SelectedIndex < len(m.state.Items)-1 {
				m.state.SelectedIndex++
			}
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "编辑"))):
			// 进入编辑模式
			return m.enterEditMode()

		case key.Matches(msg, key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "保存"))):
			// 保存前验证
			return m.saveWithValidation()

		case key.Matches(msg, key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "重置"))):
			// 重置确认
			if m.state.ErrorMessage == "按 [R] 再次确认重置配置" {
				m.resetConfig()
				m.state.SuccessMessage = "配置已重置为默认值"
				m.state.ErrorMessage = ""
			} else {
				m.state.ErrorMessage = "按 [R] 再次确认重置配置"
				m.state.SuccessMessage = ""
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	return m, nil
}

// updateEditMode 更新编辑模式
func (m *ConfigModel) updateEditMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	if editor, ok := m.state.EditorModel.(*EditorModel); ok {
		// 文本编辑器 - 先更新 editor
		var model tea.Model = editor
		var cmd tea.Cmd
		model, cmd = model.Update(msg)
		m.state.EditorModel = model

		// 检查 editor 是否已完成
		updatedEditor, ok := model.(*EditorModel)
		if ok && updatedEditor.IsDone() {
			if !updatedEditor.IsAborted() {
				// 应用编辑的值
				item := m.state.Items[m.state.SelectedIndex]
				newValue := updatedEditor.GetValue()
				log.Infof("应用编辑值: item.Key=%q, 旧值=%q, 新值=%q", item.Key, item.Value, newValue)
				item.Value = newValue
			} else {
				log.Infof("编辑器被中止，不应用更改")
			}
			m.state.EditMode = false
			m.state.EditorModel = nil
			return m, nil
		}

		return m, cmd
	}

	if selector, ok := m.state.EditorModel.(*ProviderSelectorModel); ok {
		// Provider 选择器 - 先更新 selector
		var teaModel tea.Model = selector
		var cmd tea.Cmd
		teaModel, cmd = teaModel.Update(msg)
		m.state.EditorModel = teaModel

		// 检查 selector 是否已完成
		updatedSelector, selectorOk := teaModel.(*ProviderSelectorModel)
		if selectorOk && updatedSelector.IsDone() {
			if !updatedSelector.IsAborted() {
				// 应用选择的 Provider
				item := m.state.Items[m.state.SelectedIndex]
				oldProvider := model.LlmProvider(item.Value)
				newProvider := updatedSelector.GetSelectedProvider()
				item.Value = string(newProvider)

				log.Infof("切换 Provider: %q -> %q", oldProvider, newProvider)

				// 切换 Provider 时清空 API Key 并更新 Base URL
				apiKeyCleared := false
				baseURLUpdated := false
				for _, it := range m.state.Items {
					if it.Key == "apikey" {
						oldAPIKey := it.Value
						it.Value = ""
						apiKeyCleared = true
						log.Infof("清空 API Key: %q -> %q", oldAPIKey, it.Value)
					}
					if it.Key == "baseurl" {
						oldBaseURL := it.Value
						newBaseURL := newProvider.GetDefaultBaseURL()
						it.Value = newBaseURL
						baseURLUpdated = true
						log.Infof("更新 Base URL: %q -> %q", oldBaseURL, newBaseURL)
					}
				}

				if apiKeyCleared && baseURLUpdated {
					m.state.SuccessMessage = fmt.Sprintf("已切换到 %s，Base URL 已更新，请重新配置 API Key", newProvider.GetDisplayName())
				} else if apiKeyCleared {
					m.state.SuccessMessage = fmt.Sprintf("已切换到 %s，请重新配置 API Key", newProvider.GetDisplayName())
				} else if baseURLUpdated {
					m.state.SuccessMessage = fmt.Sprintf("已切换到 %s，Base URL 已更新", newProvider.GetDisplayName())
				}
			} else {
				log.Infof("Provider 选择器被中止，不应用更改")
			}
			m.state.EditMode = false
			m.state.EditorModel = nil
			return m, nil
		}

		return m, cmd
	}

	return m, nil
}

// enterEditMode 进入编辑模式
func (m *ConfigModel) enterEditMode() (tea.Model, tea.Cmd) {
	item := m.state.Items[m.state.SelectedIndex]

	if item.Key == "provider" {
		// Provider 选择器
		selector := NewProviderSelectorModel(model.LlmProvider(item.Value))
		m.state.EditMode = true
		m.state.EditorModel = selector
		return m, nil
	}

	// 文本编辑器
	isSecret := item.IsSecret
	var editor *EditorModel

	if item.Key == "apikey" {
		editor = NewEditorModel("编辑 API Key", "API Key", item.Value, isSecret)
		editor.SetValidator(func(v string) []*ValidationError {
			cfg := config.GetLocalStoreConfig()
			return validateAPIKey(v, string(cfg.Provider))
		})
	} else if item.Key == "baseurl" {
		editor = NewEditorModel("编辑 Base URL", "Base URL", item.Value, isSecret)
		editor.SetValidator(func(v string) []*ValidationError {
			return validateBaseURL(v)
		})
	} else if item.Key == "model" {
		editor = NewEditorModel("编辑 Model", "Model", item.Value, isSecret)
	} else if item.Key == "loglevel" {
		editor = NewEditorModel("编辑日志级别", "日志级别", item.Value, isSecret)
		editor.SetValidator(func(v string) []*ValidationError {
			return validateLogLevel(v)
		})
	} else if item.Key == "logfile" {
		editor = NewEditorModel("编辑日志文件路径", "日志文件路径", item.Value, isSecret)
		editor.SetValidator(func(v string) []*ValidationError {
			return validateLogPath(v)
		})
	} else {
		editor = NewEditorModel("编辑 "+item.Label, item.Label, item.Value, isSecret)
	}

	m.state.EditMode = true
	m.state.EditorModel = editor
	return m, nil
}

// saveWithValidation 保存并验证配置
func (m *ConfigModel) saveWithValidation() (tea.Model, tea.Cmd) {
	log.Infof("=== 开始保存配置 ===")

	// 构建临时配置进行验证
	cfg := m.buildConfigFromState()
	log.Infof("构建的配置: Provider=%q APIKey=%q BaseURL=%q Model=%q",
		cfg.Provider, maskKey(cfg.APIKey), cfg.BaseURL, cfg.Model)

	// 运行验证
	errors := validateConfig(cfg)

	// 检查是否有错误
	hasError := false
	for _, e := range errors {
		if e.Severity == SeverityError {
			hasError = true
			break
		}
	}

	if hasError {
		// 有错误，显示错误信息
		var errorStrings []string
		for _, e := range errors {
			if e.Severity == SeverityError {
				errorStrings = append(errorStrings, "• "+e.Error())
			}
		}
		m.state.ErrorMessage = "验证失败:\n" + strings.Join(errorStrings, "\n")
		m.state.SuccessMessage = ""
		return m, nil
	}

	// 保存配置
	if err := m.saveConfig(); err != nil {
		m.state.ErrorMessage = fmt.Sprintf("保存失败: %v", err)
		m.state.SuccessMessage = ""
		return m, nil
	}

	// 显示成功消息和警告（如果有）
	if len(errors) > 0 {
		var warningStrings []string
		for _, e := range errors {
			if e.Severity == SeverityWarning {
				warningStrings = append(warningStrings, "• "+e.Error())
			}
		}
		if len(warningStrings) > 0 {
			m.state.SuccessMessage = "配置已保存（有警告）\n" + strings.Join(warningStrings, "\n")
		} else {
			m.state.SuccessMessage = "配置已保存"
		}
	} else {
		m.state.SuccessMessage = "配置已保存"
	}
	m.state.ErrorMessage = ""

	return m, nil
}

// buildConfigFromState 从当前状态构建配置
func (m *ConfigModel) buildConfigFromState() *config.LocalStoreConfig {
	cfg := &config.LocalStoreConfig{}

	for _, item := range m.state.Items {
		switch item.Key {
		case "provider":
			cfg.Provider = model.LlmProvider(item.Value)
		case "apikey":
			cfg.APIKey = item.Value
		case "baseurl":
			cfg.BaseURL = item.Value
		case "model":
			cfg.Model = item.Value
		case "loglevel":
			if cfg.LogConfig == nil {
				cfg.LogConfig = &config.LogConfig{}
			}
			cfg.LogConfig.Level = item.Value
		case "logfile":
			if cfg.LogConfig == nil {
				cfg.LogConfig = &config.LogConfig{}
			}
			cfg.LogConfig.File = item.Value
		}
	}

	return cfg
}

// resetConfig 重置配置为默认值
func (m *ConfigModel) resetConfig() {
	defaultCfg := config.GetDefaultConfig()
	m.state.Items[0].Value = string(defaultCfg.Provider)
	m.state.Items[1].Value = defaultCfg.APIKey
	m.state.Items[2].Value = defaultCfg.BaseURL
	m.state.Items[3].Value = defaultCfg.Model
	m.state.Items[4].Value = defaultCfg.LogConfig.Level
	m.state.Items[5].Value = defaultCfg.LogConfig.File
}

// saveConfig 保存配置
func (m *ConfigModel) saveConfig() error {
	cfg := config.GetLocalStoreConfig()

	// 打印当前状态的值（调试）
	log.Infof("保存前 - 状态中的值:")
	for _, item := range m.state.Items {
		log.Infof("  %s = %q", item.Key, item.Value)
	}

	// 更新配置项
	for _, item := range m.state.Items {
		switch item.Key {
		case "provider":
			log.Infof("更新 Provider: %q -> %q", cfg.Provider, item.Value)
			cfg.Provider = model.LlmProvider(item.Value)
		case "apikey":
			log.Infof("更新 APIKey: %q -> %q", cfg.APIKey, item.Value)
			cfg.APIKey = item.Value
		case "baseurl":
			log.Infof("更新 BaseURL: %q -> %q", cfg.BaseURL, item.Value)
			cfg.BaseURL = item.Value
		case "model":
			log.Infof("更新 Model: %q -> %q", cfg.Model, item.Value)
			cfg.Model = item.Value
		case "loglevel":
			if cfg.LogConfig == nil {
				cfg.LogConfig = &config.LogConfig{}
			}
			log.Infof("更新 LogLevel: %q -> %q", cfg.LogConfig.Level, item.Value)
			cfg.LogConfig.Level = item.Value
		case "logfile":
			if cfg.LogConfig == nil {
				cfg.LogConfig = &config.LogConfig{}
			}
			log.Infof("更新 LogFile: %q -> %q", cfg.LogConfig.File, item.Value)
			cfg.LogConfig.File = item.Value
		}
	}

	log.Infof("保存后 - 缓存的配置: Provider=%q APIKey=%q BaseURL=%q Model=%q",
		cfg.Provider, cfg.APIKey, cfg.BaseURL, cfg.Model)

	return config.SaveConfig()
}

// View bubbletea.Model 接口实现
func (m *ConfigModel) View() string {
	if m.quitting {
		return ""
	}

	// 如果在编辑模式，显示编辑器
	if m.state.EditMode {
		if editor, ok := m.state.EditorModel.(tea.Model); ok {
			return editor.View()
		}
	}

	var s string

	// 标题
	s += TitleStyle.Render("MSA 配置管理") + "\n\n"

	// API 配置组
	s += SectionTitleStyle.Render("API 配置") + "\n"
	for i, item := range m.state.Items {
		if item.Key == "provider" || item.Key == "apikey" || item.Key == "baseurl" || item.Key == "model" {
			s += m.renderItem(i, item)
		}
	}

	s += "\n"

	// 日志配置组
	s += SectionTitleStyle.Render("日志配置") + "\n"
	for i, item := range m.state.Items {
		if item.Key == "loglevel" || item.Key == "logfile" {
			s += m.renderItem(i, item)
		}
	}

	// 错误消息
	if m.state.ErrorMessage != "" {
		s += "\n" + ErrorStyle.Render(m.state.ErrorMessage) + "\n"
	}

	// 成功消息
	if m.state.SuccessMessage != "" {
		s += "\n" + SuccessStyle.Render(m.state.SuccessMessage) + "\n"
	}

	// 帮助信息
	s += "\n" + HelpStyle.Render("操作: [↑↓] 选择 | [Enter] 编辑 | [S] 保存 | [R] 重置 | [Q] 退出")

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(1, 2).
		Render(s)
}

// renderItem 渲染单个配置项
func (m *ConfigModel) renderItem(index int, item *ConfigItem) string {
	// 找到该项在所有项中的真实索引
	realIndex := -1
	for i, it := range m.state.Items {
		if it.Key == item.Key {
			realIndex = i
			break
		}
	}

	if realIndex == -1 {
		return ""
	}

	cursor := " "
	if realIndex == m.state.SelectedIndex {
		cursor = CursorStyle.Render("▸")
	}

	label := LabelStyle.Render(item.Label + ":")
	displayValue := item.DisplayFunc(item.Value)
	value := ValueStyle.Render(displayValue)

	return fmt.Sprintf("%s %s %s\n", cursor, label, value)
}

// Run 运行配置 TUI
func Run() error {
	model := NewConfigModel()
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("运行配置界面失败: %w", err)
	}
	return nil
}

// maskKey 隐藏密钥用于日志
func maskKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

// 本地版本的 ValidateConfig 函数（避免循环导入）
func validateConfig(cfg *config.LocalStoreConfig) []*ValidationError {
	var allErrors []*ValidationError

	// 验证 Provider
	if cfg.Provider != "" {
		p := model.LlmProvider(cfg.Provider)
		if _, ok := model.ProviderRegistry[p]; !ok {
			allErrors = append(allErrors, &ValidationError{
				Field:    "Provider",
				Message:  fmt.Sprintf("未知的 Provider: %s", cfg.Provider),
				Severity: SeverityError,
			})
		}
	}

	// 验证 API Key
	if cfg.APIKey != "" {
		p := model.LlmProvider(cfg.Provider)
		if err := p.ValidateAPIKey(cfg.APIKey); err != nil {
			allErrors = append(allErrors, &ValidationError{
				Field:    "API Key",
				Message:  err.Error(),
				Severity: SeverityError,
			})
		}
		if len(cfg.APIKey) < 10 {
			allErrors = append(allErrors, &ValidationError{
				Field:    "API Key",
				Message:  "API Key 长度可能不正确（少于 10 个字符）",
				Severity: SeverityWarning,
			})
		}
	}

	// 验证 Base URL
	if cfg.BaseURL != "" {
		if len(cfg.BaseURL) >= 7 {
			hasPrefix := false
			if cfg.BaseURL[:7] == "http://" {
				hasPrefix = true
			}
			if len(cfg.BaseURL) >= 8 && cfg.BaseURL[:8] == "https://" {
				hasPrefix = true
			}
			if !hasPrefix {
				allErrors = append(allErrors, &ValidationError{
					Field:    "Base URL",
					Message:  "URL 必须以 http:// 或 https:// 开头",
					Severity: SeverityError,
				})
			}
		}
	}

	// 验证日志配置
	if cfg.LogConfig != nil {
		// 验证日志级别
		if cfg.LogConfig.Level != "" {
			validLevels := map[string]bool{
				"debug": true, "info": true, "warn": true,
				"warning": true, "error": true, "fatal": true, "panic": true,
			}
			if !validLevels[cfg.LogConfig.Level] {
				allErrors = append(allErrors, &ValidationError{
					Field:    "日志级别",
					Message:  "无效的日志级别",
					Severity: SeverityError,
				})
			}
		}

		// 验证日志路径
		if cfg.LogConfig.File != "" && len(cfg.LogConfig.File) == 0 {
			allErrors = append(allErrors, &ValidationError{
				Field:    "日志路径",
				Message:  "日志路径不能为空",
				Severity: SeverityWarning,
			})
		}
	}

	return allErrors
}
