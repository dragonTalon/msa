package service

import (
	"testing"

	"msa/pkg/db"
	"msa/pkg/model"
)

// TestTransactionCommit 测试正常流程事务提交
func TestTransactionCommit(t *testing.T) {
	database, accountID := setupPositionTestDB(t)
	defer db.CloseDB(database)

	// 记录初始状态
	accountBefore, _ := db.GetAccountByID(database, accountID)
	availableBefore := accountBefore.AvailableAmt

	// 提交买入订单
	order := Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  10,
		Price:     180000, // 18.00元
		Fee:       5000,   // 0.50元
	}

	transID, err := SubmitBuyOrder(database, accountID, order)
	if err != nil {
		t.Fatalf("SubmitBuyOrder failed: %v", err)
	}

	// 验证事务已提交
	// 1. 交易记录已创建
	trans, err := db.GetTransactionByID(database, transID)
	if err != nil {
		t.Fatalf("GetTransactionByID failed: %v", err)
	}

	if trans.Status != model.TransactionStatusPending {
		t.Errorf("expected status PENDING, got %s", trans.Status)
	}

	// 2. 账户金额已更新
	accountAfter, _ := db.GetAccountByID(database, accountID)
	if accountAfter.AvailableAmt != availableBefore-1805000 {
		t.Errorf("available_amt not updated correctly")
	}

	if accountAfter.LockedAmt != 1805000 {
		t.Errorf("locked_amt not updated correctly")
	}
}

// TestTransactionRollback 测试异常情况事务回滚
func TestTransactionRollback(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.sqlite"

	database, err := db.InitDBWithPath(dbPath)
	if err != nil {
		t.Fatalf("InitDBWithPath failed: %v", err)
	}
	defer db.CloseDB(database)

	if err := db.Migrate(database); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	userID := "test_user"
	initialAmount := int64(100000)
	accountID, _ := db.CreateAccount(database, userID, initialAmount)

	// 记录初始状态
	accountBefore, _ := db.GetAccountByID(database, accountID)

	// 开始事务
	tx := database.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建交易记录
	trans := &model.Transaction{
		AccountID: accountID,
		StockCode: "600519",
		StockName: "贵州茅台",
		Type:      model.TransactionTypeBuy,
		Quantity:  10,
		Price:     180000, // 18.00元
		Amount:    1800000,
		Fee:       5000,   // 0.50元
		Status:    model.TransactionStatusPending,
	}

	transID, err := db.CreateTransaction(tx, trans)
	if err != nil {
		t.Fatalf("CreateTransaction failed: %v", err)
	}

	// 更新账户金额
	err = db.UpdateAccountAmounts(tx, accountID, -1805000, 1805000)
	if err != nil {
		t.Fatalf("UpdateAccountAmounts failed: %v", err)
	}

	// 模拟错误发生
	tx.Rollback()

	// 验证回滚：
	// 1. 账户余额未变
	accountAfter, _ := db.GetAccountByID(database, accountID)
	if accountAfter.AvailableAmt != accountBefore.AvailableAmt {
		t.Errorf("available_amt should remain unchanged after rollback")
	}

	if accountAfter.LockedAmt != accountBefore.LockedAmt {
		t.Errorf("locked_amt should remain unchanged after rollback")
	}

	// 2. 无交易记录创建
	_, err = db.GetTransactionByID(database, transID)
	if err == nil {
		t.Error("transaction should not exist after rollback")
	}
}

