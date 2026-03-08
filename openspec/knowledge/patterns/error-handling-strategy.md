# 错误处理策略

## 模式标识

- **ID**: error-handling-strategy
- **类别**: 设计模式/最佳实践
- **来源**: add-web-search-tool 实施总结

## 核心原则

分类错误处理，根据错误类型采取不同的处理策略。

## 错误分类

### 1. 参数错误（立即返回）

用户输入或参数不正确，立即返回错误，不重试。

```go
if query == "" {
    return nil, fmt.Errorf("搜索查询不能为空")
}
```

### 2. 永久错误（不重试）

无法通过重试解决的错误，立即返回：

- CAPTCHA 验证失败
- 404 页面不存在
- 权限不足
- 格式错误

```go
if engine.isCAPTCHA(html) {
    return nil, fmt.Errorf("Google 检测到自动化行为（CAPTCHA），请稍后再试")
}
```

### 3. 临时错误（重试）

网络超时、临时故障等，可以重试：

```go
var lastErr error
for attempt := 0; attempt < maxRetries; attempt++ {
    err := doSearch()
    if err == nil {
        return results, nil
    }

    // 判断是否可重试
    if isTemporaryError(err) {
        lastErr = err
        time.Sleep(backoff(attempt))
        continue
    }

    // 永久错误，立即返回
    return nil, err
}

// 所有重试都失败
return nil, fmt.Errorf("搜索失败（已重试 %d 次）: %w", maxRetries, lastErr)
```

## 清晰的错误信息

### 中英文双语

```go
return fmt.Errorf("Google 检测到自动化行为（CAPTCHA），请稍后再试 | " +
    "Google detected automated behavior (CAPTCHA), please try again later")
```

### 包含具体原因

```go
// ❌ 不好
return fmt.Errorf("搜索失败")

// ✅ 好
return fmt.Errorf("搜索失败: %w", err)
```

### 提供解决方案

```go
return fmt.Errorf("未找到 Chrome 浏览器 | Chrome browser not found. " +
    "Please install Chrome from: https://www.google.com/chrome/")
```

## 重试策略

### 退避策略

```go
func backoff(attempt int) time.Duration {
    // 指数退避：1s, 2s, 4s
    return time.Duration(1<<uint(attempt)) * time.Second
}
```

### 最大重试次数

```go
const maxRetries = 2  // 总共尝试 3 次（1 次初始 + 2 次重试）
```

### 可配置重试

```go
type RetryConfig struct {
    MaxRetries int
    Backoff    func(attempt int) time.Duration
}
```

## 错误包装

使用 `%w` 包装错误，保留错误链：

```go
return fmt.Errorf("获取页面 HTML 失败: %w", err)
```

调用者可以使用 `errors.Is()` 或 `errors.As()` 检查错误类型。

## 日志记录

### 错误级别

```go
// 参数错误：WARN
log.Warnf("搜索查询为空")

// 永久错误：ERROR
log.Errorf("搜索失败: %v", err)

// 临时错误重试：INFO
log.Infof("第 %d 次重试", attempt+1)
```

### 上下文信息

```go
log.Errorf("获取搜索结果失败（第 %d 次尝试）: %v", attempt+1, err)
```

## 用户友好的错误消息

对于终端用户，提供：

1. **清晰的错误描述**
2. **可能的原因**
3. **建议的解决方案**

```go
type UserError struct {
    Message     string
    Cause       string
    Suggestion  string
}

func (e *UserError) Error() string {
    return fmt.Sprintf("%s\n原因: %s\n建议: %s", e.Message, e.Cause, e.Suggestion)
}
```

## 相关知识

- [Google CAPTCHA 检测](../issues/google-captcha-detection.md)
- [chromedp 超时配置](../issues/chromedp-timeout-configuration.md)
- [浏览器管理器单例模式](./browser-manager-singleton.md)
