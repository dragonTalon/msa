package finance

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"msa/pkg/db"
	msadb "msa/pkg/db"
	"msa/pkg/logic/finsvc"
	"msa/pkg/logic/tools/safetool"
	"msa/pkg/model"
)

// SubmitBuyOrderParam 提交买入订单参数
type SubmitBuyOrderParam struct {
	StockCode string  `json:"stock_code" jsonschema:"description=股票代码（如 sh600000）"`
	StockName string  `json:"stock_name" jsonschema:"description=股票名称"`
	Quantity  int64   `json:"quantity" jsonschema:"description=买入数量（股）"`
	Price     float64 `json:"price" jsonschema:"description=买入价格（元/股）"`
	Fee       float64 `json:"fee" jsonschema:"description=手续费（元），默认5元"`
}

// SubmitBuyOrderTool 提交买入订单工具
type SubmitBuyOrderTool struct{}

func (t *SubmitBuyOrderTool) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(t.GetName(), t.GetDescription(), SubmitBuyOrder)
}

func (t *SubmitBuyOrderTool) GetName() string {
	return "submit_buy_order"
}

func (t *SubmitBuyOrderTool) GetDescription() string {
	return "提交买入订单并自动成交| Submit buy order and auto-fill"
}

func (t *SubmitBuyOrderTool) GetToolGroup() model.ToolGroup {
	return model.FinanceToolGroup
}

// OrderData 订单数据
type OrderData struct {
	TransactionID int64   `json:"transaction_id"`
	StockCode     string  `json:"stock_code"`
	StockName     string  `json:"stock_name"`
	Quantity      int64   `json:"quantity"`
	Price         float64 `json:"price"`
	Fee           float64 `json:"fee"`
	TotalAmount   float64 `json:"total_amount"`
}

// SubmitBuyOrder 提交买入订单
func SubmitBuyOrder(ctx context.Context, param *SubmitBuyOrderParam) (string, error) {
	return safetool.SafeExecute("submit_buy_order", fmt.Sprintf("%s %s %d股@%.4f元",
		param.StockCode, param.StockName, param.Quantity, param.Price), func() (string, error) {
		return doSubmitBuyOrder(ctx, param)
	})
}

func doSubmitBuyOrder(ctx context.Context, param *SubmitBuyOrderParam) (string, error) {
	database := msadb.GetDB()
	if database == nil {
		err := fmt.Errorf("数据库未初始化")
		return model.NewErrorResult(err.Error()), nil
	}

	account, err := getActiveAccount(database)
	if err != nil {
		return model.NewErrorResult(err.Error()), nil
	}

	// 默认手续费 5 元
	fee := param.Fee
	if fee == 0 {
		fee = 5.0
	}

	// 构建订单
	order := finsvc.Order{
		StockCode: param.StockCode,
		StockName: param.StockName,
		Quantity:  param.Quantity,
		Price:     model.YuanToHao(param.Price),
		Fee:       model.YuanToHao(fee),
		Note:      "",
	}

	// 提交订单
	transID, err := finsvc.SubmitBuyOrder(database, account.ID, order)
	if err != nil {
		return model.NewErrorResult(err.Error()), nil
	}

	// 检查是否被拒绝（余额不足）
	trans, err := db.GetTransactionByID(database, transID)
	if err != nil {
		return model.NewErrorResult(err.Error()), nil
	}

	if trans.Status == model.TransactionStatusRejected {
		err := fmt.Errorf("订单被拒绝：%s", trans.Note)
		return model.NewErrorResult(err.Error()), nil
	}

	// 自动成交
	if err := finsvc.FillOrder(database, transID); err != nil {
		return model.NewErrorResult(fmt.Sprintf("订单创建成功但成交失败: %v", err)), nil
	}

	data := &OrderData{
		TransactionID: int64(transID),
		StockCode:     param.StockCode,
		StockName:     param.StockName,
		Quantity:      param.Quantity,
		Price:         param.Price,
		Fee:           fee,
		TotalAmount:   float64(param.Quantity)*param.Price + fee,
	}

	return model.NewSuccessResult(data, "买入订单已成交"), nil
}

