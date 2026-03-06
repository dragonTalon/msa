package service

import (
	"fmt"

	"gorm.io/gorm"

	"msa/pkg/db"
	"msa/pkg/model"
)

// Position 持仓信息
type Position struct {
	StockCode   string
	StockName   string
	Quantity    int64  // 持仓数量
	Cost        int64  // 持仓成本（分）
	CurrentPrice int64 // 当前价格（分，需要外部传入）
	Value       int64  // 持仓市值（分）
	PnL         int64  // 盈亏（分）
}

// GetPosition 获取指定股票的持仓数量
// 计算：SUM(BUY) - SUM(SELL)
func GetPosition(database *gorm.DB, accountID uint, stockCode string) (int64, error) {
	// 查询已成交的买入数量
	var buyQty int64
	err := database.Model(&model.Transaction{}).
		Select("COALESCE(SUM(quantity), 0)").
		Where("account_id = ? AND stock_code = ? AND type = ? AND status = ?", accountID, stockCode, model.TransactionTypeBuy, model.TransactionStatusFilled).
		Scan(&buyQty).Error
	if err != nil {
		return 0, fmt.Errorf("failed to query buy quantity: %w", err)
	}

	// 查询已成交的卖出数量
	var sellQty int64
	err = database.Model(&model.Transaction{}).
		Select("COALESCE(SUM(quantity), 0)").
		Where("account_id = ? AND stock_code = ? AND type = ? AND status = ?", accountID, stockCode, model.TransactionTypeSell, model.TransactionStatusFilled).
		Scan(&sellQty).Error
	if err != nil {
		return 0, fmt.Errorf("failed to query sell quantity: %w", err)
	}

	return buyQty - sellQty, nil
}

// GetPositionCost 获取指定股票的持仓成本
// 计算：SUM(BUY金额) - SUM(SELL金额)
func GetPositionCost(database *gorm.DB, accountID uint, stockCode string) (int64, error) {
	// 查询已成交的买入金额（含手续费）
	var buyAmount int64
	err := database.Model(&model.Transaction{}).
		Select("COALESCE(SUM(amount + fee), 0)").
		Where("account_id = ? AND stock_code = ? AND type = ? AND status = ?", accountID, stockCode, model.TransactionTypeBuy, model.TransactionStatusFilled).
		Scan(&buyAmount).Error
	if err != nil {
		return 0, fmt.Errorf("failed to query buy amount: %w", err)
	}

	// 查询已成交的卖出金额（净收入）
	var sellAmount int64
	err = database.Model(&model.Transaction{}).
		Select("COALESCE(SUM(amount - fee), 0)").
		Where("account_id = ? AND stock_code = ? AND type = ? AND status = ?", accountID, stockCode, model.TransactionTypeSell, model.TransactionStatusFilled).
		Scan(&sellAmount).Error
	if err != nil {
		return 0, fmt.Errorf("failed to query sell amount: %w", err)
	}

	return buyAmount - sellAmount, nil
}

// GetPositionValue 获取指定股票的持仓市值
// 调用实时价格 API，计算：持仓数量 × 实时价格
func GetPositionValue(database *gorm.DB, accountID uint, stockCode string, price int64) (int64, error) {
	qty, err := GetPosition(database, accountID, stockCode)
	if err != nil {
		return 0, err
	}

	return qty * price, nil
}

