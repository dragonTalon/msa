package internal

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	searchmsa_model "msa/pkg/model"
)

// SearchEngine 搜索引擎接口
type SearchEngine interface {
	Search(ctx context.Context, query string, numResults int) ([]searchmsa_model.SearchResultItem, error)
}

// GoogleSearchEngine Google 搜索引擎实现
type GoogleSearchEngine struct {
	browser Browser
}

// NewGoogleSearchEngine 创建 Google 搜索引擎
func NewGoogleSearchEngine(browser Browser) *GoogleSearchEngine {
	return &GoogleSearchEngine{
		browser: browser,
	}
}

// Search 执行搜索
func (g *GoogleSearchEngine) Search(ctx context.Context, query string, numResults int) ([]searchmsa_model.SearchResultItem, error) {
	// 参数验证
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("搜索查询不能为空")
	}

	// 构建 Google 搜索 URL
	searchURL := g.buildSearchURL(query, numResults)
	log.Infof("执行搜索: %s", searchURL)

	// 使用重试逻辑获取搜索结果
	html, err := g.fetchWithRetry(ctx, searchURL)
	if err != nil {
		return nil, err
	}

	// 检测 CAPTCHA
	if g.isCAPTCHA(html) {
		return nil, fmt.Errorf("Google 检测到自动化行为（CAPTCHA），请稍后再试或使用代理")
	}

	// 解析搜索结果
	results, err := g.parseSearchResults(html)
	if err != nil {
		return nil, fmt.Errorf("解析搜索结果失败: %w", err)
	}

	log.Infof("成功获取 %d 条搜索结果", len(results))
	return results, nil
}

// buildSearchURL 构建 Google 搜索 URL
func (g *GoogleSearchEngine) buildSearchURL(query string, numResults int) string {
	baseURL := "https://www.google.com/search"
	params := url.Values{}
	params.Add("q", query)
	params.Add("num", fmt.Sprintf("%d", numResults))
	params.Add("hl", "zh-CN") // 设置语言为中文
	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

// fetchWithRetry 带重试的页面获取
func (g *GoogleSearchEngine) fetchWithRetry(ctx context.Context, searchURL string) (string, error) {
	const (
		maxRetries = 2
		retryDelay = 1 * time.Second
	)

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		// 检查 context 是否已取消
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		html, err := g.browser.GetPageHTML(ctx, searchURL)
		if err == nil {
			return html, nil
		}

		lastErr = err

		// 检查是否为永久错误（不重试）
		if isPermanentError(err) {
			return "", err
		}

		log.Warnf("获取搜索结果失败（第 %d 次尝试）: %v", i+1, err)
		if i < maxRetries-1 {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(retryDelay):
			}
		}
	}

	return "", fmt.Errorf("重试 %d 次后仍失败: %w", maxRetries, lastErr)
}

// isPermanentError 判断是否为永久错误（不重试）
func isPermanentError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	permanentErrors := []string{
		"captcha",
		"404",
		"403",
		"chrome not found",
		"未检测到 chrome",
	}

	for _, pe := range permanentErrors {
		if strings.Contains(errStr, pe) {
			return true
		}
	}

	return false
}

// isCAPTCHA 检测是否为 CAPTCHA 页面
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

// parseSearchResults 解析搜索结果
func (g *GoogleSearchEngine) parseSearchResults(htmlStr string) ([]searchmsa_model.SearchResultItem, error) {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return nil, fmt.Errorf("解析 HTML 失败: %w", err)
	}

	var results []searchmsa_model.SearchResultItem

	// 遍历 DOM 树查找搜索结果
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// Google 搜索结果通常在 <div class="g"> 中
			if g.isSearchResultContainer(n) {
				if result := g.extractSearchResult(n); result != nil {
					results = append(results, *result)
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)

	if len(results) == 0 {
		return nil, fmt.Errorf("未找到搜索结果，可能是页面结构变化或被反爬虫拦截")
	}

	return results, nil
}

// isSearchResultContainer 判断是否为搜索结果容器
func (g *GoogleSearchEngine) isSearchResultContainer(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}

	// 检查 class 属性
	for _, attr := range n.Attr {
		if attr.Key == "class" {
			classes := strings.Fields(attr.Val)
			for _, class := range classes {
				// Google 搜索结果容器通常是 class="g"
				if class == "g" || strings.HasPrefix(class, "g ") || strings.HasSuffix(class, " g") {
					return true
				}
			}
		}
	}

	return false
}

// extractSearchResult 从搜索结果容器中提取信息
func (g *GoogleSearchEngine) extractSearchResult(container *html.Node) *searchmsa_model.SearchResultItem {
	var result searchmsa_model.SearchResultItem
	var hasTitle, hasURL bool

	var extract func(*html.Node)
	extract = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// 提取标题和 URL（通常在 <h3> 下的 <a> 标签中）
			if n.DataAtom == atom.A || n.Data == "a" {
				if !hasTitle || !hasURL {
					for _, attr := range n.Attr {
						if attr.Key == "href" {
							// 清理 URL（Google 会添加前缀）
							result.URL = g.cleanURL(attr.Val)
							hasURL = true
						}
					}
					// 提取链接文本作为标题
					if title := g.extractTextContent(n); title != "" {
						result.Title = title
						hasTitle = true
					}
				}
			}

			// 提取摘要（通常在特定的 div 中）
			if strings.Contains(n.Data, "div") {
				for _, attr := range n.Attr {
					if attr.Key == "class" {
						// Google 摘要的 class 可能包含 "VwiC3b" 或 "st"
						if strings.Contains(attr.Val, "VwiC3b") ||
							strings.Contains(attr.Val, "st") ||
							strings.Contains(attr.Val, "aCOpRe") {
							if snippet := g.extractTextContent(n); snippet != "" && result.Snippet == "" {
								result.Snippet = snippet
							}
						}
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}

	extract(container)

	// 从 URL 中提取 source（域名）
	if result.URL != "" {
		if u, err := url.Parse(result.URL); err == nil {
			result.Source = u.Hostname()
		}
	}

	// 验证必需字段
	if !hasTitle || !hasURL {
		return nil
	}

	return &result
}

// cleanURL 清理 Google 搜索结果中的 URL
func (g *GoogleSearchEngine) cleanURL(rawURL string) string {
	// Google 搜索结果中的 URL 可能包含 /url?q= 前缀和重定向参数
	if strings.HasPrefix(rawURL, "/url?q=") {
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			return rawURL
		}
		if urlParam := parsedURL.Query().Get("url"); urlParam != "" {
			return urlParam
		}
	}

	// 移除 Google 的跟踪参数
	if u, err := url.Parse(rawURL); err == nil {
		u.RawQuery = "" // 移除查询参数
		return u.String()
	}

	return rawURL
}

// extractTextContent 提取节点的文本内容
func (g *GoogleSearchEngine) extractTextContent(n *html.Node) string {
	var content strings.Builder
	var extract func(*html.Node)

	extract = func(node *html.Node) {
		if node.Type == html.TextNode {
			text := strings.TrimSpace(node.Data)
			if text != "" {
				content.WriteString(text)
				content.WriteString(" ")
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}

	extract(n)
	return strings.TrimSpace(content.String())
}