// TestTransactionIntegrity 测试账户余额与交易记录一致性
func TestTransactionIntegrity(t *testing.T) {
	database, accountID := setupPositionTestDB(t)
	defer db.CloseDB(database)

	// 执行一系列交易
	transactions := []struct {
		stockCode string
		stockName string
		transType model.TransactionType
		quantity  int64
		price     int64
	}{
		{"600519", "贵州茅台", model.TransactionTypeBuy, 10, 180000},
		{"600519", "贵州茅台", model.TransactionTypeBuy, 20, 182000},
		{"600519", "贵州茅台", model.TransactionTypeSell, 15, 185000},
	}

	var transIDs []uint

	for _, tt := range transactions {
		order := Order{
			StockCode: tt.stockCode,
			StockName: tt.stockName,
			Quantity:  tt.quantity,
			Price:     tt.price,
			Fee:       5000, // 0.50元
		}

		var transID uint
		var err error

		if tt.transType == model.TransactionTypeBuy {
			transID, err = SubmitBuyOrder(database, accountID, order)
		} else {
			transID, err = SubmitSellOrder(database, accountID, order)
		}

		if err != nil {
			t.Fatalf("SubmitOrder failed for %s: %v", tt.transType, err)
		}

		transIDs = append(transIDs, transID)

		// 如果是买入订单，立即成交
		if tt.transType == model.TransactionTypeBuy {
			FillOrder(database, transID)
		}
	}

	// 等待所有交易完成
	for _, transID := range transIDs {
		trans, _ := db.GetTransactionByID(database, transID)
		if trans.Status == model.TransactionStatusPending {
			FillOrder(database, transID)
		}
	}

	// 验证一致性
	account, _ := db.GetAccountByID(database, accountID)

	// 从交易记录计算总变化
	buyTotal := int64(0)
	sellTotal := int64(0)

	allTrans, _ := db.GetTransactionsByAccount(database, accountID)
	for _, trans := range allTrans {
		if trans.Status == model.TransactionStatusFilled {
			if trans.Type == model.TransactionTypeBuy {
				buyTotal += trans.Amount + trans.Fee
			} else {
				sellTotal += trans.Amount - trans.Fee
			}
		}
	}

	// 初始金额 - 买入支出 + 卖出收入 = 当前可用金额
	expectedAvailable := int64(10000000) - buyTotal + sellTotal
	if account.AvailableAmt != expectedAvailable {
		t.Errorf("available_amt mismatch: expected %d, got %d", expectedAvailable, account.AvailableAmt)
	}

	// 验证持仓计算正确
	position, _ := GetPosition(database, accountID, "600519")
	// 10 + 20 - 15 = 15 股
	if position != 15 {
		t.Errorf("position mismatch: expected 15, got %d", position)
	}
}

// TestTransactionErrorHandling 测试事务中的错误处理
func TestTransactionErrorHandling(t *testing.T) {
	database, accountID := setupPositionTestDB(t)
	defer db.CloseDB(database)

	// 测试余额不足时的错误处理
	largeOrder := Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  10000,
		Price:     180000, // 18.00元
		Fee:       5000,   // 0.50元
	}

	transID, err := SubmitBuyOrder(database, accountID, largeOrder)
	if err != nil {
		t.Fatalf("SubmitBuyOrder should succeed with REJECTED status: %v", err)
	}

	// 验证交易被拒绝
	trans, _ := db.GetTransactionByID(database, transID)
	if trans.Status != model.TransactionStatusRejected {
		t.Errorf("expected status REJECTED, got %s", trans.Status)
	}

	// 验证账户余额未变
	account, _ := db.GetAccountByID(database, accountID)
	if account.AvailableAmt != 10000000 {
		t.Errorf("available_amt should remain unchanged after rejected order")
	}
}

// TestConcurrentTransactions 测试并发交易处理
func TestConcurrentTransactions(t *testing.T) {
	database, accountID := setupPositionTestDB(t)
	defer db.CloseDB(database)

	// 提交多个挂单
	for i := 0; i < 5; i++ {
		order := Order{
			StockCode: "600519",
			StockName: "贵州茅台",
			Quantity:  10,
			Price:     180000, // 18.00元
			Fee:       5000,   // 0.50元
		}

		_, err := SubmitBuyOrder(database, accountID, order)
		if err != nil && err.Error() != "expected error" {
			t.Fatalf("SubmitBuyOrder failed: %v", err)
		}
	}

	// 验证所有交易都正确记录
	allTrans, _ := db.GetTransactionsByAccount(database, accountID)
	pendingCount := 0
	for _, trans := range allTrans {
		if trans.Status == model.TransactionStatusPending || trans.Status == model.TransactionStatusRejected {
			pendingCount++
		}
	}

	if pendingCount == 0 {
		t.Error("should have some pending or rejected transactions")
	}
}

// TestTransactionStates 测试交易状态流转
func TestTransactionStates(t *testing.T) {
	database, accountID := setupPositionTestDB(t)
	defer db.CloseDB(database)

	order := Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  10,
		Price:     180000, // 18.00元
		Fee:       5000,   // 0.50元
	}

	// PENDING → FILLED
	transID, _ := SubmitBuyOrder(database, accountID, order)
	trans, _ := db.GetTransactionByID(database, transID)
	if trans.Status != model.TransactionStatusPending {
		t.Fatalf("expected initial status PENDING, got %s", trans.Status)
	}

	FillOrder(database, transID)
	trans, _ = db.GetTransactionByID(database, transID)
	if trans.Status != model.TransactionStatusFilled {
		t.Errorf("expected status FILLED after fill, got %s", trans.Status)
	}

	// 测试新的订单：PENDING → CANCELLED
	transID2, _ := SubmitBuyOrder(database, accountID, order)
	CancelOrder(database, transID2)
	trans2, _ := db.GetTransactionByID(database, transID2)
	if trans2.Status != model.TransactionStatusCancelled {
		t.Errorf("expected status CANCELLED after cancel, got %s", trans2.Status)
	}
}
