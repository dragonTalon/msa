# 问题：chromedp 依赖未正确添加

**ID**: 002  
**来源**: 2026-02-23-add-web-search-tool  
**类别**: build  
**严重级别**: medium  

## 问题描述

首次编译时出现 `no required module provides package github.com/chromedp/chromedp` 错误。

## 根本原因

在项目根目录运行 `go get` 后，需要在子包编译前运行 `go mod tidy`。

## 解决方案

```bash
go get github.com/chromedp/chromedp
go mod tidy
go build ./pkg/logic/tools/search/
```

## 预防措施

- 在添加新依赖后，始终运行 `go mod tidy`
- 验证整个项目可以编译：`go build ./...`
