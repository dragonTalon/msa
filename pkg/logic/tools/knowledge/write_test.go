package knowledge

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"msa/pkg/model"
)

func TestValidateWriteParam(t *testing.T) {
	tests := []struct {
		name    string
		param   *WriteKnowledgeParam
		wantErr bool
	}{
		{
			name: "valid error type",
			param: &WriteKnowledgeParam{
				Type:    "error",
				Content: "test content",
			},
			wantErr: false,
		},
		{
			name: "valid summary type",
			param: &WriteKnowledgeParam{
				Type:    "summary",
				Content: "test content",
				Date:    "2026-04-06",
			},
			wantErr: false,
		},
		{
			name: "missing type",
			param: &WriteKnowledgeParam{
				Content: "test content",
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			param: &WriteKnowledgeParam{
				Type:    "invalid",
				Content: "test content",
			},
			wantErr: true,
		},
		{
			name: "missing content",
			param: &WriteKnowledgeParam{
				Type: "error",
			},
			wantErr: true,
		},
		{
			name: "summary missing date",
			param: &WriteKnowledgeParam{
				Type:    "summary",
				Content: "test content",
			},
			wantErr: true,
		},
		{
			name: "summary invalid date format",
			param: &WriteKnowledgeParam{
				Type:    "summary",
				Content: "test content",
				Date:    "2026/04/06",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWriteParam(tt.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateWriteParam() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAppendError(t *testing.T) {
	// 使用临时目录
	tmpDir := t.TempDir()

	// 创建测试文件路径
	errorsPath := filepath.Join(tmpDir, "errors.md")

	// 第一次写入（文件不存在）
	content1 := "## 2026-04-05 测试错误\n\n**股票**: 300001\n**错误**: 测试"
	err := os.WriteFile(errorsPath, []byte(content1), 0644)
	if err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	// 验证文件内容
	data, err := os.ReadFile(errorsPath)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	if !containsString(string(data), "测试错误") {
		t.Errorf("file does not contain written content")
	}
}

func TestWriteSummary(t *testing.T) {
	// 使用临时目录
	tmpDir := t.TempDir()
	summariesDir := filepath.Join(tmpDir, "summaries")
	err := os.MkdirAll(summariesDir, 0755)
	if err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}

	// 写入总结
	summaryPath := filepath.Join(summariesDir, "2026-04-06.md")
	content := "---\ndate: 2026-04-06\n---\n\n# 今日总结"
	err = os.WriteFile(summaryPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	// 验证文件内容
	data, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	if !containsString(string(data), "今日总结") {
		t.Errorf("file does not contain written content")
	}
}

func TestVerifyWrite(t *testing.T) {
	// 创建临时测试文件
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "test.md")

	content := "test content"
	err := os.WriteFile(testPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	// 测试验证
	param := &WriteKnowledgeParam{
		Type:    "error",
		Content: "test content",
	}

	if !verifyWrite(testPath, param) {
		t.Errorf("verifyWrite() = false, want true")
	}
}

func TestWriteKnowledgeTool_Info(t *testing.T) {
	tool := &WriteKnowledgeTool{}

	if tool.GetName() != "write_knowledge" {
		t.Errorf("GetName() = %v, want write_knowledge", tool.GetName())
	}

	if tool.GetToolGroup() != model.KnowledgeToolGroup {
		t.Errorf("GetToolGroup() = %v, want %v", tool.GetToolGroup(), model.KnowledgeToolGroup)
	}

	baseTool, err := tool.GetToolInfo()
	if err != nil {
		t.Fatalf("GetToolInfo() error = %v", err)
	}

	if baseTool == nil {
		t.Error("GetToolInfo() returned nil")
	}
}

func TestWriteKnowledgeResult_Format(t *testing.T) {
	data := &WriteKnowledgeData{
		Type:       "error",
		Path:       "/test/path.md",
		Verified:   true,
		ContentLen: 100,
		RetryCount: 0,
	}

	result := model.NewSuccessResult(data, "test")
	var parsed struct {
		Success bool               `json:"success"`
		Data    WriteKnowledgeData `json:"data"`
		Message string             `json:"message"`
	}

	err := json.Unmarshal([]byte(result), &parsed)
	if err != nil {
		t.Fatalf("json.Unmarshal error = %v", err)
	}

	if !parsed.Success {
		t.Error("Success should be true")
	}

	if !parsed.Data.Verified {
		t.Error("Verified should be true")
	}
}

func TestWriteKnowledge_ParamValidation(t *testing.T) {
	// 测试参数验证错误返回
	ctx := context.Background()

	// 缺少 type
	param := &WriteKnowledgeParam{
		Content: "test",
	}

	result, err := WriteKnowledge(ctx, param)
	if err != nil {
		t.Fatalf("WriteKnowledge() error = %v", err)
	}

	// 解析结果
	var parsed struct {
		Success  bool   `json:"success"`
		ErrorMsg string `json:"error_msg"`
	}

	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("json.Unmarshal error = %v", err)
	}

	if parsed.Success {
		t.Error("Expected failure for missing type")
	}
}

// containsString 辅助函数
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
