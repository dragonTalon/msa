# 实现任务：add-sqlite-local-db

## ⚠️ 风险警报

基于数据库操作的最佳实践识别的风险：

### MEDIUM：数据库事务未正确处理

- **风险描述**：账户更新和交易记录未在同一事务中，可能导致数据不一致
- **缓解措施**：所有涉及金额变更的操作必须在事务中执行，使用 defer tx.Rollback() 确保异常时回滚

### MEDIUM：金额精度问题

- **风险描述**：使用浮点数存储金额可能导致累积误差
- **缓解措施**：使用 INTEGER 存储分为单位，显示时转换为元

---

## 1. 依赖和基础设置

- [x] 1.1 添加 SQLite 依赖
  - 执行 `go get modernc.org/sqlite`
  - 验证依赖添加成功

- [x] 1.2 创建数据库目录结构
  - 创建 `pkg/db/` 目录
  - 创建 `pkg/model/` 目录（如不存在）

---

## 2. 数据库初始化

- [x] 2.1 实现数据库文件路径获取函数
  - 创建 `pkg/db/db.go`
  - 实现 `getDBPath()` 函数，返回 `~/.msa/msa.sqlite`
  - 如果 `.msa` 目录不存在，自动创建

- [x] 2.2 实现数据库初始化函数
  - 实现 `InitDB()` 函数
  - 打开/创建数据库连接
  - 设置连接参数（如 WAL 模式）

- [x] 2.3 创建数据库迁移函数
  - 创建 `pkg/db/migrate.go`
  - 实现表创建 SQL（accounts、transactions）
  - 实现索引创建 SQL
  - 支持版本号管理（如需要）

---

## 3. 数据模型

- [x] 3.1 创建 Account 数据模型
  - 创建 `pkg/model/account.go`
  - 定义 `Account` 结构体
  - 定义 `AccountStatus` 常量（ACTIVE、FROZEN、CLOSED）
  - 实现金额转换方法（分 ↔ 元）

- [x] 3.2 创建 Transaction 数据模型
  - 创建 `pkg/model/transaction.go`
  - 定义 `Transaction` 结构体
  - 定义 `TransactionType` 常量（BUY、SELL）
  - 定义 `TransactionStatus` 常量（PENDING、FILLED、CANCELLED、OBSOLETE、REJECTED）

---

## 4. 账户管理

- [x] 4.1 实现账户创建
  - 创建 `pkg/db/account.go`
  - 实现 `CreateAccount(db, userID, initialAmount)` 函数
  - 验证 user_id 唯一性
  - 初始化 available_amt = initialAmount, locked_amt = 0

- [x] 4.2 实现账户查询
  - 实现 `GetAccountByUserID(db, userID)` 函数
  - 实现 `GetAccountByID(db, id)` 函数

- [x] 4.3 实现账户状态更新
  - 实现 `UpdateAccountStatus(db, id, status)` 函数
  - 验证关闭账户时余额为 0

- [x] 4.4 实现账户金额更新（内部函数）
  - 实现 `updateAccountAmounts(tx, accountID, deltaAvailable, deltaLocked)` 函数
  - 必须在事务中调用

---

## 5. 交易记录

- [x] 5.1 实现交易记录创建
  - 创建 `pkg/db/transaction.go`
  - 实现 `CreateTransaction(tx, transaction)` 函数
  - 在事务中执行

- [x] 5.2 实现交易状态更新
  - 实现 `UpdateTransactionStatus(db, id, status)` 函数
  - 实现 `UpdateTransactionsWithParent(tx, parentID, ...)` 用于部分成交

- [x] 5.3 实现交易查询
  - 实现 `GetTransactionsByAccount(db, accountID)` 函数
  - 实现 `GetTransactionsByStock(db, stockCode)` 函数
  - 实现 `GetTransactionsByType(db, transType)` 函数
  - 实现 `GetTransactionsByStatus(db, status)` 函数
  - 支持 parent_id 关联查询

