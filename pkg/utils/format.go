package utils

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
)

// PrettyJSON 美化 JSON 字符串
func PrettyJSON(data interface{}) (string, error) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(bytes), nil
}

// PrettyJSONBytes 美化 JSON 字节数组
func PrettyJSONBytes(data []byte) (string, error) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return PrettyJSON(v)
}

// CompactJSON 压缩 JSON 字符串
func CompactJSON(data interface{}) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(bytes), nil
}

// FormatObject 格式化任意对象为 JSON 字符串
func FormatObject(obj interface{}, pretty bool) (string, error) {
	if pretty {
		return PrettyJSON(obj)
	}
	return CompactJSON(obj)
}

// ValidateJSON 验证 JSON 字符串是否有效
func ValidateJSON(str string) bool {
	var v interface{}
	return json.Unmarshal([]byte(str), &v) == nil
}

// ParseJSON 解析 JSON 字符串到目标对象
func ParseJSON(data string, v interface{}) error {
	return json.Unmarshal([]byte(data), v)
}

// ParseJSONBytes 解析 JSON 字节数组到目标对象
func ParseJSONBytes(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// ToJSONString 将任意对象转换为 JSON 字符串
func ToJSONString(obj interface{}) string {
	bytes, err := json.Marshal(obj)
	if err != nil {
		log.Error("ToJSONString error: %v", err)
		return ""
	}
	return string(bytes)
}

// ToJSONBytes 将任意对象转换为 JSON 字节数组
func ToJSONBytes(obj interface{}) ([]byte, error) {
	return json.Marshal(obj)
}

// FormatArray 格式化数组为易读的字符串
func FormatArray(arr []string, separator string) string {
	if len(arr) == 0 {
		return ""
	}
	return strings.Join(arr, separator)
}

// TruncateString 截断字符串到指定长度
func TruncateString(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	return str[:maxLen] + "..."
}
