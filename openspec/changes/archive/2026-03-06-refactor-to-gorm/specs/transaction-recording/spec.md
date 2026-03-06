# 规格：transaction-recording

> 📚 **知识上下文**：基于 `add-sqlite-local-db` 变更
>
> **原有实现**：使用原生 SQL 和 `*sql.DB` 进行交易记录操作

## 修改需求

### [R-1] 创建交易记录（MODIFIED）

**原实现**：`CreateTransaction(tx *sql.Tx, transaction *model.Transaction)`

**新实现**：`CreateTransaction(tx *gorm.DB, transaction *model.Transaction)`

#### 场景：在事务中创建交易记录
- **当** 调用 `CreateTransaction(tx, trans)`
- **那么** 使用 `tx.Create(trans)` 插入记录
- **并且** 返回新创建的记录 ID
- **并且** 自动设置时间戳字段

---

### [R-2] 查询交易记录（MODIFIED）

**原实现**：使用 `db.Query()` 和手动扫描

**新实现**：使用 GORM 查询 API

#### 场景：根据账户查询
- **当** 调用 `GetTransactionsByAccount(db, accountID)`
- **那么** 使用 `db.Where("account_id = ?", accountID).Find(&transactions)`
- **并且** 返回 `[]model.Transaction`

#### 场景：根据股票代码查询
- **当** 调用 `GetTransactionsByStock(db, stockCode)`
- **那么** 使用 `db.Where("stock_code = ?", stockCode).Find(&transactions)`

#### 场景：根据状态查询
- **当** 调用 `GetTransactionsByStatus(db, status)`
- **那么** 使用 `db.Where("status = ?", status).Find(&transactions)`

#### 场景：根据 ID 查询单条
- **当** 调用 `GetTransactionByID(db, id)`
- **那么** 使用 `db.First(&trans, id)`

---

### [R-3] 更新交易状态（MODIFIED）

**原实现**：`UpdateTransactionStatus(db *sql.DB, id int64, status string)`

**新实现**：`UpdateTransactionStatus(db *gorm.DB, id int64, status string)`

#### 场景：更新为成交状态
- **当** 调用 `UpdateTransactionStatus(db, id, FILLED)`
- **那么** 使用 `db.Model(&Transaction{}).Where("id = ?", id).Update("status", "FILLED")`
- **并且** 自动更新 `updated_at`

#### 场景：更新为撤销状态
- **当** 调用 `UpdateTransactionStatus(db, id, CANCELLED)`
- **那么** 更新状态为 CANCELLED
- **并且** 记录状态变更时间

---

### [R-4] 部分成交处理（MODIFIED）

**原实现**：多条 SQL 语句更新

**新实现**：使用 GORM 在同一事务中处理

#### 场景：部分成交
- **当** 调用 `PartialFill(db, transID, filledQty)`
- **那么** 在事务中：
  1. 将原记录状态更新为 OBSOLETE
  2. 创建已成交部分的 FILLED 记录
  3. 创建剩余部分的 PENDING 记录
- **并且** 设置 `parent_id` 关联原记录

---

### [R-5] 关联查询（MODIFIED）

**原实现**：多次查询或 JOIN 语句

**新实现**：使用 GORM 的 Preload 或关联查询

#### 场景：查询交易及关联账户
- **当** 需要同时获取交易和账户信息
- **那么** 使用 `db.Preload("Account").Find(&transactions)`
- **或者** 使用 `db.Joins("Account").Find(&transactions)`

---

> 📚 **迁移说明**：
>
> **Breaking Changes**:
> - 所有函数参数类型从 `*sql.DB`/`*sql.Tx` 改为 `*gorm.DB`
>
> **Migration Path**:
> 1. 更新函数签名
> 2. 替换查询语句为 GORM 链式调用
> 3. 利用 GORM 的自动时间戳功能
