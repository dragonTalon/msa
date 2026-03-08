package skills

import (
	"os"
	"path/filepath"
	"testing"
)

// resetGlobalManager 重置全局单例（测试辅助函数）
func resetGlobalManager() {
	globalManager = nil
}

// TestGetManager 测试 GetManager 全局单例
func TestGetManager(t *testing.T) {
	// 重置全局单例
	resetGlobalManager()
	// 重置配置缓存
	configCache = nil
	configLoaded = false

	// 第一次获取
	manager1 := GetManager()
	if manager1 == nil {
		t.Fatal("GetManager() returned nil")
	}

	// 第二次获取（应该返回同一个实例）
	manager2 := GetManager()
	if manager2 == nil {
		t.Fatal("GetManager() returned nil on second call")
	}

	// 验证是同一个实例
	if manager1 != manager2 {
		t.Error("GetManager() should return the same instance")
	}

	// 验证内部字段已初始化
	if manager1.registry == nil {
		t.Error("Manager.registry should not be nil")
	}
	if manager1.loader == nil {
		t.Error("Manager.loader should not be nil")
	}
}

// TestManagerGetRegistry 测试 GetRegistry 方法
func TestManagerGetRegistry(t *testing.T) {
	resetGlobalManager()
	configCache = nil
	configLoaded = false

	manager := GetManager()
	registry := manager.GetRegistry()

	if registry == nil {
		t.Fatal("GetRegistry() returned nil")
	}

	// 验证 registry 可用（可以注册 skill）
	registry.Register(&Skill{
		Name:        "test-skill",
		Description: "Test skill",
		Priority:    5,
		path:        "/test/path.md",
	})

	skill, err := registry.Get("test-skill")
	if err != nil {
		t.Fatalf("Failed to get registered skill: %v", err)
	}
	if skill.Name != "test-skill" {
		t.Errorf("Expected skill name 'test-skill', got '%s'", skill.Name)
	}
}

// TestManagerListAllSkills 测试 ListAllSkills 方法
func TestManagerListAllSkills(t *testing.T) {
	resetGlobalManager()
	configCache = nil
	configLoaded = false

	manager := GetManager()
	registry := manager.GetRegistry()

	// 注册一些测试 skills
	testSkills := []*Skill{
		{Name: "base", Description: "Base", Priority: 10, path: "/base.md"},
		{Name: "stock", Description: "Stock", Priority: 8, path: "/stock.md"},
		{Name: "news", Description: "News", Priority: 5, path: "/news.md"},
	}

	for _, skill := range testSkills {
		registry.Register(skill)
	}

	// 获取所有 skills（包括禁用的）
	allSkills := manager.ListAllSkills()

	if len(allSkills) < len(testSkills) {
		t.Errorf("Expected at least %d skills, got %d", len(testSkills), len(allSkills))
	}

	// 验证包含我们注册的 skills
	foundCount := 0
	for _, s := range testSkills {
		for _, listed := range allSkills {
			if listed.Name == s.Name {
				foundCount++
				break
			}
		}
	}
	if foundCount != len(testSkills) {
		t.Errorf("Expected to find %d test skills, found %d", len(testSkills), foundCount)
	}
}

