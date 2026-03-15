package finance

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	msadb "msa/pkg/db"
	"msa/pkg/model"
	"msa/pkg/logic/finsvc"
)

// setupTestDB 创建测试用的临时数据库
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// 创建临时数据库文件
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// 自动迁移
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

// TestIntegration_CreateAccountFlow 测试创建账户流程
func TestIntegration_CreateAccountFlow(t *testing.T) {
	db := setupTestDB(t)

	// 测试创建账户
	accountID, err := msadb.CreateAccount(db, "test", model.YuanToHao(10000))
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// 验证账户已创建
	account, err := msadb.GetAccountByID(db, accountID)
	if err != nil {
		t.Fatalf("Failed to get account: %v", err)
	}

	if account.InitialAmount != model.YuanToHao(10000) {
		t.Errorf("Expected initial amount %d, got %d", model.YuanToHao(10000), account.InitialAmount)
	}

	if account.AvailableAmt != model.YuanToHao(10000) {
		t.Errorf("Expected available amount %d, got %d", model.YuanToHao(10000), account.AvailableAmt)
	}

	if account.Status != model.AccountStatusActive {
		t.Errorf("Expected status %s, got %s", model.AccountStatusActive, account.Status)
	}
}

// TestIntegration_BuyOrderFlow 测试买入流程
func TestIntegration_BuyOrderFlow(t *testing.T) {
	db := setupTestDB(t)

	// 创建账户
	accountID, err := msadb.CreateAccount(db, "test", model.YuanToHao(10000))
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	account, err := msadb.GetAccountByID(db, accountID)
	if err != nil {
		t.Fatalf("Failed to get account: %v", err)
	}

	// 提交买入订单
	order := finsvc.Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  100,
		Price:     model.YuanToHao(10.50),
		Fee:       model.YuanToHao(0.05),
	}
	txID, err := finsvc.SubmitBuyOrder(db, account.ID, order)
	if err != nil {
		t.Fatalf("Failed to submit buy order: %v", err)
	}

	// 验证交易已创建
	tx, err := msadb.GetTransactionByID(db, txID)
	if err != nil {
		t.Fatalf("Failed to get transaction: %v", err)
	}

	if tx.Status != model.TransactionStatusPending {
		t.Errorf("Expected status %s, got %s", model.TransactionStatusPending, tx.Status)
	}

	// 自动成交
	err = finsvc.FillOrder(db, txID)
	if err != nil {
		t.Fatalf("Failed to fill order: %v", err)
	}

	// 验证交易已成交
	tx, err = msadb.GetTransactionByID(db, txID)
	if err != nil {
		t.Fatalf("Failed to get transaction after fill: %v", err)
	}

	if tx.Status != model.TransactionStatusFilled {
		t.Errorf("Expected status %s after fill, got %s", model.TransactionStatusFilled, tx.Status)
	}

	// 验证账户余额已扣减
	account, err = msadb.GetAccountByID(db, accountID)
	if err != nil {
		t.Fatalf("Failed to get account after fill: %v", err)
	}

	expectedBalance := model.YuanToHao(10000) - model.YuanToHao(10.50)*100 - model.YuanToHao(0.05) // 手续费 0.05 元
	if account.AvailableAmt != expectedBalance {
		t.Errorf("Expected balance %d, got %d", expectedBalance, account.AvailableAmt)
	}

	// 验证锁定金额为 0
	if account.LockedAmt != 0 {
		t.Errorf("Expected locked amount 0, got %d", account.LockedAmt)
	}
}

// TestIntegration_SellOrderFlow 测试卖出流程
func TestIntegration_SellOrderFlow(t *testing.T) {
	db := setupTestDB(t)

	// 创建账户
	accountID, err := msadb.CreateAccount(db, "test", model.YuanToHao(10000))
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	account, err := msadb.GetAccountByID(db, accountID)
	if err != nil {
		t.Fatalf("Failed to get account: %v", err)
	}

	// 先买入建立持仓
	price := model.YuanToHao(10.50)
	quantity := int64(100)
	buyOrder := finsvc.Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  quantity,
		Price:     price,
		Fee:       model.YuanToHao(0.05),
	}
	txID1, err := finsvc.SubmitBuyOrder(db, account.ID, buyOrder)
	if err != nil {
		t.Fatalf("Failed to submit buy order: %v", err)
	}
	err = finsvc.FillOrder(db, txID1)
	if err != nil {
		t.Fatalf("Failed to fill buy order: %v", err)
	}

	// 获取持仓
	positions, err := finsvc.GetAllPositions(db, account.ID, nil)
	if err != nil {
		t.Fatalf("Failed to get positions: %v", err)
	}

	if len(positions) != 1 {
		t.Fatalf("Expected 1 position, got %d", len(positions))
	}

	if positions[0].Quantity != quantity {
		t.Errorf("Expected quantity %d, got %d", quantity, positions[0].Quantity)
	}

	// 提交卖出订单
	sellPrice := model.YuanToHao(11.00)
	sellOrder := finsvc.Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  quantity / 2,
		Price:     sellPrice,
		Fee:       model.YuanToHao(0.05),
	}
	txID2, err := finsvc.SubmitSellOrder(db, account.ID, sellOrder)
	if err != nil {
		t.Fatalf("Failed to submit sell order: %v", err)
	}

	// 自动成交
	err = finsvc.FillOrder(db, txID2)
	if err != nil {
		t.Fatalf("Failed to fill sell order: %v", err)
	}

	// 验证持仓已减少
	positions, err = finsvc.GetAllPositions(db, account.ID, nil)
	if err != nil {
		t.Fatalf("Failed to get positions after sell: %v", err)
	}

	if positions[0].Quantity != quantity/2 {
		t.Errorf("Expected quantity %d after sell, got %d", quantity/2, positions[0].Quantity)
	}

	// 验证账户余额增加
	account, err = msadb.GetAccountByID(db, accountID)
	if err != nil {
		t.Fatalf("Failed to get account after sell: %v", err)
	}

	// 余额 = 初始 - 买入成本 - 买入手续费 + 卖出收入 - 卖出手续费
	expectedBalance := model.YuanToHao(10000) - price*quantity - model.YuanToHao(0.05) + sellPrice*(quantity/2) - model.YuanToHao(0.05)
	if account.AvailableAmt != expectedBalance {
		t.Errorf("Expected balance %d, got %d", expectedBalance, account.AvailableAmt)
	}
}

