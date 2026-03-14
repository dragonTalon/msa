package cmd_skill

import (
	"fmt"

	"github.com/spf13/cobra"

	"msa/pkg/logic/skills"
)

// NewCommand 创建 skills 子命令
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skills",
		Short: "管理 Skills（技能）",
		Long:  `管理 Skills 命令，用于列出、查看、启用和禁用系统技能。`,
		RunE:  runSkills,
	}

	// 添加子命令
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newShowCmd())
	cmd.AddCommand(newDisableCmd())
	cmd.AddCommand(newEnableCmd())

	return cmd
}

func runSkills(cmd *cobra.Command, args []string) error {
	// 默认执行 list 命令
	return listRunE(cmd, args)
}

// listAvailableSkillsBrief 简要列出可用的 Skills
func listAvailableSkillsBrief(manager *skills.Manager) error {
	allSkills := manager.ListSkills()

	if len(allSkills) == 0 {
		return nil
	}

	for _, skill := range allSkills {
		fmt.Printf("  - %s (优先级: %d)\n", skill.Name, skill.Priority)
	}

	return nil
}