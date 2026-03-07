package finance

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"msa/pkg/db"
	msadb "msa/pkg/db"
	"msa/pkg/logic/message"
	"msa/pkg/model"
	"msa/pkg/service"
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

// SubmitBuyOrder 提交买入订单
func SubmitBuyOrder(ctx context.Context, param *SubmitBuyOrderParam) (string, error) {
	message.BroadcastToolStart("submit_buy_order", fmt.Sprintf("%s %s %d股@%.4f元",
		param.StockCode, param.StockName, param.Quantity, param.Price))

	database := msadb.GetDB()
	if database == nil {
		err := fmt.Errorf("数据库未初始化")
		message.BroadcastToolEnd("submit_buy_order", "", err)
		return "", err
	}

	account, err := getActiveAccount(database)
	if err != nil {
		message.BroadcastToolEnd("submit_buy_order", "", err)
		return "", err
	}

	// 默认手续费 5 元
	fee := param.Fee
	if fee == 0 {
		fee = 5.0
	}

	// 构建订单
	order := service.Order{
		StockCode: param.StockCode,
		StockName: param.StockName,
		Quantity:  param.Quantity,
		Price:     model.YuanToHao(param.Price),
		Fee:       model.YuanToHao(fee),
		Note:      "",
	}

	// 提交订单
	transID, err := service.SubmitBuyOrder(database, account.ID, order)
	if err != nil {
		message.BroadcastToolEnd("submit_buy_order", "", err)
		return "", err
	}

	// 检查是否被拒绝（余额不足）
	trans, err := db.GetTransactionByID(database, transID)
	if err != nil {
		message.BroadcastToolEnd("submit_buy_order", "", err)
		return "", err
	}

	if trans.Status == model.TransactionStatusRejected {
		err := fmt.Errorf("订单被拒绝：%s", trans.Note)
		message.BroadcastToolEnd("submit_buy_order", "", err)
		return "", err
	}

	// 自动成交
	if err := service.FillOrder(database, transID); err != nil {
		result := fmt.Sprintf("订单已创建（ID: %d），但成交失败：%v", transID, err)
		message.BroadcastToolEnd("submit_buy_order", result, err)
		return "", fmt.Errorf("订单创建成功但成交失败: %w", err)
	}

	result := fmt.Sprintf("买入订单已成交！\n交易ID: %d\n股票: %s (%s)\n数量: %d 股\n价格: %.2f 元/股\n手续费: %.2f 元\n总金额: %.2f 元",
		transID, param.StockName, param.StockCode, param.Quantity, param.Price, fee,
		float64(param.Quantity)*param.Price+fee)

	message.BroadcastToolEnd("submit_buy_order", result, nil)
	return result, nil
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
	message.BroadcastToolStart("submit_sell_order", fmt.Sprintf("%s %s %d股@%.2f元",
		param.StockCode, param.StockName, param.Quantity, param.Price))

	database := msadb.GetDB()
	if database == nil {
		err := fmt.Errorf("数据库未初始化")
		message.BroadcastToolEnd("submit_sell_order", "", err)
		return "", err
	}

	account, err := getActiveAccount(database)
	if err != nil {
		message.BroadcastToolEnd("submit_sell_order", "", err)
		return "", err
	}

	// 验证持仓充足
	position, err := service.GetPosition(database, account.ID, param.StockCode)
	if err != nil {
		message.BroadcastToolEnd("submit_sell_order", "", err)
		return "", err
	}

	if position < param.Quantity {
		err := fmt.Errorf("持仓不足：当前持仓 %d 股，要卖出 %d 股", position, param.Quantity)
		message.BroadcastToolEnd("submit_sell_order", "", err)
		return "", err
	}

	if position == 0 {
		err := fmt.Errorf("无持仓：请先查询持仓")
		message.BroadcastToolEnd("submit_sell_order", "", err)
		return "", err
	}

	// 默认手续费 5 元
	fee := param.Fee
	if fee == 0 {
		fee = 5.0
	}

	// 构建订单
	order := service.Order{
		StockCode: param.StockCode,
		StockName: param.StockName,
		Quantity:  param.Quantity,
		Price:     model.YuanToHao(param.Price),
		Fee:       model.YuanToHao(fee),
		Note:      "",
	}

	// 提交订单
	transID, err := service.SubmitSellOrder(database, account.ID, order)
	if err != nil {
		message.BroadcastToolEnd("submit_sell_order", "", err)
		return "", err
	}

	// 自动成交
	if err := service.FillOrder(database, transID); err != nil {
		result := fmt.Sprintf("订单已创建（ID: %d），但成交失败：%v", transID, err)
		message.BroadcastToolEnd("submit_sell_order", result, err)
		return "", fmt.Errorf("订单创建成功但成交失败: %w", err)
	}

	result := fmt.Sprintf("卖出订单已成交！\n交易ID: %d\n股票: %s (%s)\n数量: %d 股\n价格: %.2f 元/股\n手续费: %.2f 元\n收入: %.2f 元",
		transID, param.StockName, param.StockCode, param.Quantity, param.Price, fee,
		float64(param.Quantity)*param.Price-fee)

	message.BroadcastToolEnd("submit_sell_order", result, nil)
	return result, nil
}

