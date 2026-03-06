# 规格：position-calculation

> 📚 **知识上下文**：基于 `add-sqlite-local-db` 变更
>
> **原有实现**：使用原生 SQL 的 SUM 聚合函数计算持仓

## 修改需求

### [R-1] 持仓数量计算（MODIFIED）

**原实现**：`GetPosition(db *sql.DB, accountID int64, stockCode string)`

**新实现**：`GetPosition(db *gorm.DB, accountID int64, stockCode string)`

#### 场景：计算持仓数量
- **当** 调用 `GetPosition(db, accountID, stockCode)`
- **那么** 使用 GORM 的聚合查询：
  ```go
  db.Model(&Transaction{}).
      Select("COALESCE(SUM(CASE WHEN type = 'BUY' THEN quantity ELSE -quantity END), 0)").
      Where("account_id = ? AND stock_code = ? AND status = ?", accountID, stockCode, FILLED).
      Scan(&position)
  ```
- **并且** 返回持仓数量（可能为负）

#### 场景：无持仓股票
- **当** 查询从未交易过的股票
- **那么** 返回 0 而非错误

---

### [R-2] 持仓成本计算（MODIFIED）

**原实现**：使用 `SUM()` 计算买入和卖出金额

**新实现**：使用 GORM 的聚合 API

#### 场景：计算持仓成本
- **当** 调用 `GetPositionCost(db, accountID, stockCode)`
- **那么** 使用：
  ```go
  buyCost := db.Model(&Transaction{}).
      Select("COALESCE(SUM(amount + fee), 0)").
      Where("account_id = ? AND stock_code = ? AND type = ? AND status = ?", accountID, stockCode, BUY, FILLED).
      Scan(&cost)

  sellIncome := db.Model(&Transaction{}).
      Select("COALESCE(SUM(amount - fee), 0)").
      Where("account_id = ? AND stock_code = ? AND type = ? AND status = ?", accountID, stockCode, SELL, FILLED).
      Scan(&income)

  return buyCost - sellIncome, nil
  ```

---

### [R-3] 持仓市值计算（MODIFIED）

**原实现**：查询持仓后乘以价格

**新实现**：使用 GORM 查询后计算

#### 场景：计算持仓市值
- **当** 调用 `GetPositionValue(db, accountID, stockCode, currentPrice)`
- **那么** 先获取持仓数量
- **然后** 计算：`position * currentPrice`
- **并且** 返回市值（分为单位）

---

### [R-4] 账户总资产计算（MODIFIED）

**原实现**：多个 SQL 查询后汇总

**新实现**：使用 GORM 链式查询

#### 场景：计算账户总成本
- **当** 调用 `GetAccountTotalCost(db, accountID)`
- **那么** 使用 GORM 聚合查询
- **并且** 计算：可用金额 + 锁定金额 + 所有持仓成本

#### 场景：计算账户总市值
- **当** 调用 `GetAccountTotalValue(db, accountID, prices)`
- **那么** 对每只股票查询持仓并乘以当前价格
- **并且** 加上可用金额

#### 场景：计算账户盈亏
- **当** 调用 `GetAccountPnL(db, accountID, prices)`
- **那么** 计算：总市值 - 初始投入
- **并且** 返回盈亏金额和盈亏比例

---

### [R-5] 获取所有持仓（MODIFIED）

**原实现**：使用 GROUP BY 查询

**新实现**：使用 GORM 的 Group 和 Scan

#### 场景：获取所有持仓列表
- **当** 调用 `GetAllPositions(db, accountID, prices)`
- **那么** 使用：
  ```go
  db.Model(&Transaction{}).
      Select("stock_code, stock_name, COALESCE(SUM(CASE WHEN type = 'BUY' THEN quantity ELSE -quantity END), 0) as quantity").
      Where("account_id = ? AND status = ?", accountID, FILLED).
      Group("stock_code, stock_name").
      Having("quantity != 0").
      Scan(&positions)
  ```
- **并且** 过滤掉持仓为 0 的股票
- **并且** 计算每只股票的市值和盈亏

---

> 📚 **迁移说明**：
>
> **Breaking Changes**:
> - 所有函数参数类型从 `*sql.DB` 改为 `*gorm.DB`
> - 查询结果使用 GORM 的 `Scan()` 方法填充
>
> **Migration Path**:
> 1. 更新函数签名
> 2. 将 SQL 聚合查询转换为 GORM 的 `Select()` + `Scan()` 模式
> 3. 利用 GORM 的类型安全特性
