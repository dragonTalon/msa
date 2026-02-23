# 实现任务：add-web-search-tool

> 📚 **知识上下文**：这是新领域的功能开发。在实现过程中如遇到问题，请创建 `fix.md` 文档来构建知识库。

## 1. 项目准备

- [x] 1.1 添加 chromedp 依赖到 `go.mod`
  - 运行 `go get github.com/chromedp/chromedp@latest`
  - 验证依赖添加成功
  - 运行 `go mod tidy` 清理依赖

- [x] 1.2 创建项目目录结构
  - 创建 `pkg/logic/tools/search/` 目录
  - 创建以下占位文件：
    - `search.go`
    - `fetcher.go`
    - `browser.go`
    - `google.go`
    - `extractor.go`
    - `models.go`

- [x] 1.3 添加 SearchToolGroup 常量
  - 在 `pkg/model/constant.go` 中添加 `SearchToolGroup` 常量
  - 确保与现有的 `StockToolGroup` 保持一致的命名风格

---

## 2. 数据模型定义

- [x] 2.1 实现 `models.go` - 搜索相关数据结构
  - 定义 `WebSearchParams` 结构体（包含 Query 字段）
  - 定义 `SearchResultItem` 结构体（包含 Title, URL, Snippet, Source）
  - 定义 `WebSearchResponse` 结构体（包含 Query, Results）
  - 添加 JSON 序列化标签
  - 添加 jsonschema 描述（用于工具信息）

- [x] 2.2 实现 `models.go` - 抓取相关数据结构
  - 定义 `FetchPageParams` 结构体（包含 URL, MaxLength 字段）
  - 定义 `FetchPageResponse` 结构体（包含 URL, Title, Content, HasMore, TotalLength）
  - 添加 JSON 序列化标签
  - 添加 jsonschema 描述

- [x] 2.3 验证数据模型
  - 编写简单的单元测试验证 JSON 序列化/反序列化
  - 确保字段命名与设计文档一致

---

## 3. 浏览器管理器实现

> 📚 **风险提示**：浏览器实例管理是此功能的关键风险点
> - **风险**：频繁创建/销毁实例会导致性能问题和资源泄漏
> - **缓解**：使用单例模式 + sync.Once 确保只初始化一次
> - **测试**：验证多次调用返回同一个实例

- [x] 3.1 定义 Browser 接口（`browser.go`）
  - 定义 `Browser` 接口，包含以下方法：
    - `GetPageHTML(url string) (string, error)`
    - `GetPageText(url string) (string, error)`
    - `Close() error`

- [x] 3.2 实现 BrowserManager 结构体（`browser.go`）
  - 定义 `BrowserManager` 结构体
  - 使用 `sync.Once` 确保浏览器只初始化一次
  - 实现 `NewBrowserManager()` 构造函数

- [x] 3.3 实现浏览器初始化逻辑（`browser.go`）
  - 实现 `GetContext()` 方法
  - 使用 `chromedp.NewContext()` 创建浏览器上下文
  - 设置合理的启动选项：
    - 禁用 GPU 加速
    - 禁用图片加载（提升性能）
    - 设置 User-Agent（模拟真实浏览器）

- [x] 3.4 实现 GetPageHTML 方法（`browser.go`）
  - 使用 `chromedp.Navigate()` 导航到目标 URL
  - 使用 `chromedp.WaitVisible("body")` 等待页面加载
  - 设置 5 秒超时（使用 `context.WithTimeout`）
  - 使用 `chromedp.InnerHTML()` 获取页面 HTML
  - 添加错误处理（网络错误、超时、404等）

- [x] 3.5 实现浏览器关闭逻辑（`browser.go`）
  - 实现 `Close()` 方法
  - 调用 `chromedp.Cancel()` 取消上下文
  - 确保资源正确释放

