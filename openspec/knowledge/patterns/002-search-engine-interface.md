# 模式：搜索引擎接口设计模式

**ID**: 002  
**来源**: 2026-02-23-add-web-search-tool  
**类别**: design-patterns  

## 概述

搜索引擎实现必须与工具层解耦，便于未来添加其他搜索引擎。

## 接口定义

```go
type SearchEngine interface {
    Search(ctx context.Context, query string, numResults int) ([]SearchResultItem, error)
}
```

## 实现示例

```go
type GoogleSearchEngine struct {
    browser Browser
}

func NewGoogleSearchEngine(browser Browser) *GoogleSearchEngine {
    return &GoogleSearchEngine{browser: browser}
}

func (g *GoogleSearchEngine) Search(ctx context.Context, query string, numResults int) ([]SearchResultItem, error) {
    // 实现搜索逻辑
}
```

## 优势

- 添加新搜索引擎只需实现接口
- 不需要修改工具层代码
- 可以通过配置切换默认引擎
- 便于单元测试
