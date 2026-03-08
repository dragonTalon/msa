# chromedp 超时配置问题

## 问题标识

- **ID**: chromedp-timeout-configuration
- **严重程度**: ⚠️ 配置问题
- **类别**: 性能/稳定性
- **来源**: add-web-search-tool 实施总结

## 问题描述

在运行 web_search 工具时，出现 `context deadline exceeded` 错误。

## 错误日志

```
ERRO 获取页面 HTML 失败: context deadline exceeded
WARN 获取搜索结果失败（第 1 次尝试）: 获取页面 HTML 失败: context deadline exceeded
```

## 根本原因

### 1. 超时时间太短

5秒超时对于 Google 搜索页面来说不够：
- Chrome 首次启动需要 1-3 秒
- Google 搜索页面可能有额外的重定向
- JavaScript 执行需要额外时间
- 网络延迟不可预测

### 2. WaitVisible 过于严格

要求元素在视图中可见，这对于某些页面来说太严格。

## 解决方案

### 修改前

```go
timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

err = chromedp.Run(timeoutCtx,
    chromedp.Navigate(url),
    chromedp.WaitVisible("body", chromedp.ByQuery),  // 要求元素可见
    chromedp.OuterHTML("html", &html, chromedp.ByQuery),
)
```

### 修改后

```go
timeoutCtx, cancel := context.WithTimeout(ctx, 15*time.Second)  // 增加到 15 秒
defer cancel()

err = chromedp.Run(timeoutCtx,
    chromedp.Navigate(url),
    chromedp.WaitReady("body", chromedp.ByQuery),  // 只要求元素存在 DOM 中
    chromedp.Sleep(500*time.Millisecond),  // 给 JS 执行留出时间
    chromedp.OuterHTML("html", &html, chromedp.ByQuery),
)
```

## 变更说明

### 1. 超时时间：从 5 秒增加到 15 秒

- 给 Chrome 启动留出足够时间
- 给网络请求留出缓冲
- 给 JavaScript 执行留出时间

### 2. WaitReady 替代 WaitVisible

- `WaitReady`：只检查元素是否在 DOM 中存在（更快）
- `WaitVisible`：检查元素是否在视图中可见（更慢，更严格）

### 3. 添加 Sleep

给 JavaScript 执行留出 500ms 时间，避免 race condition。

## 预防措施

- 根据实际使用情况调整超时时间
- 对于关键页面，考虑使用更长的超时
- 监控实际请求时间，根据数据优化
- 考虑实现可配置的超时时间

## 教训

1. **在设置超时时间时，要考虑各种情况**
   - 冷启动时间
   - 网络波动
   - JavaScript 执行时间
   - 页面重定向

2. **WaitReady 比 WaitVisible 更适合大多数场景**
   - 除非真的需要元素可见，否则使用 WaitReady
   - WaitVisible 会等待元素在视图中渲染，更耗时

3. **适当的 Sleep 可以提高稳定性**
   - 避免 race condition
   - 给 JavaScript 执行留出时间
   - 特别是有动态内容的页面

## 相关知识

- [Google CAPTCHA 检测](./google-captcha-detection.md)
- [错误处理策略](../patterns/error-handling-strategy.md)
