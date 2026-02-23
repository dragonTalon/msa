# 模式：chromedp 浏览器配置最佳实践

**ID**: 001  
**来源**: 2026-02-23-add-web-search-tool  
**类别**: browser-automation  

## 概述

基于使用 chromedp 的经验，总结出以下浏览器配置最佳实践。

## 推荐配置

```go
opts := append(chromedp.DefaultExecAllocatorOptions[:],
    // 禁用 GPU 加速
    chromedp.Flag("disable-gpu", true),
    // 禁用扩展
    chromedp.Flag("disable-extensions", true),
    // 禁用图片加载（提升性能）
    chromedp.Flag("blink-settings", "imagesEnabled=false"),
    // 设置 User-Agent 模拟真实浏览器
    chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
)
```

## 应避免的配置

```go
// ❌ 不要使用：安全风险
chromedp.Flag("disable-web-security", true),

// ✅ 替代方案
chromedp.Flag("disable-features", "VizDisplayCompositor"),
```

## 超时设置

- 页面加载：15 秒
- 等待元素：WaitReady 而非 WaitVisible
- 添加 Sleep 给 JS 执行留出时间

## 单例模式

使用 `sync.Once` 确保浏览器只初始化一次，复用实例。
