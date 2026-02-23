package model

// WebSearchParams 网页搜索参数
type WebSearchParams struct {
	Query string `json:"query" jsonschema:"description=搜索查询词，例如: 特斯拉 最新财报 | Search query, e.g.: Tesla latest earnings"`
}

// WebSearchResponse 网页搜索响应
type WebSearchResponse struct {
	Query     string             `json:"query"`
	Results   []SearchResultItem `json:"results"`
	RequestID string             `json:"request_id,omitempty"` // 请求唯一标识，用于日志追踪
	Status    string             `json:"status,omitempty"`     // 状态: "success", "failed", "captcha"
	Error     string             `json:"error,omitempty"`      // 错误代码
	Message   string             `json:"message,omitempty"`    // 用户友好的错误信息
}

// SearchResultItem 搜索结果项
type SearchResultItem struct {
	Title   string `json:"title" jsonschema:"description=网页标题 | Page title"`
	URL     string `json:"url" jsonschema:"description=网页 URL | Page URL"`
	Snippet string `json:"snippet" jsonschema:"description=搜索结果摘要 | Search result snippet"`
	Source  string `json:"source" jsonschema:"description=来源域名 | Source domain"`
}

// FetchPageParams 抓取页面内容参数
type FetchPageParams struct {
	URL       string `json:"url" jsonschema:"description=要抓取的网页 URL | URL of the page to fetch"`
	MaxLength int    `json:"max_length,omitempty" jsonschema:"description=最大内容长度，默认 5000 | Maximum content length, default 5000"`
}

// FetchPageResponse 抓取页面内容响应
type FetchPageResponse struct {
	URL         string `json:"url" jsonschema:"description=页面 URL | Page URL"`
	Title       string `json:"title" jsonschema:"description=页面标题 | Page title"`
	Content     string `json:"content" jsonschema:"description=页面内容 | Page content"`
	HasMore     bool   `json:"has_more" jsonschema:"description=是否还有更多内容 | Whether there is more content"`
	TotalLength int    `json:"total_length" jsonschema:"description=内容总长度 | Total content length"`
}
