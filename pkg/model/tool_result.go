package model

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

// ToolResult 统一的 Tool 响应结构
type ToolResult struct {
	Success  bool        `json:"success"`
	Data     interface{} `json:"data,omitempty"`
	Message  string      `json:"message,omitempty"`   // 成功时的描述
	ErrorMsg string      `json:"error_msg,omitempty"` // 失败时的错误信息
}

// NewSuccessResult 生成成功响应
func NewSuccessResult(data interface{}, message string) string {
	result := &ToolResult{
		Success: true,
		Data:    data,
		Message: message,
	}
	return toJSON(result)
}

// NewErrorResult 生成失败响应
func NewErrorResult(message string) string {
	result := &ToolResult{
		Success:  false,
		Data:     nil,
		ErrorMsg: message,
	}
	return toJSON(result)
}

// toJSON 使用缩进格式序列化，避免单行过长被截断
func toJSON(v interface{}) string {
	bytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Errorf("toJSON error: %v", err)
		return `{"success":false,"error_msg":"JSON serialization failed"}`
	}
	return string(bytes)
}
