package todo

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// StepStatus 步骤状态
type StepStatus string

const (
	StatusPending StepStatus = "pending" // [ ] 待执行
	StatusDone    StepStatus = "done"    // [x] 成功
	StatusFailed  StepStatus = "failed"  // [!] 失败
	StatusHandled StepStatus = "handled" // [-] 已处理
	StatusSkipped StepStatus = "skipped" // [~] 跳过
)

// Step 表示 TODO 文件中的一个步骤
type Step struct {
	ID     string     // 步骤 ID，如 "1.1", "2.3"
	Status StepStatus // 步骤状态
	Line   int        // 行号
	Text   string     // 步骤文本
}

// TodoFile 表示解析后的 TODO 文件
type TodoFile struct {
	Path      string  // 文件路径
	Content   string  // 原始内容
	Steps     []*Step // 所有步骤
	StepCount Stats   // 统计信息
}

// Stats 统计信息
type Stats struct {
	Total   int // 总步骤数
	Done    int // 成功数
	Failed  int // 失败数
	Handled int // 已处理数
	Skipped int // 跳过数
	Pending int // 待执行数
}

// 状态标记正则
var statusPattern = regexp.MustCompile(`^\s*- \[([ x!\-~])\]\s*(\d+\.\d+)\s*(.*)$`)

// statusCharToStatus 将字符转换为状态
func statusCharToStatus(char string) StepStatus {
	switch char {
	case " ":
		return StatusPending
	case "x":
		return StatusDone
	case "!":
		return StatusFailed
	case "-":
		return StatusHandled
	case "~":
		return StatusSkipped
	default:
		return StatusPending
	}
}

// statusToChar 将状态转换为字符
func statusToChar(status StepStatus) string {
	switch status {
	case StatusPending:
		return " "
	case StatusDone:
		return "x"
	case StatusFailed:
		return "!"
	case StatusHandled:
		return "-"
	case StatusSkipped:
		return "~"
	default:
		return " "
	}
}

// ParseTodoFile 解析 TODO 文件
func ParseTodoFile(path string) (*TodoFile, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取 TODO 文件失败: %w", err)
	}

	return ParseTodoContent(path, string(content))
}

// ParseTodoContent 解析 TODO 内容
func ParseTodoContent(path, content string) (*TodoFile, error) {
	todo := &TodoFile{
		Path:    path,
		Content: content,
		Steps:   make([]*Step, 0),
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		matches := statusPattern.FindStringSubmatch(line)
		if matches != nil {
			step := &Step{
				ID:     matches[2],
				Status: statusCharToStatus(matches[1]),
				Line:   lineNum,
				Text:   matches[3],
			}

			// 验证步骤 ID 格式
			if !isValidStepID(step.ID) {
				return nil, fmt.Errorf("无效的步骤 ID 格式: %s (行 %d)", step.ID, lineNum)
			}

			todo.Steps = append(todo.Steps, step)
		}
	}

	// 统计
	todo.StepCount = CountSteps(todo.Steps)

	return todo, nil
}

// isValidStepID 验证步骤 ID 格式 (X.Y)
func isValidStepID(id string) bool {
	matched, _ := regexp.MatchString(`^\d+\.\d+$`, id)
	return matched
}

// CountSteps 统计步骤状态
func CountSteps(steps []*Step) Stats {
	var stats Stats
	stats.Total = len(steps)

	for _, step := range steps {
		switch step.Status {
		case StatusDone:
			stats.Done++
		case StatusFailed:
			stats.Failed++
		case StatusHandled:
			stats.Handled++
		case StatusSkipped:
			stats.Skipped++
		case StatusPending:
			stats.Pending++
		}
	}

	return stats
}

// UpdateStepStatus 更新步骤状态
func UpdateStepStatus(path, stepID string, status StepStatus, note string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取 TODO 文件失败: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	found := false
	char := statusToChar(status)

	for i, line := range lines {
		matches := statusPattern.FindStringSubmatch(line)
		if matches != nil && matches[2] == stepID {
			// 更新状态标记
			newLine := strings.Replace(line, "- ["+matches[1]+"]", "- ["+char+"]", 1)

			// 如果有备注，追加到行尾
			if note != "" {
				// 移除旧备注（如果有）
				if idx := strings.Index(newLine, " // "); idx > 0 {
					newLine = newLine[:idx]
				}
				newLine += " // " + note
			}

			lines[i] = newLine
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("未找到步骤: %s", stepID)
	}

	// 写回文件
	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("写入 TODO 文件失败: %w", err)
	}

	return nil
}

// FindStepByID 根据 ID 查找步骤
func (t *TodoFile) FindStepByID(id string) *Step {
	for _, step := range t.Steps {
		if step.ID == id {
			return step
		}
	}
	return nil
}

// GetStepsByStatus 获取指定状态的步骤
func (t *TodoFile) GetStepsByStatus(status StepStatus) []*Step {
	var result []*Step
	for _, step := range t.Steps {
		if step.Status == status {
			result = append(result, step)
		}
	}
	return result
}

// GetStepIDs 获取指定状态的步骤 ID 列表
func (t *TodoFile) GetStepIDs(status StepStatus) []string {
	steps := t.GetStepsByStatus(status)
	ids := make([]string, len(steps))
	for i, step := range steps {
		ids[i] = step.ID
	}
	return ids
}

// IsAllComplete 检查是否全部完成（包括失败已处理）
func (t *TodoFile) IsAllComplete() bool {
	for _, step := range t.Steps {
		if step.Status == StatusPending || step.Status == StatusFailed {
			return false
		}
	}
	return true
}

// HasUnhandledFail 检查是否有未处理的失败
func (t *TodoFile) HasUnhandledFail() bool {
	return t.StepCount.Failed > 0
}

// HasPending 检查是否有待执行步骤
func (t *TodoFile) HasPending() bool {
	return t.StepCount.Pending > 0
}
