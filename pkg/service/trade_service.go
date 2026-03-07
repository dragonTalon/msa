package service

import (
	"fmt"

	"gorm.io/gorm"

	"msa/pkg/db"
	"msa/pkg/model"
)

// Order 订单信息
type Order struct {
	StockCode string
	StockName string
	Quantity  int64
	Price     int64 // 价格（毫）
	Fee       int64 // 手续费（毫）
	Note      string
}

// SubmitBuyOrder 提交买入订单
// 在事务中执行：检查余额 → 创建交易记录 → 锁定金额
// 余额不足时创建 REJECTED 记录
func SubmitBuyOrder(database *gorm.DB, accountID uint, order Order) (uint, error) {
	// 开始事务
	tx := database.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 查询账户
	account, err := db.GetAccountByID(database, accountID)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("failed to get account: %w", err)
	}

	// 计算总金额
	totalAmount := order.Quantity*order.Price + order.Fee

	// 检查余额
	if account.AvailableAmt < totalAmount {
		// 余额不足，创建 REJECTED 记录
		trans := &model.Transaction{
			AccountID: accountID,
			StockCode: order.StockCode,
			StockName: order.StockName,
			Type:      model.TransactionTypeBuy,
			Quantity:  order.Quantity,
			Price:     order.Price,
			Amount:    order.Quantity * order.Price,
			Fee:       order.Fee,
			Status:    model.TransactionStatusRejected,
			Note:      "余额不足",
		}
		transID, err := db.CreateTransaction(tx, trans)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("failed to create rejected transaction: %w", err)
		}
		if err := tx.Commit().Error; err != nil {
			return 0, fmt.Errorf("failed to commit transaction: %w", err)
		}
		return transID, nil
	}

	// 创建交易记录
	trans := &model.Transaction{
		AccountID: accountID,
		StockCode: order.StockCode,
		StockName: order.StockName,
		Type:      model.TransactionTypeBuy,
		Quantity:  order.Quantity,
		Price:     order.Price,
		Amount:    order.Quantity * order.Price,
		Fee:       order.Fee,
		Status:    model.TransactionStatusPending,
		Note:      order.Note,
	}

	transID, err := db.CreateTransaction(tx, trans)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("failed to create transaction: %w", err)
	}

	// 锁定金额
	err = db.UpdateAccountAmounts(tx, accountID, -totalAmount, totalAmount)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("failed to lock amount: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return transID, nil
}

// SubmitSellOrder 提交卖出订单
// 验证持仓充足，不锁定金额
func SubmitSellOrder(database *gorm.DB, accountID uint, order Order) (uint, error) {
	// TODO: 实现持仓验证
	// 暂时跳过，等待持仓计算功能实现

	tx := database.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建交易记录
	trans := &model.Transaction{
		AccountID: accountID,
		StockCode: order.StockCode,
		StockName: order.StockName,
		Type:      model.TransactionTypeSell,
		Quantity:  order.Quantity,
		Price:     order.Price,
		Amount:    order.Quantity * order.Price,
		Fee:       order.Fee,
		Status:    model.TransactionStatusPending,
		Note:      order.Note,
	}

	transID, err := db.CreateTransaction(tx, trans)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("failed to create transaction: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return transID, nil
}

// FillOrder 订单成交处理
// 买入：减少 locked_amt
// 卖出：增加 available_amt
func FillOrder(database *gorm.DB, transID uint) error {
	tx := database.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 查询交易记录
	trans, err := db.GetTransactionByID(database, transID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get transaction: %w", err)
	}

	// 更新交易状态
	if err := db.UpdateTransactionStatus(tx, transID, model.TransactionStatusFilled); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	// 处理金额
	totalAmount := trans.GetTotalAmount()

	if trans.Type == model.TransactionTypeBuy {
		// 买入成交：减少 locked_amt（不返还到 available，已在提交时扣除）
		err = db.UpdateAccountAmounts(tx, trans.AccountID, 0, -totalAmount)
	} else {
		// 卖出成交：增加 available_amt
		sellAmount := trans.Amount - trans.Fee
		err = db.UpdateAccountAmounts(tx, trans.AccountID, sellAmount, 0)
	}

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update account amounts: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// CancelOrder 订单撤销处理
// 买入：locked_amt 返还到 available_amt
// 卖出：仅更新状态
func CancelOrder(database *gorm.DB, transID uint) error {
	tx := database.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 查询交易记录
	trans, err := db.GetTransactionByID(database, transID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get transaction: %w", err)
	}

	// 更新交易状态
	if err := db.UpdateTransactionStatus(tx, transID, model.TransactionStatusCancelled); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	// 买入订单需要解锁金额
	if trans.Type == model.TransactionTypeBuy {
		totalAmount := trans.GetTotalAmount()
		err = db.UpdateAccountAmounts(tx, trans.AccountID, totalAmount, -totalAmount)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to unlock amount: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// PartialFill 部分成交处理
// 原记录状态 → OBSOLETE
// 创建 FILLED 记录（已成交部分）
// 创建 PENDING 记录（剩余部分）
// 调整 locked_amt
func PartialFill(database *gorm.DB, transID uint, filledQty int64) error {
	tx := database.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 查询原始交易
	trans, err := db.GetTransactionByID(database, transID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get transaction: %w", err)
	}

	// 原记录标记为 OBSOLETE
	if err := db.UpdateTransactionStatus(tx, transID, model.TransactionStatusObsolete); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update original transaction: %w", err)
	}

	// 创建已成交部分记录
	filledTrans := &model.Transaction{
		AccountID: trans.AccountID,
		ParentID:  &transID,
		StockCode: trans.StockCode,
		StockName: trans.StockName,
		Type:      trans.Type,
		Quantity:  filledQty,
		Price:     trans.Price,
		Amount:    filledQty * trans.Price,
		Fee:       int64(float64(trans.Fee) * float64(filledQty) / float64(trans.Quantity)), // 按比例分摊手续费
		Status:    model.TransactionStatusFilled,
	}

	_, err = db.CreateTransaction(tx, filledTrans)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create filled transaction: %w", err)
	}

	// 创建剩余部分记录
	remainingQty := trans.Quantity - filledQty
	remainingTrans := &model.Transaction{
		AccountID: trans.AccountID,
		ParentID:  &transID,
		StockCode: trans.StockCode,
		StockName: trans.StockName,
		Type:      trans.Type,
		Quantity:  remainingQty,
		Price:     trans.Price,
		Amount:    remainingQty * trans.Price,
		Fee:       trans.Fee - filledTrans.Fee, // 剩余手续费
		Status:    model.TransactionStatusPending,
	}

	_, err = db.CreateTransaction(tx, remainingTrans)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create remaining transaction: %w", err)
	}

	// 调整锁定金额
	if trans.Type == model.TransactionTypeBuy {
		// 买入：调整 locked_amt 为剩余部分的金额
		remainingAmount := remainingTrans.GetTotalAmount()
		originalAmount := trans.GetTotalAmount()
		delta := remainingAmount - originalAmount // 负数，表示释放部分锁定
		err = db.UpdateAccountAmounts(tx, trans.AccountID, 0, delta)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to adjust locked amount: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
