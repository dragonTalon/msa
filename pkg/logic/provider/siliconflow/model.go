package siliconflow

// Model 表示一个 AI 模型
type Model struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// ModelsResponse 表示获取模型列表的响应
type ModelsResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}
