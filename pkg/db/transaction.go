package db

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"msa/pkg/model"
)

// CreateTransaction 在事务中创建交易记录
func CreateTransaction(tx *gorm.DB, trans *model.Transaction) (uint, error) {
	if err := tx.Create(trans).Error; err != nil {
		return 0, fmt.Errorf("failed to create transaction: %w", err)
	}

	return trans.ID, nil
}

// GetTransactionByID 根据ID查询单个交易记录
func GetTransactionByID(db *gorm.DB, id uint) (*model.Transaction, error) {
	var trans model.Transaction
	result := db.First(&trans, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("transaction not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get transaction: %w", result.Error)
	}

	return &trans, nil
}

// UpdateTransactionStatus 更新交易状态
func UpdateTransactionStatus(db *gorm.DB, id uint, status model.TransactionStatus) error {
	result := db.Model(&model.Transaction{}).Where("id = ?", id).Update("status", status)

	if result.Error != nil {
		return fmt.Errorf("failed to update transaction status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("transaction not found: %d", id)
	}

	return nil
}

// GetTransactionsByAccount 按账户ID查询交易记录
func GetTransactionsByAccount(db *gorm.DB, accountID uint) ([]*model.Transaction, error) {
	var transactions []*model.Transaction
	result := db.Where("account_id = ?", accountID).Order("created_at DESC").Find(&transactions)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", result.Error)
	}

	return transactions, nil
}

// GetTransactionsByStock 按股票代码查询交易记录
func GetTransactionsByStock(db *gorm.DB, stockCode string) ([]*model.Transaction, error) {
	var transactions []*model.Transaction
	result := db.Where("stock_code = ?", stockCode).Order("created_at DESC").Find(&transactions)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", result.Error)
	}

	return transactions, nil
}

// GetTransactionsByType 按交易类型查询交易记录
func GetTransactionsByType(db *gorm.DB, transType model.TransactionType) ([]*model.Transaction, error) {
	var transactions []*model.Transaction
	result := db.Where("type = ?", transType).Order("created_at DESC").Find(&transactions)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", result.Error)
	}

	return transactions, nil
}

// GetTransactionsByStatus 按交易状态查询交易记录
func GetTransactionsByStatus(db *gorm.DB, status model.TransactionStatus) ([]*model.Transaction, error) {
	var transactions []*model.Transaction
	result := db.Where("status = ?", status).Order("created_at DESC").Find(&transactions)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", result.Error)
	}

	return transactions, nil
}

// GetTransactionsByParent 按父交易ID查询子交易记录（用于部分成交）
func GetTransactionsByParent(db *gorm.DB, parentID uint) ([]*model.Transaction, error) {
	var transactions []*model.Transaction
	result := db.Where("parent_id = ?", parentID).Order("created_at ASC").Find(&transactions)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", result.Error)
	}

	return transactions, nil
}

// UpdateTransactionsWithParent 用于部分成交时更新子交易的 parent_id
func UpdateTransactionsWithParent(tx *gorm.DB, parentID uint, childIDs []uint) error {
	if len(childIDs) == 0 {
		return nil
	}

	result := tx.Model(&model.Transaction{}).Where("id IN ?", childIDs).Update("parent_id", parentID)
	if result.Error != nil {
		return fmt.Errorf("failed to update transactions parent_id: %w", result.Error)
	}

	return nil
}

// GetTodayTransactions 获取今日交易记录
func GetTodayTransactions() ([]model.Transaction, error) {
	db := GetDB()
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	var transactions []model.Transaction

	// 获取今天的开始和结束时间
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	result := db.Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).
		Order("created_at DESC").
		Find(&transactions)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to query today's transactions: %w", result.Error)
	}

	return transactions, nil
}
