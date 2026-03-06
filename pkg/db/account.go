package db

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"msa/pkg/model"
)

// CreateAccount 创建新账户
// 返回新创建的账户ID和可能的错误
func CreateAccount(db *gorm.DB, userID string, initialAmount int64) (uint, error) {
	account := &model.Account{
		UserID:        userID,
		InitialAmount: initialAmount,
		AvailableAmt:  initialAmount,
		LockedAmt:     0,
		Status:        model.AccountStatusActive,
	}

	if err := db.Create(account).Error; err != nil {
		return 0, fmt.Errorf("failed to create account: %w", err)
	}

	return account.ID, nil
}

// GetAccountByUserID 根据用户ID查询账户
func GetAccountByUserID(db *gorm.DB, userID string) (*model.Account, error) {
	var account model.Account
	result := db.Where("user_id = ?", userID).First(&account)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("account not found: %s", userID)
		}
		return nil, fmt.Errorf("failed to get account: %w", result.Error)
	}

	return &account, nil
}

// GetAccountByID 根据账户ID查询账户
func GetAccountByID(db *gorm.DB, id uint) (*model.Account, error) {
	var account model.Account
	result := db.First(&account, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("account not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get account: %w", result.Error)
	}

	return &account, nil
}

// UpdateAccountStatus 更新账户状态
func UpdateAccountStatus(db *gorm.DB, id uint, status model.AccountStatus) error {
	result := db.Model(&model.Account{}).Where("id = ?", id).Update("status", status)

	if result.Error != nil {
		return fmt.Errorf("failed to update account status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("account not found: %d", id)
	}

	return nil
}

// UpdateAccountAmounts 更新账户金额（必须在事务中调用）
// deltaAvailable: available_amt 的变化量（正数增加，负数减少）
// deltaLocked: locked_amt 的变化量（正数增加，负数减少）
func UpdateAccountAmounts(tx *gorm.DB, accountID uint, deltaAvailable int64, deltaLocked int64) error {
	result := tx.Model(&model.Account{}).Where("id = ?", accountID).Updates(map[string]interface{}{
		"available_amt": gorm.Expr("available_amt + ?", deltaAvailable),
		"locked_amt":    gorm.Expr("locked_amt + ?", deltaLocked),
	})

	if result.Error != nil {
		return fmt.Errorf("failed to update account amounts: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("account not found: %d", accountID)
	}

	return nil
}
