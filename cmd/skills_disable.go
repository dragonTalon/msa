package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"msa/pkg/logic/skills"
)

var skillsDisableCmd = &cobra.Command{
	Use:   "disable <name>",
	Short: "禁用 Skill",
	Long:  `禁用指定的 Skill，使其在对话中不被自动选择。`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSkillsDisable,
}

func runSkillsDisable(cmd *cobra.Command, args []string) error {
	name := args[0]

	// 检查是否为 base skill
	if name == "base" {
		return fmt.Errorf("不能禁用 'base' skill，它是核心系统规则")
	}

	manager := skills.GetManager()

	// 检查 Skill 是否存在
	skill, err := manager.GetSkill(name)
	if err != nil || skill == nil {
		return fmt.Errorf("skill '%s' 不存在", name)
	}

	// 禁用 Skill
	if err := manager.DisableSkill(name); err != nil {
		return err
	}

	fmt.Printf("✓ Skill '%s' 已禁用\n", name)
	fmt.Printf("提示: 使用 'msa skills enable %s' 可以重新启用\n", name)

	return nil
}