// SubmitSellOrderParam 提交卖出订单参数
type SubmitSellOrderParam struct {
	StockCode string  `json:"stock_code" jsonschema:"description=股票代码（如 sh600000）"`
	StockName string  `json:"stock_name" jsonschema:"description=股票名称"`
	Quantity  int64   `json:"quantity" jsonschema:"description=卖出数量（股）"`
	Price     float64 `json:"price" jsonschema:"description=卖出价格（元/股）"`
	Fee       float64 `json:"fee" jsonschema:"description=手续费（元），默认5元"`
}

// SubmitSellOrderTool 提交卖出订单工具
type SubmitSellOrderTool struct{}

func (t *SubmitSellOrderTool) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(t.GetName(), t.GetDescription(), SubmitSellOrder)
}

func (t *SubmitSellOrderTool) GetName() string {
	return "submit_sell_order"
}

func (t *SubmitSellOrderTool) GetDescription() string {
	return "提交卖出订单并自动成交| Submit sell order and auto-fill"
}

func (t *SubmitSellOrderTool) GetToolGroup() model.ToolGroup {
	return model.FinanceToolGroup
}

// SubmitSellOrder 提交卖出订单
func SubmitSellOrder(ctx context.Context, param *SubmitSellOrderParam) (string, error) {
	return safetool.SafeExecute("submit_sell_order", fmt.Sprintf("%s %s %d股@%.2f元",
		param.StockCode, param.StockName, param.Quantity, param.Price), func() (string, error) {
		return doSubmitSellOrder(ctx, param)
	})
}

func doSubmitSellOrder(ctx context.Context, param *SubmitSellOrderParam) (string, error) {
	database := msadb.GetDB()
	if database == nil {
		err := fmt.Errorf("数据库未初始化")
		return model.NewErrorResult(err.Error()), nil
	}

	account, err := getActiveAccount(database)
	if err != nil {
		return model.NewErrorResult(err.Error()), nil
	}

	// 验证持仓充足
	position, err := finsvc.GetPosition(database, account.ID, param.StockCode)
	if err != nil {
		return model.NewErrorResult(err.Error()), nil
	}

	if position == 0 {
		err := fmt.Errorf("无持仓：请先查询持仓")
		return model.NewErrorResult(err.Error()), nil
	}

	if position < param.Quantity {
		err := fmt.Errorf("持仓不足：当前持仓 %d 股，要卖出 %d 股", position, param.Quantity)
		return model.NewErrorResult(err.Error()), nil
	}

	// 默认手续费 5 元
	fee := param.Fee
	if fee == 0 {
		fee = 5.0
	}

	// 构建订单
	order := finsvc.Order{
		StockCode: param.StockCode,
		StockName: param.StockName,
		Quantity:  param.Quantity,
		Price:     model.YuanToHao(param.Price),
		Fee:       model.YuanToHao(fee),
		Note:      "",
	}

	// 提交订单
	transID, err := finsvc.SubmitSellOrder(database, account.ID, order)
	if err != nil {
		return model.NewErrorResult(err.Error()), nil
	}

	// 自动成交
	if err := finsvc.FillOrder(database, transID); err != nil {
		return model.NewErrorResult(fmt.Sprintf("订单创建成功但成交失败: %v", err)), nil
	}

	data := &OrderData{
		TransactionID: int64(transID),
		StockCode:     param.StockCode,
		StockName:     param.StockName,
		Quantity:      param.Quantity,
		Price:         param.Price,
		Fee:           fee,
		TotalAmount:   float64(param.Quantity)*param.Price - fee,
	}

	return model.NewSuccessResult(data, "卖出订单已成交"), nil
}

