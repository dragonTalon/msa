package finsvc

import (
	"testing"

	msadb "msa/pkg/db"
	"msa/pkg/model"
)

func TestGetPosition(t *testing.T) {
	db := setupTestDB(t)

	accountID, _ := msadb.CreateAccount(db, "test", model.YuanToHao(10000))
	account, _ := msadb.GetAccountByID(db, accountID)

	// 无持仓时
	qty, err := GetPosition(db, account.ID, "600519")
	if err != nil {
		t.Fatalf("GetPosition failed: %v", err)
	}
	if qty != 0 {
		t.Errorf("Expected 0, got %d", qty)
	}

	// 买入 100 股
	buyOrder := Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  100,
		Price:     model.YuanToHao(10),
		Fee:       model.YuanToHao(0.05),
	}
	txID, _ := SubmitBuyOrder(db, account.ID, buyOrder)
	FillOrder(db, txID)

	qty, err = GetPosition(db, account.ID, "600519")
	if err != nil {
		t.Fatalf("GetPosition failed: %v", err)
	}
	if qty != 100 {
		t.Errorf("Expected 100, got %d", qty)
	}

	// 卖出 30 股
	sellOrder := Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  30,
		Price:     model.YuanToHao(11),
		Fee:       model.YuanToHao(0.05),
	}
	sellTxID, _ := SubmitSellOrder(db, account.ID, sellOrder)
	FillOrder(db, sellTxID)

	qty, err = GetPosition(db, account.ID, "600519")
	if err != nil {
		t.Fatalf("GetPosition failed: %v", err)
	}
	if qty != 70 {
		t.Errorf("Expected 70, got %d", qty)
	}
}

func TestGetAllPositions(t *testing.T) {
	db := setupTestDB(t)

	accountID, _ := msadb.CreateAccount(db, "test", model.YuanToHao(10000))
	account, _ := msadb.GetAccountByID(db, accountID)

	// 无持仓
	positions, err := GetAllPositions(db, account.ID, nil)
	if err != nil {
		t.Fatalf("GetAllPositions failed: %v", err)
	}
	if len(positions) != 0 {
		t.Errorf("Expected 0 positions, got %d", len(positions))
	}

	// 买入两只股票
	order1 := Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  100,
		Price:     model.YuanToHao(10),
		Fee:       model.YuanToHao(0.05),
	}
	txID1, _ := SubmitBuyOrder(db, account.ID, order1)
	FillOrder(db, txID1)

	order2 := Order{
		StockCode: "000001",
		StockName: "平安银行",
		Quantity:  200,
		Price:     model.YuanToHao(15),
		Fee:       model.YuanToHao(0.05),
	}
	txID2, _ := SubmitBuyOrder(db, account.ID, order2)
	FillOrder(db, txID2)

	// 提供价格
	priceMap := PriceMap{
		"600519": model.YuanToHao(12),
		"000001": model.YuanToHao(16),
	}

	positions, err = GetAllPositions(db, account.ID, priceMap)
	if err != nil {
		t.Fatalf("GetAllPositions failed: %v", err)
	}
	if len(positions) != 2 {
		t.Errorf("Expected 2 positions, got %d", len(positions))
	}

	// 验证持仓内容
	for _, pos := range positions {
		if pos.StockCode == "600519" {
			if pos.Quantity != 100 {
				t.Errorf("Expected quantity 100, got %d", pos.Quantity)
			}
			if pos.CurrentPrice != model.YuanToHao(12) {
				t.Errorf("Expected price %d, got %d", model.YuanToHao(12), pos.CurrentPrice)
			}
		}
	}
}

func TestGetAccountTotalValue(t *testing.T) {
	db := setupTestDB(t)

	accountID, _ := msadb.CreateAccount(db, "test", model.YuanToHao(10000))
	account, _ := msadb.GetAccountByID(db, accountID)

	// 无持仓时
	value, err := GetAccountTotalValue(db, account.ID, nil)
	if err != nil {
		t.Fatalf("GetAccountTotalValue failed: %v", err)
	}
	// 只有可用余额
	if value != model.YuanToHao(10000) {
		t.Errorf("Expected %d, got %d", model.YuanToHao(10000), value)
	}

	// 买入股票
	order := Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  100,
		Price:     model.YuanToHao(10),
		Fee:       model.YuanToHao(0.05),
	}
	txID, _ := SubmitBuyOrder(db, account.ID, order)
	FillOrder(db, txID)

	// 提供价格上涨到 12 元
	priceMap := PriceMap{
		"600519": model.YuanToHao(12),
	}

	value, err = GetAccountTotalValue(db, account.ID, priceMap)
	if err != nil {
		t.Fatalf("GetAccountTotalValue failed: %v", err)
	}

	// 总价值 = 可用余额 + 持仓市值
	// 可用 = 10000 - 1000 - 0.05 = 8999.95
	// 持仓市值 = 100 * 12 = 1200
	// 总计 = 10199.95
	expectedAvailable := model.YuanToHao(10000) - model.YuanToHao(10)*100 - model.YuanToHao(0.05)
	expectedValue := expectedAvailable + model.YuanToHao(12)*100
	if value != expectedValue {
		t.Errorf("Expected %d, got %d", expectedValue, value)
	}
}

func TestGetAllPositions_ZeroQuantity(t *testing.T) {
	db := setupTestDB(t)

	accountID, _ := msadb.CreateAccount(db, "test", model.YuanToHao(10000))
	account, _ := msadb.GetAccountByID(db, accountID)

	// 买入后全部卖出
	order := Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  100,
		Price:     model.YuanToHao(10),
		Fee:       model.YuanToHao(0.05),
	}
	buyTxID, _ := SubmitBuyOrder(db, account.ID, order)
	FillOrder(db, buyTxID)

	sellOrder := Order{
		StockCode: "600519",
		StockName: "贵州茅台",
		Quantity:  100,
		Price:     model.YuanToHao(11),
		Fee:       model.YuanToHao(0.05),
	}
	sellTxID, _ := SubmitSellOrder(db, account.ID, sellOrder)
	FillOrder(db, sellTxID)

	// 持仓应为空
	positions, err := GetAllPositions(db, account.ID, nil)
	if err != nil {
		t.Fatalf("GetAllPositions failed: %v", err)
	}
	if len(positions) != 0 {
		t.Errorf("Expected 0 positions after sell all, got %d", len(positions))
	}
}
