# 提案：fix-account-summary-position-value

修复账户总览工具计算总资产时未正确计入持仓市值的问题。

## 知识上下文

### 相关历史问题

知识库中未找到与此变更直接相关的历史问题（金融/交易/持仓计算领域）。这是该领域首次记录的变更。

### 需避免的反模式

未识别到特定反模式。

### 推荐模式

未识别到特定模式。

## 为什么

`GetAccountSummary` 工具在计算总资产时存在 Bug：

1. **价格获取失败时未正确处理**：`fetchAllPrices` 返回错误时，代码调用 `BroadcastToolEnd` 但没有 `return`，导致 `priceMap` 为空继续执行
2. **持仓市值被忽略**：当 `priceMap` 为空时，`GetAccountTotalValue` 跳过所有股票，导致总资产 = 余额 + 锁定，持仓市值完全丢失
3. **用户看到错误的资产数据**：用户期望看到 `总资产 = 可用余额 + 锁定金额 + 持仓市值`

### 问题代码位置

`pkg/logic/tools/finance/position.go:207-215`

```go
if err != nil {
    // 价格获取失败不影响其他计算
    message.BroadcastToolEnd("get_account_summary", "", err)
    // ❌ 缺少 return！priceMap 保持为空
}
```

## 变更内容

- 重构 `GetAccountTotalValue` 返回更完整的 `AccountValueResult` 结构体
- 修改 `AccountSummaryData` 增加 `FailedStocks` 和 `Warning` 字段
- 修复价格获取失败时的错误处理逻辑
- 即使部分股票价格获取失败，也尽可能计算准确的资产价值

## 能力

### 新增能力

无新增能力。

### 修改能力

- `account-summary`：账户总览计算逻辑，需正确计入持仓市值并提示价格获取失败的情况

## 影响

**代码影响**：
- `pkg/logic/finsvc/position_service.go` - 新增结构体，修改函数签名
- `pkg/logic/tools/finance/position.go` - 修改数据处理和错误处理逻辑
- `pkg/logic/finsvc/position_service_test.go` - 更新测试
- `pkg/logic/tools/finance/integration_test.go` - 更新测试

**向后兼容性**：
- 函数签名变更，但仅内部调用
- API 响应新增可选字段，向后兼容

---

**知识备注**：
- 此提案使用 `openspec/knowledge/` 中的知识生成
- 最后知识同步：首次变更
- 知识库总问题数：5
- 总模式数：6
- 此领域首次记录变更，建议完成后归档 fix.md 以构建知识库
