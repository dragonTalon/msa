package deepseek

// Model DeepSeek 模型信息
type Model struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	OwnedBy string `json:"owned_by"`
}

// ModelsResponse DeepSeek 模型列表 API 响应
type ModelsResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}
