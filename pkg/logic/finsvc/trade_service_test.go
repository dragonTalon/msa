package finsvc

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	msadb "msa/pkg/db"
	"msa/pkg/model"
)

// setupTestDB 创建测试用的临时数据库
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	if err := db.AutoMigrate(&model.Account{}, &model.Transaction{}); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
		os.Remove(dbPath)
	})

	return db
}

func TestSubmitBuyOrder_Success(t *testing.T) {
	db := setupTestDB(t)

	accountID, _ := msadb.CreateAccount(db, "test", model.YuanToHao(10000))
	account, _ := msadb.GetAccountByID(db, accountID)

	order := Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  100,
		Price:     model.YuanToHao(10.50),
		Fee:       model.YuanToHao(0.05),
	}

	txID, err := SubmitBuyOrder(db, account.ID, order)
	if err != nil {
		t.Fatalf("SubmitBuyOrder failed: %v", err)
	}

	// 验证交易记录
	tx, _ := msadb.GetTransactionByID(db, txID)
	if tx.Status != model.TransactionStatusPending {
		t.Errorf("Expected status %s, got %s", model.TransactionStatusPending, tx.Status)
	}

	// 验证账户余额锁定
	account, _ = msadb.GetAccountByID(db, accountID)
	expectedAvailable := model.YuanToHao(10000) - model.YuanToHao(10.50)*100 - model.YuanToHao(0.05)
	if account.AvailableAmt != expectedAvailable {
		t.Errorf("Expected available %d, got %d", expectedAvailable, account.AvailableAmt)
	}
	if account.LockedAmt != model.YuanToHao(10.50)*100+model.YuanToHao(0.05) {
		t.Errorf("Expected locked %d, got %d", model.YuanToHao(10.50)*100+model.YuanToHao(0.05), account.LockedAmt)
	}
}

func TestSubmitBuyOrder_InsufficientBalance(t *testing.T) {
	db := setupTestDB(t)

	accountID, _ := msadb.CreateAccount(db, "test", model.YuanToHao(100))
	account, _ := msadb.GetAccountByID(db, accountID)

	order := Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  1000,
		Price:     model.YuanToHao(10),
		Fee:       model.YuanToHao(0.05),
	}

	txID, err := SubmitBuyOrder(db, account.ID, order)
	if err != nil {
		t.Fatalf("SubmitBuyOrder should not error on insufficient balance: %v", err)
	}

	// 验证订单被拒绝
	tx, _ := msadb.GetTransactionByID(db, txID)
	if tx.Status != model.TransactionStatusRejected {
		t.Errorf("Expected status %s, got %s", model.TransactionStatusRejected, tx.Status)
	}

	// 验证余额未变动
	account, _ = msadb.GetAccountByID(db, accountID)
	if account.AvailableAmt != model.YuanToHao(100) {
		t.Errorf("Expected available %d, got %d", model.YuanToHao(100), account.AvailableAmt)
	}
}

func TestSubmitSellOrder_Success(t *testing.T) {
	db := setupTestDB(t)

	accountID, _ := msadb.CreateAccount(db, "test", model.YuanToHao(10000))
	account, _ := msadb.GetAccountByID(db, accountID)

	// 先买入建立持仓
	buyOrder := Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  100,
		Price:     model.YuanToHao(10),
		Fee:       model.YuanToHao(0.05),
	}
	txID, _ := SubmitBuyOrder(db, account.ID, buyOrder)
	FillOrder(db, txID)

	// 卖出
	sellOrder := Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  50,
		Price:     model.YuanToHao(11),
		Fee:       model.YuanToHao(0.05),
	}

	sellTxID, err := SubmitSellOrder(db, account.ID, sellOrder)
	if err != nil {
		t.Fatalf("SubmitSellOrder failed: %v", err)
	}

	tx, _ := msadb.GetTransactionByID(db, sellTxID)
	if tx.Status != model.TransactionStatusPending {
		t.Errorf("Expected status %s, got %s", model.TransactionStatusPending, tx.Status)
	}
}

func TestFillOrder_Buy(t *testing.T) {
	db := setupTestDB(t)

	accountID, _ := msadb.CreateAccount(db, "test", model.YuanToHao(10000))
	account, _ := msadb.GetAccountByID(db, accountID)

	order := Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  100,
		Price:     model.YuanToHao(10),
		Fee:       model.YuanToHao(0.05),
	}
	txID, _ := SubmitBuyOrder(db, account.ID, order)

	err := FillOrder(db, txID)
	if err != nil {
		t.Fatalf("FillOrder failed: %v", err)
	}

	// 验证交易状态
	tx, _ := msadb.GetTransactionByID(db, txID)
	if tx.Status != model.TransactionStatusFilled {
		t.Errorf("Expected status %s, got %s", model.TransactionStatusFilled, tx.Status)
	}

	// 验证锁定金额清零
	account, _ = msadb.GetAccountByID(db, accountID)
	if account.LockedAmt != 0 {
		t.Errorf("Expected locked 0, got %d", account.LockedAmt)
	}

	// 验证可用余额
	expectedAvailable := model.YuanToHao(10000) - model.YuanToHao(10)*100 - model.YuanToHao(0.05)
	if account.AvailableAmt != expectedAvailable {
		t.Errorf("Expected available %d, got %d", expectedAvailable, account.AvailableAmt)
	}
}

func TestFillOrder_Sell(t *testing.T) {
	db := setupTestDB(t)

	accountID, _ := msadb.CreateAccount(db, "test", model.YuanToHao(10000))
	account, _ := msadb.GetAccountByID(db, accountID)

	// 先买入
	buyOrder := Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  100,
		Price:     model.YuanToHao(10),
		Fee:       model.YuanToHao(0.05),
	}
	buyTxID, _ := SubmitBuyOrder(db, account.ID, buyOrder)
	FillOrder(db, buyTxID)

	// 卖出
	sellOrder := Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  50,
		Price:     model.YuanToHao(11),
		Fee:       model.YuanToHao(0.05),
	}
	sellTxID, _ := SubmitSellOrder(db, account.ID, sellOrder)

	err := FillOrder(db, sellTxID)
	if err != nil {
		t.Fatalf("FillOrder failed: %v", err)
	}

	// 验证可用余额增加
	account, _ = msadb.GetAccountByID(db, accountID)
	// 初始 10000 - 买入 1000.05 + 卖出收入 (550 - 0.05)
	expectedAvailable := model.YuanToHao(10000) - model.YuanToHao(10)*100 - model.YuanToHao(0.05) + model.YuanToHao(11)*50 - model.YuanToHao(0.05)
	if account.AvailableAmt != expectedAvailable {
		t.Errorf("Expected available %d, got %d", expectedAvailable, account.AvailableAmt)
	}
}
