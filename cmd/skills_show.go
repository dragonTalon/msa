package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"msa/pkg/logic/skills"
)

var skillsShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "查看 Skill 详情",
	Long:  `查看指定 Skill 的详细信息，包括元数据、内容、来源目录和状态。`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSkillsShow,
}

func runSkillsShow(cmd *cobra.Command, args []string) error {
	name := args[0]

	manager := skills.GetManager()

	// 初始化 Skills
	if err := manager.Initialize(); err != nil {
		return err
	}

	// 获取 Skill
	skill, err := manager.GetSkill(name)
	if err != nil || skill == nil {
		fmt.Printf("错误: Skill '%s' 不存在\n", name)
		fmt.Println("\n可用的 Skills:")
		if err := listAvailableSkillsBrief(manager); err != nil {
			return err
		}
		return fmt.Errorf("skill not found: %s", name)
	}

	// 显示详细信息
	fmt.Printf("=== Skill 详情: %s ===\n\n", skill.Name)

	// 元数据
	fmt.Println("## 元数据")
	fmt.Printf("名称: %s\n", skill.Name)
	fmt.Printf("描述: %s\n", skill.Description)
	fmt.Printf("版本: %s\n", skill.Version)
	fmt.Printf("优先级: %s\n", formatPriority(skill.Priority))
	fmt.Printf("来源: %s\n", formatSource(skill.Source))
	fmt.Printf("路径: %s\n", skill.GetPath())

	// 状态
	disabled := manager.GetDisabledSkills()
	fmt.Printf("状态: %s\n", formatStatus(skill.Name, disabled))

	// 内容
	fmt.Println("\n## 内容")
	content, err := skill.GetContent()
	if err != nil {
		return fmt.Errorf("无法加载 Skill 内容: %w", err)
	}

	if content == "" {
		fmt.Println("(无内容)")
	} else {
		// 显示前 500 个字符的预览
		preview := content
		if len(preview) > 500 {
			preview = preview[:500] + "\n... (内容被截断)"
		}
		fmt.Println(preview)
	}

	fmt.Printf("\n总字符数: %d\n", len(content))

	return nil
}

func listAvailableSkillsBrief(manager *skills.Manager) error {
	allSkills := manager.ListSkills()

	if len(allSkills) == 0 {
		fmt.Println("  (没有可用的 Skills)")
		return nil
	}

	for _, skill := range allSkills {
		fmt.Printf("  - %s (优先级: %d)\n", skill.Name, skill.Priority)
	}

	return nil
}
