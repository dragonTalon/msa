package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"

	log "github.com/sirupsen/logrus"
)

// SkillMetadata 用于发送给 LLM 的轻量级 Skill 信息
type SkillMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
}

// SelectionResult LLM 选择结果
type SelectionResult struct {
	SelectedSkills []string `json:"selected_skills"`
	Reasoning      string   `json:"reasoning"`
}

// Selector LLM 自动选择器
type Selector struct {
	llm      *openai.ChatModel
	registry *Registry
}

// NewSelector 创建一个新的 Selector
func NewSelector(llm *openai.ChatModel, registry *Registry) *Selector {
	return &Selector{
		llm:      llm,
		registry: registry,
	}
}

// SelectSkills 使用 LLM 选择合适的 Skills
func (s *Selector) SelectSkills(ctx context.Context, userInput string) (*SelectionResult, error) {
	// 获取可用 skills（过滤禁用的）
	disabledMap := getDisabledSkillsMap()
	skills := s.registry.ListAvailable(disabledMap)

	// 构建轻量级 metadata 列表
	metadatas := make([]SkillMetadata, len(skills))
	for i, skill := range skills {
		metadatas[i] = SkillMetadata{
			Name:        skill.Name,
			Description: skill.Description,
			Priority:    skill.Priority,
		}
	}

	log.Infof("Skill Selector: selecting skills for input: %s", truncate(userInput, 50))
	log.Debugf("Available skills: %v", formatSkillsForLog(metadatas))

	// 构建选择 Prompt
	prompt := s.buildSelectionPrompt(metadatas, userInput)

	// 调用 LLM
	response, err := s.callLLM(ctx, prompt)
	if err != nil {
		log.Warnf("LLM selection failed, using base skill: %v", err)
		return &SelectionResult{
			SelectedSkills: []string{"base"},
			Reasoning:      "LLM selection failed, using fallback",
		}, nil
	}

	// 解析结果
	var result SelectionResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		log.Warnf("Failed to parse LLM response, using base skill: %v", err)
		return &SelectionResult{
			SelectedSkills: []string{"base"},
			Reasoning:      "Parse failed, using fallback",
		}, nil
	}

	// 确保 base skill 总是被选中
	result.SelectedSkills = s.ensureBaseSkill(result.SelectedSkills)

	// 按 priority 排序
	result.SelectedSkills = s.sortByPriority(result.SelectedSkills, skills)

	log.Infof("Skill Selector: selected skills: %v", result.SelectedSkills)
	log.Debugf("Reasoning: %s", result.Reasoning)

	return &result, nil
}

// buildSelectionPrompt 构建选择 Prompt
func (s *Selector) buildSelectionPrompt(metadatas []SkillMetadata, userInput string) string {
	var sb strings.Builder

	sb.WriteString("你是一个 skill 选择器。根据用户输入和可用的 skills，选择最合适的 skills。\n\n")
	sb.WriteString("## 可用 Skills\n\n")

	for _, meta := range metadatas {
		sb.WriteString(fmt.Sprintf("### %s (优先级: %d)\n", meta.Name, meta.Priority))
		sb.WriteString(fmt.Sprintf("%s\n\n", meta.Description))
	}

	sb.WriteString("## 用户输入\n\n")
	sb.WriteString(userInput)
	sb.WriteString("\n\n")

	sb.WriteString("## 任务\n\n")
	sb.WriteString("分析用户输入，返回应该激活的 skills（JSON格式）：\n")
	sb.WriteString("```json\n")
	sb.WriteString(`{"selected_skills": ["skill-name-1", "skill-name-2"], "reasoning": "选择理由（中文）"}`)
	sb.WriteString("\n```\n\n")

	sb.WriteString("## 规则\n\n")
	sb.WriteString("1. 总是包含 \"base\" skill（基础系统规则）\n")
	sb.WriteString("2. 可以选择多个 skills\n")
	sb.WriteString("3. 如果没有匹配的 skill，只返回 [\"base\"]\n")
	sb.WriteString("4. 高优先级的 skill 优先选择\n")

	return sb.String()
}

// callLLM 调用 LLM
func (s *Selector) callLLM(ctx context.Context, prompt string) (string, error) {
	// 构建消息
	messages := []*schema.Message{
		{
			Role:    schema.System,
			Content: "You are a skill selection assistant. Respond only with valid JSON.",
		},
		{
			Role:    schema.User,
			Content: prompt,
		},
	}

	// 调用 LLM
	response, err := s.llm.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("LLM generate failed: %w", err)
	}

	// 提取内容
	if response == nil {
		return "", fmt.Errorf("empty LLM response")
	}

	content := response.Content
	if content == "" {
		return "", fmt.Errorf("empty content in LLM response")
	}

	// 清理可能的 markdown 代码块标记
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	return content, nil
}

// ensureBaseSkill 确保 base skill 总是被选中
func (s *Selector) ensureBaseSkill(skills []string) []string {
	for _, skill := range skills {
		if skill == "base" {
			return skills
		}
	}

	// base 不在列表中，添加到开头
	return append([]string{"base"}, skills...)
}

// sortByPriority 按 priority 降序排序
func (s *Selector) sortByPriority(skillNames []string, allSkills []*Skill) []string {
	// 创建 priority 映射
	priorityMap := make(map[string]int)
	for _, skill := range allSkills {
		priorityMap[skill.Name] = skill.Priority
	}

	// 排序
	sorted := make([]string, len(skillNames))
	copy(sorted, skillNames)

	// 简单冒泡排序（因为列表通常很短）
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if priorityMap[sorted[i]] < priorityMap[sorted[j]] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

// 辅助函数
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func formatSkillsForLog(metadatas []SkillMetadata) string {
	names := make([]string, len(metadatas))
	for i, meta := range metadatas {
		names[i] = fmt.Sprintf("%s(%d)", meta.Name, meta.Priority)
	}
	return strings.Join(names, ", ")
}
