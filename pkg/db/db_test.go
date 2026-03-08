package db

import (
	"path/filepath"
	"testing"

	"gorm.io/gorm"

	"msa/pkg/model"
)

func setupTestDB(t *testing.T) *gorm.DB {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.sqlite")

	database, err := InitDBWithPath(dbPath)
	if err != nil {
		t.Fatalf("InitDBWithPath failed: %v", err)
	}

	if err := Migrate(database); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	return database
}

// TestInitDB 测试数据库初始化
func TestInitDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.sqlite")

	// 测试数据库初始化
	database, err := InitDBWithPath(dbPath)
	if err != nil {
		t.Fatalf("InitDBWithPath failed: %v", err)
	}
	defer CloseDB(database)

	// 验证连接可用
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		t.Errorf("Database ping failed: %v", err)
	}
}

// TestMigrate 测试数据库迁移
func TestMigrate(t *testing.T) {
	database := setupTestDB(t)
	defer CloseDB(database)

	// 验证 accounts 表已创建
	var tableName string
	err := database.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name='accounts'").Scan(&tableName).Error
	if err != nil {
		t.Errorf("accounts table not created: %v", err)
	}

	// 验证 transactions 表已创建
	err = database.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name='transactions'").Scan(&tableName).Error
	if err != nil {
		t.Errorf("transactions table not created: %v", err)
	}

	// 验证索引已创建
	var indexCount int64
	err = database.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name LIKE 'idx_%'").Scan(&indexCount).Error
	if err != nil {
		t.Errorf("failed to query indexes: %v", err)
	}

	// 应该有至少 8 个索引
	if indexCount < 8 {
		t.Errorf("expected at least 8 indexes, got %d", indexCount)
	}
}

// TestCreateAccount 测试账户创建
func TestCreateAccount(t *testing.T) {
	database := setupTestDB(t)
	defer CloseDB(database)

	userID := "test_user"
	initialAmount := int64(100000) // 10.00元

	accountID, err := CreateAccount(database, userID, initialAmount)
	if err != nil {
		t.Fatalf("CreateAccount failed: %v", err)
	}

	if accountID <= 0 {
		t.Errorf("expected valid account ID, got %d", accountID)
	}

	// 验证账户创建成功
	account, err := GetAccountByID(database, accountID)
	if err != nil {
		t.Fatalf("GetAccountByID failed: %v", err)
	}

	if account.UserID != userID {
		t.Errorf("expected user_id %s, got %s", userID, account.UserID)
	}

	if account.InitialAmount != initialAmount {
		t.Errorf("expected initial_amount %d, got %d", initialAmount, account.InitialAmount)
	}

	if account.AvailableAmt != initialAmount {
		t.Errorf("expected available_amt %d, got %d", initialAmount, account.AvailableAmt)
	}

	if account.LockedAmt != 0 {
		t.Errorf("expected locked_amt 0, got %d", account.LockedAmt)
	}

	if account.Status != model.AccountStatusActive {
		t.Errorf("expected status %s, got %s", model.AccountStatusActive, account.Status)
	}
}

// TestCreateAccountDuplicate 测试用户ID唯一性
func TestCreateAccountDuplicate(t *testing.T) {
	database := setupTestDB(t)
	defer CloseDB(database)

	userID := "test_user"
	initialAmount := int64(100000)

	// 创建第一个账户
	_, err := CreateAccount(database, userID, initialAmount)
	if err != nil {
		t.Fatalf("CreateAccount failed: %v", err)
	}

	// 尝试创建相同用户ID的账户
	_, err = CreateAccount(database, userID, initialAmount)
	if err == nil {
		t.Error("expected error when creating duplicate user, got nil")
	}
}

// TestGetAccountByUserID 测试按用户ID查询账户
func TestGetAccountByUserID(t *testing.T) {
	database := setupTestDB(t)
	defer CloseDB(database)

	userID := "test_user"
	initialAmount := int64(100000)

	accountID, err := CreateAccount(database, userID, initialAmount)
	if err != nil {
		t.Fatalf("CreateAccount failed: %v", err)
	}

	// 查询账户
	account, err := GetAccountByUserID(database, userID)
	if err != nil {
		t.Fatalf("GetAccountByUserID failed: %v", err)
	}

	if account.ID != accountID {
		t.Errorf("expected account ID %d, got %d", accountID, account.ID)
	}

	if account.UserID != userID {
		t.Errorf("expected user_id %s, got %s", userID, account.UserID)
	}
}

// TestGetAccountByUserIDNotFound 测试查询不存在的账户
func TestGetAccountByUserIDNotFound(t *testing.T) {
	database := setupTestDB(t)
	defer CloseDB(database)

	_, err := GetAccountByUserID(database, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent user, got nil")
	}
}

// TestUpdateAccountStatus 测试账户状态更新
func TestUpdateAccountStatus(t *testing.T) {
	database := setupTestDB(t)
	defer CloseDB(database)

	userID := "test_user"
	accountID, err := CreateAccount(database, userID, 100000)
	if err != nil {
		t.Fatalf("CreateAccount failed: %v", err)
	}

	// 冻结账户
	err = UpdateAccountStatus(database, accountID, model.AccountStatusFrozen)
	if err != nil {
		t.Fatalf("UpdateAccountStatus failed: %v", err)
	}

	account, err := GetAccountByID(database, accountID)
	if err != nil {
		t.Fatalf("GetAccountByID failed: %v", err)
	}

	if account.Status != model.AccountStatusFrozen {
		t.Errorf("expected status %s, got %s", model.AccountStatusFrozen, account.Status)
	}
}
