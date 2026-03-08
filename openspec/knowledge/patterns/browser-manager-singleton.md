# 浏览器管理器单例模式

## 模式标识

- **ID**: browser-manager-singleton
- **类别**: 设计模式
- **来源**: add-web-search-tool 实施总结

## 问题场景

需要管理 Chrome 浏览器实例用于网页抓取，同时避免：
- 频繁创建/销毁 Chrome 进程（性能开销大）
- 多个 Chrome 实例同时运行（资源浪费）
- 浏览器资源没有正确清理（内存泄漏）

## 解决方案

使用单例模式 + `sync.Once` 确保浏览器只初始化一次。

## 实现示例

```go
type BrowserManager struct {
    ctx    context.Context
    cancel context.CancelFunc
    once   sync.Once
}

var globalBrowser *BrowserManager

func GetBrowser() *BrowserManager {
    if globalBrowser == nil {
        globalBrowser = &BrowserManager{}
    }
    return globalBrowser
}

func (b *BrowserManager) GetContext() (context.Context, error) {
    var initErr error
    b.once.Do(func() {
        b.ctx, b.cancel = b.createContext()
        initErr = b.startChrome()
    })
    return b.ctx, initErr
}

func (b *BrowserManager) Close() error {
    if b.cancel != nil {
        b.cancel()
    }
    return nil
}
```

## 关键要素

### 1. sync.Once 确保只初始化一次

```go
b.once.Do(func() {
    // 只会执行一次
    b.ctx, b.cancel = b.createContext()
})
```

### 2. 延迟初始化

浏览器只在第一次使用时才启动，不是程序启动时。

### 3. 资源清理

提供 Close() 方法正确释放资源。

## 优点

1. **性能优化**：避免频繁创建/销毁 Chrome 进程
2. **资源节约**：只运行一个 Chrome 实例
3. **线程安全**：sync.Once 确保并发安全
4. **易于使用**：全局访问点

## 配置优化

### 禁用图片加载

```go
chromedp.Flag("blink-settings", "imagesEnabled", false),
```

### 禁用 GPU

```go
chromedp.Flag("disable-gpu", true),
```

### 禁用 web 安全（谨慎使用）

```go
// ❌ 不推荐：有安全风险
chromedp.Flag("disable-web-security", true),

// ✅ 推荐：使用更安全的替代方案
chromedp.Flag("disable-features", "VizDisplayCompositor"),
```

### 无头模式

```go
chromedp.Flag("headless", "new"),
```

## 应用场景

- 网页抓取
- 自动化测试
- 内容提取
- 搜索引擎接口

## 注意事项

1. **超时配置**：根据实际场景调整超时时间
2. **错误处理**：提供清晰的错误信息
3. **资源清理**：确保程序退出时调用 Close()
4. **Chrome 检测**：自动检测 Chrome 安装路径

## 相关知识

- [chromedp 超时配置](../issues/chromedp-timeout-configuration.md)
- [错误处理策略](./error-handling-strategy.md)
- [安全优先配置](./security-first-configuration.md)
