# 指南：chromedp 网页抓取实施指南

**ID**: 001  
**来源**: 2026-02-23-add-web-search-tool  
**类别**: implementation  

## 概述

本指南总结了使用 chromedp 实现网页抓取功能的经验和最佳实践。

## 架构设计

```
pkg/logic/tools/search/
├── models.go      # 数据模型定义
├── browser.go     # 浏览器管理（单例模式）
├── extractor.go   # 内容提取器
├── google.go      # Google 搜索引擎
├── search.go      # web_search 工具
└── fetcher.go     # fetch_page_content 工具
```

## 接口抽象层次

- `Browser` 接口 → BrowserManager 实现
- `SearchEngine` 接口 → GoogleSearchEngine 实现
- `Extractor` 接口 → BasicExtractor 实现

## 浏览器管理

使用单例模式管理 Chrome 实例：

```go
type BrowserManager struct {
    ctx         context.Context
    cancel      context.CancelFunc
    allocCancel context.CancelFunc
    once        sync.Once
}

func (b *BrowserManager) GetContext() (context.Context, error) {
    b.once.Do(func() {
        // 初始化浏览器
    })
    return b.ctx, nil
}
```

## 性能优化

1. **浏览器实例复用**：避免频繁创建/销毁 Chrome 进程
2. **禁用图片加载**：显著减少网络流量和内存占用
3. **合理的超时设置**：15 秒页面加载超时
4. **并发控制**：使用 `sync.Mutex` 保护共享状态

## 技术债务

1. **测试覆盖率低**：需要添加单元测试和集成测试
2. **Google 结果解析脆弱**：Google 页面结构变化会导致解析失败
3. **缺少配置选项**：超时时间、结果数量等硬编码
4. **缺少日志和监控**：难以诊断生产环境问题

## 未来改进方向

1. **财经新闻专用提取器**
2. **多搜索引擎支持**
3. **搜索结果重排序**
4. **结果缓存**
5. **HTTP 备选方案**
