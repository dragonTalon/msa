package todo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	log "github.com/sirupsen/logrus"

	"msa/pkg/logic/message"
	"msa/pkg/logic/skills"
	"msa/pkg/model"
	"msa/pkg/session"
)

// CreateTodoParam create_todo 工具的输入参数
type CreateTodoParam struct {
	SkillName string `json:"skill_name" jsonschema:"description=the name of the skill to create TODO for"`
}

// CreateTodoTool 创建 TODO 文件的工具
type CreateTodoTool struct{}

func (t *CreateTodoTool) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(t.GetName(), t.GetDescription(), CreateTodo)
}

func (t *CreateTodoTool) GetName() string {
	return "create_todo"
}

func (t *CreateTodoTool) GetDescription() string {
	return "为指定 Skill 创建 TODO 文件，记录执行步骤 | Create a TODO file for the specified skill to track execution steps"
}

func (t *CreateTodoTool) GetToolGroup() model.ToolGroup {
	return model.TodoToolGroup
}

// CreateTodoData create_todo 返回数据
type CreateTodoData struct {
	SkillName  string `json:"skill_name"`
	TodoPath   string `json:"todo_path"`
	Content    string `json:"content"`
	TotalSteps int    `json:"total_steps"`
}

// CreateTodo 创建 TODO 文件
func CreateTodo(ctx context.Context, param *CreateTodoParam) (string, error) {
	log.Infof("CreateTodo start, skill_name: %s", param.SkillName)
	message.BroadcastToolStart("create_todo", fmt.Sprintf("skill_name: %s", param.SkillName))

	if param.SkillName == "" {
		err := fmt.Errorf("skill_name is required")
		message.BroadcastToolEnd("create_todo", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	// 获取 Skill
	manager := skills.GetManager()
	sk, err := manager.GetSkill(param.SkillName)
	if err != nil {
		log.Errorf("CreateTodo: failed to get skill %s: %v", param.SkillName, err)
		message.BroadcastToolEnd("create_todo", "", err)
		return model.NewErrorResult(err.Error()), nil
	}
	if sk == nil {
		err = fmt.Errorf("skill '%s' not found", param.SkillName)
		log.Warnf("CreateTodo: %v", err)
		message.BroadcastToolEnd("create_todo", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	// 获取当前 Session
	sessionMgr := session.GetManager()
	currentSession := sessionMgr.Current()
	if currentSession == nil {
		err = fmt.Errorf("no active session")
		log.Warnf("CreateTodo: %v", err)
		message.BroadcastToolEnd("create_todo", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	// 构建 TODO 目录路径
	todoDir := filepath.Join(
		filepath.Dir(sessionMgr.GetMemoryDir()),
		"todos",
		currentSession.SessionID(),
	)

	// 确保目录存在
	if err := os.MkdirAll(todoDir, 0755); err != nil {
		log.Errorf("CreateTodo: failed to create todo directory: %v", err)
		message.BroadcastToolEnd("create_todo", "", err)
		return model.NewErrorResult(fmt.Sprintf("创建 TODO 目录失败: %v", err)), nil
	}

	// 获取模板内容
	templateContent, err := getTemplateContent(sk)
	if err != nil {
		log.Errorf("CreateTodo: failed to get template: %v", err)
		message.BroadcastToolEnd("create_todo", "", err)
		return model.NewErrorResult(fmt.Sprintf("获取模板失败: %v", err)), nil
	}

	// 替换模板变量
	content := replaceTemplateVariables(templateContent, param.SkillName, currentSession)

	// 创建 TODO 文件
	todoPath := filepath.Join(todoDir, param.SkillName+".md")
	if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
		log.Errorf("CreateTodo: failed to write todo file: %v", err)
		message.BroadcastToolEnd("create_todo", "", err)
		return model.NewErrorResult(fmt.Sprintf("写入 TODO 文件失败: %v", err)), nil
	}

	// 解析统计步骤数
	todo, err := ParseTodoContent(todoPath, content)
	if err != nil {
		log.Warnf("CreateTodo: failed to parse todo: %v", err)
	}

	totalSteps := 0
	if todo != nil {
		totalSteps = todo.StepCount.Total
	}

	data := &CreateTodoData{
		SkillName:  param.SkillName,
		TodoPath:   todoPath,
		Content:    content,
		TotalSteps: totalSteps,
	}

	resultMsg := fmt.Sprintf("创建 TODO 文件: %s (%d 步骤)", todoPath, totalSteps)
	message.BroadcastToolEnd("create_todo", resultMsg, nil)

	return model.NewSuccessResult(data, resultMsg), nil
}

// getTemplateContent 获取模板内容
func getTemplateContent(sk *skills.Skill) (string, error) {
	// 优先使用 Skill 自定义模板
	if sk.HasTodoTemplate() {
		content, err := sk.GetTodoTemplate()
		if err != nil {
			log.Warnf("getTemplateContent: failed to get skill template: %v, using default", err)
		} else {
			return content, nil
		}
	}

	// 使用默认模板
	return getDefaultTemplate()
}

// getDefaultTemplate 获取默认模板
func getDefaultTemplate() (string, error) {
	// 获取默认模板路径
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("获取可执行文件路径失败: %w", err)
	}

	// 尝试多个可能的路径
	possiblePaths := []string{
		filepath.Join(filepath.Dir(execPath), "pkg", "logic", "tools", "todo", "default-template.md"),
		filepath.Join("pkg", "logic", "tools", "todo", "default-template.md"),
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			content, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			return string(content), nil
		}
	}

	// 返回内嵌的默认模板
	return getEmbeddedDefaultTemplate(), nil
}

// getEmbeddedDefaultTemplate 返回内嵌的默认模板
func getEmbeddedDefaultTemplate() string {
	return `# TODO: {skill-name}

> 创建时间：{timestamp}
> 会话ID：{session-id}

---

## 执行状态
- [ ] 所有步骤完成
- [ ] 已验证执行结果

---

## 1. 主任务

### 步骤
- [ ] 1.1 执行步骤描述
- [ ] 1.2 执行步骤描述

### 失败处理
- 如果 1.1 失败：跳过继续执行
- 如果 1.2 失败：中止并提示用户

---

## 执行总结

### 完成情况
- 总步骤数：
- 成功数：
- 失败数：
- 跳过数：

### 结论
（执行完成后填写）
`
}

// replaceTemplateVariables 替换模板变量
func replaceTemplateVariables(template, skillName string, sess *session.Session) string {
	now := time.Now().Format("2006-01-02 15:04:05")

	result := strings.ReplaceAll(template, "{skill-name}", skillName)
	result = strings.ReplaceAll(result, "{timestamp}", now)
	result = strings.ReplaceAll(result, "{session-id}", sess.SessionID())

	return result
}
