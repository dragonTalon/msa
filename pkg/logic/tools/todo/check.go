package todo

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	log "github.com/sirupsen/logrus"

	"msa/pkg/logic/message"
	"msa/pkg/logic/skills"
	"msa/pkg/logic/tools/safetool"
	"msa/pkg/model"
)

// CheckTodoParam check_skill_todo 工具的输入参数
type CheckTodoParam struct {
	SkillName string `json:"skill_name" jsonschema:"description=the name of the skill to check"`
}

// CheckTodoTool 检查 Skill 是否需要 TODO 的工具
type CheckTodoTool struct{}

func (t *CheckTodoTool) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(t.GetName(), t.GetDescription(), CheckSkillTodo)
}

func (t *CheckTodoTool) GetName() string {
	return "check_skill_todo"
}

func (t *CheckTodoTool) GetDescription() string {
	return "检查指定 Skill 是否需要创建 TODO 列表 | Check if a skill requires creating a TODO list"
}

func (t *CheckTodoTool) GetToolGroup() model.ToolGroup {
	return model.TodoToolGroup
}

// CheckTodoData check_skill_todo 返回数据
type CheckTodoData struct {
	SkillName         string `json:"skill_name"`
	RequiresTodo      bool   `json:"requires_todo"`
	HasCustomTemplate bool   `json:"has_custom_template"`
	TemplateSource    string `json:"template_source"` // "skill" | "default"
}

// CheckSkillTodo 检查 Skill 是否需要创建 TODO
func CheckSkillTodo(ctx context.Context, param *CheckTodoParam) (string, error) {
	return safetool.SafeExecute("check_skill_todo", fmt.Sprintf("skill_name: %s", param.SkillName), func() (string, error) {
		return doCheckSkillTodo(ctx, param)
	})
}

func doCheckSkillTodo(ctx context.Context, param *CheckTodoParam) (string, error) {
	log.Infof("CheckSkillTodo start, skill_name: %s", param.SkillName)
	message.BroadcastToolStart("check_skill_todo", fmt.Sprintf("skill_name: %s", param.SkillName))

	if param.SkillName == "" {
		err := fmt.Errorf("skill_name is required")
		message.BroadcastToolEnd("check_skill_todo", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	manager := skills.GetManager()
	sk, err := manager.GetSkill(param.SkillName)
	if err != nil {
		log.Errorf("CheckSkillTodo: failed to get skill %s: %v", param.SkillName, err)
		message.BroadcastToolEnd("check_skill_todo", "", err)
		return model.NewErrorResult(err.Error()), nil
	}
	if sk == nil {
		err = fmt.Errorf("skill '%s' not found", param.SkillName)
		log.Warnf("CheckSkillTodo: %v", err)
		message.BroadcastToolEnd("check_skill_todo", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	// 检查是否需要
	requiresTodo := sk.Metadata.RequiresTodo
	hasCustomTemplate := sk.HasTodoTemplate()
	templateSource := "default"
	if hasCustomTemplate {
		templateSource = "skill"
	}

	data := &CheckTodoData{
		SkillName:         param.SkillName,
		RequiresTodo:      requiresTodo,
		HasCustomTemplate: hasCustomTemplate,
		TemplateSource:    templateSource,
	}

	resultMsg := fmt.Sprintf("Skill %s: requires_todo=%v, has_custom_template=%v",
		param.SkillName, requiresTodo, hasCustomTemplate)
	message.BroadcastToolEnd("check_skill_todo", resultMsg, nil)

	return model.NewSuccessResult(data, resultMsg), nil
}
