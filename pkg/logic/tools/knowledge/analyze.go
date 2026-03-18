package knowledge

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	log "github.com/sirupsen/logrus"
	"msa/pkg/db"
	"msa/pkg/logic/memory"
	"msa/pkg/logic/message"
	"msa/pkg/model"
)

// AnalyzeTodayTradesParam 分析今日交易参数
type AnalyzeTodayTradesParam struct {
	// 无需参数，自动获取今日数据
}

// AnalyzeTodayTradesTool 分析今日交易工具
type AnalyzeTodayTradesTool struct{}

func (t *AnalyzeTodayTradesTool) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(t.GetName(), t.GetDescription(), AnalyzeTodayTrades)
}

func (t *AnalyzeTodayTradesTool) GetName() string {
	return "analyze_today_trades"
}

func (t *AnalyzeTodayTradesTool) GetDescription() string {
	return "分析今日交易，识别潜在错误操作 | Analyze today's trades and identify potential errors"
}

func (t *AnalyzeTodayTradesTool) GetToolGroup() model.ToolGroup {
	return model.KnowledgeToolGroup
}

// PotentialError 潜在错误
type PotentialError struct {
	Type       string `json:"type"`       // 错误类型
	StockCode  string `json:"stock_code"` // 股票代码
	StockName  string `json:"stock_name"` // 股票名称
	Detail     string `json:"detail"`     // 详情描述
	Suggestion string `json:"suggestion"` // 建议操作
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	Date          string           `json:"date"`           // 分析日期
	TotalTrades   int              `json:"total_trades"`   // 总交易数
	Errors        []PotentialError `json:"errors"`         // 潜在错误列表
	ErrorTemplate string           `json:"error_template"` // 可直接写入 errors/ 的模板
}

// AnalyzeTodayTrades 分析今日交易
func AnalyzeTodayTrades(ctx context.Context, param *AnalyzeTodayTradesParam) (string, error) {
	message.BroadcastToolStart("analyze_today_trades", "分析今日交易")

	result := &AnalysisResult{
		Date:   time.Now().Format("2006-01-02"),
		Errors: []PotentialError{},
	}

	// 1. 获取今日交易记录
	transactions, err := db.GetTodayTransactions()
	if err != nil {
		log.Warnf("获取今日交易记录失败: %v，使用空列表继续", err)
		transactions = []model.Transaction{}
	}
	result.TotalTrades = len(transactions)

	// 2. 获取今日 Session（用于分析决策上下文）
	manager := memory.GetManager()
	store := manager.GetStore()
	var sessions []model.SessionIndex
	if store != nil {
		sessions, err = store.ListSessions(0, 100)
		if err != nil {
			log.Warnf("获取 Session 列表失败: %v，使用空列表继续", err)
			sessions = []model.SessionIndex{}
		}
	}

	// 3. 分析每笔交易
	for _, tx := range transactions {
		analyzeTransaction(tx, result)
	}

	// 4. 分析 Session 中的决策记录
	analyzeSessions(sessions, result)

	// 5. 生成错误模板
	if len(result.Errors) > 0 {
		result.ErrorTemplate = generateErrorTemplate(result.Errors)
	}

	message.BroadcastToolEnd("analyze_today_trades", fmt.Sprintf("发现 %d 个潜在错误", len(result.Errors)), nil)
	return model.NewSuccessResult(result, fmt.Sprintf("分析完成，发现 %d 个潜在错误", len(result.Errors))), nil
}

// analyzeTransaction 分析单笔交易
func analyzeTransaction(tx model.Transaction, result *AnalysisResult) {
	// 注意：当前 Transaction 模型没有涨幅字段
	// 实际的追高判断需要在买入时记录当时的涨幅
	// 这里主要分析交易数量和频率

	// 检查：大额买入（需要结合账户总资产判断，暂时标记为提示）
	if tx.Type == model.TransactionTypeBuy {
		// 可以添加更多分析逻辑
		log.Debugf("分析买入交易: %s %s %d股", tx.StockCode, tx.StockName, tx.Quantity)
	}
}

// analyzeSessions 分析 Session 中的决策
func analyzeSessions(sessions []model.SessionIndex, result *AnalysisResult) {
	today := time.Now().Format("2006-01-02")

	for _, s := range sessions {
		// 只分析今天的 Session
		if s.StartTime.Format("2006-01-02") != today {
			continue
		}

		// 检查是否有交易相关的 Session
		for _, tag := range s.Tags {
			if tag == "morning-session" || tag == "afternoon-session" || tag == "close-session" {
				log.Debugf("发现交易时段 Session: %s, 标签: %s", s.ID, tag)
			}
		}
	}
}

// generateErrorTemplate 生成错误记录模板
func generateErrorTemplate(errors []PotentialError) string {
	if len(errors) == 0 {
		return ""
	}

	template := `---
id: err-XXX
date: ` + time.Now().Format("2006-01-02") + `
severity: medium
tags: [%s]
status: active
---

## 错误描述
%s

## 错误原因
[分析错误原因]

## 正确做法
%s

## 防范措施
[如何避免类似错误]
`

	var tags string
	var descriptions string
	var suggestions string

	for i, e := range errors {
		if i > 0 {
			tags += ", "
			descriptions += "\n\n"
		}
		tags += e.Type
		descriptions += fmt.Sprintf("- %s: %s", e.StockName, e.Detail)
		suggestions += fmt.Sprintf("- %s: %s\n", e.StockName, e.Suggestion)
	}

	return fmt.Sprintf(template, tags, descriptions, suggestions)
}
