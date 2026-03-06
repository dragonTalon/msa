# 规格：amount-locking

金额锁定机制，用于挂单时锁定资金，确保资金安全。

## 增加需求

### 需求：买入时锁定金额

系统必须在提交买入订单时锁定相应金额。

#### 场景：提交买入订单
- **当** 用户提交买入订单
- **并且** available_amt >= amount + fee
- **那么** 系统应在事务中执行：
  - 创建交易记录（status=PENDING）
  - available_amt 减少 (amount + fee)
  - locked_amt 增加 (amount + fee)
- **并且** 确保操作的原子性

#### 场景：余额不足
- **当** 用户提交买入订单
- **并且** available_amt < amount + fee
- **那么** 系统应拒绝订单
- **并且** 创建交易记录（status=REJECTED）
- **并且** note 字段记录"余额不足"
- **并且** 不修改账户金额

### 需求：卖出时不锁定金额

卖出订单不锁定金额，但需验证持仓充足。

#### 场景：提交卖出订单
- **当** 用户提交卖出订单
- **并且** 持仓数量 >= 卖出数量
- **那么** 系统应创建交易记录（status=PENDING）
- **并且** 不修改 available_amt 和 locked_amt

#### 场景：持仓不足
- **当** 用户提交卖出订单
- **并且** 持仓数量 < 卖出数量
- **那么** 系统应拒绝订单
- **并且** 创建交易记录（status=REJECTED）
- **并且** note 字段记录"持仓不足"

### 需求：成交时解锁金额

订单成交时需要正确处理锁定金额。

#### 场景：买入成交
- **当** 买入订单成交（status: PENDING → FILLED）
- **那么** 系统应减少 locked_amt（amount + fee）
- **并且** available_amt 保持不变（已在提交时扣除）

#### 场景：卖出成交
- **当** 卖出订单成交（status: PENDING → FILLED）
- **那么** 系统应增加 available_amt（amount - fee）
- **并且** locked_amt 保持不变（卖出时未锁定）

### 需求：撤销时解锁金额

订单撤销时需要将锁定金额返还。

#### 场景：买入订单撤销
- **当** 买入订单撤销（status: PENDING → CANCELLED）
- **那么** 系统应在事务中执行：
  - 更新交易状态为 CANCELLED
  - locked_amt 减少（amount + fee）
  - available_amt 增加（amount + fee）
- **并且** 确保操作的原子性

#### 场景：卖出订单撤销
- **当** 卖出订单撤销（status: PENDING → CANCELLED）
- **那么** 系统应更新交易状态为 CANCELLED
- **并且** 不修改账户金额（卖出时未锁定）

### 需求：部分成交时的锁定调整

部分成交时需要正确调整锁定金额。

#### 场景：买入部分成交
- **当** 买入订单部分成交（如 100 股成交 50 股）
- **那么** 系统应：
  - 将原记录状态更新为 OBSOLETE
  - 创建已成交部分记录（status=FILLED, parent_id=原id）
  - 创建剩余部分记录（status=PENDING, parent_id=原id）
  - locked_amt 调整为剩余部分的锁定金额
- **并且** 已成交部分的锁定金额被释放（不返还到 available）

#### 场景：卖出部分成交
- **当** 卖出订单部分成交
- **那么** 系统应：
  - 将原记录状态更新为 OBSOLETE
  - 创建已成交部分记录（status=FILLED）
  - 创建剩余部分记录（status=PENDING）
  - available_amt 增加已成交部分的金额
- **并且** 剩余部分继续等待成交

### 需求：锁定金额验证

系统必须确保锁定金额的一致性。

#### 场景：验证锁定金额
- **当** 系统执行任何涉及金额的操作前
- **那么** 应验证：total_amount >= available_amt + locked_amt
- **并且** 验证：available_amt >= 0, locked_amt >= 0
- **并且** 如果验证失败，拒绝操作并记录错误
