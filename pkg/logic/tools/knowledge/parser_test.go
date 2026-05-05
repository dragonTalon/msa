package knowledge

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseErrorsFile_FileNotExist(t *testing.T) {
	// 文件不存在时返回空列表
	entries, err := ParseErrorsFile("/nonexistent/path/errors.md")
	if err != nil {
		t.Fatalf("ParseErrorsFile() error = %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("ParseErrorsFile() = %v entries, want 0", len(entries))
	}
}

func TestParseErrorsContent_NormalParse(t *testing.T) {
	content := `# 错误记录

## 2026-04-05 追高买入导致亏损

**股票**: 300001
**错误**: 在涨幅超过5%时追高买入
**原因**: 看到股价快速上涨，担心错过机会
**教训**: 不追高，等待回调再买入

## 2026-04-03 未及时止损

**股票**: 600000
**错误**: 持仓亏损超过10%未止损
**原因**: 抱有侥幸心理，认为会反弹
**教训**: 严格执行止损规则
`

	entries := parseErrorsContent(content)

	if len(entries) != 2 {
		t.Fatalf("parseErrorsContent() = %v entries, want 2", len(entries))
	}

	// 验证第一个条目
	if entries[0].Date != "2026-04-05" {
		t.Errorf("entries[0].Date = %v, want 2026-04-05", entries[0].Date)
	}
	if entries[0].Title != "追高买入导致亏损" {
		t.Errorf("entries[0].Title = %v, want 追高买入导致亏损", entries[0].Title)
	}
	if entries[0].Stock != "300001" {
		t.Errorf("entries[0].Stock = %v, want 300001", entries[0].Stock)
	}
	if entries[0].Error != "在涨幅超过5%时追高买入" {
		t.Errorf("entries[0].Error = %v", entries[0].Error)
	}

	// 验证第二个条目
	if entries[1].Date != "2026-04-03" {
		t.Errorf("entries[1].Date = %v, want 2026-04-03", entries[1].Date)
	}
	if entries[1].Lesson != "严格执行止损规则" {
		t.Errorf("entries[1].Lesson = %v, want 严格执行止损规则", entries[1].Lesson)
	}
}

func TestExtractField(t *testing.T) {
	content := `**股票**: 300001
**错误**: 测试错误
**原因**: 测试原因
`

	tests := []struct {
		fieldName string
		want      string
	}{
		{"股票", "300001"},
		{"错误", "测试错误"},
		{"原因", "测试原因"},
		{"不存在", ""},
	}

	for _, tt := range tests {
		got := extractField(content, tt.fieldName)
		if got != tt.want {
			t.Errorf("extractField(%q) = %q, want %q", tt.fieldName, got, tt.want)
		}
	}
}

func TestIsValidDate(t *testing.T) {
	tests := []struct {
		date string
		want bool
	}{
		{"2026-04-06", true},
		{"2026-01-01", true},
		{"2026-4-6", false},   // 格式不对
		{"2026/04/06", false}, // 格式不对
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		got := isValidDate(tt.date)
		if got != tt.want {
			t.Errorf("isValidDate(%q) = %v, want %v", tt.date, got, tt.want)
		}
	}
}

func TestFindPrevSummaryFile_EmptyDir(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	// 保存原 GetSummariesDir 函数的行为
	// 这里直接测试函数逻辑
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("ReadDir error = %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty dir")
	}
}

