package finance

import (
	"testing"

	"msa/pkg/model"
)

// TestFormatHaoToYuan tests the formatHaoToYuan function
func TestFormatHaoToYuan(t *testing.T) {
	tests := []struct {
		name     string
		hao      int64
		expected string
	}{
		{"零", 0, "0.0000"},
		{"一毫", 1, "0.0001"},
		{"一厘", 10, "0.0010"},
		{"一分", 100, "0.0100"},
		{"一角", 1000, "0.1000"},
		{"一元", 10000, "1.0000"},
		{"百元", 1000000, "100.0000"},
		{"千元", 10000000, "1000.0000"},
		{"万元", 100000000, "10000.0000"},
		{"负数", -10000, "-1.0000"},
		{"小数", 1234, "0.1234"},
		{"大额", 12345678900, "1234567.8900"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatHaoToYuan(tt.hao)
			if result != tt.expected {
				t.Errorf("formatHaoToYuan(%d) = %s, want %s", tt.hao, result, tt.expected)
			}
		})
	}
}

// TestGetActiveAccount_NoDatabase tests getActiveAccount with nil database
func TestGetActiveAccount_NoDatabase(t *testing.T) {
	account, err := getActiveAccount(nil)
	if err == nil {
		t.Error("getActiveAccount() with nil DB should return error")
	}
	if account != nil {
		t.Error("getActiveAccount() with nil DB should return nil account")
	}
}

// TestGetActiveAccount_NoAccount tests getActiveAccount with empty database
func TestGetActiveAccount_NoAccount(t *testing.T) {
	// This test requires a test database setup
	// For now, we skip it as it needs DB initialization
	t.Skip("Skipping - requires test database setup")
}

// TestFetchCurrentPrice_InvalidCode tests fetchCurrentPrice with invalid code
func TestFetchCurrentPrice_InvalidCode(t *testing.T) {
	// Use an invalid stock code
	_, err := fetchCurrentPrice("")
	if err == nil {
		t.Error("fetchCurrentPrice() with empty code should return error")
	}
}

// TestFetchCurrentPrice_RealAPI_Note documents that real API testing requires mocking
func TestFetchCurrentPrice_RealAPI_Note(t *testing.T) {
	t.Skip("Skipping real API test - requires HTTP mock setup")

	// To properly test fetchCurrentPrice, we need to:
	// 1. Mock the FetchStockData function
	// 2. Set up test responses for various scenarios
	// 3. Test error handling and price extraction
}

// TestGetOrderStatusText tests the GetOrderStatusText function
func TestGetOrderStatusText(t *testing.T) {
	tests := []struct {
		name     string
		status   model.TransactionStatus
		expected string
	}{
		{"待成交", model.TransactionStatusPending, "待成交"},
		{"已成交", model.TransactionStatusFilled, "已成交"},
		{"已拒绝", model.TransactionStatusRejected, "已拒绝"},
		{"未知状态", model.TransactionStatus("unknown"), "未知状态"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getOrderStatusText(tt.status)
			if result != tt.expected {
				t.Errorf("getOrderStatusText(%s) = %s, want %s", tt.status, result, tt.expected)
			}
		})
	}
}

// TestGetTransactionTypeText tests the GetTransactionTypeText function
func TestGetTransactionTypeText(t *testing.T) {
	tests := []struct {
		name     string
		txType   model.TransactionType
		expected string
	}{
		{"买入", model.TransactionTypeBuy, "买入"},
		{"卖出", model.TransactionTypeSell, "卖出"},
		{"未知类型", model.TransactionType("unknown"), "未知"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTransactionTypeText(tt.txType)
			if result != tt.expected {
				t.Errorf("getTransactionTypeText(%s) = %s, want %s", tt.txType, result, tt.expected)
			}
		})
	}
}
