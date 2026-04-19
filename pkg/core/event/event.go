// pkg/core/event/event.go
package event

// EventType 定义 pipeline 中所有可能的事件类型
type EventType int

const (
	// 文本流
	EventTextChunk EventType = iota // LLM 输出的普通文本片段
	EventTextDone                   // 本轮文本输出完毕

	// 工具调用
	EventToolStart  // 工具开始执行（携带工具名、参数）
	EventToolResult // 工具执行完成（携带结果）
	EventToolError  // 工具执行失败

	// 思考过程（reasoning model 支持）
	EventThinking // <thinking> 内容片段

	// 会话控制
	EventRoundDone // 一轮对话完成（agent 不再调用工具）
	EventError     // 不可恢复的错误，pipeline 终止
)

// Event 是 pipeline 中流动的最小单元
type Event struct {
	Type EventType

	// EventTextChunk / EventThinking
	Text string

	// EventToolStart
	Tool ToolCall

	// EventToolResult / EventToolError
	Result ToolResult

	// EventError
	Err error
}

// ToolCall 描述一次工具调用请求
type ToolCall struct {
	ID    string
	Name  string
	Input string // JSON 字符串
}

// ToolResult 描述一次工具调用的结果
type ToolResult struct {
	ToolCallID string
	Name       string
	Output     string
	IsError    bool
}