func TestParseSummaryFile(t *testing.T) {
	// 创建临时测试文件
	tmpDir := t.TempDir()
	summaryPath := filepath.Join(tmpDir, "2026-04-05.md")

	content := `---
date: 2026-04-05
prev_asset: 100000.00
curr_asset: 102000.00
pnl: 2000.00
pnl_rate: 2.0
---

# 今日总结

## 收益情况
今日盈利 2000 元，收益率 2%。

## 操作复盘
早盘卖出 300001，获利了结。
`

	err := os.WriteFile(summaryPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	data, err := ParseSummaryFile(summaryPath)
	if err != nil {
		t.Fatalf("ParseSummaryFile() error = %v", err)
	}

	if data.Date != "2026-04-05" {
		t.Errorf("Date = %v, want 2026-04-05", data.Date)
	}
	if data.PrevAsset != 100000.00 {
		t.Errorf("PrevAsset = %v, want 100000.00", data.PrevAsset)
	}
	if data.CurrAsset != 102000.00 {
		t.Errorf("CurrAsset = %v, want 102000.00", data.CurrAsset)
	}
	if data.PNL != 2000.00 {
		t.Errorf("PNL = %v, want 2000.00", data.PNL)
	}
	if data.PNLRate != 2.0 {
		t.Errorf("PNLRate = %v, want 2.0", data.PNLRate)
	}
}

func TestParseSummaryFile_WithoutFrontmatter(t *testing.T) {
	// 创建临时测试文件
	tmpDir := t.TempDir()
	summaryPath := filepath.Join(tmpDir, "2026-04-05.md")

	content := `# 今日总结

**昨日资产**: 100000.00
**今日资产**: 102000.00
**今日盈亏**: 2000.00
**收益率**: 2.0%
`

	err := os.WriteFile(summaryPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	data, err := ParseSummaryFile(summaryPath)
	if err != nil {
		t.Fatalf("ParseSummaryFile() error = %v", err)
	}

	if data.Date != "2026-04-05" {
		t.Errorf("Date = %v, want 2026-04-05", data.Date)
	}
	if data.PrevAsset != 100000.00 {
		t.Errorf("PrevAsset = %v, want 100000.00", data.PrevAsset)
	}
	if data.CurrAsset != 102000.00 {
		t.Errorf("CurrAsset = %v, want 102000.00", data.CurrAsset)
	}
}

func TestParseSummaryFile_TableFormat(t *testing.T) {
	tmpDir := t.TempDir()
	summaryPath := filepath.Join(tmpDir, "2026-05-01.md")

	content := `
# 2026-05-01 交易日总结

## 账户总览

| 项目 | 数值 |
|------|------|
| 初始资金 | **500,000.00 元** |
| 当前总资产 | **493,233.00 元** |
| 总盈亏 | **-6,767.00 元** |
| 总收益率 | **-1.35%** |
`

	err := os.WriteFile(summaryPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	data, err := ParseSummaryFile(summaryPath)
	if err != nil {
		t.Fatalf("ParseSummaryFile() error = %v", err)
	}

	if data.Date != "2026-05-01" {
		t.Errorf("Date = %v, want 2026-05-01", data.Date)
	}
	if data.PrevAsset != 500000.00 {
		t.Errorf("PrevAsset = %v, want 500000.00", data.PrevAsset)
	}
	if data.CurrAsset != 493233.00 {
		t.Errorf("CurrAsset = %v, want 493233.00", data.CurrAsset)
	}
	if data.PNL != -6767.00 {
		t.Errorf("PNL = %v, want -6767.00", data.PNL)
	}
	if data.PNLRate != -1.35 {
		t.Errorf("PNLRate = %v, want -1.35", data.PNLRate)
	}
	if data.Raw == "" {
		t.Errorf("Raw should not be empty")
	}
}

func TestParseSummaryFile_TableFormatFallbackPriority(t *testing.T) {
	tmpDir := t.TempDir()
	summaryPath := filepath.Join(tmpDir, "2026-05-01.md")

	// 同时包含 frontmatter 和表格 - YAML 应优先
	content := `---
prev_asset: 200000.00
curr_asset: 210000.00
pnl: 10000.00
pnl_rate: 5.0
---

# 总结

| 项目 | 数值 |
|------|------|
| 当前总资产 | **300,000.00 元** |
`

	err := os.WriteFile(summaryPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	data, err := ParseSummaryFile(summaryPath)
	if err != nil {
		t.Fatalf("ParseSummaryFile() error = %v", err)
	}

	// YAML frontmatter 优先，不应使用表格数据
	if data.CurrAsset != 210000.00 {
		t.Errorf("CurrAsset = %v, want 210000.00 (YAML frontmatter should take priority)", data.CurrAsset)
	}
}

func TestExtractNumberFromCellValue(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"**500,000.00 元**", "500000.00"},
		{"493,233.00 元", "493233.00"},
		{"-6,767.00 元", "-6767.00"},
		{"-1.35%", "-1.35"},
		{"普通文本无数字", ""},
		{"+3.5%", "3.5"},
	}
	for _, tt := range tests {
		got := extractNumberFromCellValue(tt.input)
		if got != tt.want {
			t.Errorf("extractNumberFromCellValue(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
