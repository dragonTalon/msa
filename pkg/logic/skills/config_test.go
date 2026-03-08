package skills

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestGetDefaultConfig 测试 getDefaultConfig 函数
func TestGetDefaultConfig(t *testing.T) {
	config := getDefaultConfig()

	if config == nil {
		t.Fatal("Expected non-nil config")
	}

	if len(config.DisableSkills) != 0 {
		t.Errorf("Expected empty DisableSkills, got %v", config.DisableSkills)
	}
}

// TestGetDisabledSkillsMap 测试 getDisabledSkillsMap 函数
func TestGetDisabledSkillsMap(t *testing.T) {
	// 重置配置缓存
	configCache = nil
	configLoaded = false

	// 测试无配置文件的情况（应该返回空 map）
	disabledMap := getDisabledSkillsMap()

	if disabledMap == nil {
		t.Fatal("Expected non-nil map")
	}

	if len(disabledMap) > 0 {
		t.Errorf("Expected empty map, got %v", disabledMap)
	}
}

// TestSaveAndLoadConfig 测试配置的保存和加载
func TestSaveAndLoadConfig(t *testing.T) {
	// 创建临时目录作为用户主目录
	tmpHome := t.TempDir()

	// 保存原始环境变量
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// 设置临时主目录
	os.Setenv("HOME", tmpHome)

	// 重置配置缓存
	configCache = nil
	configLoaded = false

	// 测试保存配置
	disabledSkills := []string{"skill-a", "skill-b", "skill-c"}
	err := saveConfig(disabledSkills)
	if err != nil {
		t.Fatalf("saveConfig failed: %v", err)
	}

	// 验证文件是否存在
	configPath := filepath.Join(tmpHome, configDir, configFile)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// 重置缓存以强制重新加载
	configCache = nil
	configLoaded = false

	// 测试加载配置
	loadedConfig, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig failed: %v", err)
	}

	if len(loadedConfig.DisableSkills) != len(disabledSkills) {
		t.Errorf("Expected %d disabled skills, got %d", len(disabledSkills), len(loadedConfig.DisableSkills))
	}

	// 验证内容正确
	for i, expected := range disabledSkills {
		if i >= len(loadedConfig.DisableSkills) {
			t.Errorf("Missing disabled skill: %s", expected)
			continue
		}
		if loadedConfig.DisableSkills[i] != expected {
			t.Errorf("Expected skill %q at index %d, got %q", expected, i, loadedConfig.DisableSkills[i])
		}
	}
}

// TestSaveConfigEmptyList 测试保存空的禁用列表
func TestSaveConfigEmptyList(t *testing.T) {
	// 创建临时目录作为用户主目录
	tmpHome := t.TempDir()

	// 保存原始环境变量
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// 设置临时主目录
	os.Setenv("HOME", tmpHome)

	// 重置配置缓存
	configCache = nil
	configLoaded = false

	// 保存空列表
	err := saveConfig([]string{})
	if err != nil {
		t.Fatalf("saveConfig with empty list failed: %v", err)
	}

	// 验证文件存在
	configPath := filepath.Join(tmpHome, configDir, configFile)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// 验证文件内容
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	// 空列表时应该删除该字段
	if _, exists := config["disable_skills"]; exists {
		t.Error("disable_skills field should be removed when empty")
	}
}

// TestLoadConfigCorruptedJSON 测试加载损坏的 JSON 文件
func TestLoadConfigCorruptedJSON(t *testing.T) {
	// 创建临时目录作为用户主目录
	tmpHome := t.TempDir()

	// 保存原始环境变量
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// 设置临时主目录
	os.Setenv("HOME", tmpHome)

	// 创建配置目录
	configPath := filepath.Join(tmpHome, configDir, configFile)
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// 写入损坏的 JSON
	if err := os.WriteFile(configPath, []byte("{ invalid json }"), 0644); err != nil {
		t.Fatalf("Failed to write corrupted config: %v", err)
	}

	// 重置配置缓存
	configCache = nil
	configLoaded = false

	// 应该返回默认配置而不是报错
	config, err := loadConfig()
	if err != nil {
		t.Errorf("loadConfig should not return error for corrupted JSON, got: %v", err)
	}

	if config == nil {
		t.Fatal("Expected non-nil config")
	}

	// 应该返回默认配置
	if len(config.DisableSkills) != 0 {
		t.Errorf("Expected default config (empty), got %v", config.DisableSkills)
	}
}

// TestLoadConfigWithExistingFields 测试加载包含其他字段的配置文件
func TestLoadConfigWithExistingFields(t *testing.T) {
	// 创建临时目录作为用户主目录
	tmpHome := t.TempDir()

	// 保存原始环境变量
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// 设置临时主目录
	os.Setenv("HOME", tmpHome)

	// 创建配置目录
	configPath := filepath.Join(tmpHome, configDir, configFile)
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// 写入包含其他字段的配置
	existingConfig := map[string]interface{}{
		"other_field":    "some_value",
		"disable_skills": []string{"skill-x"},
	}
	data, _ := json.MarshalIndent(existingConfig, "", "  ")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// 重置配置缓存
	configCache = nil
	configLoaded = false

	// 加载配置
	config, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig failed: %v", err)
	}

	if len(config.DisableSkills) != 1 || config.DisableSkills[0] != "skill-x" {
		t.Errorf("Expected [skill-x], got %v", config.DisableSkills)
	}

	// 保存新的禁用列表（应该保留其他字段）
	newDisabledSkills := []string{"skill-y"}
	if err := saveConfig(newDisabledSkills); err != nil {
		t.Fatalf("saveConfig failed: %v", err)
	}

	// 验证其他字段仍然存在
	data, err = os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var reloadedConfig map[string]interface{}
	if err := json.Unmarshal(data, &reloadedConfig); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	if reloadedConfig["other_field"] != "some_value" {
		t.Error("other_field should be preserved")
	}
}

// TestConfigCaching 测试配置缓存机制
func TestConfigCaching(t *testing.T) {
	// 创建临时目录作为用户主目录
	tmpHome := t.TempDir()

	// 保存原始环境变量
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// 设置临时主目录
	os.Setenv("HOME", tmpHome)

	// 重置配置缓存
	configCache = nil
	configLoaded = false

	// 先保存一个配置文件，这样缓存才会生效
	if err := saveConfig([]string{"skill-test"}); err != nil {
		t.Fatalf("saveConfig failed: %v", err)
	}

	// 第一次加载
	config1, err := loadConfig()
	if err != nil {
		t.Fatalf("First loadConfig failed: %v", err)
	}

	// 第二次加载（应该使用缓存）
	config2, err := loadConfig()
	if err != nil {
		t.Fatalf("Second loadConfig failed: %v", err)
	}

	// 验证是同一个对象
	if config1 != config2 {
		t.Error("Expected cached config to be the same object")
	}

	// 验证缓存标志
	if !configLoaded {
		t.Error("configLoaded should be true after loading")
	}

	// 验证缓存内容正确
	if len(config1.DisableSkills) != 1 || config1.DisableSkills[0] != "skill-test" {
		t.Errorf("Expected [skill-test], got %v", config1.DisableSkills)
	}
}
