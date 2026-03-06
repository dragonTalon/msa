package model

import (
	"fmt"

	"gorm.io/gorm"
)

// AccountStatus 账户状态类型
type AccountStatus string

const (
	// AccountStatusActive 正常状态
	AccountStatusActive AccountStatus = "ACTIVE"
	// AccountStatusFrozen 冻结状态
	AccountStatusFrozen AccountStatus = "FROZEN"
	// AccountStatusClosed 已关闭
	AccountStatusClosed AccountStatus = "CLOSED"
)

// Account 账户模型
// 所有金额字段以分为单位存储，显示时需要转换
type Account struct {
	gorm.Model
	UserID        string       `gorm:"type:TEXT;uniqueIndex;not null" db:"user_id"`
	InitialAmount int64        `gorm:"type:INTEGER;not null" db:"initial_amount"` // 投入金额（分）
	AvailableAmt  int64        `gorm:"type:INTEGER;not null;default:0" db:"available_amt"`  // 可用金额（分）
	LockedAmt     int64        `gorm:"type:INTEGER;not null;default:0" db:"locked_amt"`     // 锁定金额（分）
	Status        AccountStatus `gorm:"type:TEXT;not null;default:'ACTIVE'" db:"status"`
}

// YuanToFen 将元转换为分
func YuanToFen(yuan float64) int64 {
	return int64(yuan * 100)
}

// FenToYuan 将分转换为元
func FenToYuan(fen int64) float64 {
	return float64(fen) / 100
}

// FormatAmount 格式化金额为元字符串（保留2位小数）
func FormatAmount(fen int64) string {
	yuan := FenToYuan(fen)
	return fmt.Sprintf("%.2f", yuan)
}
