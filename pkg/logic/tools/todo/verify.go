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

// VerifyTodoParam verify_todo_completion 工具的输入参数
type VerifyTodoParam struct {
	TodoPath string `json:"todo_path" jsonschema:"description=the path of the TODO file to verify"`
}

// VerifyTodoTool 验证 TODO 完成情况的工具
type VerifyTodoTool struct{}

func (t *VerifyTodoTool) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(t.GetName(), t.GetDescription(), VerifyTodoCompletion)
}

func (t *VerifyTodoTool) GetName() string {
	return "verify_todo_completion"
}

func (t *VerifyTodoTool) GetDescription() string {
	return "验证 TODO 文件的完成情况，返回各状态统计和建议 | Verify the completion status of a TODO file"
}

func (t *VerifyTodoTool) GetToolGroup() model.ToolGroup {
	return model.TodoToolGroup
}

// VerifyTodoData verify_todo_completion 返回数据
type VerifyTodoData struct {
	TodoPath         string   `json:"todo_path"`
	AllComplete      bool     `json:"all_complete"`
	HasPending       bool     `json:"has_pending"`
	HasUnhandledFail bool     `json:"has_unhandled_fail"`
	TotalSteps       int      `json:"total_steps"`
	DoneSteps        int      `json:"done_steps"`
	FailedSteps      []string `json:"failed_steps,omitempty"`
	HandledSteps     []string `json:"handled_steps,omitempty"`
	SkippedSteps     []string `json:"skipped_steps,omitempty"`
	PendingSteps     []string `json:"pending_steps,omitempty"`
	Suggestion       string   `json:"suggestion"`
}

// VerifyTodoCompletion 验证 TODO 完成情况
func VerifyTodoCompletion(ctx context.Context, param *VerifyTodoParam) (string, error) {
	return safetool.SafeExecute("verify_todo_completion", param.TodoPath, func() (string, error) {
		return doVerifyTodoCompletion(ctx, param)
	})
}

func doVerifyTodoCompletion(ctx context.Context, param *VerifyTodoParam) (string, error) {
	log.Infof("VerifyTodoCompletion start, path: %s", param.TodoPath)
	message.BroadcastToolStart("verify_todo_completion", param.TodoPath)

	if param.TodoPath == "" {
		err := fmt.Errorf("todo_path is required")
		message.BroadcastToolEnd("verify_todo_completion", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	// 自动补全路径前缀
	todoPath := ResolveTodoPath(param.TodoPath)

	// 解析 TODO 文件
	todo, err := ParseTodoFile(todoPath)
	if err != nil {
		log.Errorf("VerifyTodoCompletion: failed to parse todo: %v", err)
		message.BroadcastToolEnd("verify_todo_completion", "", err)
		return model.NewErrorResult(fmt.Sprintf("解析 TODO 文件失败: %v", err)), nil
	}

	// 获取各状态步骤列表
	failedSteps := todo.GetStepIDs(StatusFailed)
	handledSteps := todo.GetStepIDs(StatusHandled)
	skippedSteps := todo.GetStepIDs(StatusSkipped)
	pendingSteps := todo.GetStepIDs(StatusPending)

	// 生成建议
	suggestion := generateSuggestion(todo)

	data := &VerifyTodoData{
		TodoPath:         todoPath,
		AllComplete:      todo.IsAllComplete(),
		HasPending:       todo.HasPending(),
		HasUnhandledFail: todo.HasUnhandledFail(),
		TotalSteps:       todo.StepCount.Total,
		DoneSteps:        todo.StepCount.Done,
		FailedSteps:      failedSteps,
		HandledSteps:     handledSteps,
		SkippedSteps:     skippedSteps,
		PendingSteps:     pendingSteps,
		Suggestion:       suggestion,
	}

	resultMsg := fmt.Sprintf("验证完成: 总步骤 %d, 成功 %d, 待执行 %d, 失败 %d",
		todo.StepCount.Total, todo.StepCount.Done, todo.StepCount.Pending, todo.StepCount.Failed)
	message.BroadcastToolEnd("verify_todo_completion", resultMsg, nil)

	return model.NewSuccessResult(data, resultMsg), nil
}

// generateSuggestion 生成下一步建议
func generateSuggestion(todo *TodoFile) string {
	if todo.IsAllComplete() {
		return "所有步骤已完成，可以调用 fill_todo_summary 填写执行总结"
	}

	if todo.HasUnhandledFail() {
		failedSteps := todo.GetStepIDs(StatusFailed)
		return fmt.Sprintf("步骤 %v 失败但未处理，请读取失败处理逻辑并执行", failedSteps)
	}

	if todo.HasPending() {
		pendingSteps := todo.GetStepIDs(StatusPending)
		return fmt.Sprintf("以下步骤待执行: %v", pendingSteps)
	}

	return "请继续执行"
}
