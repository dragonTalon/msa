package search

import (
	"context"
	"fmt"
	"msa/pkg/logic/message"
	internal "msa/pkg/logic/tools/search/internal"
	"msa/pkg/model"
	mas_utils "msa/pkg/utils"
	"net/url"
	"sync"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	log "github.com/sirupsen/logrus"
)

// FetcherTool 页面抓取工具
type FetcherTool struct {
	browser   internal.Browser
	extractor internal.Extractor
}

// NewFetcherTool 创建页面抓取工具
func NewFetcherTool(browser internal.Browser, extractor internal.Extractor) *FetcherTool {
	return &FetcherTool{
		browser:   browser,
		extractor: extractor,
	}
}

// GetToolInfo 获取工具信息
func (f *FetcherTool) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(f.GetName(), f.GetDescription(), FetchPageContent)
}

// GetName 获取工具名称
func (f *FetcherTool) GetName() string {
	return "fetch_page_content"
}

// GetDescription 获取工具描述
func (f *FetcherTool) GetDescription() string {
	return "获取网页的详细内容。输入 URL，返回页面的正文文本 | Fetch detailed content of a webpage. Input URL, return the main text content"
}

// GetToolGroup 获取工具组
func (f *FetcherTool) GetToolGroup() model.ToolGroup {
	return model.SearchToolGroup
}

// FetchPageContent 抓取页面内容
func FetchPageContent(ctx context.Context, param *model.FetchPageParams) (string, error) {
	log.Infof("FetchPageContent start, url: %s", param.URL)

	// 使用公共函数记录工具调用开始
	message.BroadcastToolStart("fetch_page_content", fmt.Sprintf("url: %s", param.URL))

	// 参数校验
	if param.URL == "" {
		err := fmt.Errorf("URL 不能为空 | URL cannot be empty")
		message.BroadcastToolEnd("fetch_page_content", "", err)
		return "", err
	}

	// 验证 URL 格式
	if _, err := url.Parse(param.URL); err != nil {
		err = fmt.Errorf("无效的 URL 格式: %w | invalid URL format: %w", err, err)
		message.BroadcastToolEnd("fetch_page_content", "", err)
		return "", err
	}

	// 使用共享的浏览器实例（懒加载单例，线程安全）
	fetcherOnce.Do(func() {
		defaultFetcherTool = NewFetcherTool(GetSharedBrowser(), internal.NewBasicExtractor())
	})

	// 获取页面 HTML
	html, err := defaultFetcherTool.browser.GetPageHTML(ctx, param.URL)
	if err != nil {
		log.Errorf("获取页面失败: %v", err)
		message.BroadcastToolEnd("fetch_page_content", "", err)
		return "", err
	}

	// 提取内容
	title, content, err := defaultFetcherTool.extractor.Extract(html)
	if err != nil {
		log.Errorf("提取内容失败: %v", err)
		message.BroadcastToolEnd("fetch_page_content", "", err)
		return "", err
	}

	// 设置默认长度限制
	maxLength := param.MaxLength
	if maxLength <= 0 {
		maxLength = 5000
	}

	if content == "" {
		content = html
	}

	// 计算长度
	totalLength := len(content)
	hasMore := totalLength > maxLength
	returnedLength := maxLength

	// 截断内容
	if hasMore {
		content = content[:maxLength]
	} else {
		returnedLength = totalLength
	}

	// 构建响应
	response := &model.FetchPageResponse{
		URL:         param.URL,
		Title:       title,
		Content:     content,
		HasMore:     hasMore,
		TotalLength: totalLength,
	}

	resultJSON := mas_utils.ToJSONString(response)
	message.BroadcastToolEnd("fetch_page_content", fmt.Sprintf("获取 %d 字符（共 %d 字符）", returnedLength, totalLength), nil)
	log.Debugf("页面抓取成功，返回 %d/%d 字符", returnedLength, totalLength)

	return resultJSON, nil
}

// 全局默认抓取工具实例
var (
	defaultFetcherTool *FetcherTool
	fetcherOnce        sync.Once
)
