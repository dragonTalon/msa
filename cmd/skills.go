package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"msa/pkg/logic/skills"
)

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "管理 Skills（技能）",
	Long:  `管理 Skills 命令，用于列出、查看、启用和禁用系统技能。`,
	RunE:  runSkills,
}

func init() {
	// 添加子命令
	skillsCmd.AddCommand(skillsListCmd)
	skillsCmd.AddCommand(skillsShowCmd)
	skillsCmd.AddCommand(skillsDisableCmd)
	skillsCmd.AddCommand(skillsEnableCmd)

	// 注册到根命令
	rootCmd.AddCommand(skillsCmd)
}

func runSkills(cmd *cobra.Command, args []string) error {
	// 默认执行 list 命令
	return skillsListCmd.RunE(cmd, args)
}

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
