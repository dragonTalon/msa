# 实现任务：fix-account-summary-position-value

## ⚠️ 风险警报

基于本次变更分析的风险：

### 中等：函数签名变更影响未知调用方

- **缓解措施**：已确认仅内部调用，但仍需运行完整测试验证

### 低：用户可能忽略 Warning 字段

- **缓解措施**：Warning 字段包含人类可读信息，LLM 可解析并告知用户

## 1. 修改 position_service.go

- [x] 1.1 新增 `AccountValueResult` 结构体

> 📚 **知识上下文**：使用结果结构体模式，返回多个相关值

```go
type AccountValueResult struct {
    TotalValue    int64
    AvailableAmt  int64
    LockedAmt     int64
    PositionValue int64
    FailedStocks  []string
}
```

- [x] 1.2 修改 `GetAccountTotalValue` 函数返回类型

> 📚 **知识上下文**：错误降级模式 - 返回可用数据 + 失败信息，而非完全失败

- 从 `(int64, error)` 改为 `(*AccountValueResult, error)`
- 在函数内部记录 `FailedStocks`（有持仓但 priceMap 中无价格的股票）
- 直接返回 `PositionValue` 而非通过减法推算

## 2. 更新测试文件

- [x] 2.1 更新 `position_service_test.go` 中的 `TestGetAccountTotalValue`

- 适配新的返回类型 `*AccountValueResult`
- 验证 `TotalValue`、`PositionValue`、`FailedStocks` 字段

- [x] 2.2 更新 `integration_test.go` 中调用 `GetAccountTotalValue` 的部分

- 适配新的返回类型

## 3. 修改 position.go (Tool 层)

- [x] 3.1 修改 `AccountSummaryData` 结构体

```go
type AccountSummaryData struct {
    // ... 现有字段 ...
    FailedStocks []string `json:"failed_stocks,omitempty"`
    Warning      string   `json:"warning,omitempty"`
}
```

- [x] 3.2 修改 `GetAccountSummary` 函数

> ⚠️ **反模式警告**：不要在错误处理后继续执行而不清理状态

- 移除第 209 行错误的 `BroadcastToolEnd` 调用
- 使用新的 `AccountValueResult` 结构
- 根据 `FailedStocks` 生成 `Warning` 信息
- 确保 `BroadcastToolEnd` 只在函数末尾调用一次

## 4. 验证

- [x] 4.1 运行 `go vet ./...` 验证编译

- [x] 4.2 运行单元测试 `go test ./pkg/logic/finsvc/...`

- [x] 4.3 运行集成测试 `go test ./pkg/logic/tools/finance/...`

## 基于知识的测试

### 错误处理后继续执行回归测试

> 📚 **来源**：本次发现的 Bug

**历史问题**：
`fetchAllPrices` 返回错误时，调用 `BroadcastToolEnd` 但没有 `return`，导致 `priceMap` 为空继续执行。

**测试场景**：
1. Mock `fetchAllPrices` 返回错误
2. 调用 `GetAccountSummary`
3. 验证函数正常返回（不 panic、不中断）
4. 验证 `FailedStocks` 包含预期股票
5. 验证 `Warning` 包含提示信息

**预期行为**：
- 函数成功返回，`TotalValue` = 可用余额 + 锁定金额
- `PositionValue` = 0（因为价格获取失败）
- `FailedStocks` 包含所有持仓股票
- `Warning` 提示"无法获取持仓价格"

### 部分价格失败测试

**测试场景**：
1. Mock 账户持有股票 A 和 B
2. Mock `fetchAllPrices` 只返回股票 A 的价格
3. 调用 `GetAccountSummary`

**预期行为**：
- `PositionValue` = 股票 A 的市值
- `FailedStocks` = `["B"]`
- `Warning` 提示"部分股票价格获取失败"
