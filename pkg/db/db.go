package db

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// getDBPath 返回数据库文件的完整路径
// 路径为 ~/.msa/msa.sqlite
// 如果目录不存在，会自动创建
func getDBPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	dbDir := filepath.Join(homeDir, ".msa")
	dbPath := filepath.Join(dbDir, "msa.sqlite")

	// 如果目录不存在，创建目录
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create directory %s: %w", dbDir, err)
		}
	}

	return dbPath, nil
}

// InitDB 初始化数据库连接
// 如果数据库文件不存在，会自动创建
// 返回数据库连接和可能的错误
func InitDB() (*gorm.DB, error) {
	dbPath, err := getDBPath()
	if err != nil {
		return nil, err
	}

	return openDB(dbPath)
}

// InitDBWithPath 使用指定路径初始化数据库连接（用于测试）
func InitDBWithPath(dbPath string) (*gorm.DB, error) {
	return openDB(dbPath)
}

// openDB 打开数据库连接的内部函数
func openDB(dbPath string) (*gorm.DB, error) {
	// 确保目录存在
	dbDir := filepath.Dir(dbPath)
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dbDir, err)
		}
	}

	// 配置 SQLite DSN 参数
	// ?mode=rwc: 读写模式，不存在则创建
	// &_journal_mode=WAL: 启用 WAL 模式
	// &_timeout=5000: 设置 5 秒超时
	// &_foreign_keys=1: 启用外键约束
	// &_cache=shared: 共享缓存模式
	dsn := dbPath + "?mode=rwc&_journal_mode=WAL&_timeout=5000&_foreign_keys=1&_cache=shared"

	// 打开数据库连接，如果不存在会自动创建
	// 使用纯 Go SQLite 驱动 glebarez/sqlite，不需要 CGO
	database, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 获取底层 sql.DB 以配置连接池
	sqlDB, err := database.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 配置连接池参数
	// SQLite 可以支持多个连接，特别是 WAL 模式下
	sqlDB.SetMaxOpenConns(5) // 允许多个连接
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(time.Minute * 5)

	// 设置额外的 PRAGMA 选项以优化性能和避免死锁
	pragmas := []string{
		"PRAGMA busy_timeout = 5000",  // 设置忙等待超时
		"PRAGMA synchronous = NORMAL", // 降低同步级别
		"PRAGMA cache_size = -64000",  // 64MB 缓存
	}

	for _, pragma := range pragmas {
		if err := database.Exec(pragma).Error; err != nil {
			CloseDB(database)
			return nil, fmt.Errorf("failed to set pragma %s: %w", pragma, err)
		}
	}

	// 验证连接
	if err := sqlDB.Ping(); err != nil {
		CloseDB(database)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return database, nil
}

// CloseDB 关闭数据库连接
func CloseDB(db *gorm.DB) error {
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
