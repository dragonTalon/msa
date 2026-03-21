package skills

import (
	"context"
	"fmt"
	"msa/pkg/utils"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"

	"msa/pkg/model"

	log "github.com/sirupsen/logrus"
)

// SkillInfo 用于发送给 LLM 的轻量级 Skill 信息
type SkillInfo struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	Priority      int    `json:"priority"`
	Pattern       string `json:"pattern"`
	HasReferences bool   `json:"has_references"`
	HasAssets     bool   `json:"has_assets"`
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
// history: 对话历史，用于让 LLM 理解上下文（如"它"指代什么）
func (s *Selector) SelectSkills(ctx context.Context, userInput string, history []model.Message) (*SelectionResult, error) {
	// 获取可用 skills（过滤禁用的）
	disabledMap := getDisabledSkillsMap()
	skills := s.registry.ListAvailable(disabledMap)

	// 构建轻量级 metadata 列表
	metadataList := make([]SkillInfo, len(skills))
	for i, skill := range skills {
		metadataList[i] = SkillInfo{
			Name:          skill.Name,
			Description:   skill.Description,
			Priority:      skill.Priority,
			Pattern:       string(skill.Metadata.Pattern),
			HasReferences: skill.HasReferences(),
			HasAssets:     skill.HasAssets(),
		}
	}

	log.Infof("Skill Selector: selecting skills for input: %s", truncate(userInput, 50))
	log.Debugf("Available skills: %v", formatSkillsForLog(metadataList))

	// 构建选择 Prompt（包含最近的对话历史作为上下文）
	promptMessages, err := s.buildSelectionPrompt(ctx, metadataList, userInput, history)
	if err != nil {
		log.Warnf("Failed to build selection prompt, using base skill: %v", err)
		return &SelectionResult{
			SelectedSkills: []string{},
			Reasoning:      "Prompt build failed, using fallback",
		}, nil
	}

	// 调用 LLM
	response, err := s.callLLM(ctx, promptMessages)
	if err != nil {
		log.Warnf("LLM selection failed, using base skill: %v", err)
		return &SelectionResult{
			SelectedSkills: []string{},
			Reasoning:      "LLM selection failed, using fallback",
		}, nil
	}

	// 解析结果
	var result SelectionResult
	if err := utils.ParseJSON(response, &result); err != nil {
		log.Warnf("Failed to parse LLM response, using base skill: %v", err)
		return &SelectionResult{
			SelectedSkills: []string{},
			Reasoning:      "Parse failed, using fallback",
		}, nil
	}
	if len(result.SelectedSkills) == 0 {
		log.Warnf("LLM response is empty, using base skill %s", utils.ToJSONString(result))
	}

	log.Infof("Skill Selector: selected skills: %v", len(result.SelectedSkills))
	log.Debugf("Reasoning: %s", result.Reasoning)

	// 按 priority 排序
	result.SelectedSkills = s.sortByPriority(result.SelectedSkills, skills)

	return &result, nil
}

// selectionSystemTemplate 系统消息模板
const selectionSystemTemplate = `你是一个 skill 选择器。根据用户输入和可用的 skills，选择最合适的 skills。

## 可用 Skills

{{range .skills}}### {{.Name}} (优先级: {{.Priority}})
{{.Description}}

{{end}}{{if .history}}## 最近对话上下文

{{range .history}}{{if eq .Role "user"}}用户: {{.Content}}
{{else if eq .Role "assistant"}}助手: {{.Content}}
{{end}}{{end}}
{{end}}## 任务

分析用户输入，返回应该激活的 skills（JSON格式）：
` + "```json" + `
{"selected_skills": ["skill-name-1", "skill-name-2"], "reasoning": "选择理由（中文）"}
` + "```" + `

## 规则

1. 可以选择多个 skills
2. 如果没有匹配的 skill，只返回 [""]
3. 高优先级的 skill 优先选择`

// buildSelectionPrompt 使用 eino prompt.FromMessages 构建选择 Prompt
func (s *Selector) buildSelectionPrompt(ctx context.Context, metadatas []SkillInfo, userInput string, history []model.Message) ([]*schema.Message, error) {
	hasHistory := len(history) > 0
	template := prompt.FromMessages(schema.GoTemplate,
		schema.SystemMessage(selectionSystemTemplate),
		schema.UserMessage("## 用户输入\n\n{{.user_input}}"),
	)

	// 截取最近 6 条历史，并转换为模板可用的结构
	historySnippet := buildHistorySnippet(history)

	return template.Format(ctx, map[string]any{
		"skills":      metadatas,
		"history":     historySnippet,
		"has_history": hasHistory,
		"user_input":  userInput,
	})
}

// historyItem 用于模板渲染的对话历史条目
type historyItem struct {
	Role    string
	Content string
}

// buildHistorySnippet 截取最近 6 条历史并转换为模板可用结构
func buildHistorySnippet(history []model.Message) []historyItem {
	if len(history) == 0 {
		return nil
	}
	startIdx := 0
	if len(history) > 6 {
		startIdx = len(history) - 6
	}
	items := make([]historyItem, 0, len(history)-startIdx)
	for _, msg := range history[startIdx:] {
		switch msg.Role {
		case model.RoleUser:
			items = append(items, historyItem{Role: "user", Content: truncate(msg.Content, 200)})
		case model.RoleAssistant:
			items = append(items, historyItem{Role: "assistant", Content: truncate(msg.Content, 200)})
		}
	}
	return items
}

// callLLM 调用 LLM
func (s *Selector) callLLM(ctx context.Context, messages []*schema.Message) (string, error) {
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

func formatSkillsForLog(metadatas []SkillInfo) string {
	names := make([]string, len(metadatas))
	for i, meta := range metadatas {
		names[i] = fmt.Sprintf("%s(%d)", meta.Name, meta.Priority)
	}
	return strings.Join(names, ", ")
}
