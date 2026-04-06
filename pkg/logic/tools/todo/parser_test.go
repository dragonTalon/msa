package todo

import (
	"os"
	"path/filepath"
	"testing"
)

const testTodoContent = `# TODO: test-skill

> 创建时间：2026-04-06 12:00:00
> 会话ID：2026-04-06_test123

---

## 执行状态
- [ ] 所有步骤完成
- [ ] 已验证执行结果

---

## 1. 信息收集

### 步骤
- [ ] 1.1 读取历史错误
- [x] 1.2 读取用户知识库
- [!] 1.3 获取账户状态
- [-] 1.4 处理失败
- [~] 1.5 跳过此步骤

### 失败处理
- 如果 1.1 失败：跳过继续执行
- 如果 1.2 失败：使用默认配置
`

func TestParseTodoContent(t *testing.T) {
	todo, err := ParseTodoContent("test.md", testTodoContent)
	if err != nil {
		t.Fatalf("ParseTodoContent failed: %v", err)
	}

	// 验证步骤数量（5个任务步骤，不含执行状态的复选框）
	if len(todo.Steps) != 5 {
		t.Errorf("Expected 5 steps, got %d", len(todo.Steps))
	}

	// 验证统计
	if todo.StepCount.Total != 5 {
		t.Errorf("Expected total 5, got %d", todo.StepCount.Total)
	}

	if todo.StepCount.Pending != 1 {
		t.Errorf("Expected 1 pending, got %d", todo.StepCount.Pending)
	}

	if todo.StepCount.Done != 1 {
		t.Errorf("Expected 1 done, got %d", todo.StepCount.Done)
	}

	if todo.StepCount.Failed != 1 {
		t.Errorf("Expected 1 failed, got %d", todo.StepCount.Failed)
	}

	if todo.StepCount.Handled != 1 {
		t.Errorf("Expected 1 handled, got %d", todo.StepCount.Handled)
	}

	if todo.StepCount.Skipped != 1 {
		t.Errorf("Expected 1 skipped, got %d", todo.StepCount.Skipped)
	}
}

func TestStepStatus(t *testing.T) {
	tests := []struct {
		id       string
		expected StepStatus
	}{
		{"1.1", StatusPending},
		{"1.2", StatusDone},
		{"1.3", StatusFailed},
		{"1.4", StatusHandled},
		{"1.5", StatusSkipped},
	}

	todo, _ := ParseTodoContent("test.md", testTodoContent)

	for _, tt := range tests {
		step := todo.FindStepByID(tt.id)
		if step == nil {
			t.Errorf("Step %s not found", tt.id)
			continue
		}
		if step.Status != tt.expected {
			t.Errorf("Step %s: expected %s, got %s", tt.id, tt.expected, step.Status)
		}
	}
}

func TestIsValidStepID(t *testing.T) {
	validIDs := []string{"1.1", "1.2", "2.1", "10.5", "99.99"}
	invalidIDs := []string{"1", "1.", ".1", "a.1", "1.a", "1.1.1", ""}

	for _, id := range validIDs {
		if !isValidStepID(id) {
			t.Errorf("Expected %s to be valid", id)
		}
	}

	for _, id := range invalidIDs {
		if isValidStepID(id) {
			t.Errorf("Expected %s to be invalid", id)
		}
	}
}

func TestUpdateStepStatus(t *testing.T) {
	// 创建临时文件
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")
	os.WriteFile(tmpFile, []byte(testTodoContent), 0644)

	// 测试更新状态
	err := UpdateStepStatus(tmpFile, "1.1", StatusDone, "测试完成")
	if err != nil {
		t.Fatalf("UpdateStepStatus failed: %v", err)
	}

	// 验证更新
	todo, err := ParseTodoFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseTodoFile failed: %v", err)
	}

	step := todo.FindStepByID("1.1")
	if step == nil {
		t.Fatal("Step 1.1 not found")
	}
	if step.Status != StatusDone {
		t.Errorf("Expected status done, got %s", step.Status)
	}
}

func TestIsAllComplete(t *testing.T) {
	// 测试未完成
	todo, _ := ParseTodoContent("test.md", testTodoContent)
	if todo.IsAllComplete() {
		t.Error("Expected not all complete")
	}

	// 测试全部完成
	allDoneContent := `# TODO: test
- [x] 1.1 Step 1
- [x] 1.2 Step 2
- [-] 1.3 Step 3 (handled)
- [~] 1.4 Step 4 (skipped)
`
	todo2, _ := ParseTodoContent("test.md", allDoneContent)
	if !todo2.IsAllComplete() {
		t.Error("Expected all complete")
	}
}

func TestHasUnhandledFail(t *testing.T) {
	// 测试有未处理失败
	todo, _ := ParseTodoContent("test.md", testTodoContent)
	if !todo.HasUnhandledFail() {
		t.Error("Expected has unhandled fail")
	}

	// 测试无未处理失败
	noFailContent := `# TODO: test
- [x] 1.1 Step 1
- [ ] 1.2 Step 2
`
	todo2, _ := ParseTodoContent("test.md", noFailContent)
	if todo2.HasUnhandledFail() {
		t.Error("Expected no unhandled fail")
	}
}

func TestGetStepsByStatus(t *testing.T) {
	todo, _ := ParseTodoContent("test.md", testTodoContent)

	pendingSteps := todo.GetStepsByStatus(StatusPending)
	if len(pendingSteps) != 1 {
		t.Errorf("Expected 1 pending step, got %d", len(pendingSteps))
	}

	doneSteps := todo.GetStepsByStatus(StatusDone)
	if len(doneSteps) != 1 {
		t.Errorf("Expected 1 done step, got %d", len(doneSteps))
	}

	failedSteps := todo.GetStepsByStatus(StatusFailed)
	if len(failedSteps) != 1 {
		t.Errorf("Expected 1 failed step, got %d", len(failedSteps))
	}

	handledSteps := todo.GetStepsByStatus(StatusHandled)
	if len(handledSteps) != 1 {
		t.Errorf("Expected 1 handled step, got %d", len(handledSteps))
	}

	skippedSteps := todo.GetStepsByStatus(StatusSkipped)
	if len(skippedSteps) != 1 {
		t.Errorf("Expected 1 skipped step, got %d", len(skippedSteps))
	}
}

func TestStatusToChar(t *testing.T) {
	tests := []struct {
		status   StepStatus
		expected string
	}{
		{StatusPending, " "},
		{StatusDone, "x"},
		{StatusFailed, "!"},
		{StatusHandled, "-"},
		{StatusSkipped, "~"},
	}

	for _, tt := range tests {
		result := statusToChar(tt.status)
		if result != tt.expected {
			t.Errorf("Status %s: expected %q, got %q", tt.status, tt.expected, result)
		}
	}
}
