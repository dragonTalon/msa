# 规格：account-summary

账户总览计算逻辑，正确计入持仓市值并提示价格获取失败的情况。

## 修改需求

### 需求：总资产计算必须包含持仓市值

总资产必须正确计算为：可用余额 + 锁定金额 + 持仓市值。

#### 场景：正常计算总资产

- **当** 账户有可用余额 10000 元、锁定金额 0 元、持有股票 A 100 股（当前价格 12 元）
- **那么** 总资产 = 10000 + 0 + 1200 = 11200 元

#### 场景：无持仓时只计算余额

- **当** 账户有可用余额 10000 元、锁定金额 500 元、无任何持仓
- **那么** 总资产 = 10000 + 500 + 0 = 10500 元

---

### 需求：GetAccountTotalValue 返回完整信息

`GetAccountTotalValue` 必须返回 `AccountValueResult` 结构体，包含总资产、持仓市值、价格获取失败的股票列表。

#### 场景：返回完整计算结果

- **当** 调用 `GetAccountTotalValue(db, accountID, priceMap)`
- **那么** 返回 `AccountValueResult` 包含：
  - `TotalValue`: 总资产（可用 + 锁定 + 持仓市值）
  - `AvailableAmt`: 可用余额
  - `LockedAmt`: 锁定金额
  - `PositionValue`: 持仓市值
  - `FailedStocks`: 价格获取失败的股票代码列表

#### 场景：记录价格获取失败的股票

- **当** 账户持有股票 A 和股票 B，但 priceMap 中只有股票 A 的价格
- **那么** `FailedStocks` 包含 `["B"]`，`PositionValue` 只计算股票 A 的市值

---

### 需求：部分价格失败时仍计算可用市值

即使部分股票价格获取失败，系统仍应计算能获取价格的股票市值，并明确提示哪些股票价格缺失。

#### 场景：部分价格失败

- **当** 账户持有股票 A（100 股，价格 12 元）和股票 B（200 股，价格未知）
- **那么** `PositionValue` = 1200 元（只计算股票 A）
- **那么** `FailedStocks` = `["B"]`
- **那么** 返回成功（而非错误），但在 `Warning` 字段提示"部分股票价格获取失败"

#### 场景：全部价格失败

- **当** 账户持有股票 A 和股票 B，但两者的价格都无法获取
- **那么** `PositionValue` = 0
- **那么** `FailedStocks` = `["A", "B"]`
- **那么** `Warning` 提示"无法获取持仓价格，市值未计入"

---

### 需求：AccountSummaryData 包含失败信息

`AccountSummaryData` 必须包含 `FailedStocks` 和 `Warning` 字段，让用户知晓数据不完整。

#### 场景：返回数据包含失败信息

- **当** 有股票价格获取失败
- **那么** `AccountSummaryData.FailedStocks` 包含失败股票代码列表
- **那么** `AccountSummaryData.Warning` 包含人类可读的警告信息

#### 场景：无失败时字段为空

- **当** 所有股票价格都成功获取
- **那么** `FailedStocks` 为空数组 `[]`
- **那么** `Warning` 为空字符串 `""`

---

### 需求：移除错误的 BroadcastToolEnd 调用

价格获取失败不应触发 `BroadcastToolEnd` 作为错误返回，而应作为警告信息继续执行。

#### 场景：价格失败不中断流程

- **当** `fetchAllPrices` 返回错误或部分股票价格获取失败
- **那么** 不调用 `message.BroadcastToolEnd` 作为错误返回
- **那么** 继续执行，将失败信息记录在返回数据中

---

### 需求：向后兼容

API 响应新增字段为可选字段，不影响现有调用方。

#### 场景：新增字段可选

- **当** 调用方解析 `AccountSummaryData` JSON 响应
- **那么** `FailedStocks` 和 `Warning` 字段可被忽略
- **那么** 现有字段 `TotalAssets`、`AvailableAmt`、`PositionValue` 等保持不变
