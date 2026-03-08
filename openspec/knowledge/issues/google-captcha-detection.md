# Google CAPTCHA 检测限制

## 问题标识

- **ID**: google-captcha-detection
- **严重程度**: 🚫 已知限制
- **类别**: 外部服务限制
- **来源**: add-web-search-tool 实施总结

## 问题描述

Google 检测到自动化行为，返回 CAPTCHA 验证页面，导致搜索失败。

## 根本原因

Google 使用多种技术检测自动化行为：

1. **浏览器指纹检测**：检测 headless Chrome 的特征
2. **行为模式分析**：检测非人类的浏览模式
3. **IP 信誉评分**：频繁请求可能触发限制
4. **Cookie 和浏览历史**：新浏览器配置缺少正常使用痕迹

## 为什么难以完全避免

- Google 的反爬虫技术持续进化
- 完全模拟人类行为非常困难
- 即使使用真实浏览器，headless 模式仍有可识别特征
- IP 限制可能在多次请求后触发

## 错误信息示例

```
ERRO 搜索失败: Google 检测到自动化行为（CAPTCHA），请稍后再试或使用代理
```

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

## 可能的缓解措施

### 方案 A：使用 Google Custom Search API

**优点**：
- 官方 API，稳定可靠
- 无需担心 CAPTCHA
- 合规

**缺点**：
- 有费用
- 有配额限制

```go
type GoogleCSEClient struct {
    apiKey string
    cx      string
}

func (c *GoogleCSEClient) Search(query string) ([]SearchResultItem, error) {
    // 使用官方 API
}
```

### 方案 B：添加 Bing 搜索作为备选

**优点**：
- Bing 对自动化更友好
- 有官方 API

**缺点**：
- 需要额外实现

```go
type BingSearchEngine struct {
    apiKey string
}

func (b *BingSearchEngine) Search(query string, numResults int) ([]SearchResultItem, error) {
    // Bing 对自动化更友好
}
```

### 方案 C：使用 SerpAPI 等第三方服务

**优点**：
- 专门处理 Google 搜索结果
- 返回结构化数据

**缺点**：
- 有费用

```go
type SerpAPIClient struct {
    apiKey string
}

func (s *SerpAPIClient) Search(query string) ([]SearchResultItem, error) {
    // 专门处理 Google 搜索结果
}
```

### 方案 D：使用代理 IP

**优点**：
- 可以轮换 IP

**缺点**：
- 增加成本和复杂度
- 使用住宅代理更贵

## 预防措施

- 避免短时间内多次搜索
- 提供用户友好的错误提示
- 考虑实现搜索引擎降级策略

## 当前状态

- ✅ CAPTCHA 检测功能正常工作
- ✅ 错误信息清晰，用户知道问题所在
- ⚠️ 这是一个**已知限制**，不是 bug
- 📝 建议使用官方 API 或其他搜索引擎作为生产方案

## 教训

1. 对于重要的生产功能，官方 API 是最可靠的选择
2. 爬虫技术需要持续维护和更新
3. 依赖免费搜索 API 有固有的不稳定性
4. 在设计阶段就应该评估 CAPTCHA 风险

## 设计文档预见性

这个问题在设计文档的 **风险 2：Google 反爬虫** 中已经预见并记录了缓解措施。这说明设计阶段的风险评估是准确的。

## 相关知识

- [chromedp 超时配置](./chromedp-timeout-configuration.md)
- [错误处理策略](../patterns/error-handling-strategy.md)
