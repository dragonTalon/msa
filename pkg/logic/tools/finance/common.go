package finance

import (
	"errors"
	"fmt"
	"strconv"

	"msa/pkg/logic/tools/stock"
	"msa/pkg/model"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
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
// 当前价为空时（如停牌），降级使用昨收价（PrevClose）
func fetchCurrentPrice(stockCode string) (int64, error) {
	resp, err := stock.FetchStockData(stockCode)
	if err != nil {
		return 0, err
	}

	// 优先使用当前价，停牌时降级使用昨收价
	priceStr := resp.CurrentPrice
	usedFallback := false
	if priceStr == "" {
		priceStr = resp.PrevClose
		usedFallback = true
	}

	if priceStr == "" {
		return 0, fmt.Errorf("股票 %s 当前价和昨收价均为空，无法获取价格", stockCode)
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0, fmt.Errorf("解析价格失败: %w", err)
	}

	if usedFallback {
		log.Warnf("股票 %s 当前价为空（可能停牌），使用昨收价 %s 计算市值", stockCode, priceStr)
	}

	return model.YuanToHao(price), nil
}

// fetchAllPrices 批量获取股票价格
// 返回 map[stockCode]price (毫)
// 任意一只股票价格获取失败，则返回 error，避免市值计算不完整
func fetchAllPrices(stockCodes []string) (map[string]int64, error) {
	prices := make(map[string]int64)
	var failedCodes []string
	for _, stockCode := range stockCodes {
		price, err := fetchCurrentPrice(stockCode)
		if err != nil {
			log.Errorf("获取股票价格失败: stockCode=%s, err=%v", stockCode, err)
			failedCodes = append(failedCodes, stockCode)
			continue
		}
		prices[stockCode] = price
	}
	if len(failedCodes) > 0 {
		return nil, fmt.Errorf("以下股票价格获取失败: %v", failedCodes)
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
