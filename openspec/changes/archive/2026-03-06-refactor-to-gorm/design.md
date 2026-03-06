# 设计：refactor-to-gorm

## 上下文

当前 MSA 项目使用原生 SQL (`database/sql`) 操作 SQLite 数据库。在 `add-sqlite-local-db` 变更的实现过程中，发现了以下问题：

1. **测试超时**：服务层测试因 SQLite 连接配置问题导致超时
2. **并发问题**：WAL 模式配置不当导致事务处理超时
3. **代码冗长**：需要手写大量 SQL 语句和扫描代码
4. **维护成本高**：表结构变更需要修改多处 SQL 代码
5. **类型转换繁琐**：手动处理数据库字段到 Go 结构体的映射

> 📚 **相关历史问题**：
> - **测试超时**：`pkg/service/integration_test.go` 中的测试在 30 秒后超时
> - **连接配置**：原生 SQL 的连接池配置在并发场景下表现不佳

## 目标 / 非目标

**目标：**
- 使用 GORM 替代原生 SQL 操作，简化代码
- 解决测试超时和并发问题
- 提高代码可维护性和类型安全性
- 利用 GORM 的 AutoMigrate 功能自动处理表结构变更
- 减少约 40-50% 的数据库操作代码量

**非目标：**
- 修改现有业务逻辑
- 改变数据库 schema（表结构保持不变）
- 更改 API 接口（服务层接口保持兼容）

## 推荐模式

### GORM 事务模式

**描述**：
使用 GORM 的 `db.Begin()` 和 `defer tx.Rollback()` 模式确保事务安全。

**适用场景**：
- 所有涉及金额变更的操作
- 多步骤数据库操作
- 需要原子性保证的业务逻辑

**实现**：
```go
func SubmitBuyOrder(db *gorm.DB, accountID int64, order Order) (int64, error) {
    tx := db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    // 业务逻辑...

    if err != nil {
        tx.Rollback()
        return 0, err
    }

    if err := tx.Commit().Error; err != nil {
        return 0, err
    }

    return transID, nil
}
```

---

### GORM 错误处理模式

**描述**：
统一处理 GORM 返回的错误，包括 `gorm.ErrRecordNotFound`。

**适用场景**：
- 所有数据库查询操作
- 需要区分错误类型的场景

**实现**：
```go
func GetAccountByID(db *gorm.DB, id int64) (*model.Account, error) {
    var account model.Account
    result := db.First(&account, id)
    if result.Error != nil {
        if errors.Is(result.Error, gorm.ErrRecordNotFound) {
            return nil, ErrAccountNotFound
        }
        return nil, result.Error
    }
    return &account, nil
}
```

---

### GORM 钩子函数模式

**描述**：
使用 GORM 的钩子函数处理模型生命周期事件，避免在业务代码中散布验证逻辑。

**适用场景**：
- 创建/更新前的数据验证
- 创建/更新后的触发逻辑

**实现**：
```go
func (a *Account) BeforeCreate(tx *gorm.DB) error {
    // 验证 user_id 唯一性
    var count int64
    tx.Model(&Account{}).Where("user_id = ?", a.UserID).Count(&count)
    if count > 0 {
        return errors.New("user_id already exists")
    }
    return nil
}
```

---

### GORM 表达式模式

**描述**：
使用 `gorm.Expr()` 执行原子更新操作，避免并发问题。

**适用场景**：
- 金额增加/减少操作
- 需要原子更新的字段

**实现**：
```go
func UpdateAccountAmounts(tx *gorm.DB, accountID int64, deltaAvailable, deltaLocked int64) error {
    return tx.Model(&Account{}).Where("id = ?", accountID).Updates(map[string]interface{}{
        "available_amt": gorm.Expr("available_amt + ?", deltaAvailable),
        "locked_amt":    gorm.Expr("locked_amt + ?", deltaLocked),
    }).Error
}
```

---

## 避免的反模式

### 反模式 1：忽略 GORM 错误

**问题所在**：
GORM 的链式调用不返回错误，需要检查最终结果的 Error 字段。

**影响**：
静默失败，数据不一致。

**不要这样做**：
```go
// 不要这样做 - 没有检查错误
db.Create(&account)
db.Where("id = ?", id).First(&account)
```

**应该这样做**：
```go
// 应该这样做 - 始终检查错误
if err := db.Create(&account).Error; err != nil {
    return err
}

if err := db.Where("id = ?", id).First(&account).Error; err != nil {
    return err
}
```

---

### 反模式 2：N+1 查询问题

**问题所在**：
在循环中执行查询导致性能问题。

**影响**：
数据库压力增大，响应变慢。

