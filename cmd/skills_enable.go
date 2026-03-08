package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"msa/pkg/logic/skills"
)

var skillsEnableCmd = &cobra.Command{
	Use:   "enable <name>",
	Short: "启用 Skill",
	Long:  `启用被禁用的 Skill，使其在对话中可以被选择。`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSkillsEnable,
}

func runSkillsEnable(cmd *cobra.Command, args []string) error {
	name := args[0]

	manager := skills.GetManager()

	// 检查 Skill 是否存在
	skill, err := manager.GetSkill(name)
	if err != nil || skill == nil {
		return fmt.Errorf("skill '%s' 不存在", name)
	}

	// 检查是否已启用
	if !manager.IsDisabled(name) {
		fmt.Printf("Skill '%s' 已经是启用状态\n", name)
		return nil
	}

	// 启用 Skill
	if err := manager.EnableSkill(name); err != nil {
		return err
	}

	fmt.Printf("✓ Skill '%s' 已启用\n", name)

	return nil
}