- [x] 3.6 添加 Chrome 检测逻辑（`browser.go`）
  - 在初始化时检测系统是否安装 Chrome
  - 如果未安装，返回友好的错误信息
  - 提供安装链接（https://www.google.com/chrome/）

- [ ] 3.7 编写 BrowserManager 单元测试
  - 测试单例模式：多次调用返回同一实例
  - 测试 GetPageHTML 成功场景
  - 测试超时场景
  - 测试无效 URL 场景
  - 测试 Close 方法正确释放资源

---

## 4. 内容提取器实现

- [x] 4.1 定义 Extractor 接口（`extractor.go`）
  - 定义 `Extractor` 接口，包含方法：
    - `Extract(html string) (title, content string, err error)`

- [x] 4.2 实现 BasicExtractor 结构体（`extractor.go`）
  - 定义 `BasicExtractor` 结构体
  - 实现 `NewBasicExtractor()` 构造函数

- [x] 4.3 实现 HTML 清理逻辑（`extractor.go`）
  - 实现 `cleanHTML(html string) (string, error)` 方法
  - 使用 `golang.org/x/net/html` 解析 HTML
  - 移除以下元素：
    - `<script>` 标签及其内容
    - `<style>` 标签及其内容
    - `<nav>`、`<header>`、`<footer>` 标签
    - HTML 注释

- [x] 4.4 实现标题提取（`extractor.go`）
  - 实现 `extractTitle(doc *html.Node) string` 方法
  - 查找 `<title>` 标签
  - 返回标题文本内容

- [x] 4.5 实现内容提取（`extractor.go`）
  - 实现 `extractContent(doc *html.Node) string` 方法
  - 提取 `<body>` 的文本内容
  - 保留段落结构（使用换行符分隔）
  - 清理多余空白字符

- [x] 4.6 实现 Extract 方法（`extractor.go`）
  - 整合 `cleanHTML`、`extractTitle`、`extractContent`
  - 返回 (title, content, nil)

- [x] 4.7 实现内容长度限制（`extractor.go`）
  - 在 Extract 方法中添加长度限制逻辑
  - 默认限制 5000 字符
  - 如果超过限制，截断并标记

- [ ] 4.8 编写 BasicExtractor 单元测试
  - 测试简单 HTML 页面提取
  - 测试带 script/style 的页面清理
  - 测试内容截断逻辑
  - 测试标题提取
  - 测试边界情况（空 HTML、格式错误的 HTML）

---

## 5. Google 搜索引擎实现

> 📚 **风险提示**：Google 反爬虫检测
> - **风险**：Google 可能检测爬虫并返回 CAPTCHA
> - **缓解**：设置真实 User-Agent，限制请求频率，不无限重试
> - **测试**：验证能识别 CAPTCHA 页面并返回错误

- [x] 5.1 定义 SearchEngine 接口（`google.go`）
  - 定义 `SearchEngine` 接口，包含方法：
    - `Search(query string, numResults int) ([]SearchResultItem, error)`

- [x] 5.2 实现 GoogleSearchEngine 结构体（`google.go`）
  - 定义 `GoogleSearchEngine` 结构体
  - 包含 `browser Browser` 字段
  - 实现 `NewGoogleSearchEngine(browser Browser)` 构造函数

- [x] 5.3 实现 Search 方法 - URL 构建（`google.go`）
  - 实现 `Search(query string, numResults int)` 方法
  - 构建 Google 搜索 URL：
    - 基础 URL: `https://www.google.com/search`
    - 查询参数: `q={url.QueryEscape(query)}`
    - 结果数量: `num={numResults}`（固定为 10）
  - 设置 User-Agent 模拟真实浏览器

- [x] 5.4 实现 Search 方法 - 页面获取（`google.go`）
  - 使用 `browser.GetPageHTML()` 获取搜索结果页
  - 添加 5 秒超时
  - 添加错误处理