// TestManagerDisableEnableSkill 测试 DisableSkill 和 EnableSkill 方法
func TestManagerDisableEnableSkill(t *testing.T) {
	// 创建临时目录作为用户主目录
	tmpHome := t.TempDir()

	// 保存原始环境变量
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// 设置临时主目录
	os.Setenv("HOME", tmpHome)

	// 重置全局状态
	resetGlobalManager()
	configCache = nil
	configLoaded = false

	manager := GetManager()
	registry := manager.GetRegistry()

	// 注册测试 skills
	testSkills := []*Skill{
		{Name: "skill-a", Description: "A", Priority: 5, path: "/a.md"},
		{Name: "skill-b", Description: "B", Priority: 5, path: "/b.md"},
	}
	for _, skill := range testSkills {
		registry.Register(skill)
	}

	// 测试初始状态：没有禁用的 skills
	disabled := manager.GetDisabledSkills()
	if len(disabled) != 0 {
		t.Errorf("Expected no disabled skills initially, got %v", disabled)
	}

	// 测试禁用 skill-a
	err := manager.DisableSkill("skill-a")
	if err != nil {
		t.Fatalf("DisableSkill failed: %v", err)
	}

	// 验证禁用状态
	if !manager.IsDisabled("skill-a") {
		t.Error("skill-a should be disabled")
	}
	if manager.IsDisabled("skill-b") {
		t.Error("skill-b should not be disabled")
	}

	// 验证禁用列表
	disabled = manager.GetDisabledSkills()
	if len(disabled) != 1 || disabled[0] != "skill-a" {
		t.Errorf("Expected [skill-a], got %v", disabled)
	}

	// 验证 ListSkills 不包含禁用的 skill
	availableSkills := manager.ListSkills()
	for _, s := range availableSkills {
		if s.Name == "skill-a" {
			t.Error("skill-a should not be in available skills list")
		}
	}

	// 测试重复禁用（应该无错误）
	err = manager.DisableSkill("skill-a")
	if err != nil {
		t.Errorf("Disabling already disabled skill should not error: %v", err)
	}

	// 测试启用 skill-a
	err = manager.EnableSkill("skill-a")
	if err != nil {
		t.Fatalf("EnableSkill failed: %v", err)
	}

	// 验证启用状态
	if manager.IsDisabled("skill-a") {
		t.Error("skill-a should not be disabled after enabling")
	}

	// 验证禁用列表为空
	disabled = manager.GetDisabledSkills()
	if len(disabled) != 0 {
		t.Errorf("Expected no disabled skills, got %v", disabled)
	}

	// 测试启用未禁用的 skill（应该无错误）
	err = manager.EnableSkill("skill-b")
	if err != nil {
		t.Errorf("Enabling non-disabled skill should not error: %v", err)
	}
}

// TestManagerDisableMultipleSkills 测试禁用多个 skills
func TestManagerDisableMultipleSkills(t *testing.T) {
	// 创建临时目录作为用户主目录
	tmpHome := t.TempDir()

	// 保存原始环境变量
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// 设置临时主目录
	os.Setenv("HOME", tmpHome)

	// 重置全局状态
	resetGlobalManager()
	configCache = nil
	configLoaded = false

	manager := GetManager()

	// 禁用多个 skills
	skillsToDisable := []string{"skill-1", "skill-2", "skill-3"}
	for _, name := range skillsToDisable {
		if err := manager.DisableSkill(name); err != nil {
			t.Fatalf("DisableSkill(%s) failed: %v", name, err)
		}
	}

	// 验证所有都被禁用
	disabled := manager.GetDisabledSkills()
	if len(disabled) != len(skillsToDisable) {
		t.Errorf("Expected %d disabled skills, got %d", len(skillsToDisable), len(disabled))
	}

	// 验证每个都在禁用列表中
	for _, name := range skillsToDisable {
		if !manager.IsDisabled(name) {
			t.Errorf("skill %s should be disabled", name)
		}
	}

	// 启用其中一个
	if err := manager.EnableSkill("skill-2"); err != nil {
		t.Fatalf("EnableSkill failed: %v", err)
	}

	// 验证只有两个被禁用
	disabled = manager.GetDisabledSkills()
	if len(disabled) != 2 {
		t.Errorf("Expected 2 disabled skills, got %d", len(disabled))
	}

	// 验证 skill-2 不在禁用列表中
	if manager.IsDisabled("skill-2") {
		t.Error("skill-2 should not be disabled")
	}
}

