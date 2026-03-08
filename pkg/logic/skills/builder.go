package skills

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"
)

// PromptVars 系统 Prompt 中的模板变量
type PromptVars struct {
	Time    string // 当前实时时间
	Weekday string // 今天星期几
	Role    string // 角色定义（如 "专业股票分析助手"）
	Style   string // 语气风格（如 "理性、专业、客观且严谨"）
}

// systemPromptTemplate 系统配置和角色定义的模板
const systemPromptTemplate = `【系统配置】
当前实时时间：{{.Time}}（系统实时时间，请以此为准）
今天是：{{.Weekday}}
**必须**: 你对时间的理解需以此为准,同时你的判断的时间基准也以此为准

#【角色定义】
你是专业的{{.Role}}，需使用{{.Style}}的语气与用户交流，交流内容需贴合股票投资服务场景，保持专业度与亲和力平衡。
`

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
// vars: 模板变量，用于填充系统配置和角色定义中的动态内容
func (b *Builder) BuildSystemPrompt(selectedSkills []string, vars PromptVars) (string, error) {
	var parts []string

	// 渲染基础角色定义和系统配置模板
	tmpl, err := template.New("system").Parse(systemPromptTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse system prompt template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", fmt.Errorf("failed to render system prompt template: %w", err)
	}
	parts = append(parts, buf.String())

	// 构建 Skill List 汇总
	type skillEntry struct {
		name string
		desc string
	}
	var entries []skillEntry
	for _, skillName := range selectedSkills {
		skill, err := b.registry.Get(skillName)
		if err != nil || skill == nil {
			log.Warnf("Skill not found: %s", skillName)
			continue
		}
		entries = append(entries, skillEntry{
			name: skill.Name,
			desc: skill.Description,
		})
	}

	if len(entries) > 0 {
		// 输出 Skill List 汇总
		var listBuf strings.Builder
		listBuf.WriteString("## Skill List\n")
		for _, e := range entries {
			listBuf.WriteString(fmt.Sprintf("- name: %s\n  describe: %s\n", e.name, e.desc))
		}
		parts = append(parts, listBuf.String())
	}

	// 添加基础工具规则
	parts = append(parts, `
## 【工具调用约束】
1. **重点**所有工具名均采用下划线命名法（snake_case）
`)

	return strings.Join(parts, "\n"), nil
}