// TestIntegration_InsufficientBalance 测试余额不足场景
func TestIntegration_InsufficientBalance(t *testing.T) {
	db := setupTestDB(t)

	// 创建账户，只有 100 元
	accountID, err := msadb.CreateAccount(db, "test", model.YuanToHao(100))
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	account, err := msadb.GetAccountByID(db, accountID)
	if err != nil {
		t.Fatalf("Failed to get account: %v", err)
	}

	// 尝试买入 1000 股，每股 10 元，需要 10000 元，余额不足
	order := finsvc.Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  1000,
		Price:     model.YuanToHao(10),
		Fee:       model.YuanToHao(0.05),
	}
	_, err = finsvc.SubmitBuyOrder(db, account.ID, order)
	// SubmitBuyOrder 在余额不足时不返回错误，而是创建 REJECTED 记录
	// 我们验证返回的是交易 ID
	if err != nil {
		t.Errorf("Expected no error (rejected order creates transaction), got %v", err)
	}

	// 验证订单被拒绝
	transactions, err := msadb.GetTransactionsByAccount(db, account.ID)
	if err != nil {
		t.Fatalf("Failed to get transactions: %v", err)
	}

	if len(transactions) != 1 {
		t.Fatalf("Expected 1 transaction, got %d", len(transactions))
	}

	if transactions[0].Status != model.TransactionStatusRejected {
		t.Errorf("Expected status %s, got %s", model.TransactionStatusRejected, transactions[0].Status)
	}

	// 验证余额未变动
	account, err = msadb.GetAccountByID(db, accountID)
	if err != nil {
		t.Fatalf("Failed to get account: %v", err)
	}

	if account.AvailableAmt != model.YuanToHao(100) {
		t.Errorf("Expected balance %d, got %d", model.YuanToHao(100), account.AvailableAmt)
	}
}

// TestIntegration_AccountTotalValue 测试账户总价值计算
func TestIntegration_AccountTotalValue(t *testing.T) {
	db := setupTestDB(t)

	// 创建账户
	accountID, err := msadb.CreateAccount(db, "test", model.YuanToHao(10000))
	if err != nil {
		t.Fatalf("Failed to create account: %v", err)
	}

	// 买入股票
	price := model.YuanToHao(10.00)
	quantity := int64(100)
	order := finsvc.Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  quantity,
		Price:     price,
		Fee:       model.YuanToHao(0.05),
	}
	txID, err := finsvc.SubmitBuyOrder(db, accountID, order)
	if err != nil {
		t.Fatalf("Failed to submit buy order: %v", err)
	}
	err = finsvc.FillOrder(db, txID)
	if err != nil {
		t.Fatalf("Failed to fill order: %v", err)
	}

	// 计算总价值
	priceMap := make(finsvc.PriceMap)
	priceMap["600519"] = model.YuanToHao(11.00) // 当前价格涨到 11 元

	totalValue, err := finsvc.GetAccountTotalValue(db, accountID, priceMap)
	if err != nil {
		t.Fatalf("Failed to get total value: %v", err)
	}

	// 总价值 = 可用余额 + 持仓市值
	// 可用余额 = 10000 - 1000 - 0.05 = 8999.95 元 = 89999500 毫
	// 持仓市值 = 100 * 11 = 1100 元 = 11000000 毫
	// 总价值 = 89999500 + 11000000 = 100999500 毫
	expectedValue := model.YuanToHao(10000) - price*quantity - model.YuanToHao(0.05) + model.YuanToHao(11.00)*quantity
	if totalValue != expectedValue {
		t.Errorf("Expected total value %d, got %d", expectedValue, totalValue)
	}
}
