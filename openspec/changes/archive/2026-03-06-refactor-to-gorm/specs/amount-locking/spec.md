# 规格：amount-locking

> 📚 **知识上下文**：基于 `add-sqlite-local-db` 变更
>
> **原有实现**：使用原生 SQL 事务手动管理金额锁定

## 修改需求

### [R-1] 买入订单锁定金额（MODIFIED）

**原实现**：`SubmitBuyOrder(db *sql.DB, accountID int64, order Order)`

**新实现**：`SubmitBuyOrder(db *gorm.DB, accountID int64, order Order)`

#### 场景：正常买入订单
- **当** 调用 `SubmitBuyOrder(db, accountID, order)`
- **那么** 在 GORM 事务中：
  1. 查询账户余额：`tx.First(&account, accountID)`
  2. 检查余额是否充足
  3. 创建交易记录：`tx.Create(&transaction)`
  4. 锁定金额：`tx.Model(&Account{}).Where("id = ?", accountID).Updates(...)`
  5. 提交事务：`tx.Commit()`
- **并且** 返回交易记录 ID

#### 场景：余额不足
- **当** 账户可用金额小于订单金额
- **那么** 创建状态为 REJECTED 的交易记录
- **并且** note 字段设置为 "余额不足"
- **并且** 不修改账户金额

---

### [R-2] 卖出订单验证持仓（MODIFIED）

**原实现**：`SubmitSellOrder(db *sql.DB, accountID int64, order Order)`

**新实现**：`SubmitSellOrder(db *gorm.DB, accountID int64, order Order)`

#### 场景：正常卖出订单
- **当** 调用 `SubmitSellOrder(db, accountID, order)`
- **那么** 验证持仓充足
- **并且** 创建 PENDING 状态的交易记录
- **并且** 卖出订单不锁定金额（只有买入锁定）

#### 场景：持仓不足
- **当** 持仓数量小于订单数量
- **那么** 创建 REJECTED 状态的交易记录
- **并且** note 设置为 "持仓不足"

---

### [R-3] 订单成交解锁（MODIFIED）

**原实现**：`FillOrder(db *sql.DB, transactionID int64)`

**新实现**：`FillOrder(db *gorm.DB, transactionID int64)`

#### 场景：买入订单成交
- **当** 买入订单成交
- **那么** 在事务中：
  1. 更新交易状态为 FILLED
  2. 减少锁定金额：`tx.Model(&Account{}).Where("id = ?", accountID).Update("locked_amt", gorm.Expr("locked_amt - ?", amount))`
  3. 提交事务

#### 场景：卖出订单成交
- **当** 卖出订单成交
- **那么** 在事务中：
  1. 更新交易状态为 FILLED
  2. 增加可用金额：`tx.Model(&Account{}).Where("id = ?", accountID).Update("available_amt", gorm.Expr("available_amt + ?", amount))`
  3. 提交事务

---

### [R-4] 订单撤销解锁（MODIFIED）

**原实现**：`CancelOrder(db *sql.DB, transactionID int64)`

**新实现**：`CancelOrder(db *gorm.DB, transactionID int64)`

#### 场景：撤销买入订单
- **当** 撤销买入订单
- **那么** 在事务中：
  1. 更新交易状态为 CANCELLED
  2. 锁定金额返还：`tx.Model(&Account{}).Where("id = ?", accountID).Updates(map[string]interface{}{
       "available_amt": gorm.Expr("available_amt + ?", lockedAmount),
       "locked_amt": gorm.Expr("locked_amt - ?", lockedAmount),
     })`
  3. 提交事务

#### 场景：撤销卖出订单
- **当** 撤销卖出订单
- **那么** 仅更新交易状态为 CANCELLED
- **并且** 不调整账户金额

---

### [R-5] 部分成交调整锁定（MODIFIED）

**原实现**：`PartialFill(db *sql.DB, transID int64, filledQty int64)`

**新实现**：`PartialFill(db *gorm.DB, transID int64, filledQty int64)`

#### 场景：部分成交
- **当** 订单部分成交
- **那么** 在事务中：
  1. 将原记录状态改为 OBSOLETE
  2. 创建已成交部分的 FILLED 记录
  3. 创建剩余部分的 PENDING 记录
  4. 调整锁定金额（对于买入订单）
  5. 设置 parent_id 关联
  6. 提交事务

#### 场景：计算新的锁定金额
- **当** 调整买入订单的锁定金额
- **那么** 使用 GORM 表达式：
  ```go
  tx.Model(&Account{}).Where("id = ?", accountID).Update("locked_amt",
      gorm.Expr("locked_amt - ?", unlockedAmount))
  ```

---

> 📚 **迁移说明**：
>
> **Breaking Changes**:
> - 所有函数参数类型从 `*sql.DB` 改为 `*gorm.DB`
> - 事务处理从 `*sql.Tx` 改为 GORM 事务（`*gorm.DB`）
>
> **Migration Path**:
> 1. 更新函数签名
> 2. 使用 GORM 事务 API 替代原生事务
> 3. 使用 `gorm.Expr()` 执行原子更新操作
> 4. 利用 GORM 的自动时间戳和错误处理
