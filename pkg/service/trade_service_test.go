package service

import (
	"testing"

	"gorm.io/gorm"

	"msa/pkg/db"
	"msa/pkg/model"
)

func setupTestServiceDB(t *testing.T) (*gorm.DB, uint) {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.sqlite"

	database, err := db.InitDBWithPath(dbPath)
	if err != nil {
		t.Fatalf("InitDBWithPath failed: %v", err)
	}

	if err := db.Migrate(database); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	userID := "test_user"
	initialAmount := int64(10000000) // 1000.00元 - 足够用于测试

	accountID, err := db.CreateAccount(database, userID, initialAmount)
	if err != nil {
		t.Fatalf("CreateAccount failed: %v", err)
	}

	return database, accountID
}

// TestSubmitBuyOrder 测试买入订单提交
func TestSubmitBuyOrder(t *testing.T) {
	database, accountID := setupTestServiceDB(t)
	defer db.CloseDB(database)

	order := Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  10,
		Price:     180000, // 18.00元
		Fee:       5000,   // 0.50元
		Note:      "测试买入",
	}

	transID, err := SubmitBuyOrder(database, accountID, order)
	if err != nil {
		t.Fatalf("SubmitBuyOrder failed: %v", err)
	}

	if transID <= 0 {
		t.Errorf("expected valid transaction ID, got %d", transID)
	}

	// 验证交易记录
	trans, err := db.GetTransactionByID(database, transID)
	if err != nil {
		t.Fatalf("GetTransactionByID failed: %v", err)
	}

	if trans.Status != model.TransactionStatusPending {
		t.Errorf("expected status PENDING, got %s", trans.Status)
	}

	// 验证金额已锁定
	account, err := db.GetAccountByID(database, accountID)
	if err != nil {
		t.Fatalf("GetAccountByID failed: %v", err)
	}

	expectedLocked := int64(1805000) // 10 * 180000 + 5000
	if account.LockedAmt != expectedLocked {
		t.Errorf("expected locked_amt %d, got %d", expectedLocked, account.LockedAmt)
	}

	expectedAvailable := int64(10000000 - 1805000)
	if account.AvailableAmt != expectedAvailable {
		t.Errorf("expected available_amt %d, got %d", expectedAvailable, account.AvailableAmt)
	}
}

// TestSubmitBuyOrderInsufficientBalance 测试余额不足的买入订单
func TestSubmitBuyOrderInsufficientBalance(t *testing.T) {
	database, accountID := setupTestServiceDB(t)
	defer db.CloseDB(database)

	// 创建一个超过余额的订单
	order := Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  1000,
		Price:     180000, // 18.00元
		Fee:       5000,
	}

	transID, err := SubmitBuyOrder(database, accountID, order)
	if err != nil {
		t.Fatalf("SubmitBuyOrder should succeed with REJECTED status: %v", err)
	}

	// 验证交易被拒绝
	trans, err := db.GetTransactionByID(database, transID)
	if err != nil {
		t.Fatalf("GetTransactionByID failed: %v", err)
	}

	if trans.Status != model.TransactionStatusRejected {
		t.Errorf("expected status REJECTED, got %s", trans.Status)
	}

	if trans.Note != "余额不足" {
		t.Errorf("expected note '余额不足', got %s", trans.Note)
	}

	// 验证账户余额未变
	account, _ := db.GetAccountByID(database, accountID)
	if account.AvailableAmt != 10000000 {
		t.Errorf("available_amt should remain 10000000, got %d", account.AvailableAmt)
	}

	if account.LockedAmt != 0 {
		t.Errorf("locked_amt should remain 0, got %d", account.LockedAmt)
	}
}

// TestSubmitSellOrder 测试卖出订单提交
func TestSubmitSellOrder(t *testing.T) {
	database, accountID := setupTestServiceDB(t)
	defer db.CloseDB(database)

	order := Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  10,
		Price:     180000, // 18.00元
		Fee:       5000,   // 0.50元
	}

	transID, err := SubmitSellOrder(database, accountID, order)
	if err != nil {
		t.Fatalf("SubmitSellOrder failed: %v", err)
	}

	// 验证交易记录
	trans, err := db.GetTransactionByID(database, transID)
	if err != nil {
		t.Fatalf("GetTransactionByID failed: %v", err)
	}

	if trans.Type != model.TransactionTypeSell {
		t.Errorf("expected type SELL, got %s", trans.Type)
	}

	if trans.Status != model.TransactionStatusPending {
		t.Errorf("expected status PENDING, got %s", trans.Status)
	}

	// 卖出不锁定金额
	account, _ := db.GetAccountByID(database, accountID)
	if account.LockedAmt != 0 {
		t.Errorf("locked_amt should remain 0 for sell order, got %d", account.LockedAmt)
	}
}

// TestFillOrder_Buy 测试买入订单成交
func TestFillOrder_Buy(t *testing.T) {
	database, accountID := setupTestServiceDB(t)
	defer db.CloseDB(database)

	// 先提交买入订单
	order := Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  10,
		Price:     180000, // 18.00元
		Fee:       5000,   // 0.50元
	}

	transID, _ := SubmitBuyOrder(database, accountID, order)

	// 验证锁定前状态
	account, _ := db.GetAccountByID(database, accountID)
	lockedBefore := account.LockedAmt

	// 订单成交
	err := FillOrder(database, transID)
	if err != nil {
		t.Fatalf("FillOrder failed: %v", err)
	}

	// 验证交易状态
	trans, _ := db.GetTransactionByID(database, transID)
	if trans.Status != model.TransactionStatusFilled {
		t.Errorf("expected status FILLED, got %s", trans.Status)
	}

	// 验证锁定金额已减少
	account, _ = db.GetAccountByID(database, accountID)
	if account.LockedAmt != lockedBefore-1805000 {
		t.Errorf("locked_amt should decrease by 1805000, got %d", account.LockedAmt)
	}

	// available_amt 不变
	if account.AvailableAmt != 10000000-1805000 {
		t.Errorf("available_amt should not change on fill, got %d", account.AvailableAmt)
	}
}

