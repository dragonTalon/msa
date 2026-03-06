# 实现任务：refactor-to-gorm

## ⚠️ 风险警报

基于历史问题识别的风险：

### MEDIUM：事务处理不当导致数据不一致

- **风险描述**：GORM 事务与原生 SQL 事务使用方式不同，可能误用
- **缓解措施**：统一使用 `tx.Begin()` 和 `defer tx.Rollback()` 模式，确保所有数据库操作使用事务实例

### MEDIUM：N+1 查询问题

- **风险描述**：使用 GORM 的关联查询时可能产生 N+1 查询
- **缓解措施**：使用 `Preload` 预加载关联数据，避免循环中查询

### LOW：性能开销

- **风险描述**：ORM 会引入一定性能开销
- **缓解措施**：对于复杂查询可以使用原生 SQL 作为补充

---

## 1. 依赖和基础设置

- [x] 1.1 添加 GORM 依赖
  - 执行 `go get gorm.io/gorm`
  - 执行 `go get gorm.io/driver/sqlite`
  - 验证依赖添加成功

- [x] 1.2 备份现有数据库代码
  - 确认当前实现已提交到 git
  - 创建 feature 分支（如需要）

---

## 2. 模型定义更新

- [x] 2.1 更新 Account 模型
  - 在 `pkg/model/account.go` 中嵌入 `gorm.Model`
  - 保留现有字段和业务方法
  - 添加 GORM 标签：
    ```go
    type Account struct {
        gorm.Model
        UserID        string       `gorm:"type:TEXT;uniqueIndex;not null"`
        InitialAmount int64        `gorm:"type:INTEGER;not null"`
        AvailableAmt  int64        `gorm:"type:INTEGER;not null;default:0"`
        LockedAmt     int64        `gorm:"type:INTEGER;not null;default:0"`
        Status        AccountStatus `gorm:"type:TEXT;not null;default:'ACTIVE'"`
    }
    ```

- [x] 2.2 更新 Transaction 模型
  - 在 `pkg/model/transaction.go` 中嵌入 `gorm.Model`
  - 保留现有字段和常量定义
  - 添加 GORM 标签：
    ```go
    type Transaction struct {
        gorm.Model
        AccountID  int64              `gorm:"type:INTEGER;not null;index:idx_account_stock"`
        StockCode  string             `gorm:"type:TEXT;not null;index:idx_account_stock"`
        StockName  string             `gorm:"type:TEXT;not null"`
        Type       TransactionType    `gorm:"type:TEXT;not null"`
        Quantity   int64              `gorm:"type:INTEGER;not null"`
        Price      int64              `gorm:"type:INTEGER;not null"`
        Amount     int64              `gorm:"type:INTEGER;not null"`
        Fee        int64              `gorm:"type:INTEGER;not null;default:0"`
        Status     TransactionStatus  `gorm:"type:TEXT;not null;index"`
        Note       string             `gorm:"type:TEXT"`
        ParentID   *uint              `gorm:"type:INTEGER;index"`
    }
    ```

- [x] 2.3 验证模型定义
  - 确保所有标签正确
  - 确保索引定义符合需求
  - 确保所有标签正确
  - 确保索引定义符合需求

---

## 3. 数据库初始化重构

- [x] 3.1 重构 `InitDB()` 函数
  - 修改返回类型从 `*sql.DB` 到 `*gorm.DB`
  - 使用 `gorm.Open()` 替代 `sql.Open()`
  - 配置 SQLite driver

- [x] 3.2 配置连接池参数
  - 获取底层 `*sql.DB`：`db.DB()`
  - 设置 `SetMaxOpenConns(1)` 适配 SQLite
  - 设置 `SetMaxIdleConns(1)`
  - 设置 `SetConnMaxLifetime(time.Hour)`

- [x] 3.3 更新 `InitDBWithPath()` 函数
  - 修改返回类型从 `*sql.DB` 到 `*gorm.DB`
  - 支持自定义路径用于测试

- [x] 3.4 重构 `CloseDB()` 函数
  - 修改参数类型从 `*sql.DB` 到 `*gorm.DB`
  - 添加数据库优化：`db.Exec("PRAGMA optimize")`

---

## 4. 数据库迁移重构

- [x] 4.1 重构 `Migrate()` 函数
  - 修改参数类型从 `*sql.DB` 到 `*gorm.DB`
  - 使用 `db.AutoMigrate()` 替代手写 SQL
  - 传入模型：`&model.Account{}, &model.Transaction{}`

- [x] 4.2 移除手写 SQL
  - 删除 `CREATE TABLE` 语句
  - 删除 `CREATE INDEX` 语句（由 GORM 标签定义）

---

## 5. 账户操作重构

- [x] 5.1 重构 `CreateAccount()`
  - 修改参数类型从 `*sql.DB` 到 `*gorm.DB`
  - 使用 `db.Create(&account)` 替代 INSERT
  - 返回新创建记录的 ID

- [x] 5.2 重构 `GetAccountByUserID()`
  - 修改参数类型从 `*sql.DB` 到 `*gorm.DB`
  - 使用 `db.Where("user_id = ?", userID).First(&account)`
  - 处理 `gorm.ErrRecordNotFound`

- [x] 5.3 重构 `GetAccountByID()`
  - 修改参数类型从 `*sql.DB` 到 `*gorm.DB`
  - 使用 `db.First(&account, id)`

