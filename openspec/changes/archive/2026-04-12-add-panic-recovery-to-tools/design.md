# 设计：add-panic-recovery-to-tools

## 上下文

### 背景

当前 MSA 项目中的工具入口函数缺乏统一的 panic recovery 机制。当工具内部发生 panic 时（如 HTML 解析异常、空指针引用等），整个会话会崩溃，用户体验极差。

### 当前状态

- 所有工具入口函数直接暴露给 eino Agent
- 工具函数签名：`func(ctx context.Context, param *T) (string, error)`
- 错误处理：返回 JSON 格式的 `ToolResult`，但无 panic 保护

### 约束

- 必须保持向后兼容（函数签名不变）
- 使用现有的 `model.ToolResult` 结构
- 使用 `logrus` 进行日志记录
- 使用 `message.BroadcastToolEnd` 通知 UI

> 📚 **相关历史问题**：#005 Google CAPTCHA 检测、#004 浏览器超时配置

## 目标 / 非目标

**目标：**
- 为所有工具入口函数添加 panic recovery 保护
- 提供统一的辅助函数，减少重复代码
- 记录完整的 panic 日志，便于调试
- 保持向后兼容

**非目标：**
- 不改变工具的错误处理逻辑（正常 error 处理保持不变）
- 不引入新的依赖
- 不改变 `ToolResult` 结构

## 推荐模式

### 错误处理策略

**模式ID**：p003
**来源问题**：add-web-search-tool 实施总结

**描述**：
分类错误处理，根据错误类型采取不同的处理策略。参数错误立即返回，永久错误不重试，临时错误可重试。

**适用场景**：
- 所有工具入口函数
- 区分 panic 和正常 error

**实现**：
```go
// SafeExecute 提供统一的 panic recovery 保护
func SafeExecute(toolName string, params string, fn func() (string, error)) (result string, err error) {
    defer func() {
        if r := recover(); r != nil {
            // 记录完整日志
            log.Errorf("Tool [%s] panic recovered: %v\n%s", toolName, r, debug.Stack())

            // 广播工具结束
            errMsg := fmt.Errorf("工具内部错误: %v", r)
            message.BroadcastToolEnd(toolName, "", errMsg)

            // 返回错误 JSON
            result = model.NewErrorResult(fmt.Sprintf("工具内部错误: %v", r))
            err = nil // 不返回 error，让 LLM 继续处理
        }
    }()

    return fn()
}
```

### 浏览器管理器单例模式

**模式ID**：p004
**来源问题**：add-web-search-tool 实施总结

**描述**：
浏览器实例复用，独立 Tab 隔离并发任务。已有的浏览器管理实现可作为参考。

**适用场景**：
- search/fetcher 工具已有类似保护模式
- 可借鉴其隔离和恢复思想

## 避免的反模式

### 无 Panic Recovery 机制

**反模式ID**：ap001
**来源问题**：当前代码状态

**问题所在**：
工具函数内部任何 panic 都会导致整个会话崩溃。

**影响**：
- 用户体验差
- 无法优雅降级
- 难以调试（panic 信息可能丢失）

**不要这样做**：
```go
// 不要这样做 - 无保护
func WebSearch(ctx context.Context, param *model.WebSearchParams) (string, error) {
    // 直接执行，无 panic recovery
    return doSearch(ctx, param)
}
```

**应该这样做**：
```go
// 应该这样做 - 有保护
func WebSearch(ctx context.Context, param *model.WebSearchParams) (string, error) {
    return SafeExecute("web_search", fmt.Sprintf("query: %s", param.Query), func() (string, error) {
        return doSearch(ctx, param)
    })
}
```

### 掩盖真正的错误

**反模式ID**：ap002

**问题所在**：
Panic recovery 可能掩盖真正的编程错误。

**影响**：
- 难以发现和修复 bug
- 可能隐藏严重问题

**缓解措施**：
- 记录完整的堆栈信息
- 使用 ERROR 级别日志
- 在开发环境可以考虑重新 panic

## 设计决策

### 决策 1：辅助函数位置

**选择**：在 `pkg/logic/tools/basic.go` 中添加 `SafeExecute` 函数

**理由**：
- `basic.go` 已包含工具相关的基础定义
- 所有工具都可以方便地引用
- 保持代码组织的一致性

### 决策 2：返回值处理

**选择**：Panic 恢复后返回 `(string, nil)`，错误信息包装在 JSON 中

**理由**：
- 保持与现有工具的错误处理模式一致
- LLM 可以理解错误 JSON 并继续对话
- 不破坏 eino Agent 的调用约定

### 决策 3：日志级别

**选择**：使用 ERROR 级别记录 panic

**理由**：
- Panic 是严重的异常情况
- 需要引起开发者的注意
- 便于监控和告警

### 决策 4：堆栈信息

**选择**：使用 `runtime/debug.Stack()` 获取完整堆栈

**理由**：
- 便于定位问题
- 不增加额外依赖
- 标准库支持

## 风险 / 权衡

| 风险 | 缓解措施 |
|------|----------|
| 修改 24 个函数可能引入 bug | 统一辅助函数，减少重复代码；逐个模块修改并测试 |
| Panic recovery 可能掩盖错误 | 记录完整堆栈；使用 ERROR 级别日志 |
| 性能影响 | defer 开销极小，可忽略不计 |
| 遗漏某些工具 | 使用代码审查确保所有工具都被覆盖 |

## 迁移计划

### 阶段 1：添加辅助函数

1. 在 `pkg/logic/tools/basic.go` 中添加 `SafeExecute` 函数
2. 添加单元测试

### 阶段 2：修改工具函数（按优先级）

**高优先级**（已知问题风险高）：
1. `search/search.go` - WebSearch
2. `search/fetcher.go` - FetchPageContent

**中优先级**：
3. `stock/company.go` - GetStockCompanyCode
4. `stock/company_info.go` - GetStockCompanyInfo
5. `stock/company_k.go` - GetStockCompanyK

**低优先级**（风险较低）：
6. `finance/account.go` - 3 个函数
7. `finance/trade.go` - 3 个函数
8. `finance/position.go` - 2 个函数
9. `skill/skill.go` - 3 个函数
10. `todo/*.go` - 5 个函数
11. `knowledge/*.go` - 2 个函数

### 验证方式

- `go vet ./...` 验证编译问题
- 运行现有测试确保无回归
- 手动测试 search 工具的异常处理