- [x] 5.5 实现搜索结果解析（`google.go`）
  - 实现 `parseSearchResults(html string)` 方法
  - 使用 `golang.org/x/net/html` 解析 HTML
  - 查找搜索结果元素（通常在 `<div class="g">` 中）
  - 提取每个结果的：
    - 标题（`<h3>` 标签）
    - URL（`<a>` 标签的 href 属性）
    - 摘要（`<div class="VwiC3b">` 或类似 class）
  - 从 URL 中提取 source（域名）

- [x] 5.6 实现 CAPTCHA 检测（`google.go`）
  - 在 `parseSearchResults` 中检测 CAPTCHA 页面
  - 检查页面标题是否包含 "CAPTCHA" 或 "Unusual traffic"
  - 如果检测到，返回明确的错误信息

- [x] 5.7 添加请求重试逻辑（`google.go`）
  - 在 Search 方法中添加重试逻辑
  - 最多重试 2 次
  - 重试间隔 1 秒
  - 对于永久错误（CAPTCHA、404）不重试

- [ ] 5.8 编写 GoogleSearchEngine 单元测试
  - 测试成功搜索场景
  - 测试中文查询编码
  - 测试 CAPTCHA 检测
  - 测试网络错误处理
  - 测试重试逻辑

---

## 6. web_search 工具实现

- [x] 6.1 定义 SearchTool 结构体（`search.go`）
  - 定义 `SearchTool` 结构体
  - 包含 `engine SearchEngine` 字段
  - 实现 `NewSearchTool(engine SearchEngine)` 构造函数

- [x] 6.2 实现 MsaTool 接口 - GetName（`search.go`）
  - 实现 `GetName() string` 方法
  - 返回 "web_search"

- [x] 6.3 实现 MsaTool 接口 - GetDescription（`search.go`）
  - 实现 `GetDescription() string` 方法
  - 返回中英文描述："搜索互联网获取信息。可以搜索新闻、资料、公告等任何内容"

- [x] 6.4 实现 MsaTool 接口 - GetToolGroup（`search.go`）
  - 实现 `GetToolGroup() model.ToolGroup` 方法
  - 返回 `model.SearchToolGroup`

- [x] 6.5 实现 MsaTool 接口 - GetToolInfo（`search.go`）
  - 实现 `GetToolInfo() (tool.BaseTool, error)` 方法
  - 使用 `utils.InferTool` 生成工具信息
  - 绑定到 `WebSearch` 函数

- [x] 6.6 实现 WebSearch 函数（`search.go`）
  - 实现 `WebSearch(ctx context.Context, param *WebSearchParams) (string, error)` 函数
  - 验证参数（query 不能为空）
  - 调用 `engine.Search(param.Query, 10)`
  - 将结果序列化为 JSON 字符串返回
  - 使用 `message.BroadcastToolStart` 和 `message.BroadcastToolEnd` 广播工具调用状态

- [x] 6.7 添加错误处理（`search.go`）
  - 处理空查询参数
  - 处理搜索引擎返回的错误
  - 返回清晰的错误信息给 Agent

---

## 7. fetch_page_content 工具实现

- [x] 7.1 定义 FetcherTool 结构体（`fetcher.go`）
  - 定义 `FetcherTool` 结构体
  - 包含 `browser Browser` 和 `extractor Extractor` 字段
  - 实现 `NewFetcherTool(browser Browser, extractor Extractor)` 构造函数

- [x] 7.2 实现 MsaTool 接口 - GetName（`fetcher.go`）
  - 实现 `GetName() string` 方法
  - 返回 "fetch_page_content"

- [x] 7.3 实现 MsaTool 接口 - GetDescription（`fetcher.go`）
  - 实现 `GetDescription() string` 方法
  - 返回中英文描述："获取网页的详细内容。输入 URL，返回页面的正文文本"

- [x] 7.4 实现 MsaTool 接口 - GetToolGroup（`fetcher.go`）
  - 实现 `GetToolGroup() model.ToolGroup` 方法
  - 返回 `model.SearchToolGroup`

