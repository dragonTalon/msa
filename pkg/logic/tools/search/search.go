package search

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	log "github.com/sirupsen/logrus"
	"msa/pkg/logic/message"
	internal "msa/pkg/logic/tools/search/internal"
	"msa/pkg/model"
	mas_utils "msa/pkg/utils"
)

// SearchTool 搜索工具
type SearchTool struct {
	router *internal.SearchRouter
}

// NewSearchTool 创建搜索工具（使用路由器）
func NewSearchTool(router *internal.SearchRouter) *SearchTool {
	return &SearchTool{
		router: router,
	}
}

// NewSearchToolWithEngine 创建搜索工具（使用单个引擎，向后兼容）
func NewSearchToolWithEngine(engine internal.SearchEngine) *SearchTool {
	// 创建一个包装器，将 SearchEngine 转换为 SearchRouter
	return &SearchTool{
		router: nil, // 将在运行时初始化
	}
}

// GetToolInfo 获取工具信息
func (s *SearchTool) GetToolInfo() (tool.BaseTool, error) {
	return utils.InferTool(s.GetName(), s.GetDescription(), WebSearch)
}

// GetName 获取工具名称
func (s *SearchTool) GetName() string {
	return "web_search"
}

// GetDescription 获取工具描述
func (s *SearchTool) GetDescription() string {
	return "搜索互联网获取信息。可以搜索新闻、资料、公告等任何内容 | Search the internet for information. Can search for news, data, announcements, and any other content"
}

// GetToolGroup 获取工具组
func (s *SearchTool) GetToolGroup() model.ToolGroup {
	return model.SearchToolGroup
}

// WebSearch 执行网页搜索
func WebSearch(ctx context.Context, param *model.WebSearchParams) (string, error) {
	log.Debugf("WebSearch start, query: %s", param.Query)

	// 使用公共函数记录工具调用开始
	message.BroadcastToolStart("web_search", fmt.Sprintf("query: %s", param.Query))

	// 参数校验
	if param.Query == "" {
		err := fmt.Errorf("搜索查询不能为空 | search query cannot be empty")
		message.BroadcastToolEnd("web_search", "", err)
		return "", err
	}

	// 初始化搜索路由器（如果未初始化）
	if defaultSearchRouter == nil {
		browser := internal.NewBrowserManager()
		searchLogger := internal.NewSearchLogger(log.StandardLogger())
		defaultSearchRouter = internal.NewSearchRouter(browser, searchLogger)
	}
	// 执行搜索（支持自动降级）
	searchResult := defaultSearchRouter.Search(ctx, param.Query, 10)

	// 构建响应
	response := &model.WebSearchResponse{
		Query:     param.Query,
		Results:   searchResult.Results,
		RequestID: searchResult.RequestID,
		Status:    searchResult.Status,
		Error:     searchResult.Error,
		Message:   searchResult.Message,
	}

	// 记录结果
	resultJSON := mas_utils.ToJSONString(response)

	if searchResult.Status == "success" {
		message.BroadcastToolEnd("web_search",
			fmt.Sprintf("找到 %d 条结果 (引擎: %s)", len(searchResult.Results), searchResult.UsedEngine), nil)
		log.Debugf("搜索成功，返回 %d 条结果 (引擎: %s, RequestID: %s)",
			len(searchResult.Results), searchResult.UsedEngine, searchResult.RequestID)
	} else {
		// 失败不返回错误，而是通过 status 字段告知
		message.BroadcastToolEnd("web_search",
			fmt.Sprintf("搜索失败: %s", searchResult.Message), nil)
		log.Warnf("搜索失败: %s (RequestID: %s)", searchResult.Error, searchResult.RequestID)
	}

	return resultJSON, nil
}

// GetDefaultRouter 获取默认路由器（用于测试）
func GetDefaultRouter() *internal.SearchRouter {
	return defaultSearchRouter
}

// SetDefaultRouter 设置默认路由器（用于测试）
func SetDefaultRouter(router *internal.SearchRouter) {
	defaultSearchRouter = router
}

// 全局默认搜索路由器实例
var defaultSearchRouter *internal.SearchRouter
