package skills

import (
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// Manager Skill 管理器（对外 API）
type Manager struct {
	registry *Registry
	loader   *Loader
}

var globalManager *Manager

// GetManager 获取 Manager 单例
func GetManager() *Manager {
	if globalManager != nil {
		return globalManager
	}

	// 初始化 Manager
	registry := NewRegistry()

	// 获取当前工作目录的 skills 目录
	builtinDir := filepath.Join("pkg", "logic", "skills", "plugs")
	userDir := filepath.Join(".msa", "skills")

	loader := NewLoader(builtinDir, userDir, registry)

	globalManager = &Manager{
		registry: registry,
		loader:   loader,
	}

	log.Infof("Skills manager initialized (builtin: %s, user: %s)", builtinDir, userDir)

	return globalManager
}

// Initialize 初始化并加载所有 Skills
func (m *Manager) Initialize() error {
	log.Info("Initializing skills...")

	if err := m.loader.LoadAll(); err != nil {
		log.Warnf("Some skills failed to load: %v", err)
		// 不返回错误，允许部分加载
	}

	skills := m.registry.ListAll()
	log.Infof("Skills initialized: loaded %d skills", len(skills))

	return nil
}

// GetSkill 根据 name 获取 Skill
func (m *Manager) GetSkill(name string) (*Skill, error) {
	return m.registry.Get(name)
}

// ListSkills 返回所有可用的 Skills
func (m *Manager) ListSkills() []*Skill {
	disabledMap := getDisabledSkillsMap()
	return m.registry.ListAvailable(disabledMap)
}

// ListAllSkills 返回所有 Skills（包括禁用的）
func (m *Manager) ListAllSkills() []*Skill {
	return m.registry.ListAll()
}

// IsDisabled 检查 Skill 是否被禁用
func (m *Manager) IsDisabled(name string) bool {
	config, err := loadConfig()
	if err != nil {
		return false
	}

	for _, disabledName := range config.DisableSkills {
		if disabledName == name {
			return true
		}
	}

	return false
}

// DisableSkill 禁用指定 Skill
func (m *Manager) DisableSkill(name string) error {
	config, err := loadConfig()
	if err != nil {
		config = getDefaultConfig()
	}

	// 检查是否已在列表中
	for _, disabledName := range config.DisableSkills {
		if disabledName == name {
			log.Infof("Skill %s is already disabled", name)
			return nil
		}
	}

	// 添加到禁用列表
	config.DisableSkills = append(config.DisableSkills, name)

	return saveConfig(config.DisableSkills)
}

// EnableSkill 启用指定 Skill
func (m *Manager) EnableSkill(name string) error {
	config, err := loadConfig()
	if err != nil {
		// 配置不存在，说明没有禁用的 Skills
		log.Infof("Skill %s is not disabled", name)
		return nil
	}

	// 从禁用列表中移除
	newList := make([]string, 0, len(config.DisableSkills))
	found := false
	for _, disabledName := range config.DisableSkills {
		if disabledName != name {
			newList = append(newList, disabledName)
		} else {
			found = true
		}
	}

	if !found {
		log.Infof("Skill %s is not disabled", name)
		return nil
	}

	return saveConfig(newList)
}

// GetDisabledSkills 获取所有禁用的 Skills
func (m *Manager) GetDisabledSkills() []string {
	config, err := loadConfig()
	if err != nil {
		return []string{}
	}

	return config.DisableSkills
}

// GetRegistry 获取内部 Registry（供 Selector 和 Builder 使用）
func (m *Manager) GetRegistry() *Registry {
	return m.registry
}