- [x] 5.4 重构 `UpdateAccountStatus()`
  - 修改参数类型从 `*sql.DB` 到 `*gorm.DB`
  - 使用 `db.Model(&Account{}).Where("id = ?", id).Update("status", status)`

- [x] 5.5 重构 `UpdateAccountAmounts()`
  - 修改参数类型从 `*sql.Tx` 到 `*gorm.DB`
  - 使用 `gorm.Expr()` 进行原子更新
  - 确保在事务中调用

---

## 6. 交易操作重构

- [x] 6.1 重构 `CreateTransaction()`
  - 修改参数类型从 `*sql.Tx` 到 `*gorm.DB`
  - 使用 `tx.Create(&trans)`

- [x] 6.2 重构 `GetTransactionByID()`
  - 修改参数类型从 `*sql.DB` 到 `*gorm.DB`
  - 使用 `db.First(&trans, id)`

- [x] 6.3 重构 `UpdateTransactionStatus()`
  - 修改参数类型从 `*sql.DB` 到 `*gorm.DB`
  - 使用 `db.Model(&Transaction{}).Where("id = ?", id).Update("status", status)`

- [x] 6.4 重构 `GetTransactionsByAccount()`
  - 修改参数类型从 `*sql.DB` 到 `*gorm.DB`
  - 使用 `db.Where("account_id = ?", accountID).Find(&transactions)`

- [x] 6.5 重构其他查询函数
  - `GetTransactionsByStock()`
  - `GetTransactionsByType()`
  - `GetTransactionsByStatus()`
  - 使用 GORM 链式查询

---

## 7. 服务层更新

- [x] 7.1 更新 `pkg/service/trade_service.go`
  - 修改所有函数参数类型从 `*sql.DB` 到 `*gorm.DB`
  - 更新事务处理代码（GORM 事务模式）
  - 使用 `gorm.Expr()` 进行金额更新

- [x] 7.2 更新 `pkg/service/position_service.go`
  - 修改所有函数参数类型从 `*sql.DB` 到 `*gorm.DB`
  - 使用 `db.Select().Scan()` 进行聚合查询

- [x] 7.3 更新 `pkg/service/account_service.go`
  - 修改所有函数参数类型从 `*sql.DB` 到 `*gorm.DB`

---

## 8. 测试更新

- [x] 8.1 更新 `pkg/db/db_test.go`
  - 修改测试函数使用 `*gorm.DB`
  - 确保所有测试通过

- [x] 8.2 更新 `pkg/service/position_service_test.go`
  - 修改测试使用 `*gorm.DB`
  - 修复测试超时问题

- [x] 8.3 更新 `pkg/service/trade_service_test.go`
  - 修改测试使用 `*gorm.DB`
  - 验证事务处理正确

- [x] 8.4 更新 `pkg/service/integration_test.go`
  - 修改测试使用 `*gorm.DB`
  - 验证并发问题已解决

---

## 9. 清理和文档

- [x] 9.1 移除原生 SQL 依赖（如不再需要）
  - 检查 `go.mod` 是否仍需要 `modernc.org/sqlite`
  - （GORM driver 可能仍依赖它）

- [x] 9.2 更新代码注释
  - 说明使用 GORM 进行数据库操作
  - 添加 GORM 特定注释（如钩子函数）

- [x] 9.3 更新 README（如需要）
  - 说明使用 GORM 进行数据库操作
  - 更新架构描述

---

## 10. 验证和部署

- [x] 10.1 运行所有测试
  - 执行 `go test ./pkg/...`
  - 确保所有测试通过
  - 验证测试超时问题已解决

- [x] 10.2 手动测试
  - 创建账户
  - 提交订单
  - 成交订单
  - 查询持仓
  - 验证数据一致性

- [x] 10.3 性能对比
  - 对比重构前后的性能
  - 确保性能可接受

---

## 基于知识的测试

### 测试超时回归测试

> 📚 **来源**：基于 `add-sqlite-local-db` 实现中的测试超时问题

**历史问题**：
`pkg/service/integration_test.go` 中的测试在 30 秒后超时，原因是 SQLite 连接配置不当。

**测试场景**：
1. 运行所有服务层测试
2. 验证测试在合理时间内完成（< 10 秒）
3. 验证无死锁或连接竞争

**预期行为**：
- 所有测试通过
- 测试执行时间显著减少
- 无超时错误

---

### 事务回滚回归测试

> 📚 **来源**：基于数据库操作最佳实践

**历史经验**：
未正确处理事务回滚导致数据不一致。

**测试场景**：
1. 开始买入订单事务
2. 创建交易记录
3. 更新账户锁定金额
4. 模拟错误发生
5. 验证事务回滚
6. 验证账户余额未变
7. 验证无交易记录创建

**预期行为**：
- 账户 available_amt 和 locked_amt 保持不变
- 数据库中无新增交易记录

---

### 并发安全回归测试

> 📚 **来源**：基于 SQLite 并发问题

**历史问题**：
SQLite 连接配置不当导致并发场景下的事务超时。

**测试场景**：
1. 并发提交多个订单
2. 验证所有订单正确处理
3. 验证无连接竞争

**预期行为**：
- 所有订单正确记录
- 无超时错误
- 数据一致性保证
