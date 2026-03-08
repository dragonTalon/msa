package model

import (
	"gorm.io/gorm"
)

// TransactionType 交易类型
type TransactionType string

const (
	// TransactionTypeBuy 买入
	TransactionTypeBuy TransactionType = "BUY"
	// TransactionTypeSell 卖出
	TransactionTypeSell TransactionType = "SELL"
)

// TransactionStatus 交易状态
type TransactionStatus string

const (
	// TransactionStatusPending 挂单中
	TransactionStatusPending TransactionStatus = "PENDING"
	// TransactionStatusFilled 已成交
	TransactionStatusFilled TransactionStatus = "FILLED"
	// TransactionStatusCancelled 已撤销
	TransactionStatusCancelled TransactionStatus = "CANCELLED"
	// TransactionStatusObsolete 已作废（部分成交后原记录）
	TransactionStatusObsolete TransactionStatus = "OBSOLETE"
	// TransactionStatusRejected 已拒绝
	TransactionStatusRejected TransactionStatus = "REJECTED"
)

// Transaction 交易记录模型
// 所有金额字段以毫为单位存储
type Transaction struct {
	gorm.Model
	AccountID uint              `gorm:"type:INTEGER;not null;index:idx_account_stock,priority:1;index" db:"account_id"`
	ParentID  *uint             `gorm:"type:INTEGER;index" db:"parent_id"` // 部分成交时关联原记录
	StockCode string            `gorm:"type:TEXT;not null;index:idx_account_stock,priority:2" db:"stock_code"`
	StockName string            `gorm:"type:TEXT;not null" db:"stock_name"`
	Type      TransactionType   `gorm:"type:TEXT;not null;index" db:"type"`
	Quantity  int64             `gorm:"type:INTEGER;not null" db:"quantity"`      // 交易数量
	Price     int64             `gorm:"type:INTEGER;not null" db:"price"`         // 交易价格（毫）
	Amount    int64             `gorm:"type:INTEGER;not null" db:"amount"`        // 交易金额（毫）= quantity × price
	Fee       int64             `gorm:"type:INTEGER;not null;default:0" db:"fee"` // 手续费（毫）
	Status    TransactionStatus `gorm:"type:TEXT;not null;index;default:'PENDING'" db:"status"`
	Note      string            `gorm:"type:TEXT" db:"note"` // 备注
}

// GetTotalAmount 获取总金额（交易金额 + 手续费）
func (t *Transaction) GetTotalAmount() int64 {
	return t.Amount + t.Fee
}
