package db

import (
	"sync"

	"gorm.io/gorm"
	log "github.com/sirupsen/logrus"
)

var (
	// globalDB 全局数据库连接
	globalDB *globalDBConn

	// once 确保 DB 只初始化一次
	once sync.Once

	// initErr 记录初始化错误
	initErr error
)

// globalDBConn 包装 *gorm.DB，提供类型安全
type globalDBConn struct {
	db *gorm.DB
}

// InitGlobalDB 初始化全局数据库连接（非阻塞）
// 如果初始化失败，仅记录警告，不阻塞程序启动
func InitGlobalDB() {
	once.Do(func() {
		db, err := InitDB()
		if err != nil {
			initErr = err
			log.Warnf("SQLite 初始化失败: %v", err)
			return
		}

		// 执行数据库迁移
		if err := Migrate(db); err != nil {
			initErr = err
			log.Warnf("数据库迁移失败: %v，关闭连接", err)
			CloseDB(db)
			return
		}

		globalDB = &globalDBConn{db: db}
		log.Info("SQLite 初始化成功")
	})
}

// GetDB 获取全局数据库连接
// 返回 nil 表示 DB 未初始化或初始化失败
func GetDB() *gorm.DB {
	if globalDB == nil {
		return nil
	}
	return globalDB.db
}

// IsDBAvailable 检查数据库是否可用
func IsDBAvailable() bool {
	return globalDB != nil
}

// CloseGlobalDB 关闭全局数据库连接
// 执行 PRAGMA optimize 优化数据库
func CloseGlobalDB() error {
	if globalDB == nil {
		return nil
	}

	db := globalDB.db
	if db != nil {
		// 优化数据库
		db.Exec("PRAGMA optimize;")

		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}

	return nil
}