// GetTransactionsParam 查询交易记录参数
type GetTransactionsParam struct {
	StockCode *string `json:"stock_code,omitempty" jsonschema:"description=按股票筛选（可选）"`
	Type      *string `json:"type,omitempty" jsonschema:"description=按类型筛选 BUY/SELL（可选）"`
	Status    *string `json:"status,omitempty" jsonschema:"description=按状态筛选（可选）"`
	Limit     *int    `json:"limit,omitempty" jsonschema:"description=返回数量限制（可选，默认50）"`
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

// GetTransactions 查询交易记录
func GetTransactions(ctx context.Context, param *GetTransactionsParam) (string, error) {
	message.BroadcastToolStart("get_transactions", "")

	database := msadb.GetDB()
	if database == nil {
		err := fmt.Errorf("数据库未初始化")
		message.BroadcastToolEnd("get_transactions", "", err)
		return "", err
	}

	account, err := getActiveAccount(database)
	if err != nil {
		message.BroadcastToolEnd("get_transactions", "", err)
		return "", err
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

	// 默认限制 50 条
	limit := 50
	if param.Limit != nil {
		limit = *param.Limit
	}
	query = query.Order("created_at DESC").Limit(limit)

	var transactions []*model.Transaction
	if err := query.Find(&transactions).Error; err != nil {
		message.BroadcastToolEnd("get_transactions", "", err)
		return "", err
	}

	if len(transactions) == 0 {
		result := "暂无交易记录"
		message.BroadcastToolEnd("get_transactions", result, nil)
		return result, nil
	}

	// 格式化输出
	result := fmt.Sprintf("交易记录（共 %d 条）：\n\n", len(transactions))
	for _, trans := range transactions {
		result += fmt.Sprintf("交易ID: %d\n", trans.ID)
		result += fmt.Sprintf("  股票: %s (%s)\n", trans.StockName, trans.StockCode)
		result += fmt.Sprintf("  类型: %s\n", trans.Type)
		result += fmt.Sprintf("  数量: %d 股\n", trans.Quantity)
		result += fmt.Sprintf("  价格: %s 元/股\n", formatHaoToYuan(trans.Price))
		result += fmt.Sprintf("  金额: %s 元\n", formatHaoToYuan(trans.Amount))
		result += fmt.Sprintf("  手续费: %s 元\n", formatHaoToYuan(trans.Fee))
		result += fmt.Sprintf("  状态: %s\n", trans.Status)
		result += fmt.Sprintf("  时间: %s\n", trans.CreatedAt.Format("2006-01-02 15:04:05"))
		result += "\n"
	}

	message.BroadcastToolEnd("get_transactions", fmt.Sprintf("获取 %d 条交易记录", len(transactions)), nil)
	return result, nil
}
