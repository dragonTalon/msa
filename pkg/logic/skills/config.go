package skills

import (
	"encoding/json"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// SkillsConfig Skill 相关配置
type SkillsConfig struct {
	DisableSkills []string `json:"disable_skills"`
}

// configDir 用户配置目录
const configDir = ".msa"
const configFile = "config.json"

// configCache 配置缓存
var configCache *SkillsConfig
var configLoaded bool

// loadConfig 加载配置文件
func loadConfig() (*SkillsConfig, error) {
	// 如果已加载，返回缓存
	if configLoaded && configCache != nil {
		return configCache, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Warnf("Failed to get home directory: %v", err)
		return getDefaultConfig(), nil
	}

	configPath := filepath.Join(homeDir, configDir, configFile)

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Infof("Config file does not exist: %s, using defaults", configPath)
			return getDefaultConfig(), nil
		}
		log.Warnf("Failed to read config file: %v", err)
		return getDefaultConfig(), nil
	}

	var config struct {
		DisableSkills []string `json:"disable_skills"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		log.Warnf("Failed to parse config file: %v, using defaults", err)
		return getDefaultConfig(), nil
	}

	skillsConfig := &SkillsConfig{
		DisableSkills: config.DisableSkills,
	}

	// 缓存配置
	configCache = skillsConfig
	configLoaded = true

	log.Debugf("Loaded config with %d disabled skills", len(skillsConfig.DisableSkills))

	return skillsConfig, nil
}

// saveConfig 保存配置文件
func saveConfig(disabledSkills []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(homeDir, configDir, configFile)

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	// 读取现有配置（如果存在）
	var existingConfig map[string]interface{}
	if data, err := os.ReadFile(configPath); err == nil {
		json.Unmarshal(data, &existingConfig)
	} else if !os.IsNotExist(err) {
		return err
	}

	// 更新 disable_skills 字段
	if existingConfig == nil {
		existingConfig = make(map[string]interface{})
	}
	existingConfig["disable_skills"] = disabledSkills

	// 如果 disable_skills 为空，可以删除该字段（可选）
	if len(disabledSkills) == 0 {
		delete(existingConfig, "disable_skills")
	}

	// 写入文件
	data, err := json.MarshalIndent(existingConfig, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return err
	}

	// 更新缓存
	if configCache != nil {
		configCache.DisableSkills = disabledSkills
	}

	log.Infof("Saved config with %d disabled skills", len(disabledSkills))

	return nil
}

// getDefaultConfig 返回默认配置
func getDefaultConfig() *SkillsConfig {
	return &SkillsConfig{
		DisableSkills: []string{},
	}
}

// getDisabledSkillsMap 获取禁用的 Skills 映射表
func getDisabledSkillsMap() map[string]bool {
	config, err := loadConfig()
	if err != nil {
		log.Warnf("Failed to load config: %v", err)
		return make(map[string]bool)
	}

	disabledMap := make(map[string]bool)
	for _, name := range config.DisableSkills {
		disabledMap[name] = true
	}

	return disabledMap
}
