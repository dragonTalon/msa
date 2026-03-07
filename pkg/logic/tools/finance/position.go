package finance

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	msadb "msa/pkg/db"
	"msa/pkg/logic/message"
	"msa/pkg/model"
	"msa/pkg/service"
)

// GetPositionsParam 查询持仓参数
type GetPositionsParam struct{}

// GetPositionsTool 查询持仓列表工具
type GetPositionsTool struct{}

func (t *GetPositionsTool) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(t.GetName(), t.GetDescription(), GetPositions)
}

func (t *GetPositionsTool) GetName() string {
	return "get_positions"
}

func (t *GetPositionsTool) GetDescription() string {
	return "获取当前所有持仓（股票、数量、成本、市值、盈亏）| Get all current positions (stock, quantity, cost, value, profit/loss)"
}

func (t *GetPositionsTool) GetToolGroup() model.ToolGroup {
	return model.FinanceToolGroup
}

// GetPositions 获取持仓列表
func GetPositions(ctx context.Context, param *GetPositionsParam) (string, error) {
	message.BroadcastToolStart("get_positions", "")

	database := msadb.GetDB()
	if database == nil {
		err := fmt.Errorf("数据库未初始化")
		message.BroadcastToolEnd("get_positions", "", err)
		return "", err
	}

	account, err := getActiveAccount(database)
	if err != nil {
		message.BroadcastToolEnd("get_positions", "", err)
		return "", err
	}

	// 获取所有持仓股票代码
	var stockCodes []string
	err = database.Model(&model.Transaction{}).
		Select("DISTINCT stock_code").
		Where("account_id = ? AND status = ?", account.ID, model.TransactionStatusFilled).
		Pluck("stock_code", &stockCodes).Error
	if err != nil {
		message.BroadcastToolEnd("get_positions", "", err)
		return "", err
	}

	if len(stockCodes) == 0 {
		result := "当前无持仓"
		message.BroadcastToolEnd("get_positions", result, nil)
		return result, nil
	}

	// 批量获取价格
	prices, err := fetchAllPrices(stockCodes)
	if err != nil {
		message.BroadcastToolEnd("get_positions", "", err)
		return "", err
	}

	// 转换为 service.PriceMap 格式
	priceMap := make(service.PriceMap)
	for code, price := range prices {
		priceMap[code] = price
	}

	// 获取所有持仓
	positions, err := service.GetAllPositions(database, account.ID, priceMap)
	if err != nil {
		message.BroadcastToolEnd("get_positions", "", err)
		return "", err
	}

	if len(positions) == 0 {
		result := "当前无持仓"
		message.BroadcastToolEnd("get_positions", result, nil)
		return result, nil
	}

	// 格式化输出
	result := fmt.Sprintf("持仓列表（共 %d 只）：\n\n", len(positions))
	for _, pos := range positions {
		pnlStatus := "亏损"
		if pos.PnL >= 0 {
			pnlStatus = "盈利"
		}
		result += fmt.Sprintf("%s (%s)\n", pos.StockName, pos.StockCode)
		result += fmt.Sprintf("  持仓数量: %d 股\n", pos.Quantity)
		result += fmt.Sprintf("  持仓成本: %s 元\n", formatHaoToYuan(pos.Cost))
		result += fmt.Sprintf("  当前价格: %s 元/股\n", formatHaoToYuan(pos.CurrentPrice))
		result += fmt.Sprintf("  持仓市值: %s 元\n", formatHaoToYuan(pos.Value))
		result += fmt.Sprintf("  盈亏: %s 元 (%s)\n", formatHaoToYuan(pos.PnL), pnlStatus)
		result += "\n"
	}

	message.BroadcastToolEnd("get_positions", fmt.Sprintf("获取 %d 只持仓", len(positions)), nil)
	return result, nil
}

// GetAccountSummaryParam 查询账户总览参数
type GetAccountSummaryParam struct{}

// GetAccountSummaryTool 查询账户总览工具
type GetAccountSummaryTool struct{}

func (t *GetAccountSummaryTool) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(t.GetName(), t.GetDescription(), GetAccountSummary)
}

func (t *GetAccountSummaryTool) GetName() string {
	return "get_account_summary"
}

func (t *GetAccountSummaryTool) GetDescription() string {
	return "获取账户总览（总资产、可用余额、持仓市值、总盈亏）| Get account summary (total assets, available balance, position value, profit/loss)"
}

func (t *GetAccountSummaryTool) GetToolGroup() model.ToolGroup {
	return model.FinanceToolGroup
}

// GetAccountSummary 获取账户总览
func GetAccountSummary(ctx context.Context, param *GetAccountSummaryParam) (string, error) {
	message.BroadcastToolStart("get_account_summary", "")

	database := msadb.GetDB()
	if database == nil {
		err := fmt.Errorf("数据库未初始化")
		message.BroadcastToolEnd("get_account_summary", "", err)
		return "", err
	}

	account, err := getActiveAccount(database)
	if err != nil {
		message.BroadcastToolEnd("get_account_summary", "", err)
		return "", err
	}

	// 获取所有持仓股票代码
	var stockCodes []string
	err = database.Model(&model.Transaction{}).
		Select("DISTINCT stock_code").
		Where("account_id = ? AND status = ?", account.ID, model.TransactionStatusFilled).
		Pluck("stock_code", &stockCodes).Error
	if err != nil {
		message.BroadcastToolEnd("get_account_summary", "", err)
		return "", err
	}

	// 批量获取价格
	priceMap := make(service.PriceMap)
	if len(stockCodes) > 0 {
		prices, err := fetchAllPrices(stockCodes)
		if err != nil {
			// 价格获取失败不影响其他计算
			message.BroadcastToolEnd("get_account_summary", "", err)
		} else {
			for code, price := range prices {
				priceMap[code] = price
			}
		}
	}

	// 计算总市值
	totalValue, err := service.GetAccountTotalValue(database, account.ID, priceMap)
	if err != nil {
		message.BroadcastToolEnd("get_account_summary", "", err)
		return "", err
	}

	// 计算总盈亏
	pnl := totalValue - account.InitialAmount
	pnlRatio := 0.0
	if account.InitialAmount > 0 {
		pnlRatio = float64(pnl) / float64(account.InitialAmount) * 100
	}

	// 格式化输出
	pnlStatus := "亏损"
	if pnl >= 0 {
		pnlStatus = "盈利"
	}

	result := fmt.Sprintf("账户总览：\n\n")
	result += fmt.Sprintf("总资产: %s 元\n", formatHaoToYuan(totalValue))
	result += fmt.Sprintf("可用余额: %s 元\n", formatHaoToYuan(account.AvailableAmt))
	result += fmt.Sprintf("锁定金额: %s 元\n", formatHaoToYuan(account.LockedAmt))
	result += fmt.Sprintf("持仓市值: %s 元\n", formatHaoToYuan(totalValue-account.AvailableAmt-account.LockedAmt))
	result += fmt.Sprintf("总成本: %s 元\n", formatHaoToYuan(account.InitialAmount))
	result += fmt.Sprintf("总盈亏: %s 元 (%.2f%%，%s)\n", formatHaoToYuan(pnl), pnlRatio, pnlStatus)

	message.BroadcastToolEnd("get_account_summary", result, nil)
	return result, nil
}
