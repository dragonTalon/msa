package finance

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	msadb "msa/pkg/db"
	"msa/pkg/logic/finsvc"
	"msa/pkg/logic/message"
	"msa/pkg/model"
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

// PositionItem 持仓项
type PositionItem struct {
	StockCode    string `json:"stock_code"`
	StockName    string `json:"stock_name"`
	Quantity     int64  `json:"quantity"`
	Cost         string `json:"cost"`
	CurrentPrice string `json:"current_price"`
	Value        string `json:"value"`
	PnL          string `json:"pnl"`
	PnLStatus    string `json:"pnl_status"`
}

// PositionsData 持仓数据
type PositionsData struct {
	Total int            `json:"total"`
	Items []PositionItem `json:"items"`
}

// GetPositions 获取持仓列表
func GetPositions(ctx context.Context, param *GetPositionsParam) (string, error) {
	message.BroadcastToolStart("get_positions", "")

	database := msadb.GetDB()
	if database == nil {
		err := fmt.Errorf("数据库未初始化")
		message.BroadcastToolEnd("get_positions", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	account, err := getActiveAccount(database)
	if err != nil {
		message.BroadcastToolEnd("get_positions", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	// 获取所有持仓股票代码
	var stockCodes []string
	err = database.Model(&model.Transaction{}).
		Select("DISTINCT stock_code").
		Where("account_id = ? AND status = ?", account.ID, model.TransactionStatusFilled).
		Pluck("stock_code", &stockCodes).Error
	if err != nil {
		message.BroadcastToolEnd("get_positions", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	if len(stockCodes) == 0 {
		message.BroadcastToolEnd("get_positions", "当前无持仓", nil)
		return model.NewSuccessResult(&PositionsData{Total: 0, Items: []PositionItem{}}, "当前无持仓"), nil
	}

	// 批量获取价格
	prices, err := fetchAllPrices(stockCodes)
	if err != nil {
		message.BroadcastToolEnd("get_positions", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	// 转换为 finsvc.PriceMap 格式
	priceMap := make(finsvc.PriceMap)
	for code, price := range prices {
		priceMap[code] = price
	}

	// 获取所有持仓
	positions, err := finsvc.GetAllPositions(database, account.ID, priceMap)
	if err != nil {
		message.BroadcastToolEnd("get_positions", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	if len(positions) == 0 {
		message.BroadcastToolEnd("get_positions", "当前无持仓", nil)
		return model.NewSuccessResult(&PositionsData{Total: 0, Items: []PositionItem{}}, "当前无持仓"), nil
	}

	// 构建返回数据
	items := make([]PositionItem, 0, len(positions))
	for _, pos := range positions {
		pnlStatus := "亏损"
		if pos.PnL >= 0 {
			pnlStatus = "盈利"
		}
		items = append(items, PositionItem{
			StockCode:    pos.StockCode,
			StockName:    pos.StockName,
			Quantity:     pos.Quantity,
			Cost:         formatHaoToYuan(pos.Cost),
			CurrentPrice: formatHaoToYuan(pos.CurrentPrice),
			Value:        formatHaoToYuan(pos.Value),
			PnL:          formatHaoToYuan(pos.PnL),
			PnLStatus:    pnlStatus,
		})
	}

	data := &PositionsData{
		Total: len(positions),
		Items: items,
	}

	message.BroadcastToolEnd("get_positions", fmt.Sprintf("获取 %d 只持仓", len(positions)), nil)
	return model.NewSuccessResult(data, fmt.Sprintf("获取 %d 只持仓", len(positions))), nil
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

// AccountSummaryData 账户总览数据
type AccountSummaryData struct {
	TotalAssets   string  `json:"total_assets"`
	AvailableAmt  string  `json:"available_amt"`
	LockedAmt     string  `json:"locked_amt"`
	PositionValue string  `json:"position_value"`
	InitialAmount string  `json:"initial_amount"`
	TotalPnL      string  `json:"total_pnl"`
	PnLRatio      float64 `json:"pnl_ratio"`
	PnLStatus     string  `json:"pnl_status"`
}

// GetAccountSummary 获取账户总览
func GetAccountSummary(ctx context.Context, param *GetAccountSummaryParam) (string, error) {
	message.BroadcastToolStart("get_account_summary", "")

	database := msadb.GetDB()
	if database == nil {
		err := fmt.Errorf("数据库未初始化")
		message.BroadcastToolEnd("get_account_summary", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	account, err := getActiveAccount(database)
	if err != nil {
		message.BroadcastToolEnd("get_account_summary", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	// 获取所有持仓股票代码
	var stockCodes []string
	err = database.Model(&model.Transaction{}).
		Select("DISTINCT stock_code").
		Where("account_id = ? AND status = ?", account.ID, model.TransactionStatusFilled).
		Pluck("stock_code", &stockCodes).Error
	if err != nil {
		message.BroadcastToolEnd("get_account_summary", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	// 批量获取价格
	priceMap := make(finsvc.PriceMap)
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
	totalValue, err := finsvc.GetAccountTotalValue(database, account.ID, priceMap)
	if err != nil {
		message.BroadcastToolEnd("get_account_summary", "", err)
		return model.NewErrorResult(err.Error()), nil
	}

	// 计算总盈亏
	pnl := totalValue - account.InitialAmount
	pnlRatio := 0.0
	if account.InitialAmount > 0 {
		pnlRatio = float64(pnl) / float64(account.InitialAmount) * 100
	}

	pnlStatus := "亏损"
	if pnl >= 0 {
		pnlStatus = "盈利"
	}

	data := &AccountSummaryData{
		TotalAssets:   formatHaoToYuan(totalValue),
		AvailableAmt:  formatHaoToYuan(account.AvailableAmt),
		LockedAmt:     formatHaoToYuan(account.LockedAmt),
		PositionValue: formatHaoToYuan(totalValue - account.AvailableAmt - account.LockedAmt),
		InitialAmount: formatHaoToYuan(account.InitialAmount),
		TotalPnL:      formatHaoToYuan(pnl),
		PnLRatio:      pnlRatio,
		PnLStatus:     pnlStatus,
	}

	message.BroadcastToolEnd("get_account_summary", "获取账户总览成功", nil)
	return model.NewSuccessResult(data, "获取账户总览成功"), nil
}
