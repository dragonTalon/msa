# 规格：account-management

> 📚 **知识上下文**：基于 `add-sqlite-local-db` 变更
>
> **原有实现**：使用原生 SQL 和 `*sql.DB` 进行账户操作

## 修改需求

### [R-1] 创建账户（MODIFIED）

**原实现**：`CreateAccount(db *sql.DB, userID string, initialAmount int64)`

**新实现**：`CreateAccount(db *gorm.DB, userID string, initialAmount int64)`

#### 场景：创建新账户
- **当** 调用 `CreateAccount(db, userID, initialAmount)`
- **那么** 使用 `db.Create(&account)` 插入记录
- **并且** 自动设置 `created_at` 和 `updated_at` 字段
- **并且** 返回新创建的账户 ID

#### 场景：重复 user_id
- **当** 创建已存在的 user_id 账户
- **那么** 返回唯一约束错误
- **并且** 错误信息包含 user_id

---

### [R-2] 查询账户（MODIFIED）

**原实现**：使用 `db.QueryRow()` 和 `Scan()`

**新实现**：使用 GORM 的查询 API

#### 场景：根据 user_id 查询
- **当** 调用 `GetAccountByUserID(db, userID)`
- **那么** 使用 `db.Where("user_id = ?", userID).First(&account)`
- **并且** 返回 `*model.Account` 或 `gorm.ErrRecordNotFound`

#### 场景：根据 ID 查询
- **当** 调用 `GetAccountByID(db, id)`
- **那么** 使用 `db.First(&account, id)`
- **并且** 返回 `*model.Account` 或错误

---

### [R-3] 更新账户状态（MODIFIED）

**原实现**：使用 `db.Exec()` 执行 UPDATE 语句

**新实现**：使用 GORM 的更新 API

#### 场景：冻结账户
- **当** 调用 `UpdateAccountStatus(db, id, FROZEN)`
- **那么** 使用 `db.Model(&Account{}).Where("id = ?", id).Update("status", "FROZEN")`
- **并且** 自动更新 `updated_at` 字段

#### 场景：关闭账户
- **当** 调用 `UpdateAccountStatus(db, id, CLOSED)`
- **那么** 验证余额为 0
- **并且** 如果余额不为 0，返回错误

---

### [R-4] 更新账户金额（MODIFIED）

**原实现**：`UpdateAccountAmounts(tx *sql.Tx, accountID int64, deltaAvailable, deltaLocked int64)`

**新实现**：`UpdateAccountAmounts(tx *gorm.DB, accountID int64, deltaAvailable, deltaLocked int64)`

#### 场景：锁定金额
- **当** 提交买入订单时
- **那么** 减少可用金额，增加锁定金额
- **并且** 使用 `tx.Model(&Account{}).Where("id = ?", accountID).Updates(map[string]interface{}{...})`

#### 场景：解锁金额
- **当** 订单成交或撤销时
- **那么** 调整锁定金额和可用金额
- **并且** 确保更新是原子的

---

> 📚 **迁移说明**：
>
> **Breaking Changes**:
> - 所有函数的 `db` 参数类型从 `*sql.DB` 改为 `*gorm.DB`
> - 事务参数从 `*sql.Tx` 改为 `*gorm.DB`（GORM 使用相同的类型）
>
> **Migration Path**:
> 1. 更新函数签名
> 2. 替换原生 SQL 查询为 GORM API
> 3. 更新错误处理以使用 GORM 错误类型
