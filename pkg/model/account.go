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
// 所有金额字段以毫为单位存储，显示时需要转换
type Account struct {
	gorm.Model
	UserID        string        `gorm:"type:TEXT;uniqueIndex;not null" db:"user_id"`
	InitialAmount int64         `gorm:"type:INTEGER;not null" db:"initial_amount"`          // 投入金额（毫）
	AvailableAmt  int64         `gorm:"type:INTEGER;not null;default:0" db:"available_amt"` // 可用金额（毫）
	LockedAmt     int64         `gorm:"type:INTEGER;not null;default:0" db:"locked_amt"`    // 锁定金额（毫）
	Status        AccountStatus `gorm:"type:TEXT;not null;default:'ACTIVE'" db:"status"`
}

// YuanToHao 将元转换为hao
func YuanToHao(yuan float64) int64 {
	return int64(yuan * 10000)
}

// HaoToYuan 将hao转换为元
func HaoToYuan(hao int64) float64 {
	return float64(hao) / 10000
}

// FormatAmount 格式化金额为元字符串（保留4位小数）
func FormatAmount(fen int64) string {
	yuan := HaoToYuan(fen)
	return fmt.Sprintf("%.4f", yuan)
}
