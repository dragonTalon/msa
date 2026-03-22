# 设计：fix-account-summary-position-value

## 上下文

### 当前状态

`GetAccountSummary` 工具存在 Bug：价格获取失败时，持仓市值被完全忽略，导致返回的总资产不准确。

**问题代码流程**：

```
GetAccountSummary
    │
    ├── fetchAllPrices(stockCodes)
    │       │
    │       └── 返回 prices map 或 error
    │
    ├── ❌ 如果 error，调用 BroadcastToolEnd 但没有 return
    │       → priceMap 保持为空
    │
    └── GetAccountTotalValue(db, accountID, priceMap)
            │
            └── 遍历 stockCodes
                    │
                    └── if !priceMap[code] { continue }  // 跳过所有股票
                            │
                            └── positionValue = 0
                                    │
                                    └── totalValue = AvailableAmt + LockedAmt + 0
```

### 约束

- 金额以"毫"为单位（1 元 = 10000 毫）
- 价格获取可能因网络问题失败
- 需要向后兼容现有 API 响应格式

## 目标 / 非目标

**目标：**
- 重构 `GetAccountTotalValue` 返回完整计算信息
- 修复价格获取失败时的错误处理逻辑
- 让用户知道数据是否完整

**非目标：**
- 不改变价格获取机制本身
- 不改变数据库结构
- 不添加重试逻辑

## 推荐模式

### 错误降级模式（Graceful Degradation）

**描述**：
当部分数据获取失败时，返回可用数据而非完全失败。在返回中明确标注哪些数据缺失。

**适用场景**：
- 外部数据获取可能失败
- 部分数据仍有价值
- 用户需要知道数据完整性

**实现**：

```go
// 不要这样做：失败时返回错误
if err != nil {
    return nil, err
}

// 应该这样做：失败时返回可用数据 + 警告
result := &AccountValueResult{
    TotalValue:    available + locked + positionValue,
    PositionValue: positionValue,
    FailedStocks:  failedStocks,  // 记录哪些失败
}
return result, nil
```

### 结果结构体模式（Result Object Pattern）

**描述**：
使用结构体返回多个相关值，而非 tuple 或多个返回值。

**适用场景**：
- 函数需要返回多个相关值
- 未来可能扩展返回内容
- 需要自文档化的返回类型

**实现**：

```go
// AccountValueResult 账户价值计算结果
type AccountValueResult struct {
    TotalValue    int64    // 总资产
    AvailableAmt  int64    // 可用余额
    LockedAmt     int64    // 锁定金额
    PositionValue int64    // 持仓市值
    FailedStocks  []string // 价格获取失败的股票
}

// 替代原来的
// func GetAccountTotalValue(...) (int64, error)
func GetAccountTotalValue(...) (*AccountValueResult, error)
```

## 避免的反模式

### 错误处理后继续执行但不清理状态

**问题所在**：
错误处理分支没有 return，导致后续代码使用未初始化或空的数据。

**影响**：
计算结果错误，但用户不知道。

**不要这样做**：

```go
if err != nil {
    log.Error(err)
    // 没有 return！
}
// 继续使用可能为空的 priceMap
```

**应该这样做**：

```go
if err != nil {
    // 方案 A：返回错误
    return nil, err

    // 方案 B：记录失败但继续（降级处理）
    // 将失败信息记录下来，在结果中标注
}
```

### 隐式计算导致精度丢失

**问题所在**：
通过减法推算持仓市值，可能因中间计算导致精度问题。

**影响**：
`PositionValue = TotalValue - AvailableAmt - LockedAmt` 可能不精确。

**不要这样做**：

```go
positionValue := totalValue - account.AvailableAmt - account.LockedAmt
```

**应该这样做**：

```go
// 直接返回计算过程中的 positionValue
result := &AccountValueResult{
    PositionValue: positionValue,  // 直接使用计算值
}
```

## 设计决策

### 决策 1：新增 AccountValueResult 结构体

**理由**：需要返回多个相关值（总资产、持仓市值、失败股票列表），使用结构体比多个返回值更清晰且易于扩展。

**位置**：`pkg/logic/finsvc/position_service.go`

### 决策 2：修改 GetAccountTotalValue 返回类型

**变更**：
```go
// 之前
func GetAccountTotalValue(database *gorm.DB, accountID uint, prices PriceMap) (int64, error)

// 之后
func GetAccountTotalValue(database *gorm.DB, accountID uint, prices PriceMap) (*AccountValueResult, error)
```

**理由**：调用者需要知道持仓市值和失败股票，而非仅总资产。

### 决策 3：AccountSummaryData 新增可选字段

**新增字段**：
```go
type AccountSummaryData struct {
    // ... 现有字段 ...

    FailedStocks []string `json:"failed_stocks,omitempty"` // 价格获取失败的股票
    Warning      string   `json:"warning,omitempty"`       // 人类可读的警告
}
```

**理由**：让用户知道数据是否完整，使用 `omitempty` 保证向后兼容。

### 决策 4：价格获取失败时不调用 BroadcastToolEnd 作为错误

**变更**：移除 `position.go:209` 的 `message.BroadcastToolEnd("get_account_summary", "", err)` 调用。

**理由**：价格获取失败是"警告"而非"错误"，应记录在返回数据中而非中断流程。

## 风险 / 权衡

| 风险 | 缓解措施 |
|------|----------|
| 函数签名变更可能影响未知的调用方 | 检查所有调用点，确认仅内部调用 |
| 用户可能忽略 Warning 字段 | Warning 字段包含人类可读信息，LLM 可解析并告知用户 |
| 部分价格失败时总资产可能被低估 | FailedStocks 明确列出哪些股票未计入 |

## 迁移计划

无需数据库迁移。代码变更按以下顺序进行：

1. 在 `position_service.go` 新增 `AccountValueResult` 结构体
2. 修改 `GetAccountTotalValue` 函数实现和返回类型
3. 更新 `position_service_test.go` 测试
4. 更新 `integration_test.go` 测试
5. 修改 `AccountSummaryData` 结构体
6. 修改 `GetAccountSummary` 函数使用新返回
7. 运行 `go vet ./...` 验证编译
8. 运行测试验证功能
