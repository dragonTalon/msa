package model

import "msa/pkg/utils"

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
	return utils.ToJSONString(result)
}

// NewErrorResult 生成失败响应
func NewErrorResult(message string) string {
	result := &ToolResult{
		Success:  false,
		Data:     nil,
		ErrorMsg: message,
	}
	return utils.ToJSONString(result)
}
