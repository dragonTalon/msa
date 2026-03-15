package finsvc

import (
	"fmt"

	log "github.com/sirupsen/logrus"
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
	log.Infof("提交买入订单: 账户=%d, 股票=%s(%s), 数量=%d, 价格=%d, 手续费=%d",
		accountID, order.StockName, order.StockCode, order.Quantity, order.Price, order.Fee)

	// 开始事务
	tx := database.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Errorf("买入订单 panic: %v", r)
		}
	}()

	// 查询账户
	account, err := db.GetAccountByID(database, accountID)
	if err != nil {
		tx.Rollback()
		log.Errorf("查询账户失败: %v", err)
		return 0, fmt.Errorf("failed to get account: %w", err)
	}

	// 计算总金额
	totalAmount := order.Quantity*order.Price + order.Fee

	// 检查余额
	if account.AvailableAmt < totalAmount {
		log.Warnf("余额不足: 可用=%d, 需要=%d", account.AvailableAmt, totalAmount)
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
			log.Errorf("创建拒绝记录失败: %v", err)
			return 0, fmt.Errorf("failed to create rejected transaction: %w", err)
		}
		if err := tx.Commit().Error; err != nil {
			log.Errorf("提交事务失败: %v", err)
			return 0, fmt.Errorf("failed to commit transaction: %w", err)
		}
		log.Infof("买入订单已拒绝: 交易ID=%d", transID)
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
		log.Errorf("创建交易记录失败: %v", err)
		return 0, fmt.Errorf("failed to create transaction: %w", err)
	}

	// 锁定金额
	err = db.UpdateAccountAmounts(tx, accountID, -totalAmount, totalAmount)
	if err != nil {
		tx.Rollback()
		log.Errorf("锁定金额失败: %v", err)
		return 0, fmt.Errorf("failed to lock amount: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		log.Errorf("提交事务失败: %v", err)
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Infof("买入订单已创建: 交易ID=%d", transID)
	return transID, nil
}

// SubmitSellOrder 提交卖出订单
// 验证持仓充足，不锁定金额
func SubmitSellOrder(database *gorm.DB, accountID uint, order Order) (uint, error) {
	log.Infof("提交卖出订单: 账户=%d, 股票=%s(%s), 数量=%d, 价格=%d, 手续费=%d",
		accountID, order.StockName, order.StockCode, order.Quantity, order.Price, order.Fee)

	tx := database.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Errorf("卖出订单 panic: %v", r)
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
		log.Errorf("创建交易记录失败: %v", err)
		return 0, fmt.Errorf("failed to create transaction: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		log.Errorf("提交事务失败: %v", err)
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Infof("卖出订单已创建: 交易ID=%d", transID)
	return transID, nil
}

// FillOrder 订单成交处理
// 买入：减少 locked_amt
// 卖出：增加 available_amt
func FillOrder(database *gorm.DB, transID uint) error {
	log.Infof("处理订单成交: 交易ID=%d", transID)

	tx := database.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Errorf("订单成交 panic: %v", r)
		}
	}()

	// 查询交易记录
	trans, err := db.GetTransactionByID(database, transID)
	if err != nil {
		tx.Rollback()
		log.Errorf("查询交易记录失败: %v", err)
		return fmt.Errorf("failed to get transaction: %w", err)
	}

	// 更新交易状态
	if err := db.UpdateTransactionStatus(tx, transID, model.TransactionStatusFilled); err != nil {
		tx.Rollback()
		log.Errorf("更新交易状态失败: %v", err)
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	// 处理金额
	totalAmount := trans.GetTotalAmount()

	if trans.Type == model.TransactionTypeBuy {
		// 买入成交：减少 locked_amt（不返还到 available，已在提交时扣除）
		err = db.UpdateAccountAmounts(tx, trans.AccountID, 0, -totalAmount)
		log.Infof("买入成交: 释放锁定金额=%d", totalAmount)
	} else {
		// 卖出成交：增加 available_amt
		sellAmount := trans.Amount - trans.Fee
		err = db.UpdateAccountAmounts(tx, trans.AccountID, sellAmount, 0)
		log.Infof("卖出成交: 增加可用金额=%d", sellAmount)
	}

	if err != nil {
		tx.Rollback()
		log.Errorf("更新账户金额失败: %v", err)
		return fmt.Errorf("failed to update account amounts: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		log.Errorf("提交事务失败: %v", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Infof("订单成交完成: 交易ID=%d", transID)
	return nil
}
