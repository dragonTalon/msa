# 设计：add-sqlite-local-db

## 上下文

当前项目是一个 Go 语言开发的股票管理工具，通过 API 获取股票数据。项目缺乏本地持久化能力，无法保存账户信息、交易记录和持仓数据。

**当前状态**：
- 所有数据通过 API 实时获取
- 无本地数据存储
- 无交易历史记录
- 无持仓管理功能

**约束**：
- 单用户本地应用，SQLite 完全够用
- 需要事务原子性保证
- 需要支持金额锁定机制
- 持仓市值需要实时查询计算

**利益相关者**：
- 终端用户：需要查看账户余额、交易历史、持仓情况
- 系统：需要确保数据一致性和完整性

---

## 目标 / 非目标

**目标：**
- 添加 SQLite 本地数据库支持
- 实现账户管理（创建、查询、状态管理）
- 实现交易记录管理（创建、状态流转、部分成交）
- 实现金额锁定机制（挂单时锁定资金）
- 实现持仓计算（数量、成本、市值）

**非目标：**
- 不支持多用户并发交易
- 不支持分布式事务
- 不实现复杂的风控规则
- 不实现自动交易功能

---

## 推荐模式

知识库中暂无 SQLite 相关的推荐模式。本变更将建立数据库操作的基础模式。

### 事务操作模式

**描述**：所有涉及金额变更的操作必须在数据库事务中执行，确保原子性。

**适用场景**：
- 账户创建
- 买入订单提交
- 卖出订单提交
- 订单成交/撤销

**实现模式**：
```go
func (s *Service) SubmitBuyOrder(tx *sql.Tx, accountID int64, order BuyOrder) error {
    // 1. 开始事务（由调用者传入或在此创建）
    // 2. 检查余额
    // 3. 创建交易记录
    // 4. 更新账户金额
    // 5. 提交事务
}
```

---

## 避免的反模式

知识库中暂无相关反模式。以下是基于最佳实践应避免的做法：

### 反模式1：金额字段使用浮点数

**问题所在**：浮点数存在精度问题，可能导致金额计算错误。

**影响**：累积误差可能导致账户余额不准确。

**应该这样做**：
```go
// 使用 DECIMAL 类型存储金额
// SQLite 中使用 TEXT 存储精确小数，或使用 INTEGER 存储分
type Amount struct {
    Value int64  // 存储分为单位
}
```

### 反模式2：未使用事务

**问题所在**：账户更新和交易记录分开执行，可能导致数据不一致。

**影响**：账户余额与实际交易不匹配。

**应该这样做**：
```go
tx, err := db.Begin()
if err != nil {
    return err
}
defer tx.Rollback() // 如果未提交，自动回滚

// 所有操作在事务中执行
// ...
return tx.Commit()
```

### 反模式3：硬编码 SQL 语句

**问题所在**：SQL 字符串散落在代码中，难以维护。

**影响**：修改表结构需要多处修改，容易遗漏。

**应该这样做**：
```go
// 使用预定义的 SQL 常量
const (
    sqlCreateAccount = `INSERT INTO accounts (...) VALUES (...)`
    sqlUpdateAccount = `UPDATE accounts SET ... WHERE id = ?`
)
```

---

## 设计决策

### 1. SQLite 驱动选择

**决策**：使用 `modernc.org/sqlite`

**理由**：
- 纯 Go 实现，无 CGO 依赖
- 跨平台编译更容易
- 性能满足单用户场景需求

**备选方案**：`github.com/mattn/go-sqlite3`（需要 CGO，性能更好但编译复杂）

### 2. 金额存储方式

**决策**：使用 INTEGER 存储分为单位

**理由**：
- 避免浮点数精度问题
- SQLite 对整数支持更好
- 简单高效

**实现**：
- 存储时：金额 × 100
- 显示时：金额 ÷ 100

### 3. 数据库表结构

#### accounts 表
```sql
CREATE TABLE accounts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT UNIQUE NOT NULL,
    initial_amount INTEGER NOT NULL,  -- 以分为单位
    available_amt INTEGER NOT NULL,    -- 以分为单位
    locked_amt INTEGER NOT NULL,       -- 以分为单位
    status TEXT NOT NULL,              -- 'ACTIVE', 'FROZEN', 'CLOSED'
    created_at INTEGER NOT NULL,       -- Unix 时间戳
    updated_at INTEGER NOT NULL        -- Unix 时间戳
);
```