**不要这样做**：
```go
// 不要这样做 - N+1 查询
var transactions []Transaction
db.Find(&transactions)
for _, trans := range transactions {
    db.Preload("Account").First(&trans, trans.ID)
}
```

**应该这样做**：
```go
// 应该这样做 - 使用 Preload
var transactions []Transaction
db.Preload("Account").Find(&transactions)
```

---

### 反模式 3：混合使用事务和原始 DB

**问题所在**：
在事务中混用 `tx` 和原始 `db`，导致事务隔离失效。

**影响**：
数据不一致，部分提交。

**不要这样做**：
```go
// 不要这样做 - 混用 tx 和 db
tx := db.Begin()
tx.Create(&transaction)
db.Model(&Account{}).Update(...) // 使用了 db 而不是 tx！
tx.Commit()
```

**应该这样做**：
```go
// 应该这样做 - 统一使用 tx
tx := db.Begin()
tx.Create(&transaction)
tx.Model(&Account{}).Update(...) // 使用 tx
tx.Commit()
```

---

### 反模式 4：过度使用自动迁移

**问题所在**：
在生产环境中启用 AutoMigrate 可能导致意外的 schema 变更。

**影响**：
数据丢失，生产事故。

**应该这样做**：
- 开发环境：使用 AutoMigrate
- 生产环境：使用显式的迁移脚本或版本控制

---

## 设计决策

### 决策 1：使用 GORM v2

**理由**：
- GORM v2 是当前主流版本，文档完善
- 支持更强的类型安全
- 更好的性能优化

**替代方案**：
- GORM v1：已过时，不再维护
- 其他 ORM（如 ent）：学习成本高，不符合项目需求

---

### 决策 2：保留 gorm.Model

**理由**：
- 自动管理 ID、CreatedAt、UpdatedAt、DeletedAt
- 支持软删除功能
- 与 GORM 生态兼容

**替代方案**：
- 自定义 ID：需要手动管理时间戳

---

### 决策 3：使用 gorm.Expr() 进行金额更新

**理由**：
- 原子操作，避免并发问题
- 数据库层面保证一致性

**替代方案**：
- 先读后写：存在竞态条件
- 数据库锁：降低并发性能

---

### 决策 4：配置 SQLite 连接池

**理由**：
- 解决测试超时问题
- 提高并发性能

**配置**：
```go
sqlDB, _ := db.DB()
sqlDB.SetMaxOpenConns(1)    // SQLite 推荐单连接
sqlDB.SetMaxIdleConns(1)
sqlDB.SetConnMaxLifetime(time.Hour)
```

---

## 风险 / 权衡

### 风险 1：学习曲线

**描述**：团队需要熟悉 GORM 的 API 和最佳实践。

**缓解措施**：
- 提供代码示例和文档
- 代码审查确保正确使用

---

### 风险 2：性能开销

**描述**：ORM 会引入一定的性能开销。

**缓解措施**：
- GORM 已经做了大量优化
- 对于复杂查询，可以使用原生 SQL 作为补充

---

### 风险 3：迁移复杂度

**描述**：需要重写所有数据库操作代码。

**缓解措施**：
- 渐进式迁移，先迁移核心功能
- 充分的测试覆盖

---

## 迁移计划

### 阶段 1：准备工作

1. 添加 GORM 依赖
2. 更新模型定义，添加 GORM 标签
3. 创建新的数据库初始化函数（与现有函数共存）

### 阶段 2：核心迁移

4. 迁移 `pkg/db/db.go` - 初始化函数
5. 迁移 `pkg/db/migrate.go` - 使用 AutoMigrate
6. 迁移 `pkg/db/account.go` - 账户操作
7. 迁移 `pkg/db/transaction.go` - 交易操作

### 阶段 3：服务层迁移

8. 更新 `pkg/service/trade_service.go`
9. 更新 `pkg/service/position_service.go`
10. 更新 `pkg/service/account_service.go`

### 阶段 4：测试和清理

11. 运行所有测试，确保功能正常
12. 移除原生 SQL 相关代码
13. 更新测试代码

### 回滚策略

- 保留原生 SQL 实现作为备份，直到 GORM 版本稳定
- 使用 git 分支管理，便于快速回滚
- 每个阶段完成后进行验证

---

## 目录结构

```
pkg/
├── db/
│   ├── db.go              # GORM 初始化
│   ├── migrate.go         # AutoMigrate
│   ├── account.go         # Account CRUD (GORM)
│   └── transaction.go     # Transaction CRUD (GORM)
├── model/
│   ├── account.go         # 添加 GORM 标签
│   └── transaction.go     # 添加 GORM 标签
└── service/
    ├── trade_service.go       # 使用 GORM 接口
    ├── position_service.go    # 使用 GORM 接口
    └── account_service.go     # 使用 GORM 接口
```
