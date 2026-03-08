package skills

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Builder Prompt 构建器
type Builder struct {
	registry *Registry
}

// NewBuilder 创建一个新的 Builder
func NewBuilder(registry *Registry) *Builder {
	return &Builder{
		registry: registry,
	}
}

// BuildSystemPrompt 构建完整的 System Prompt
func (b *Builder) BuildSystemPrompt(selectedSkills []string) (string, error) {
	var parts []string

	// 添加基础角色定义和系统配置
	// 注意：这些变量会在运行时被替换
	parts = append(parts, fmt.Sprintf(`【系统配置】
当前实时时间：%%s.%%
今天是：%%s.%%

#【角色定义】
你是专业的%%s，需使用%%s的语气与用户交流，交流内容需贴合股票投资服务场景，保持专业度与亲和力平衡。
`))

	// 添加各个 Skill 的内容
	for _, skillName := range selectedSkills {
		content, err := b.getSkillContent(skillName)
		if err != nil {
			log.Warnf("Failed to get skill content for %s: %v", skillName, err)
			continue
		}
		parts = append(parts, fmt.Sprintf("\n## Skill: %s\n\n%s", skillName, content))
	}

	// 添加基础工具规则
	parts = append(parts, `

## 【工具调用约束】
1. **重点**所有工具名均采用下划线命名法（snake_case）
`)

	return strings.Join(parts, "\n"), nil
}

// getSkillContent 获取 Skill 内容
func (b *Builder) getSkillContent(name string) (string, error) {
	skill, err := b.registry.Get(name)
	if err != nil || skill == nil {
		return "", fmt.Errorf("skill not found: %s", name)
	}

	content, err := skill.GetContent()
	if err != nil {
		return "", fmt.Errorf("failed to load skill content: %w", err)
	}

	return content, nil
}
