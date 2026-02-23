# 问题：浏览器超时时间过短

**ID**: 004  
**来源**: 2026-02-23-add-web-search-tool  
**类别**: performance  
**严重级别**: medium  

## 问题描述

在运行 web_search 工具时，出现 `context deadline exceeded` 错误。

## 根本原因

1. **5秒超时太短**：对于 Google 搜索页面，5秒可能不够
2. **WaitVisible 过于严格**：要求元素在视图中可见，这对于某些页面来说太严格

## 解决方案

```go
// 修改后
timeoutCtx, cancel := context.WithTimeout(ctx, 15*time.Second)  // 增加到 15 秒
defer cancel()

err = chromedp.Run(timeoutCtx,
    chromedp.Navigate(url),
    chromedp.WaitReady("body", chromedp.ByQuery),  // 只要求元素存在 DOM 中
    chromedp.Sleep(500*time.Millisecond),  // 给 JS 执行留出时间
    chromedp.OuterHTML("html", &html, chromedp.ByQuery),
)
```

## 预防措施

- 根据实际使用情况调整超时时间
- 对于关键页面，考虑使用更长的超时
- WaitReady 比 WaitVisible 更适合大多数场景
- 适当的 Sleep 可以提高稳定性
