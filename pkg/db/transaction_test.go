package db

import (
	"testing"

	"gorm.io/gorm"

	"msa/pkg/model"
)

// setupTestDBWithAccount 创建测试数据库和账户
func setupTestDBWithAccount(t *testing.T) (*gorm.DB, uint) {
	database := setupTestDB(t)

	userID := "test_user"
	initialAmount := int64(100000)

	accountID, err := CreateAccount(database, userID, initialAmount)
	if err != nil {
		t.Fatalf("CreateAccount failed: %v", err)
	}

	return database, accountID
}

// TestCreateTransaction 测试交易记录创建
func TestCreateTransaction(t *testing.T) {
	database, accountID := setupTestDBWithAccount(t)
	defer CloseDB(database)

	// 开始事务
	tx := database.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	trans := &model.Transaction{
		AccountID: accountID,
		StockCode: "600519",
		StockName: "贵州茅台",
		Type:      model.TransactionTypeBuy,
		Quantity:  100,
		Price:     180000, // 18.00元
		Amount:    18000000,
		Fee:       5000,   // 0.50元
		Status:    model.TransactionStatusPending,
		Note:      "测试交易",
	}

	transID, err := CreateTransaction(tx, trans)
	if err != nil {
		t.Fatalf("CreateTransaction failed: %v", err)
	}

	if err := tx.Commit().Error; err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	if transID <= 0 {
		t.Errorf("expected valid transaction ID, got %d", transID)
	}
}

// TestGetTransactionByID 测试按ID查询交易
func TestGetTransactionByID(t *testing.T) {
	database, accountID := setupTestDBWithAccount(t)
	defer CloseDB(database)

	// 创建交易记录
	tx := database.Begin()
	trans := &model.Transaction{
		AccountID: accountID,
		StockCode: "600519",
		StockName: "贵州茅台",
		Type:      model.TransactionTypeBuy,
		Quantity:  100,
		Price:     180000,
		Amount:    18000000,
		Status:    model.TransactionStatusPending,
	}
	transID, _ := CreateTransaction(tx, trans)
	tx.Commit()

	// 查询交易
	fetchedTrans, err := GetTransactionByID(database, transID)
	if err != nil {
		t.Fatalf("GetTransactionByID failed: %v", err)
	}

	if fetchedTrans.ID != transID {
		t.Errorf("expected transaction ID %d, got %d", transID, fetchedTrans.ID)
	}

	if fetchedTrans.StockCode != "600519" {
		t.Errorf("expected stock_code 600519, got %s", fetchedTrans.StockCode)
	}

	if fetchedTrans.Type != model.TransactionTypeBuy {
		t.Errorf("expected type BUY, got %s", fetchedTrans.Type)
	}
}

// TestUpdateTransactionStatus 测试交易状态更新
func TestUpdateTransactionStatus(t *testing.T) {
	database, accountID := setupTestDBWithAccount(t)
	defer CloseDB(database)

	// 创建交易记录
	tx := database.Begin()
	trans := &model.Transaction{
		AccountID: accountID,
		StockCode: "600519",
		StockName: "贵州茅台",
		Type:      model.TransactionTypeBuy,
		Quantity:  100,
		Price:     180000,
		Amount:    18000000,
		Status:    model.TransactionStatusPending,
	}
	transID, _ := CreateTransaction(tx, trans)
	tx.Commit()

	// 更新状态为已成交
	err := UpdateTransactionStatus(database, transID, model.TransactionStatusFilled)
	if err != nil {
		t.Fatalf("UpdateTransactionStatus failed: %v", err)
	}

	// 验证状态已更新
	fetchedTrans, _ := GetTransactionByID(database, transID)
	if fetchedTrans.Status != model.TransactionStatusFilled {
		t.Errorf("expected status FILLED, got %s", fetchedTrans.Status)
	}
}