// GetTransactionsParam 查询交易记录参数
type GetTransactionsParam struct {
	StockCode *string `json:"stock_code,omitempty" jsonschema:"description=按股票筛选（可选）"`
	Type      *string `json:"type,omitempty" jsonschema:"description=按类型筛选 BUY/SELL（可选）"`
	Status    *string `json:"status,omitempty" jsonschema:"description=按状态筛选（可选）"`
	Limit     *int    `json:"limit,omitempty" jsonschema:"description=返回数量限制（可选，默认50）"`
	DateFrom  *string `json:"date_from,omitempty" jsonschema:"description=起始日期 YYYY-MM-DD（可选）"`
	DateTo    *string `json:"date_to,omitempty" jsonschema:"description=截止日期 YYYY-MM-DD（可选）"`
}

// GetTransactionsTool 查询交易记录工具
type GetTransactionsTool struct{}

func (t *GetTransactionsTool) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(t.GetName(), t.GetDescription(), GetTransactions)
}

func (t *GetTransactionsTool) GetName() string {
	return "get_transactions"
}

func (t *GetTransactionsTool) GetDescription() string {
	return "查询交易记录（可按股票、类型、状态筛选）| Get transaction records (can filter by stock, type, status)"
}

func (t *GetTransactionsTool) GetToolGroup() model.ToolGroup {
	return model.FinanceToolGroup
}

// TransactionData 交易记录数据
type TransactionData struct {
	Total int64             `json:"total"`
	Items []TransactionItem `json:"items"`
}

// TransactionItem 交易记录项
type TransactionItem struct {
	ID        int64  `json:"id"`
	StockCode string `json:"stock_code"`
	StockName string `json:"stock_name"`
	Type      string `json:"type"`
	Quantity  int64  `json:"quantity"`
	Price     string `json:"price"`
	Amount    string `json:"amount"`
	Fee       string `json:"fee"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// GetTransactions 查询交易记录
func GetTransactions(ctx context.Context, param *GetTransactionsParam) (string, error) {
	return safetool.SafeExecute("get_transactions", "", func() (string, error) {
		return doGetTransactions(ctx, param)
	})
}

func doGetTransactions(ctx context.Context, param *GetTransactionsParam) (string, error) {
	database := msadb.GetDB()
	if database == nil {
		err := fmt.Errorf("数据库未初始化")
		return model.NewErrorResult(err.Error()), nil
	}

	account, err := getActiveAccount(database)
	if err != nil {
		return model.NewErrorResult(err.Error()), nil
	}

	// 构建查询
	query := database.Model(&model.Transaction{}).Where("account_id = ?", account.ID)

	if param.StockCode != nil {
		query = query.Where("stock_code = ?", *param.StockCode)
	}
	if param.Type != nil {
		query = query.Where("type = ?", *param.Type)
	}
	if param.Status != nil {
		query = query.Where("status = ?", *param.Status)
	}
	if param.DateFrom != nil && *param.DateFrom != "" {
		query = query.Where("created_at >= ?", *param.DateFrom+" 00:00:00")
	}
	if param.DateTo != nil && *param.DateTo != "" {
		query = query.Where("created_at <= ?", *param.DateTo+" 23:59:59")
	}

	// 默认限制 50 条
	limit := 50
	if param.Limit != nil {
		limit = *param.Limit
	}
	query = query.Order("created_at DESC").Limit(limit)

	var transactions []*model.Transaction
	if err := query.Find(&transactions).Error; err != nil {
		return model.NewErrorResult(err.Error()), nil
	}

	// 构建返回数据
	items := make([]TransactionItem, 0, len(transactions))
	for _, trans := range transactions {
		items = append(items, TransactionItem{
			ID:        int64(trans.ID),
			StockCode: trans.StockCode,
			StockName: trans.StockName,
			Type:      string(trans.Type),
			Quantity:  trans.Quantity,
			Price:     formatHaoToYuan(trans.Price),
			Amount:    formatHaoToYuan(trans.Amount),
			Fee:       formatHaoToYuan(trans.Fee),
			Status:    string(trans.Status),
			CreatedAt: trans.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	data := &TransactionData{
		Total: int64(len(transactions)),
		Items: items,
	}

	return model.NewSuccessResult(data, fmt.Sprintf("获取 %d 条交易记录", len(transactions))), nil
}
