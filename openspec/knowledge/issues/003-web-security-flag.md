# 问题：不安全的 disable-web-security 标志

**ID**: 003  
**来源**: 2026-02-23-add-web-search-tool  
**类别**: security  
**严重级别**: high  

## 问题描述

代码中使用了 `chromedp.Flag("disable-web-security", true)` 标志。

## 根本原因

这是一个过度配置，从其他项目复制而来，没有仔细评估是否真的需要。

## 影响

- **安全风险**：禁用 web 安全会关闭同源策略（CORS）检查
- **不必要**：在我们的场景中，chromedp 直接访问网页，不需要处理跨域问题
- **误导性**：可能让维护者认为这是必需的

## 解决方案

移除 `disable-web-security` 标志，替换为更合适的选项：

```go
// 移除
chromedp.Flag("disable-web-security", true),

// 替换为
chromedp.Flag("disable-features", "VizDisplayCompositor"),  // 禁用后台网络限制
```

## 预防措施

- 理解每个 Chrome 标志的实际作用
- 只添加真正需要的选项
- 避免从其他项目盲目复制配置
- 安全优先：不要禁用安全特性，除非有明确理由
