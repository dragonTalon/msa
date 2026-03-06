package db

import (
	"gorm.io/gorm"

	"msa/pkg/model"
)

// Migrate 使用 GORM AutoMigrate 创建数据库表和索引
// 自动根据模型结构体定义生成表结构
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Account{},
		&model.Transaction{},
	)
}
