# 问题：Google CAPTCHA 检测

**ID**: 005  
**来源**: 2026-02-23-add-web-search-tool  
**类别**: integration  
**严重级别**: high  

## 问题描述

Google 检测到自动化行为，返回 CAPTCHA 验证页面，导致搜索失败。

## 根本原因

Google 使用多种技术检测自动化行为：
1. 浏览器指纹检测：检测 headless Chrome 的特征
2. 行为模式分析：检测非人类的浏览模式
3. IP 信誉评分：频繁请求可能触发限制
4. Cookie 和浏览历史：新浏览器配置缺少正常使用痕迹

## 为什么难以完全避免

- Google 的反爬虫技术持续进化
- 完全模拟人类行为非常困难
- 即使使用真实浏览器，headless 模式仍有可识别特征
- IP 限制可能在多次请求后触发

## 当前实现

代码已经正确检测并报告 CAPTCHA：

```go
func (g *GoogleSearchEngine) isCAPTCHA(html string) bool {
    lowerHTML := strings.ToLower(html)
    captchaIndicators := []string{
        "captcha",
        "unusual traffic",
        "请证明您不是机器人",
        "verify you are not a robot",
    }

    for _, indicator := range captchaIndicators {
        if strings.Contains(lowerHTML, indicator) {
            return true
        }
    }
    return false
}
```

## 推荐的解决方案

对于生产环境，建议采用以下方案之一：

**方案 A：使用 Google Custom Search API**
- 优点：合规、稳定
- 缺点：有费用、有配额限制

**方案 B：添加 Bing 搜索作为备选**
- 优点：Bing 对自动化更友好
- 缺点：搜索结果质量可能不同

**方案 C：使用 SerpAPI 等第三方服务**
- 优点：专门处理搜索结果，返回结构化数据
- 缺点：有费用

## 当前状态

- ✅ CAPTCHA 检测功能正常工作
- ✅ 错误信息清晰，用户知道问题所在
- ⚠️ 这是一个**已知限制**，不是 bug
- 📝 建议使用官方 API 或其他搜索引擎作为生产方案
