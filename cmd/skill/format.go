package cmd_skill

import (
	"fmt"

	"msa/pkg/logic/skills"
)

// formatPriority 格式化优先级显示
func formatPriority(priority int) string {
	switch {
	case priority >= 9:
		return fmt.Sprintf("%d (最高)", priority)
	case priority >= 7:
		return fmt.Sprintf("%d (高)", priority)
	case priority >= 5:
		return fmt.Sprintf("%d (中)", priority)
	case priority >= 3:
		return fmt.Sprintf("%d (低)", priority)
	default:
		return fmt.Sprintf("%d (最低)", priority)
	}
}

// formatSource 格式化来源显示
func formatSource(source skills.SkillSource) string {
	return source.String()
}

// formatStatus 格式化状态显示
func formatStatus(name string, disabled []string) string {
	for _, d := range disabled {
		if d == name {
			return "禁用"
		}
	}
	return "启用"
}