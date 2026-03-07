package service

import (
	"fmt"

	"gorm.io/gorm"

	"msa/pkg/db"
	"msa/pkg/model"
)

// AccountService 账户服务
type AccountService struct {
	db *gorm.DB
}

// NewAccountService 创建账户服务实例
func NewAccountService(database *gorm.DB) *AccountService {
	return &AccountService{db: database}
}

// CreateAccount 创建新账户
func (s *AccountService) CreateAccount(userID string, initialAmount float64) (*model.Account, error) {
	// 检查用户是否已存在
	_, err := db.GetAccountByUserID(s.db, userID)
	if err == nil {
		return nil, fmt.Errorf("user already exists: %s", userID)
	}

	// 转换金额为毫
	initialAmtHao := model.YuanToHao(initialAmount)

	// 创建账户
	accountID, err := db.CreateAccount(s.db, userID, initialAmtHao)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	// 返回创建的账户
	return db.GetAccountByID(s.db, accountID)
}

// GetAccount 获取账户信息
func (s *AccountService) GetAccount(userID string) (*model.Account, error) {
	return db.GetAccountByUserID(s.db, userID)
}

// GetAccountByID 根据ID获取账户信息
func (s *AccountService) GetAccountByID(accountID uint) (*model.Account, error) {
	return db.GetAccountByID(s.db, accountID)
}

// FreezeAccount 冻结账户
func (s *AccountService) FreezeAccount(userID string) error {
	account, err := db.GetAccountByUserID(s.db, userID)
	if err != nil {
		return err
	}

	return db.UpdateAccountStatus(s.db, account.ID, model.AccountStatusFrozen)
}

// UnfreezeAccount 解冻账户
func (s *AccountService) UnfreezeAccount(userID string) error {
	account, err := db.GetAccountByUserID(s.db, userID)
	if err != nil {
		return err
	}

	return db.UpdateAccountStatus(s.db, account.ID, model.AccountStatusActive)
}

// CloseAccount 关闭账户
func (s *AccountService) CloseAccount(userID string) error {
	account, err := db.GetAccountByUserID(s.db, userID)
	if err != nil {
		return err
	}

	// 验证余额为0
	if account.AvailableAmt != 0 || account.LockedAmt != 0 {
		return fmt.Errorf("cannot close account with non-zero balance")
	}

	return db.UpdateAccountStatus(s.db, account.ID, model.AccountStatusClosed)
}
