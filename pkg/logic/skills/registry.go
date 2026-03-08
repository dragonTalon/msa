package skills

import (
	"sort"
	"sync"

	log "github.com/sirupsen/logrus"
)

// Registry 技能注册表
type Registry struct {
	skillsByName map[string]*Skill
	mu           sync.RWMutex
}

// NewRegistry 创建一个新的 Registry
func NewRegistry() *Registry {
	return &Registry{
		skillsByName: make(map[string]*Skill),
	}
}

// Register 注册一个 Skill
func (r *Registry) Register(skill *Skill) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.skillsByName[skill.Name] = skill
	log.Infof("Registered skill: %s (priority: %d, source: %s)", skill.Name, skill.Priority, skill.Source)
}

// Get 根据 name 获取 Skill
func (r *Registry) Get(name string) (*Skill, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skill, exists := r.skillsByName[name]
	if !exists {
		return nil, nil
	}
	return skill, nil
}

// ListAll 返回所有 Skills
func (r *Registry) ListAll() []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skills := make([]*Skill, 0, len(r.skillsByName))
	for _, skill := range r.skillsByName {
		skills = append(skills, skill)
	}

	// 按 priority 降序排序
	sort.Slice(skills, func(i, j int) bool {
		return skills[i].Priority > skills[j].Priority
	})

	return skills
}

// ListAvailable 返回可用的 Skills（排除禁用的）
func (r *Registry) ListAvailable(disabledSkills map[string]bool) []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skills := make([]*Skill, 0, len(r.skillsByName))
	for _, skill := range r.skillsByName {
		if !disabledSkills[skill.Name] {
			skills = append(skills, skill)
		}
	}

	// 按 priority 降序排序
	sort.Slice(skills, func(i, j int) bool {
		return skills[i].Priority > skills[j].Priority
	})

	return skills
}
