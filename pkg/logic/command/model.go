package command

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"msa/pkg/config"
	"msa/pkg/logic/provider"
	"msa/pkg/logic/skills"
	"msa/pkg/model"

	log "github.com/sirupsen/logrus"
)

func init() {
	RegisterCommand(&ListModel{})
	RegisterCommand(&ConfigCommand{})
	RegisterCommand(&SkillsCommand{})
	// SetModel 命令已被交互式选择器替代，使用 /models 或 /model 命令
	// RegisterCommand(&SetModel{})
}

type ListModel struct {
}

func (l *ListModel) Name() string {
	return "models"
}

func (l *ListModel) Description() string {
	return "List all models"
}

func (l *ListModel) Run(ctx context.Context, args []string) (*model.CmdResult, error) {
	p := provider.GetProvider()
	if p == nil {
		return nil, fmt.Errorf("provider not found")
	}
	models, err := p.ListModels(ctx)
	if err != nil {
		return nil, err
	}

	// 转换为 SelectorItem
	var items []*model.SelectorItem
	for _, m := range models {
		items = append(items, &model.SelectorItem{
			Name:        m.Name,
			Description: m.Description,
		})
	}

	return &model.CmdResult{
		Code: 0,
		Msg:  "success",
		Type: "selector", // 返回 selector 类型，启动交互式选择器
		Data: items,
	}, nil
}

func (l *ListModel) ToSelect(items []*model.SelectorItem) (*model.BaseSelector, error) {
	// 按名称排序，保证顺序一致
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	return &model.BaseSelector{
		Items:         items,
		FilteredItems: items, // 初始时显示所有项
		Cursor:        0,
		ViewportTop:   0,
		ViewportSize:  15, // 默认显示15行
		SearchQuery:   "", // 初始搜索为空
		OnConfirm: func(selected string) error {
			// 保存选中的模型到配置
			config.GetLocalStoreConfig().Model = selected
			err := config.SaveConfig()
			if err != nil {
				log.Errorf("保存配置失败: %v", err)
				return err
			}
			log.Infof("已选择模型: %s", selected)
			return nil
		},
	}, nil
}

// ConfigCommand 配置管理命令
type ConfigCommand struct{}

func (c *ConfigCommand) Name() string {
	return "config"
}

func (c *ConfigCommand) Description() string {
	return "Show current configuration or open config TUI (run 'msa config' in terminal)"
}

func (c *ConfigCommand) Run(ctx context.Context, args []string) (*model.CmdResult, error) {
	cfg := config.GetLocalStoreConfig()

	// 构建配置信息
	configInfo := fmt.Sprintf(`当前配置:

Provider: %s
Model: %s
Base URL: %s
API Key: %s
日志级别: %s
日志文件: %s

💡 提示: 运行 'msa config' 命令打开配置管理界面`,
		cfg.Provider,
		cfg.Model,
		cfg.BaseURL,
		maskAPIKey(cfg.APIKey),
		getLogLevelDisplay(cfg),
		getLogFileDisplay(cfg),
	)

	return &model.CmdResult{
		Code: 0,
		Msg:  "success",
		Type: "message",
		Data: configInfo,
	}, nil
}

func (c *ConfigCommand) ToSelect(items []*model.SelectorItem) (*model.BaseSelector, error) {
	return nil, fmt.Errorf("config command does not support selector mode")
}

// maskAPIKey 隐藏 API Key 显示
func maskAPIKey(apiKey string) string {
	if apiKey == "" {
		return "(未设置)"
	}
	if len(apiKey) <= 8 {
		return apiKey
	}
	return apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
}

// getLogLevelDisplay 获取日志级别显示
func getLogLevelDisplay(cfg *config.LocalStoreConfig) string {
	if cfg.LogConfig != nil && cfg.LogConfig.Level != "" {
		return cfg.LogConfig.Level
	}
	return "info (默认)"
}

// getLogFileDisplay 获取日志文件显示
func getLogFileDisplay(cfg *config.LocalStoreConfig) string {
	if cfg.LogConfig != nil && cfg.LogConfig.File != "" {
		return cfg.LogConfig.File
	}
	return "(未设置，输出到标准输出)"
}

// SkillsCommand Skills 列表命令
type SkillsCommand struct{}

func (s *SkillsCommand) Name() string {
	return "skills"
}

func (s *SkillsCommand) Description() string {
	return "List all available Skills"
}

func (s *SkillsCommand) Run(ctx context.Context, args []string) (*model.CmdResult, error) {
	// 初始化 Skills Manager
	manager := skills.GetManager()
	if err := manager.Initialize(); err != nil {
		log.Warnf("Failed to initialize skills: %v", err)
		return &model.CmdResult{
			Code: 1,
			Msg:  "failed to initialize skills",
			Type: "message",
			Data: "无法加载 Skills: " + err.Error(),
		}, nil
	}

	// 获取所有 skills
	allSkills := manager.ListSkills()
	disabled := manager.GetDisabledSkills()

	if len(allSkills) == 0 {
		return &model.CmdResult{
			Code: 0,
			Msg:  "success",
			Type: "message",
			Data: "没有可用的 Skills",
		}, nil
	}

	// 创建禁用 map
	disabledMap := make(map[string]bool)
	for _, name := range disabled {
		disabledMap[name] = true
	}

	// 按优先级排序
	sort.Slice(allSkills, func(i, j int) bool {
		return allSkills[i].Priority > allSkills[j].Priority
	})

	// 构建输出
	var lines []string
	lines = append(lines, "📋 可用 Skills:")
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("%-20s %-8s %-10s %-10s", "Name", "Priority", "Source", "Status"))
	lines = append(lines, strings.Repeat("-", 50))

	for _, skill := range allSkills {
		name := skill.Name
		priority := formatPriority(skill.Priority)
		source := formatSource(skill.Source)
		status := "启用"
		if disabledMap[name] {
			status = "禁用"
		}

		lines = append(lines, fmt.Sprintf("%-20s %-8s %-10s %-10s", name, priority, source, status))
	}

	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("总计: %d 个 Skills", len(allSkills)))
	lines = append(lines, "")
	lines = append(lines, "💡 提示:")
	lines = append(lines, "  • 使用 /skills: <name1>,<name2> 手动指定 Skills")
	lines = append(lines, "  • 使用 'msa skills show <name>' 查看详情")
	lines = append(lines, "  • 使用 'msa skills disable/enable <name>' 管理状态")

	return &model.CmdResult{
		Code: 0,
		Msg:  "success",
		Type: "message",
		Data: strings.Join(lines, "\n"),
	}, nil
}

func (s *SkillsCommand) ToSelect(items []*model.SelectorItem) (*model.BaseSelector, error) {
	return nil, fmt.Errorf("skills command does not support selector mode")
}

// formatPriority 格式化优先级显示
func formatPriority(priority int) string {
	switch {
	case priority >= 10:
		return fmt.Sprintf("%d (最高)", priority)
	case priority >= 8:
		return fmt.Sprintf("%d (高)", priority)
	case priority >= 5:
		return fmt.Sprintf("%d (中)", priority)
	default:
		return fmt.Sprintf("%d (低)", priority)
	}
}

// formatSource 格式化来源显示
func formatSource(source skills.SkillSource) string {
	return source.String()
}
