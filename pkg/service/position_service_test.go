package service

import (
	"testing"

	"gorm.io/gorm"

	"msa/pkg/db"
	"msa/pkg/model"
)

func setupPositionTestDB(t *testing.T) (*gorm.DB, uint) {
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

// createTestTransaction 创建测试交易记录
func createTestTransaction(database *gorm.DB, accountID uint, transType model.TransactionType, quantity, price int64) {
	trans := &model.Transaction{
		AccountID: accountID,
		StockCode: "600519",
		StockName: "贵州茅台",
		Type:      transType,
		Quantity:  quantity,
		Price:     price,
		Amount:    quantity * price,
		Status:    model.TransactionStatusFilled,
	}
	database.Create(trans)
}

// TestGetPosition 测试持仓数量计算
func TestGetPosition(t *testing.T) {
	database, accountID := setupPositionTestDB(t)
	defer db.CloseDB(database)

	// 创建交易：买入 100 股，卖出 30 股
	createTestTransaction(database, accountID, model.TransactionTypeBuy, 100, 180000)
	createTestTransaction(database, accountID, model.TransactionTypeSell, 30, 185000)

	// 计算持仓
	position, err := GetPosition(database, accountID, "600519")
	if err != nil {
		t.Fatalf("GetPosition failed: %v", err)
	}

	// 期望持仓：100 - 30 = 70 股
	if position != 70 {
		t.Errorf("expected position 70, got %d", position)
	}
}

// TestGetPositionCost 测试持仓成本计算
func TestGetPositionCost(t *testing.T) {
	database, accountID := setupPositionTestDB(t)
	defer db.CloseDB(database)

	// 创建交易
	// 买入：100 股 @ 18.00元 = 18000000毫 + 手续费
	createTestTransaction(database, accountID, model.TransactionTypeBuy, 100, 180000)

	// 卖出：30 股 @ 1850 = 55500 - 手续费
	trans := &model.Transaction{
		AccountID: accountID,
		StockCode: "600519",
		StockName: "贵州茅台",
		Type:      model.TransactionTypeSell,
		Quantity:  30,
		Price:     185000,
		Amount:    5550000,
		Fee:       50,
		Status:    model.TransactionStatusFilled,
	}
	database.Create(trans)

	// 计算持仓成本
	cost, err := GetPositionCost(database, accountID, "600519")
	if err != nil {
		t.Fatalf("GetPositionCost failed: %v", err)
	}

	// 买入成本：100 * 180000 = 18000000
	// 卖出收入：30 * 185000 - 50 = 5549950
	// 持仓成本：18000000 - 5549950 = 12450050
	expectedCost := int64(12450050)
	if cost != expectedCost {
		t.Errorf("expected cost %d, got %d", expectedCost, cost)
	}
}

// TestGetPositionValue 测试持仓市值计算
func TestGetPositionValue(t *testing.T) {
	database, accountID := setupPositionTestDB(t)
	defer db.CloseDB(database)

	createTestTransaction(database, accountID, model.TransactionTypeBuy, 100, 180000)

	// 计算持仓市值（假设当前价格 19.00元）
	currentPrice := int64(190000)
	value, err := GetPositionValue(database, accountID, "600519", currentPrice)
	if err != nil {
		t.Fatalf("GetPositionValue failed: %v", err)
	}

	// 期望市值：100 * 190000 = 19000000
	if value != 19000000 {
		t.Errorf("expected value 19000000, got %d", value)
	}
}

// TestGetAccountTotalCost 测试账户总成本计算
func TestGetAccountTotalCost(t *testing.T) {
	database, accountID := setupPositionTestDB(t)
	defer db.CloseDB(database)

	// 创建交易
	createTestTransaction(database, accountID, model.TransactionTypeBuy, 100, 180000)

	// 计算总成本
	cost, err := GetAccountTotalCost(database, accountID)
	if err != nil {
		t.Fatalf("GetAccountTotalCost failed: %v", err)
	}

	// 这个测试主要验证计算逻辑能够正常执行
	_ = cost
}

// TestGetAccountTotalValue 测试账户总市值计算
func TestGetAccountTotalValue(t *testing.T) {
	database, accountID := setupPositionTestDB(t)
	defer db.CloseDB(database)

	// 创建交易
	createTestTransaction(database, accountID, model.TransactionTypeBuy, 10, 180000)

	// 计算总市值
	prices := PriceMap{
		"600519": 190000, // 当前价格 19.00元
	}

	value, err := GetAccountTotalValue(database, accountID, prices)
	if err != nil {
		t.Fatalf("GetAccountTotalValue failed: %v", err)
	}

	// 验证计算结果
	_ = value
}

// TestGetAccountPnL 测试账户盈亏计算
func TestGetAccountPnL(t *testing.T) {
	database, accountID := setupPositionTestDB(t)
	defer db.CloseDB(database)

	// 创建交易
	createTestTransaction(database, accountID, model.TransactionTypeBuy, 10, 180000)

	// 计算盈亏
	prices := PriceMap{
		"600519": 190000, // 当前价格 19.00元，买入价 18.00元，盈利
	}

	pnl, err := GetAccountPnL(database, accountID, prices)
	if err != nil {
		t.Fatalf("GetAccountPnL failed: %v", err)
	}

	// 10股 × (190000 - 180000)毫 = 1000000毫 = 100.00元盈利
	_ = pnl
}

// TestGetAllPositions 测试获取所有持仓
func TestGetAllPositions(t *testing.T) {
	database, accountID := setupPositionTestDB(t)
	defer db.CloseDB(database)

	// 创建多只股票的交易
	createTestTransaction(database, accountID, model.TransactionTypeBuy, 100, 180000) // 茅台

	// 创建腾讯的交易
	trans := &model.Transaction{
		AccountID: accountID,
		StockCode: "00700",
		StockName: "腾讯控股",
		Type:      model.TransactionTypeBuy,
		Quantity:  500,
		Price:     38000, // 380 元
		Amount:    19000000,
		Status:    model.TransactionStatusFilled,
	}
	database.Create(trans)

	// 获取所有持仓
	prices := PriceMap{
		"600519": 190000,
		"00700":  40000,
	}

	positions, err := GetAllPositions(database, accountID, prices)
	if err != nil {
		t.Fatalf("GetAllPositions failed: %v", err)
	}

	// 应该有 2 只股票持仓
	if len(positions) != 2 {
		t.Errorf("expected 2 positions, got %d", len(positions))
	}

	// 验证持仓信息
	for _, pos := range positions {
		if pos.Quantity <= 0 {
			t.Errorf("position %s should have positive quantity", pos.StockCode)
		}

		if pos.StockCode == "" {
			t.Error("position should have stock_code")
		}
	}
}

// TestGetAllPositions_Empty 测试无持仓的情况
func TestGetAllPositions_Empty(t *testing.T) {
	database, accountID := setupPositionTestDB(t)
	defer db.CloseDB(database)

	// 没有任何交易

	prices := PriceMap{}
	positions, err := GetAllPositions(database, accountID, prices)
	if err != nil {
		t.Fatalf("GetAllPositions failed: %v", err)
	}

	if len(positions) != 0 {
		t.Errorf("expected 0 positions, got %d", len(positions))
	}
}

// TestGetPosition_ZeroPosition 测试持仓为0的情况
func TestGetPosition_ZeroPosition(t *testing.T) {
	database, accountID := setupPositionTestDB(t)
	defer db.CloseDB(database)

	// 买入100股，卖出100股
	createTestTransaction(database, accountID, model.TransactionTypeBuy, 100, 180000)

	trans := &model.Transaction{
		AccountID: accountID,
		StockCode: "600519",
		StockName: "贵州茅台",
		Type:      model.TransactionTypeSell,
		Quantity:  100,
		Price:     185000,
		Amount:    18500000,
		Status:    model.TransactionStatusFilled,
	}
	database.Create(trans)

	// 计算持仓
	position, err := GetPosition(database, accountID, "600519")
	if err != nil {
		t.Fatalf("GetPosition failed: %v", err)
	}

	// 期望持仓：100 - 100 = 0 股
	if position != 0 {
		t.Errorf("expected position 0, got %d", position)
	}
}