---

## 6. 金额锁定机制

- [x] 6.1 实现买入订单提交
  - 创建 `pkg/service/trade_service.go`
  - 实现 `SubmitBuyOrder(db, accountID, order)` 函数
  - 在事务中：检查余额 → 创建交易记录 → 锁定金额
  - 余额不足时创建 REJECTED 记录

- [x] 6.2 实现卖出订单提交
  - 实现 `SubmitSellOrder(db, accountID, order)` 函数
  - 验证持仓充足
  - 不锁定金额

- [x] 6.3 实现订单成交处理
  - 实现 `FillOrder(db, transactionID)` 函数
  - 买入：减少 locked_amt
  - 卖出：增加 available_amt

- [x] 6.4 实现订单撤销处理
  - 实现 `CancelOrder(db, transactionID)` 函数
  - 买入：locked_amt 返还到 available_amt
  - 卖出：仅更新状态

- [x] 6.5 实现部分成交处理
  - 实现 `PartialFill(db, transactionID, filledQty)` 函数
  - 原记录状态 → OBSOLETE
  - 创建 FILLED 记录（已成交部分）
  - 创建 PENDING 记录（剩余部分）
  - 调整 locked_amt

---

## 7. 持仓计算

- [x] 7.1 实现持仓数量计算
  - 实现 `GetPosition(db, accountID, stockCode)` 函数
  - 计算：SUM(BUY) - SUM(SELL)

- [x] 7.2 实现持仓成本计算
  - 实现 `GetPositionCost(db, accountID, stockCode)` 函数
  - 计算：SUM(BUY金额) - SUM(SELL金额)

- [x] 7.3 实现持仓市值计算
  - 实现 `GetPositionValue(db, accountID, stockCode, priceAPI)` 函数
  - 调用实时价格 API
  - 计算：持仓数量 × 实时价格

- [x] 7.4 实现账户总资产计算
  - 实现 `GetAccountTotalCost(db, accountID)` 函数
  - 实现 `GetAccountTotalValue(db, accountID, priceAPI)` 函数
  - 实现 `GetAccountPnL(db, accountID, priceAPI)` 函数

---

## 8. 服务层封装

- [x] 8.1 创建账户服务
  - 创建 `pkg/service/account_service.go`
  - 封装账户操作的业务逻辑
  - 处理错误和异常情况

- [x] 8.2 完善交易服务
  - 完善 `pkg/service/trade_service.go`
  - 封装交易流程
  - 确保事务正确使用

---

## 9. 测试

- [x] 9.1 单元测试 - 数据库初始化
  - 测试数据库文件创建
  - 测试目录自动创建
  - 测试表创建

- [x] 9.2 单元测试 - 账户管理
  - 测试账户创建
  - 测试 user_id 唯一性约束
  - 测试账户查询
  - 测试状态更新

- [x] 9.3 单元测试 - 交易记录
  - 测试交易记录创建
  - 测试状态流转
  - 测试部分成交

- [x] 9.4 单元测试 - 金额锁定
  - 测试买入锁定
  - 测试卖出不锁定
  - 测试成交解锁
  - 测试撤销解锁
  - 测试部分成交调整

- [x] 9.5 单元测试 - 持仓计算
  - 测试持仓数量计算
  - 测试持仓成本计算
  - 测试盈亏计算

- [x] 9.6 集成测试 - 事务完整性
  - 测试正常流程事务提交
  - 测试异常情况事务回滚
  - 验证账户余额与交易记录一致性

---

## 10. 文档和清理

- [x] 10.1 添加代码注释
  - 为导出函数添加文档注释
  - 说明金额单位（分）和转换方法

- [x] 10.2 更新 README（如需要）
  - 说明数据库文件位置
  - 说明如何备份数据

---

## 基于知识的测试

### 事务回滚测试

> 📚 **最佳实践**
> 来源：数据库操作通用经验

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