// TestManagerGetSkill 测试 GetSkill 方法
func TestManagerGetSkill(t *testing.T) {
	resetGlobalManager()
	configCache = nil
	configLoaded = false

	manager := GetManager()
	registry := manager.GetRegistry()

	// 注册一个测试 skill
	testSkill := &Skill{
		Name:        "test-get-skill",
		Description: "Test GetSkill",
		Priority:    5,
		path:        "/test/get.md",
	}
	registry.Register(testSkill)

	// 测试获取存在的 skill
	skill, err := manager.GetSkill("test-get-skill")
	if err != nil {
		t.Fatalf("GetSkill failed: %v", err)
	}
	if skill.Name != "test-get-skill" {
		t.Errorf("Expected skill name 'test-get-skill', got '%s'", skill.Name)
	}

	// 测试获取不存在的 skill（返回 nil, nil）
	skill, err = manager.GetSkill("non-existent-skill")
	if err != nil {
		t.Errorf("GetSkill should not return error for non-existent skill: %v", err)
	}
	if skill != nil {
		t.Error("GetSkill should return nil for non-existent skill")
	}
}

// TestManagerInitialize 测试 Initialize 方法
func TestManagerInitialize(t *testing.T) {
	resetGlobalManager()
	configCache = nil
	configLoaded = false

	manager := GetManager()

	// Initialize 应该不会返回错误（即使 skills 目录不存在）
	err := manager.Initialize()
	if err != nil {
		t.Errorf("Initialize should not return error: %v", err)
	}

	// 验证 registry 可以列出 skills（即使为空）
	skills := manager.ListAllSkills()
	// 不验证具体数量，因为可能已经有其他测试注册的 skills
	t.Logf("Initialized with %d skills", len(skills))
}

// TestManagerWithExistingConfig 测试 Manager 与现有配置文件的交互
func TestManagerWithExistingConfig(t *testing.T) {
	// 创建临时目录作为用户主目录
	tmpHome := t.TempDir()

	// 保存原始环境变量
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// 设置临时主目录
	os.Setenv("HOME", tmpHome)

	// 创建预置的配置文件
	configPath := filepath.Join(tmpHome, configDir, configFile)
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// 写入预置的禁用列表
	preexistingDisabled := []string{"pre-disabled-1", "pre-disabled-2"}
	if err := saveConfig(preexistingDisabled); err != nil {
		t.Fatalf("Failed to save preset config: %v", err)
	}

	// 重置全局状态（会加载配置）
	resetGlobalManager()
	configCache = nil
	configLoaded = false

	manager := GetManager()

	// 验证预置的禁用状态被正确加载
	if !manager.IsDisabled("pre-disabled-1") {
		t.Error("pre-disabled-1 should be disabled from config")
	}
	if !manager.IsDisabled("pre-disabled-2") {
		t.Error("pre-disabled-2 should be disabled from config")
	}

	// 添加新的禁用 skill
	if err := manager.DisableSkill("new-disabled"); err != nil {
		t.Fatalf("DisableSkill failed: %v", err)
	}

	// 验证所有三个都被禁用
	disabled := manager.GetDisabledSkills()
	if len(disabled) != 3 {
		t.Errorf("Expected 3 disabled skills, got %d", len(disabled))
	}
}

// TestManagerConcurrentAccess 测试 Manager 的并发访问安全性
func TestManagerConcurrentAccess(t *testing.T) {
	resetGlobalManager()
	configCache = nil
	configLoaded = false

	manager := GetManager()
	registry := manager.GetRegistry()

	// 注册测试 skills
	for i := 0; i < 10; i++ {
		registry.Register(&Skill{
			Name:        "concurrent-test-" + string(rune('0'+i)),
			Description: "Concurrent test",
			Priority:    5,
			path:        "/test.md",
		})
	}

	// 并发读取测试
	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func() {
			_ = manager.ListAllSkills()
			_ = manager.ListSkills()
			_, _ = manager.GetSkill("concurrent-test-0")
			done <- true
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 5; i++ {
		<-done
	}
}
