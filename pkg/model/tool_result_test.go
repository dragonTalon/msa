package model

import (
	"encoding/json"
	"testing"
)

func TestNewSuccessResult(t *testing.T) {
	data := map[string]interface{}{
		"id":   1,
		"name": "test",
	}
	result := NewSuccessResult(data, "操作成功")

	// 验证是有效的 JSON
	var parsed ToolResult
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Errorf("NewSuccessResult 返回无效的 JSON: %v", err)
	}

	// 验证字段
	if !parsed.Success {
		t.Error("Success 应该为 true")
	}
	if parsed.Message != "操作成功" {
		t.Errorf("Message 应该为 '操作成功', 实际为 '%s'", parsed.Message)
	}
	if parsed.ErrorMsg != "" {
		t.Error("ErrorMsg 应该为空")
	}
}

func TestNewErrorResult(t *testing.T) {
	result := NewErrorResult("参数错误")

	// 验证是有效的 JSON
	var parsed ToolResult
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Errorf("NewErrorResult 返回无效的 JSON: %v", err)
	}

	// 验证字段
	if parsed.Success {
		t.Error("Success 应该为 false")
	}
	if parsed.ErrorMsg != "参数错误" {
		t.Errorf("ErrorMsg 应该为 '参数错误', 实际为 '%s'", parsed.ErrorMsg)
	}
	if parsed.Message != "" {
		t.Error("Message 应该为空")
	}
	if parsed.Data != nil {
		t.Error("Data 应该为 nil")
	}
}

func TestToolResult_JSONFields(t *testing.T) {
	tests := []struct {
		name     string
		result   *ToolResult
		wantJSON string
	}{
		{
			name: "成功响应包含 data 和 message",
			result: &ToolResult{
				Success: true,
				Data:    map[string]int{"count": 5},
				Message: "查询成功",
			},
			wantJSON: `{"success":true,"data":{"count":5},"message":"查询成功"}`,
		},
		{
			name: "失败响应包含 error_msg",
			result: &ToolResult{
				Success:  false,
				ErrorMsg: "数据库连接失败",
			},
			wantJSON: `{"success":false,"data":null,"error_msg":"数据库连接失败"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.result)
			if err != nil {
				t.Errorf("JSON 序列化失败: %v", err)
			}
			// 验证包含必要字段
			var parsed map[string]interface{}
			if err := json.Unmarshal(data, &parsed); err != nil {
				t.Errorf("JSON 反序列化失败: %v", err)
			}
			if _, ok := parsed["success"]; !ok {
				t.Error("JSON 缺少 success 字段")
			}
		})
	}
}
