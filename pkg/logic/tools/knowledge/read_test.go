package knowledge

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"msa/pkg/model"
)

func TestReadKnowledge_TypeErrors(t *testing.T) {
	// 创建临时测试目录
	tmpDir := t.TempDir()

	// 创建错误文件
	errorsContent := `# 错误记录

## 2026-04-05 追高买入导致亏损

**股票**: 300001
**错误**: 在涨幅超过5%时追高买入
**原因**: 看到股价快速上涨，担心错过机会
**教训**: 不追高，等待回调再买入
`
	errorsPath := filepath.Join(tmpDir, "errors.md")
	err := os.WriteFile(errorsPath, []byte(errorsContent), 0644)
	if err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	// 覆盖路径函数（通过设置环境变量或临时目录）
	// 由于无法直接 mock，这里测试解析逻辑
	entries, err := ParseErrorsFile(errorsPath)
	if err != nil {
		t.Fatalf("ParseErrorsFile error: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("len(entries) = %v, want 1", len(entries))
	}
}

func TestReadKnowledge_TypeSummary(t *testing.T) {
	// 创建临时测试目录
	tmpDir := t.TempDir()
	summariesDir := filepath.Join(tmpDir, "summaries")
	err := os.MkdirAll(summariesDir, 0755)
	if err != nil {
		t.Fatalf("MkdirAll error: %v", err)
	}

	// 创建总结文件
	summaryContent := `---
date: 2026-04-05
prev_asset: 100000.00
curr_asset: 102000.00
pnl: 2000.00
pnl_rate: 2.0
---

# 今日总结
`
	summaryPath := filepath.Join(summariesDir, "2026-04-05.md")
	err = os.WriteFile(summaryPath, []byte(summaryContent), 0644)
	if err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	// 测试解析
	data, err := ParseSummaryFile(summaryPath)
	if err != nil {
		t.Fatalf("ParseSummaryFile error: %v", err)
	}

	if data.Date != "2026-04-05" {
		t.Errorf("Date = %v, want 2026-04-05", data.Date)
	}
}

func TestReadKnowledge_TypeAll(t *testing.T) {
	// 测试 type="all" 时能正确返回两者
	// 这里只测试解析逻辑，因为路径函数依赖于用户目录

	// 测试错误解析
	errorsContent := `# 错误记录

## 2026-04-05 测试错误
**股票**: 300001
**错误**: 测试
**原因**: 测试
**教训**: 测试
`
	entries := parseErrorsContent(errorsContent)
	if len(entries) != 1 {
		t.Errorf("parseErrorsContent returned %d entries, want 1", len(entries))
	}

	// 测试总结解析
	summaryContent := `---
date: 2026-04-05
prev_asset: 100000.00
curr_asset: 102000.00
pnl: 2000.00
pnl_rate: 2.0
---`

	data := &SummaryData{Date: "2026-04-05"}
	parseSummaryFields(summaryContent, data)

	if data.PrevAsset != 100000.00 {
		t.Errorf("PrevAsset = %v, want 100000.00", data.PrevAsset)
	}
}

func TestReadKnowledge_FileNotExist(t *testing.T) {
	// 测试文件不存在时不报错
	entries, err := ParseErrorsFile("/nonexistent/path/errors.md")
	if err != nil {
		t.Fatalf("ParseErrorsFile error = %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("ParseErrorsFile returned %d entries, want 0", len(entries))
	}
}

func TestReadKnowledgeTool_Info(t *testing.T) {
	tool := &ReadKnowledgeTool{}

	if tool.GetName() != "read_knowledge" {
		t.Errorf("GetName() = %v, want read_knowledge", tool.GetName())
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

func TestReadKnowledgeResult_Format(t *testing.T) {
	// 测试返回值格式
	data := &ReadKnowledgeData{
		Errors: []ErrorEntry{
			{Date: "2026-04-05", Title: "测试错误"},
		},
		PrevSummary: &SummaryData{
			Date:    "2026-04-04",
			PNL:     2000.0,
			PNLRate: 2.0,
		},
		HasErrors:  true,
		HasSummary: true,
	}

	result := model.NewSuccessResult(data, "test")
	var parsed struct {
		Success bool              `json:"success"`
		Data    ReadKnowledgeData `json:"data"`
		Message string            `json:"message"`
	}

	err := json.Unmarshal([]byte(result), &parsed)
	if err != nil {
		t.Fatalf("json.Unmarshal error = %v", err)
	}

	if !parsed.Success {
		t.Error("Success should be true")
	}

	if !parsed.Data.HasErrors {
		t.Error("HasErrors should be true")
	}

	if !parsed.Data.HasSummary {
		t.Error("HasSummary should be true")
	}
}
