# 安全优先配置原则

## 模式标识

- **ID**: security-first-configuration
- **类别**: 设计原则/最佳实践
- **来源**: add-web-search-tool 实施总结

## 核心原则

**永远不要禁用安全特性，除非有明确理由。**

## 常见问题

### 问题：disable-web-security 标志

**不安全的配置**：
```go
chromedp.Flag("disable-web-security", true),
```

**为什么有问题**：
- 关闭同源策略（CORS）检查
- 允许跨域请求
- 可能导致安全漏洞
- 违反浏览器安全模型

**安全替代方案**：
```go
// 使用更具体的配置
chromedp.Flag("disable-features", "VizDisplayCompositor"),
```

## 评估标准

在添加任何配置前，问自己：

### 1. 这个配置是必须的吗？

**不是必须的示例**：
```go
// ❌ 不必要：禁用 web 安全
chromedp.Flag("disable-web-security", true),

// ✅ 必要：禁用图片（性能优化）
chromedp.Flag("blink-settings", "imagesEnabled", false),
```

### 2. 理解每个选项的安全影响

**安全影响评估**：
- `disable-web-security`：关闭 CORS，高风险 ❌
- `disable-gpu`：禁用 GPU 加速，低风险 ✅
- `headless`：无头模式，无安全影响 ✅

### 3. 使用最小权限原则

**只启用必需的功能**：
```go
// ✅ 好：只添加需要的配置
opts := []chromedp.ExecAllocatorOption{
    chromedp.Flag("headless", new),
    chromedp.Flag("disable-gpu", true),
    chromedp.Flag("blink-settings", "imagesEnabled", false),
    chromedp.UserAgent("MSA Bot/1.0"),
}

// ❌ 不好：复制粘贴大量配置，不理解每个选项的作用
opts := []chromedp.ExecAllocatorOption{
    chromedp.Flag("headless", new),
    chromedp.Flag("disable-gpu", true),
    chromedp.Flag("disable-web-security", true),  // 危险！
    chromedp.Flag("disable-features", "VizDisplayCompositor"),
    chromedp.Flag("disable-dev-shm-usage", true),
    // ... 更多配置
}
```

## Chrome 标志参考

### 安全相关（谨慎使用）

| 标志 | 安全影响 | 建议 |
|------|---------|------|
| `disable-web-security` | 高 | ❌ 避免使用 |
| `ignore-certificate-errors` | 高 | ❌ 避免使用 |
| `disable-features` | 低 | ✅ 可以使用 |
| `disable-gpu` | 无 | ✅ 推荐 |

### 性能相关（安全使用）

| 标志 | 影响 | 建议 |
|------|------|------|
| `blink-settings: imagesEnabled` | 禁用图片 | ✅ 节省带宽 |
| `headless` | 无头模式 | ✅ 服务器环境 |
| `disable-gpu` | 禁用 GPU | ✅ 服务器环境 |

## 安全检查清单

在添加新配置前，确认：

- [ ] 我理解这个配置的作用
- [ ] 我评估了安全风险
- [ ] 没有更安全的替代方案
- [ ] 这个配置是必需的
- [ ] 我添加了注释说明为什么需要

## 示例：添加注释

```go
opts := []chromedp.ExecAllocatorOption{
    chromedp.Flag("headless", "new"),
    chromedp.Flag("disable-gpu", true),
    // 禁用后台网络限制，避免超时
    chromedp.Flag("disable-features", "VizDisplayCompositor"),
    // 禁用图片加载以减少带宽和内存使用
    chromedp.Flag("blink-settings", "imagesEnabled", false),
    // 设置 User-Agent 以便服务器识别
    chromedp.UserAgent("MSA Bot/1.0 (+https://example.com/bot)"),
}
```

## 教训

1. **不要从其他项目盲目复制配置**
   - 每个项目有不同的需求
   - 复制的配置可能包含不必要的选项

2. **理解每个配置的作用**
   - 查阅官方文档
   - 测试配置的效果
   - 评估安全影响

3. **安全优先**
   - 默认使用安全选项
   - 只在有明确理由时使用不安全选项
   - 记录使用不安全选项的原因

## 相关知识

- [浏览器管理器单例模式](./browser-manager-singleton.md)
- [错误处理策略](./error-handling-strategy.md)