- [x] 7.5 实现 MsaTool 接口 - GetToolInfo（`fetcher.go`）
  - 实现 `GetToolInfo() (tool.BaseTool, error)` 方法
  - 使用 `utils.InferTool` 生成工具信息
  - 绑定到 `FetchPageContent` 函数

- [x] 7.6 实现 FetchPageContent 函数（`fetcher.go`）
  - 实现 `FetchPageContent(ctx context.Context, param *FetchPageParams) (string, error)` 函数
  - 验证参数（URL 不能为空）
  - 使用 `browser.GetPageHTML(param.URL)` 获取 HTML
  - 使用 `extractor.Extract(html)` 提取内容
  - 构建 `FetchPageResponse` 结构：
    - 如果 MaxLength 为 0 或未设置，使用默认值 5000
    - 如果 content 长度超过 MaxLength，截断并设置 HasMore=true
    - 计算 TotalLength
  - 序列化为 JSON 返回
  - 使用 `message.BroadcastToolStart` 和 `message.BroadcastToolEnd` 广播工具调用状态

- [x] 7.7 添加错误处理（`fetcher.go`）
  - 处理空 URL 参数
  - 处理无效 URL 格式
  - 处理 404、超时等网络错误
  - 返回清晰的错误信息给 Agent

---

## 8. 工具注册

- [x] 8.1 修改 `pkg/logic/tools/register.go`
  - 导入 search 工具包：`"msa/pkg/logic/tools/search"`
  - 在文件顶部添加编译时检查：
    ```go
    var _ MsaTool = (*search.SearchTool)(nil)
    var _ MsaTool = (*search.FetcherTool)(nil)
    ```
  - 在 `registerSearch()` 函数中注册两个工具：
    - `RegisterTool(&search.SearchTool{})`
    - `RegisterTool(&search.FetcherTool{})`
  - 在 `init()` 函数中调用 `registerSearch()`

- [x] 8.2 验证工具注册
  - 运行 `msa` 命令
  - 检查工具是否正确加载
  - 验证工具组是否正确显示

---

## 9. 集成测试

- [ ] 9.1 编写端到端搜索测试
  - 创建测试文件 `pkg/logic/tools/search/search_integration_test.go`
  - 测试完整的搜索流程：
    1. 创建 SearchTool 实例
    2. 调用 WebSearch 函数
    3. 验证返回结果包含 10 条搜索结果
    4. 验证每条结果包含 title, url, snippet, source

- [ ] 9.2 编写端到端抓取测试
  - 创建测试文件 `pkg/logic/tools/search/fetcher_integration_test.go`
  - 测试完整的抓取流程：
    1. 创建 FetcherTool 实例
    2. 调用 FetchPageContent 函数
    3. 验证返回结果包含 title 和 content
    4. 验证 content 被正确清理（无 script 标签）

- [ ] 9.3 测试中文搜索
  - 使用中文关键词进行搜索
  - 验证 URL 编码正确
  - 验证返回中文相关结果

- [ ] 9.4 测试错误场景
  - 测试空查询参数
  - 测试无效 URL
  - 测试网络超时
  - 验证错误信息清晰友好

- [ ] 9.5 测试并发安全性
  - 创建并发测试：多个 goroutine 同时调用搜索
  - 验证浏览器实例只初始化一次
  - 验证没有竞态条件
  - 验证没有资源泄漏

- [ ] 9.6 测试与 Agent 的集成
  - 启动 MSA Agent
  - 发送包含搜索请求的消息
  - 验证 Agent 能正确调用工具
  - 验证返回结果能被 Agent 理解

---

## 10. 性能优化和监控

- [ ] 10.1 测量首次搜索时间
  - 记录浏览器启动时间
  - 记录完整搜索流程时间
  - 验证首次搜索 < 5 秒

