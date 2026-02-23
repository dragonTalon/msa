# 模式：错误处理和重试策略

**ID**: 003  
**来源**: 2026-02-23-add-web-search-tool  
**类别**: error-handling  

## 概述

系统必须优雅地处理各种错误情况，并提供清晰的反馈。

## 错误分类

1. **参数错误**：立即返回
2. **永久错误**（CAPTCHA、404）：不重试
3. **临时错误**（网络超时）：重试 2 次

## 重试实现

```go
func fetchWithRetry(ctx context.Context, url string) (string, error) {
    const maxRetries = 2
    
    for i := 0; i < maxRetries; i++ {
        result, err := fetch(ctx, url)
        if err == nil {
            return result, nil
        }
        
        if isPermanentError(err) {
            return "", err
        }
        
        time.Sleep(1 * time.Second)
    }
    
    return "", fmt.Errorf("重试 %d 次后仍失败", maxRetries)
}
```

## 错误信息

- 中英文双语错误提示
- 包含具体的错误原因
- 提供解决方案（如 Chrome 安装链接）
