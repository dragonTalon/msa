package skills

import (
	"sync"
	"testing"
)

// TestRegistryRegisterAndQuery 测试注册和查询功能
func TestRegistryRegisterAndQuery(t *testing.T) {
	registry := NewRegistry()

	// 创建测试 skills
	skill1 := &Skill{
		Name:     "skill-1",
		Priority: 5,
		Source:   SkillSourceBuiltin,
	}

	skill2 := &Skill{
		Name:     "skill-2",
		Priority: 8,
		Source:   SkillSourceUser,
	}

	// 测试注册
	registry.Register(skill1)
	registry.Register(skill2)

	// 测试查询存在的 skill
	retrieved, err := registry.Get("skill-1")
	if err != nil {
		t.Fatalf("Failed to get skill-1: %v", err)
	}
	if retrieved == nil {
		t.Fatal("Expected skill to be returned, got nil")
	}

	if retrieved.Name != "skill-1" {
		t.Errorf("Expected skill name 'skill-1', got '%s'", retrieved.Name)
	}

	// 测试查询不存在的 skill
	retrieved, err = registry.Get("non-existent")
	if err != nil {
		t.Errorf("Expected no error when getting non-existent skill, got: %v", err)
	}
	if retrieved != nil {
		t.Error("Expected nil skill when getting non-existent skill")
	}

	// 测试列出所有 skills
	allSkills := registry.ListAll()
	if len(allSkills) != 2 {
		t.Errorf("Expected 2 skills, got %d", len(allSkills))
	}
}

// TestRegistryConcurrentSafety 测试并发安全性
func TestRegistryConcurrentSafety(t *testing.T) {
	registry := NewRegistry()

	const numGoroutines = 100
	const numSkills = 50

	// 创建多个 goroutine 进行注册和查询
	var wg sync.WaitGroup

	// 注册 goroutines
	for i := 0; i < numGoroutines/2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < numSkills; j++ {
				skill := &Skill{
					Name:     string(rune('a' + (j % 26))),
					Priority: j,
					Source:   SkillSourceBuiltin,
				}
				registry.Register(skill)
			}
		}(i)
	}

	// 查询 goroutines
	for i := 0; i < numGoroutines/2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < numSkills; j++ {
				skillName := string(rune('a' + (j % 26)))
				_, _ = registry.Get(skillName)
			}
			_ = registry.ListAll()
			_ = registry.ListAvailable(make(map[string]bool))
		}(i)
	}

	wg.Wait()

	// 验证最终状态
	allSkills := registry.ListAll()
	if len(allSkills) == 0 {
		t.Error("Registry should contain skills after concurrent operations")
	}

	t.Logf("Registry contains %d skills after concurrent operations", len(allSkills))
}

// TestRegistryListAvailable 测试过滤禁用的 skills
func TestRegistryListAvailable(t *testing.T) {
	registry := NewRegistry()

	// 注册多个 skills
	skills := []*Skill{
		{Name: "base", Priority: 10, Source: SkillSourceBuiltin},
		{Name: "skill-a", Priority: 5, Source: SkillSourceBuiltin},
		{Name: "skill-b", Priority: 3, Source: SkillSourceUser},
		{Name: "skill-c", Priority: 7, Source: SkillSourceUser},
	}

	for _, skill := range skills {
		registry.Register(skill)
	}

	// 测试无禁用的情况
	disabledMap := make(map[string]bool)
	available := registry.ListAvailable(disabledMap)

	if len(available) != 4 {
		t.Errorf("Expected 4 available skills, got %d", len(available))
	}

	// 测试有禁用的情况
	disabledMap = map[string]bool{
		"skill-a": true,
		"skill-c": true,
	}

	available = registry.ListAvailable(disabledMap)

	if len(available) != 2 {
		t.Errorf("Expected 2 available skills, got %d", len(available))
	}

	// 验证返回的 skills 按优先级排序
	names := make([]string, len(available))
	for i, skill := range available {
		names[i] = skill.Name
	}

	// 应该是 base(10), skill-b(3)
	if names[0] != "base" {
		t.Errorf("Expected first skill to be 'base', got '%s'", names[0])
	}
	if names[1] != "skill-b" {
		t.Errorf("Expected second skill to be 'skill-b', got '%s'", names[1])
	}
}

// TestRegistryDuplicateRegistration 测试重复注册
func TestRegistryDuplicateRegistration(t *testing.T) {
	registry := NewRegistry()

	skill1 := &Skill{
		Name:     "test-skill",
		Priority: 5,
		Source:   SkillSourceBuiltin,
	}

	skill2 := &Skill{
		Name:     "test-skill", // 相同名称
		Priority: 8,
		Source:   SkillSourceUser,
	}

	registry.Register(skill1)
	registry.Register(skill2) // 应该覆盖第一个

	// 验证是第二个 skill
	retrieved, _ := registry.Get("test-skill")

	if retrieved.Priority != 8 {
		t.Errorf("Expected priority 8, got %d", retrieved.Priority)
	}

	if retrieved.Source != SkillSourceUser {
		t.Errorf("Expected source User, got %v", retrieved.Source)
	}
}

// TestRegistryEmpty 测试空注册表
func TestRegistryEmpty(t *testing.T) {
	registry := NewRegistry()

	// 测试查询空的注册表
	skill, err := registry.Get("non-existent")
	if err != nil {
		t.Errorf("Expected no error when getting from empty registry, got: %v", err)
	}
	if skill != nil {
		t.Error("Expected nil skill when getting from empty registry")
	}

	// 测试列出空的注册表
	allSkills := registry.ListAll()
	if len(allSkills) != 0 {
		t.Errorf("Expected empty list, got %d skills", len(allSkills))
	}

	// 测试列出可用的（空）
	available := registry.ListAvailable(make(map[string]bool))
	if len(available) != 0 {
		t.Errorf("Expected empty available list, got %d skills", len(available))
	}
}
