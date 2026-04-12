package todo

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	log "github.com/sirupsen/logrus"

	"msa/pkg/logic/message"
	"msa/pkg/logic/tools/safetool"
	"msa/pkg/model"
)

// FillSummaryParam fill_todo_summary 工具的输入参数
type FillSummaryParam struct {
	TodoPath   string `json:"todo_path" jsonschema:"description=the path of the TODO file"`
	Summary    string `json:"summary" jsonschema:"description=the summary content in markdown format"`
	Conclusion string `json:"conclusion" jsonschema:"description=the final conclusion"`
}

// FillSummaryTool 填写执行总结的工具
type FillSummaryTool struct{}

func (t *FillSummaryTool) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(t.GetName(), t.GetDescription(), FillTodoSummary)
}

func (t *FillSummaryTool) GetName() string {
	return "fill_todo_summary"
}

func (t *FillSummaryTool) GetDescription() string {
	return "填写 TODO 文件的执行总结部分 | Fill the execution summary section of the TODO file"
}

func (t *FillSummaryTool) GetToolGroup() model.ToolGroup {
	return model.TodoToolGroup
}

// FillSummaryData fill_todo_summary 返回数据
type FillSummaryData struct {
	TodoPath string `json:"todo_path"`
	Success  bool   `json:"success"`
	Message  string `json:"message"`
}

// FillTodoSummary 填写执行总结
func FillTodoSummary(ctx context.Context, param *FillSummaryParam) (string, error) {
	return safetool.SafeExecute("fill_todo_summary", param.TodoPath, func() (string, error) {
		return doFillTodoSummary(ctx, param)
	})
}

func doFillTodoSummary(ctx context.Context, param *FillSummaryParam) (string, error) {
	log.Infof("FillTodoSummary start, path: %s", param.TodoPath)
	message.BroadcastToolStart("fill_todo_summary", param.TodoPath)

	if param.TodoPath == "" {
		err := fmt.Errorf("todo_path is required")
		message.BroadcastToolEnd("fill_todo_summary", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	// 解析 TODO 文件获取统计信息
	todo, err := ParseTodoFile(param.TodoPath)
	if err != nil {
		log.Errorf("FillTodoSummary: failed to parse todo: %v", err)
		message.BroadcastToolEnd("fill_todo_summary", "", err)
		return model.NewErrorResult(fmt.Sprintf("解析 TODO 文件失败: %v", err)), nil
	}

	// 构建总结内容
	summaryContent := buildSummaryContent(todo, param.Summary, param.Conclusion)

	// 更新文件
	if err := updateSummarySection(param.TodoPath, summaryContent); err != nil {
		log.Errorf("FillTodoSummary: failed to update summary: %v", err)
		message.BroadcastToolEnd("fill_todo_summary", "", err)
		return model.NewErrorResult(fmt.Sprintf("更新总结失败: %v", err)), nil
	}

	data := &FillSummaryData{
		TodoPath: param.TodoPath,
		Success:  true,
		Message:  "执行总结已填写",
	}

	message.BroadcastToolEnd("fill_todo_summary", "执行总结已填写", nil)

	return model.NewSuccessResult(data, "执行总结已填写"), nil
}

// buildSummaryContent 构建总结内容
func buildSummaryContent(todo *TodoFile, summary, conclusion string) string {
	var sb strings.Builder

	sb.WriteString("## 执行总结\n\n")
	sb.WriteString("### 完成情况\n")
	sb.WriteString(fmt.Sprintf("- 总步骤数：%d\n", todo.StepCount.Total))
	sb.WriteString(fmt.Sprintf("- 成功数：%d\n", todo.StepCount.Done))
	sb.WriteString(fmt.Sprintf("- 失败数：%d\n", todo.StepCount.Failed))
	sb.WriteString(fmt.Sprintf("- 跳过数：%d\n\n", todo.StepCount.Skipped))

	if summary != "" {
		sb.WriteString("### 执行详情\n")
		sb.WriteString(summary)
		sb.WriteString("\n\n")
	}

	sb.WriteString("### 结论\n")
	if conclusion != "" {
		sb.WriteString(conclusion)
	} else {
		sb.WriteString("（执行完成）")
	}
	sb.WriteString("\n")

	return sb.String()
}

// updateSummarySection 更新执行总结部分
func updateSummarySection(path, newContent string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	inSummarySection := false
	summaryStarted := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// 检测执行总结部分的开始
		if strings.HasPrefix(line, "## 执行总结") {
			inSummarySection = true
			summaryStarted = false
			continue
		}

		// 如果在执行总结部分
		if inSummarySection {
			// 检测下一个 ## 标题（非执行总结的子标题）
			if strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ") {
				// 结束执行总结部分，插入新内容
				if !summaryStarted {
					newLines = append(newLines, newContent)
					summaryStarted = true
				}
				inSummarySection = false
				newLines = append(newLines, line)
				continue
			}

			// 跳过原有的执行总结内容，只保留第一次
			if !summaryStarted {
				newLines = append(newLines, newContent)
				summaryStarted = true
			}
			continue
		}

		newLines = append(newLines, line)
	}

	// 如果文件末尾是执行总结
	if inSummarySection && !summaryStarted {
		newLines = append(newLines, newContent)
	}

	// 写回文件
	newFileContent := strings.Join(newLines, "\n")
	return os.WriteFile(path, []byte(newFileContent), 0644)
}

// isSummarySubSection 检查是否是执行总结的子标题
func isSummarySubSection(line string) bool {
	return strings.HasPrefix(line, "### ")
}

// findSummaryEnd 找到执行总结部分的结束位置
func findSummaryEnd(scanner *bufio.Scanner, lines []string, startIndex int) int {
	for i := startIndex + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") && !isSummarySubSection(lines[i]) {
			return i
		}
	}
	return len(lines)
}
