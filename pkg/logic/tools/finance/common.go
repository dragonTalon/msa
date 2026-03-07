package finance

import (
	"errors"
	"fmt"
	"strconv"

	"gorm.io/gorm"
	"msa/pkg/logic/tools/stock"
	"msa/pkg/model"
)

// getActiveAccount 获取当前活跃账户
// 全局只有一个 ACTIVE 状态的账户
func getActiveAccount(db *gorm.DB) (*model.Account, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库连接为空")
	}
	var account model.Account
	err := db.Where("status = ?", model.AccountStatusActive).
		First(&account).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("未找到活跃账户，请先创建账户\n创建账户需要提供：initial_amount（初始金额，单位：元）\n例如：创建一个初始资金 10 万元的账户")
		}
		return nil, fmt.Errorf("查询账户失败: %w", err)
	}
	return &account, nil
}

// fetchCurrentPrice 获取股票当前价格（毫）
// 复用 stock/FetchStockData()
func fetchCurrentPrice(stockCode string) (int64, error) {
	resp, err := stock.FetchStockData(stockCode)
	if err != nil {
		return 0, err
	}
	// resp.CurrentPrice 是字符串，如 "10.50"
	price, err := strconv.ParseFloat(resp.CurrentPrice, 64)
	if err != nil {
		return 0, fmt.Errorf("解析价格失败: %w", err)
	}
	return model.YuanToHao(price), nil
}

// fetchAllPrices 批量获取股票价格
// 返回 map[stockCode]price (毫)
// 如果某只股票获取失败，跳过该股票
func fetchAllPrices(stockCodes []string) (map[string]int64, error) {
	prices := make(map[string]int64)
	for _, stockCode := range stockCodes {
		price, err := fetchCurrentPrice(stockCode)
		if err != nil {
			// 跳过获取失败的股票
			continue
		}
		prices[stockCode] = price
	}
	return prices, nil
}

// formatHaoToYuan 毫转元格式化
// 保留两位小数
func formatHaoToYuan(hao int64) string {
	return model.FormatAmount(hao)
}

// formatHaoToYuanFloat 毫转元浮点数
func formatHaoToYuanFloat(hao int64) float64 {
	return model.HaoToYuan(hao)
}

// getOrderStatusText 获取订单状态文本
func getOrderStatusText(status model.TransactionStatus) string {
	switch status {
	case model.TransactionStatusPending:
		return "待成交"
	case model.TransactionStatusFilled:
		return "已成交"
	case model.TransactionStatusRejected:
		return "已拒绝"
	default:
		return "未知状态"
	}
}

// getTransactionTypeText 获取交易类型文本
func getTransactionTypeText(txType model.TransactionType) string {
	switch txType {
	case model.TransactionTypeBuy:
		return "买入"
	case model.TransactionTypeSell:
		return "卖出"
	default:
		return "未知"
	}
}