// GetAccountTotalCost 获取账户总成本
// 计算：available_amt + locked_amt + 持仓成本
func GetAccountTotalCost(database *gorm.DB, accountID uint) (int64, error) {
	// 查询账户
	account, err := db.GetAccountByID(database, accountID)
	if err != nil {
		return 0, err
	}

	// 查询所有持仓的已成交买入金额（含手续费）
	var buyAmount int64
	err = database.Model(&model.Transaction{}).
		Select("COALESCE(SUM(amount + fee), 0)").
		Where("account_id = ? AND type = ? AND status = ?", accountID, model.TransactionTypeBuy, model.TransactionStatusFilled).
		Scan(&buyAmount).Error
	if err != nil {
		return 0, fmt.Errorf("failed to query buy amount: %w", err)
	}

	// 查询所有持仓的已成交卖出金额（净收入）
	var sellAmount int64
	err = database.Model(&model.Transaction{}).
		Select("COALESCE(SUM(amount - fee), 0)").
		Where("account_id = ? AND type = ? AND status = ?", accountID, model.TransactionTypeSell, model.TransactionStatusFilled).
		Scan(&sellAmount).Error
	if err != nil {
		return 0, fmt.Errorf("failed to query sell amount: %w", err)
	}

	positionCost := buyAmount - sellAmount
	totalCost := account.AvailableAmt + account.LockedAmt + positionCost

	return totalCost, nil
}

// GetAccountTotalValue 获取账户总市值
// 计算：available_amt + locked_amt + 持仓市值（需要外部提供价格）
type PriceMap map[string]int64 // stock_code -> price (分)

func GetAccountTotalValue(database *gorm.DB, accountID uint, prices PriceMap) (int64, error) {
	// 查询账户
	account, err := db.GetAccountByID(database, accountID)
	if err != nil {
		return 0, err
	}

	// 查询所有持仓股票
	var stockCodes []string
	err = database.Model(&model.Transaction{}).
		Select("DISTINCT stock_code").
		Where("account_id = ? AND status = ?", accountID, model.TransactionStatusFilled).
		Pluck("stock_code", &stockCodes).Error
	if err != nil {
		return 0, fmt.Errorf("failed to query stocks: %w", err)
	}

	positionValue := int64(0)
	for _, stockCode := range stockCodes {
		price, ok := prices[stockCode]
		if !ok {
			continue // 跳过没有价格的股票
		}

		qty, err := GetPosition(database, accountID, stockCode)
		if err != nil {
			return 0, err
		}

		positionValue += qty * price
	}

	totalValue := account.AvailableAmt + account.LockedAmt + positionValue
	return totalValue, nil
}

// GetAccountPnL 获取账户盈亏
// 计算：总市值 - initial_amount
func GetAccountPnL(database *gorm.DB, accountID uint, prices PriceMap) (int64, error) {
	// 查询账户
	account, err := db.GetAccountByID(database, accountID)
	if err != nil {
		return 0, err
	}

	totalValue, err := GetAccountTotalValue(database, accountID, prices)
	if err != nil {
		return 0, err
	}

	pnl := totalValue - account.InitialAmount
	return pnl, nil
}

// GetAllPositions 获取所有持仓
func GetAllPositions(database *gorm.DB, accountID uint, prices PriceMap) ([]*Position, error) {
	// 查询所有有交易的股票（聚合查询）
	type StockInfo struct {
		StockCode string
		StockName string
	}

	var stockInfos []StockInfo
	err := database.Model(&model.Transaction{}).
		Select("DISTINCT stock_code, stock_name").
		Where("account_id = ? AND status = ?", accountID, model.TransactionStatusFilled).
		Order("stock_code").
		Scan(&stockInfos).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query stocks: %w", err)
	}

	var positions []*Position
	for _, info := range stockInfos {
		qty, err := GetPosition(database, accountID, info.StockCode)
		if err != nil {
			return nil, err
		}

		// 只返回有持仓的股票
		if qty <= 0 {
			continue
		}

		cost, err := GetPositionCost(database, accountID, info.StockCode)
		if err != nil {
			return nil, err
		}

		price, ok := prices[info.StockCode]
		if !ok {
			price = 0 // 没有价格信息
		}

		value := qty * price
		pnl := value - cost

		positions = append(positions, &Position{
			StockCode:    info.StockCode,
			StockName:    info.StockName,
			Quantity:     qty,
			Cost:         cost,
			CurrentPrice: price,
			Value:        value,
			PnL:          pnl,
		})
	}

	return positions, nil
}
