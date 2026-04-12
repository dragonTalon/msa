package todo

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	log "github.com/sirupsen/logrus"

	"msa/pkg/logic/message"
	"msa/pkg/logic/tools/safetool"
	"msa/pkg/model"
)

// UpdateTodoParam update_todo_step 工具的输入参数
type UpdateTodoParam struct {
	TodoPath string `json:"todo_path" jsonschema:"description=the path of the TODO file"`
	StepID   string `json:"step_id" jsonschema:"description=the step ID to update (e.g. 1.1, 2.3)"`
	Status   string `json:"status" jsonschema:"description=the new status: done, failed, handled, skipped"`
	Note     string `json:"note,omitempty" jsonschema:"description=optional note about the status change"`
}

// UpdateTodoTool 更新 TODO 步骤状态的工具
type UpdateTodoTool struct{}

func (t *UpdateTodoTool) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(t.GetName(), t.GetDescription(), UpdateTodoStep)
}

func (t *UpdateTodoTool) GetName() string {
	return "update_todo_step"
}

func (t *UpdateTodoTool) GetDescription() string {
	return "更新 TODO 文件中指定步骤的状态 | Update the status of a step in the TODO file"
}

func (t *UpdateTodoTool) GetToolGroup() model.ToolGroup {
	return model.TodoToolGroup
}

// UpdateTodoData update_todo_step 返回数据
type UpdateTodoData struct {
	TodoPath  string `json:"todo_path"`
	StepID    string `json:"step_id"`
	Status    string `json:"status"`
	NewStatus string `json:"new_status"` // 更新后的状态标记
}

// UpdateTodoStep 更新 TODO 步骤状态
func UpdateTodoStep(ctx context.Context, param *UpdateTodoParam) (string, error) {
	return safetool.SafeExecute("update_todo_step", fmt.Sprintf("step: %s, status: %s", param.StepID, param.Status), func() (string, error) {
		return doUpdateTodoStep(ctx, param)
	})
}

func doUpdateTodoStep(ctx context.Context, param *UpdateTodoParam) (string, error) {
	log.Infof("UpdateTodoStep start, path: %s, step: %s, status: %s", param.TodoPath, param.StepID, param.Status)
	message.BroadcastToolStart("update_todo_step", fmt.Sprintf("step: %s, status: %s", param.StepID, param.Status))

	if param.TodoPath == "" {
		err := fmt.Errorf("todo_path is required")
		message.BroadcastToolEnd("update_todo_step", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	if param.StepID == "" {
		err := fmt.Errorf("step_id is required")
		message.BroadcastToolEnd("update_todo_step", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	if param.Status == "" {
		err := fmt.Errorf("status is required")
		message.BroadcastToolEnd("update_todo_step", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	// 验证状态值
	var stepStatus StepStatus
	switch param.Status {
	case "done":
		stepStatus = StatusDone
	case "failed":
		stepStatus = StatusFailed
	case "handled":
		stepStatus = StatusHandled
	case "skipped":
		stepStatus = StatusSkipped
	default:
		err := fmt.Errorf("invalid status: %s (must be one of: done, failed, handled, skipped)", param.Status)
		message.BroadcastToolEnd("update_todo_step", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	// 更新步骤状态
	if err := UpdateStepStatus(param.TodoPath, param.StepID, stepStatus, param.Note); err != nil {
		log.Errorf("UpdateTodoStep: failed to update step: %v", err)
		message.BroadcastToolEnd("update_todo_step", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	data := &UpdateTodoData{
		TodoPath:  param.TodoPath,
		StepID:    param.StepID,
		Status:    param.Status,
		NewStatus: string(stepStatus),
	}

	resultMsg := fmt.Sprintf("步骤 %s 状态更新为: %s", param.StepID, param.Status)
	message.BroadcastToolEnd("update_todo_step", resultMsg, nil)

	return model.NewSuccessResult(data, resultMsg), nil
}
