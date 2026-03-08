package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"msa/pkg/logic/skills"
)

var (
	showBuiltin   bool
	showUser      bool
	outputJSONFmt bool
)

var skillsListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有 Skills",
	Long:  `列出所有可用的 Skills，包括内置 Skills 和用户自定义 Skills。`,
	RunE:  runSkillsList,
}

func init() {
	skillsListCmd.Flags().BoolVar(&showBuiltin, "builtin", false, "只显示内置 Skills")
	skillsListCmd.Flags().BoolVar(&showUser, "user", false, "只显示用户自定义 Skills")
	skillsListCmd.Flags().BoolVar(&outputJSONFmt, "json", false, "以 JSON 格式输出")
}

func runSkillsList(cmd *cobra.Command, args []string) error {
	manager := skills.GetManager()

	// 初始化 Skills
	if err := manager.Initialize(); err != nil {
		return err
	}

	// 获取禁用的 Skills
	disabled := manager.GetDisabledSkills()

	// 获取所有 Skills
	var allSkills []*skills.Skill
	if showBuiltin {
		// 只显示内置 Skills
		allSkills = filterSkillsBySource(manager.ListAllSkills(), skills.SkillSourceBuiltin)
	} else if showUser {
		// 只显示用户 Skills
		allSkills = filterSkillsBySource(manager.ListAllSkills(), skills.SkillSourceUser)
	} else {
		// 显示所有 Skills
		allSkills = manager.ListAllSkills()
	}

	// 输出
	if outputJSONFmt {
		return outputJSONFormat(allSkills, disabled)
	}

	return outputTable(allSkills, disabled)
}

func filterSkillsBySource(allSkills []*skills.Skill, source skills.SkillSource) []*skills.Skill {
	var filtered []*skills.Skill
	for _, skill := range allSkills {
		if skill.Source == source {
			filtered = append(filtered, skill)
		}
	}
	return filtered
}

func outputTable(allSkills []*skills.Skill, disabled []string) error {
	if len(allSkills) == 0 {
		fmt.Println("没有找到任何 Skills。")
		return nil
	}

	// 表格头
	fmt.Printf("%-20s %-10s %-10s %-10s\n", "Name", "Priority", "Source", "Status")
	fmt.Println("─────────────────────────────────────────────────────────")

	// 表格内容
	for _, skill := range allSkills {
		name := skill.Name
		priority := formatPriority(skill.Priority)
		source := formatSource(skill.Source)
		status := formatStatus(skill.Name, disabled)

		fmt.Printf("%-20s %-10s %-10s %-10s\n", name, priority, source, status)
	}

	fmt.Printf("\n总计: %d 个 Skills\n", len(allSkills))
	return nil
}

func outputJSONFormat(allSkills []*skills.Skill, disabled []string) error {
	output := struct {
		Skills   []skillJSON `json:"skills"`
		Disabled []string    `json:"disabled"`
	}{
		Skills:   make([]skillJSON, 0, len(allSkills)),
		Disabled: disabled,
	}

	for _, skill := range allSkills {
		output.Skills = append(output.Skills, skillJSON{
			Name:        skill.Name,
			Description: skill.Description,
			Version:     skill.Version,
			Priority:    skill.Priority,
			Source:      skill.Source.String(),
		})
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

type skillJSON struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Priority    int    `json:"priority"`
	Source      string `json:"source"`
}