#### transactions 表
```sql
CREATE TABLE transactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id INTEGER NOT NULL,
    parent_id INTEGER,                 -- 部分成交时关联原记录
    stock_code TEXT NOT NULL,
    stock_name TEXT NOT NULL,
    type TEXT NOT NULL,                -- 'BUY', 'SELL'
    quantity INTEGER NOT NULL,
    price INTEGER NOT NULL,            -- 以分为单位
    amount INTEGER NOT NULL,           -- quantity × price
    fee INTEGER NOT NULL DEFAULT 0,    -- 以分为单位
    status TEXT NOT NULL,              -- 'PENDING', 'FILLED', 'CANCELLED', 'OBSOLETE', 'REJECTED'
    note TEXT,
    created_at INTEGER NOT NULL,       -- Unix 时间戳
    FOREIGN KEY (account_id) REFERENCES accounts(id),
    FOREIGN KEY (parent_id) REFERENCES transactions(id)
);
```

### 4. 数据库文件路径

**决策**：数据库文件存放在用户主目录下的 `.msa` 隐藏目录中

**文件路径**：
```
~/.msa/msa.sqlite
```

**实现代码**：
```go
import (
    "os"
    "path/filepath"
)

func getDBPath() (string, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }

    dbDir := filepath.Join(homeDir, ".msa")
    dbPath := filepath.Join(dbDir, "msa.sqlite")

    // 如果目录不存在，创建目录
    if _, err := os.Stat(dbDir); os.IsNotExist(err) {
        if err := os.MkdirAll(dbDir, 0755); err != nil {
            return "", err
        }
    }

    return dbPath, nil
}
```

**理由**：
- 遵循 Unix/Linux 配置文件存放惯例（隐藏目录）
- 跨平台兼容（macOS、Linux、Windows）
- 用户数据集中存放，便于管理
- 数据独立于应用目录，应用升级不影响数据

### 5. 索引设计

```sql
-- accounts 表
CREATE UNIQUE INDEX idx_accounts_user_id ON accounts(user_id);

-- transactions 表
CREATE INDEX idx_trans_account ON transactions(account_id);
CREATE INDEX idx_trans_stock ON transactions(stock_code);
CREATE INDEX idx_trans_type ON transactions(type);
CREATE INDEX idx_trans_status ON transactions(status);
CREATE INDEX idx_trans_created ON transactions(created_at);
CREATE INDEX idx_trans_parent ON transactions(parent_id);
CREATE INDEX idx_trans_account_status ON transactions(account_id, status);
```

### 6. 目录结构

```
pkg/
├── db/
│   ├── db.go              # 数据库初始化、连接管理
│   ├── migrate.go         # 数据库迁移
│   ├── account.go         # 账户相关数据库操作
│   └── transaction.go     # 交易相关数据库操作
├── model/
│   ├── account.go         # Account 数据模型
│   └── transaction.go     # Transaction 数据模型
└── service/
    ├── account_service.go # 账户业务逻辑
    └── trade_service.go   # 交易业务逻辑
```

---

## 风险 / 权衡

### 已知风险

1. **[SQLite 写锁]** → SQLite 使用数据库级写锁，并发写入性能受限
   - **缓解**：单用户场景，写入不频繁，影响可控

2. **[数据迁移]** → 未来如果需要升级数据库版本，需要处理迁移
   - **缓解**：设计迁移脚本，支持版本号管理

3. **[API 依赖]** → 持仓市值计算依赖实时价格 API
   - **缓解**：API 失败时使用缓存价格或提示用户

### 权衡

1. **精度 vs 性能**：使用整数存储金额（精度）vs 使用浮点数（性能）
   - **决策**：选择精度，金融场景准确性更重要

2. **实时计算 vs 缓存**：持仓市值实时查询 vs 定时缓存
   - **决策**：实时查询，确保数据最新

3. **单表 vs 多表**：所有交易存储在一张表 vs 分表存储
   - **决策**：单表存储，简化设计，满足当前需求

---

## 迁移计划

### 部署步骤

1. **依赖添加**
   ```bash
   go get modernc.org/sqlite
   ```

2. **数据库初始化**
   - 首次运行时自动创建数据库文件
   - 执行表创建 SQL
   - 创建索引

3. **现有代码适配**
   - 交易相关逻辑添加数据库操作
   - 保持向后兼容，数据库为可选功能

### 回滚策略

- 数据库文件独立存储，删除不影响原有功能
- 可通过配置禁用数据库功能
- 提供数据库导出功能，便于数据备份

### 数据备份

建议定期备份 SQLite 数据库文件：
```bash
cp ~/.msa/msa.sqlite ~/.msa/msa.sqlite.backup.$(date +%Y%m%d)
```
