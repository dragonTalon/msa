package safetool

import (
	"encoding/json"
	"msa/pkg/model"
	"strings"
	"testing"
)

// TestSafeExecute_NormalExecution 测试正常执行场景
func TestSafeExecute_NormalExecution(t *testing.T) {
	tests := []struct {
		name       string
		toolName   string
		params     string
		fn         func() (string, error)
		wantResult string
		wantErr    bool
	}{
		{
			name:     "正常返回成功结果",
			toolName: "test_tool",
			params:   "test_params",
			fn: func() (string, error) {
				return model.NewSuccessResult("test data", "success"), nil
			},
			wantResult: "test data",
			wantErr:    false,
		},
		{
			name:     "正常返回错误结果",
			toolName: "test_tool",
			params:   "test_params",
			fn: func() (string, error) {
				return model.NewErrorResult("参数错误"), nil
			},
			wantResult: "",
			wantErr:    false,
		},
		{
			name:     "函数返回 error",
			toolName: "test_tool",
			params:   "test_params",
			fn: func() (string, error) {
				return "", &testError{"some error"}
			},
			wantResult: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SafeExecute(tt.toolName, tt.params, tt.fn)

			if (err != nil) != tt.wantErr {
				t.Errorf("SafeExecute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result == "" {
				t.Error("SafeExecute() result is empty")
			}

			// 验证返回的 JSON 格式
			if !tt.wantErr {
				var toolResult model.ToolResult
				if err := json.Unmarshal([]byte(result), &toolResult); err != nil {
					t.Errorf("SafeExecute() 返回无效的 JSON: %v", err)
				}
			}
		})
	}
}

// TestSafeExecute_PanicRecovery 测试 panic 恢复场景
func TestSafeExecute_PanicRecovery(t *testing.T) {
	tests := []struct {
		name         string
		toolName     string
		params       string
		panicValue   interface{}
		wantErrMsg   bool
		errMsgPrefix string
	}{
		{
			name:         "panic string 恢复",
			toolName:     "panic_tool",
			params:       "panic_params",
			panicValue:   "something went wrong",
			wantErrMsg:   true,
			errMsgPrefix: "工具内部错误:",
		},
		{
			name:         "panic error 恢复",
			toolName:     "panic_tool",
			params:       "panic_params",
			panicValue:   &testError{"error panic"},
			wantErrMsg:   true,
			errMsgPrefix: "工具内部错误:",
		},
		{
			name:         "panic 其他类型恢复",
			toolName:     "panic_tool",
			params:       "panic_params",
			panicValue:   12345,
			wantErrMsg:   true,
			errMsgPrefix: "工具内部错误:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := func() (string, error) {
				panic(tt.panicValue)
			}

			// SafeExecute 应该捕获 panic 并返回错误 JSON，而不是让测试崩溃
			result, err := SafeExecute(tt.toolName, tt.params, fn)

			// panic 恢复后，err 应该是 nil（错误信息包装在 result JSON 中）
			if err != nil {
				t.Errorf("SafeExecute() 应该返回 nil error，但返回了: %v", err)
			}

			// 验证返回的是错误 JSON
			var toolResult model.ToolResult
			if err := json.Unmarshal([]byte(result), &toolResult); err != nil {
				t.Errorf("SafeExecute() 返回无效的 JSON: %v", err)
				return
			}

			if toolResult.Success {
				t.Error("SafeExecute() panic 后不应该返回 success=true")
			}

			if !strings.Contains(toolResult.ErrorMsg, tt.errMsgPrefix) {
				t.Errorf("SafeExecute() 错误信息应该包含 '%s'，实际为: %s", tt.errMsgPrefix, toolResult.ErrorMsg)
			}
		})
	}
}

// TestSafeExecute_WithNilFunction 测试传入 nil 函数的情况
func TestSafeExecute_WithNilFunction(t *testing.T) {
	// 注意：传入 nil 函数会导致 panic，SafeExecute 应该捕获它
	defer func() {
		// 验证测试本身没有因为 panic 而崩溃
	}()

	// 这个测试会触发 panic（nil 函数调用），但 SafeExecute 应该捕获它
	result, err := SafeExecute("nil_tool", "params", nil)

	// 如果执行到这里，说明 panic 被正确捕获
	if err != nil {
		t.Errorf("SafeExecute() 应该返回 nil error，但返回了: %v", err)
	}

	var toolResult model.ToolResult
	if err := json.Unmarshal([]byte(result), &toolResult); err != nil {
		t.Errorf("SafeExecute() 返回无效的 JSON: %v", err)
		return
	}

	if toolResult.Success {
		t.Error("SafeExecute() nil 函数不应该返回 success=true")
	}
}

// TestSafeExecute_ConcurrentCalls 测试并发调用安全性
func TestSafeExecute_ConcurrentCalls(t *testing.T) {
	const goroutines = 10
	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// 一半的 goroutine 正常执行，一半触发 panic
			if id%2 == 0 {
				result, err := SafeExecute("concurrent_tool", "params", func() (string, error) {
					return model.NewSuccessResult(id, "success"), nil
				})
				if err != nil {
					t.Errorf("goroutine %d: unexpected error: %v", id, err)
				}
				if result == "" {
					t.Errorf("goroutine %d: empty result", id)
				}
			} else {
				result, err := SafeExecute("concurrent_panic_tool", "params", func() (string, error) {
					panic("concurrent panic")
				})
				if err != nil {
					t.Errorf("goroutine %d: should return nil error after panic recovery", id)
				}
				var toolResult model.ToolResult
				if err := json.Unmarshal([]byte(result), &toolResult); err != nil {
					t.Errorf("goroutine %d: invalid JSON: %v", id, err)
				}
				if toolResult.Success {
					t.Errorf("goroutine %d: should not return success=true after panic", id)
				}
			}
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < goroutines; i++ {
		<-done
	}
}

// testError 是一个简单的错误类型，用于测试
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