// TestGetTransactionsByAccount 测试按账户查询交易
func TestGetTransactionsByAccount(t *testing.T) {
	database, accountID := setupTestDBWithAccount(t)
	defer CloseDB(database)

	// 创建多笔交易
	for i := 0; i < 3; i++ {
		tx := database.Begin()
		trans := &model.Transaction{
			AccountID: accountID,
			StockCode: "600519",
			StockName: "贵州茅台",
			Type:      model.TransactionTypeBuy,
			Quantity:  int64(100 * (i + 1)),
			Price:     180000,
			Amount:    int64(18000000 * (i + 1)),
			Status:    model.TransactionStatusFilled,
		}
		CreateTransaction(tx, trans)
		tx.Commit()
	}

	// 查询交易
	transactions, err := GetTransactionsByAccount(database, accountID)
	if err != nil {
		t.Fatalf("GetTransactionsByAccount failed: %v", err)
	}

	if len(transactions) != 3 {
		t.Errorf("expected 3 transactions, got %d", len(transactions))
	}
}

// TestGetTransactionsByStock 测试按股票代码查询交易
func TestGetTransactionsByStock(t *testing.T) {
	database, accountID := setupTestDBWithAccount(t)
	defer CloseDB(database)

	// 创建交易
	tx := database.Begin()
	trans := &model.Transaction{
		AccountID: accountID,
		StockCode: "600519",
		StockName: "贵州茅台",
		Type:      model.TransactionTypeBuy,
		Quantity:  100,
		Price:     180000,
		Amount:    18000000,
		Status:    model.TransactionStatusFilled,
	}
	CreateTransaction(tx, trans)
	tx.Commit()

	// 按股票代码查询
	transactions, err := GetTransactionsByStock(database, "600519")
	if err != nil {
		t.Fatalf("GetTransactionsByStock failed: %v", err)
	}

	if len(transactions) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(transactions))
	}
}

// TestGetTransactionsByType 测试按交易类型查询
func TestGetTransactionsByType(t *testing.T) {
	database, accountID := setupTestDBWithAccount(t)
	defer CloseDB(database)

	// 创建买入交易
	tx := database.Begin()
	trans := &model.Transaction{
		AccountID: accountID,
		StockCode: "600519",
		StockName: "贵州茅台",
		Type:      model.TransactionTypeBuy,
		Quantity:  100,
		Price:     180000,
		Amount:    18000000,
		Status:    model.TransactionStatusFilled,
	}
	CreateTransaction(tx, trans)
	tx.Commit()

	// 按类型查询
	transactions, err := GetTransactionsByType(database, model.TransactionTypeBuy)
	if err != nil {
		t.Fatalf("GetTransactionsByType failed: %v", err)
	}

	if len(transactions) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(transactions))
	}

	if transactions[0].Type != model.TransactionTypeBuy {
		t.Errorf("expected type BUY, got %s", transactions[0].Type)
	}
}

// TestGetTransactionsByParent 测试查询部分成交的子交易
func TestGetTransactionsByParent(t *testing.T) {
	database, accountID := setupTestDBWithAccount(t)
	defer CloseDB(database)

	// 创建父交易
	tx := database.Begin()
	parentTrans := &model.Transaction{
		AccountID: accountID,
		StockCode: "600519",
		StockName: "贵州茅台",
		Type:      model.TransactionTypeBuy,
		Quantity:  100,
		Price:     180000,
		Amount:    18000000,
		Status:    model.TransactionStatusObsolete,
	}
	parentID, _ := CreateTransaction(tx, parentTrans)
	tx.Commit()

	// 创建子交易
	tx = database.Begin()
	childTrans := &model.Transaction{
		AccountID: accountID,
		ParentID:  &parentID,
		StockCode: "600519",
		StockName: "贵州茅台",
		Type:      model.TransactionTypeBuy,
		Quantity:  50,
		Price:     180000,
		Amount:    9000000,
		Status:    model.TransactionStatusFilled,
	}
	CreateTransaction(tx, childTrans)
	tx.Commit()

	// 按父ID查询
	childTransactions, err := GetTransactionsByParent(database, parentID)
	if err != nil {
		t.Fatalf("GetTransactionsByParent failed: %v", err)
	}

	if len(childTransactions) != 1 {
		t.Errorf("expected 1 child transaction, got %d", len(childTransactions))
	}

	if childTransactions[0].ParentID == nil || *childTransactions[0].ParentID != parentID {
		t.Errorf("child transaction should have parent_id %d", parentID)
	}
}