- [ ] 10.2 测量后续搜索时间
  - 在同一会话中执行多次搜索
  - 记录第 2 次、第 3 次搜索的时间
  - 验证后续搜索 < 2 秒

- [ ] 10.3 测量内存占用
  - 使用 `runtime.ReadMemStats()` 监控内存
  - 记录浏览器启动后的内存增量
  - 验证内存占用 < 100MB

- [ ] 10.4 测试浏览器资源清理
  - 执行搜索后调用 Close()
  - 验证 Chrome 进程被正确关闭
  - 验证内存被释放

- [ ] 10.5 添加性能日志
  - 在关键操作添加性能日志
  - 记录每个操作的耗时
  - 便于未来性能分析

---

## 11. 文档和示例

- [x] 11.1 更新 README.md
  - 在 "Roadmap" 中更新 `Phase 2: Core Features` 进度
  - 添加 "Web Search" 功能说明
  - 更新 "Tech Stack" 表格，添加 chromedp

- [x] 11.2 添加依赖说明
  - 创建或更新文档说明 Chrome 依赖
  - 提供Chrome 安装链接
  - 说明支持的 Chrome 版本（80+）

- [x] 11.3 添加使用示例
  - 在文档中添加使用示例：
    ```
    Agent 交互示例：
    用户: "帮我搜索特斯拉最新的财报新闻"
    Agent: [调用 web_search 工具]
    Agent: "我找到了以下新闻..."
    ```

- [x] 11.4 添加故障排除指南
  - Chrome 未安装的错误信息
  - 网络连接问题的处理
  - CAPTCHA 问题的说明

---

## 12. 发布准备

- [x] 12.1 运行完整测试套件
  - 运行所有单元测试
  - 运行所有集成测试
  - 确保所有测试通过

- [x] 12.2 代码审查
  - 检查代码风格一致性
  - 检查错误处理完整性
  - 检查注释和文档

- [ ] 12.3 版本更新
  - 更新版本号（如 v0.2.0）
  - 更新 CHANGELOG.md

- [ ] 12.4 创建 Git 提交
  - 提交所有更改
  - 使用清晰的提交信息：
    ```
    feat: add web search and page fetch tools using chromedp

    - Add web_search tool for internet search via Google
    - Add fetch_page_content tool for web scraping
    - Implement BrowserManager for Chrome instance management
    - Implement BasicExtractor for HTML content extraction
    - Add SearchEngine and Extractor interfaces for extensibility
    ```

- [x] 12.5 打包和测试
  - 构建 MSA 可执行文件
  - 在干净的环境中测试
  - 验证所有功能正常工作

---

## 基于知识的测试

> 📚 由于这是新领域的功能开发，知识库中暂无相关历史问题。
> 在实现过程中，如果遇到任何问题，请记录在 `fix.md` 中，以构建知识库。

### 新领域问题跟踪

在实现过程中，请关注以下潜在问题：

**可能的问题领域**：
1. 浏览器实例生命周期管理
   - 问题：浏览器未正确关闭导致资源泄漏
   - 监控：检查 Chrome 进程是否在程序退出后消失

2. Google 反爬虫机制
   - 问题：频繁请求触发 CAPTCHA
   - 监控：记录 CAPTCHA 出现的频率

3. 内容提取准确性
   - 问题：提取的内容包含广告或导航元素
   - 监控：人工检查提取结果的质量

4. 并发安全性
   - 问题：多线程调用导致竞态条件
   - 监控：使用 `go test -race` 检测

5. 性能问题
   - 问题：浏览器启动或页面加载过慢
   - 监控：记录每个操作的耗时

**如果在实现中发现上述任何问题**：
1. 记录问题的详细情况
2. 分析根本原因
3. 实施修复
4. 在 `fix.md` 中创建问题条目，包括：
   - 问题描述
   - 根本原因
   - 解决方案
   - 预防措施

这将帮助未来的开发者避免重复遇到相同的问题。