// TestFillOrder_Sell 测试卖出订单成交
func TestFillOrder_Sell(t *testing.T) {
	database, accountID := setupTestServiceDB(t)
	defer db.CloseDB(database)

	// 先提交卖出订单
	order := Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  10,
		Price:     180000, // 18.00元
		Fee:       5000,   // 0.50元
	}

	transID, _ := SubmitSellOrder(database, accountID, order)

	// 记录成交前状态
	account, _ := db.GetAccountByID(database, accountID)
	availableBefore := account.AvailableAmt

	// 订单成交
	err := FillOrder(database, transID)
	if err != nil {
		t.Fatalf("FillOrder failed: %v", err)
	}

	// 验证可用金额增加
	account, _ = db.GetAccountByID(database, accountID)
	expectedIncrease := int64(1795000) // 10 * 180000 - 5000
	if account.AvailableAmt != availableBefore+expectedIncrease {
		t.Errorf("available_amt should increase by %d, got %d", expectedIncrease, account.AvailableAmt)
	}
}

// TestCancelOrder_Buy 测试买入订单撤销
func TestCancelOrder_Buy(t *testing.T) {
	database, accountID := setupTestServiceDB(t)
	defer db.CloseDB(database)

	// 先提交买入订单
	order := Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  10,
		Price:     180000, // 18.00元
		Fee:       5000,   // 0.50元
	}

	transID, _ := SubmitBuyOrder(database, accountID, order)

	// 记录撤销前状态
	account, _ := db.GetAccountByID(database, accountID)
	availableBefore := account.AvailableAmt

	// 撤销订单
	err := CancelOrder(database, transID)
	if err != nil {
		t.Fatalf("CancelOrder failed: %v", err)
	}

	// 验证交易状态
	trans, _ := db.GetTransactionByID(database, transID)
	if trans.Status != model.TransactionStatusCancelled {
		t.Errorf("expected status CANCELLED, got %s", trans.Status)
	}

	// 验证金额已返还
	account, _ = db.GetAccountByID(database, accountID)
	if account.LockedAmt != 0 {
		t.Errorf("locked_amt should be 0 after cancel, got %d", account.LockedAmt)
	}

	if account.AvailableAmt != availableBefore+1805000 {
		t.Errorf("available_amt should be restored, got %d", account.AvailableAmt)
	}
}

// TestCancelOrder_Sell 测试卖出订单撤销
func TestCancelOrder_Sell(t *testing.T) {
	database, accountID := setupTestServiceDB(t)
	defer db.CloseDB(database)

	// 先提交卖出订单
	order := Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  10,
		Price:     180000, // 18.00元
		Fee:       5000,   // 0.50元
	}

	transID, _ := SubmitSellOrder(database, accountID, order)

	// 记录撤销前状态
	account, _ := db.GetAccountByID(database, accountID)
	availableBefore := account.AvailableAmt

	// 撤销订单
	err := CancelOrder(database, transID)
	if err != nil {
		t.Fatalf("CancelOrder failed: %v", err)
	}

	// 卖出订单撤销不改变金额
	account, _ = db.GetAccountByID(database, accountID)
	if account.AvailableAmt != availableBefore {
		t.Errorf("available_amt should not change on sell cancel, got %d", account.AvailableAmt)
	}
}

// TestPartialFill 测试部分成交处理
func TestPartialFill(t *testing.T) {
	database, accountID := setupTestServiceDB(t)
	defer db.CloseDB(database)

	// 先提交买入订单
	order := Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  100,
		Price:     180000, // 18.00元
		Fee:       50000,  // 5.00元
	}

	transID, _ := SubmitBuyOrder(database, accountID, order)

	// 记录部分成交前状态
	account, _ := db.GetAccountByID(database, accountID)
	lockedBefore := account.LockedAmt

	// 部分成交 50 股
	err := PartialFill(database, transID, 50)
	if err != nil {
		t.Fatalf("PartialFill failed: %v", err)
	}

	// 验证原记录状态
	trans, _ := db.GetTransactionByID(database, transID)
	if trans.Status != model.TransactionStatusObsolete {
		t.Errorf("original transaction should be OBSOLETE, got %s", trans.Status)
	}

	// 验证子交易
	children, _ := db.GetTransactionsByParent(database, transID)
	if len(children) != 2 {
		t.Errorf("expected 2 child transactions, got %d", len(children))
	}

	// 验证锁定金额调整
	account, _ = db.GetAccountByID(database, accountID)
	// 原锁定：100 * 180000 + 50000 = 18050000
	// 剩余锁定：50 * 180000 + 25000 = 9025000
	// 差异：-9025000
	expectedLocked := lockedBefore - 9025000
	if account.LockedAmt != expectedLocked {
		t.Errorf("locked_amt should be %d after partial fill, got %d", expectedLocked, account.LockedAmt)
	}
}
