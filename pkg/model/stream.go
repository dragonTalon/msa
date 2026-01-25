package model

type StreamMsgType string

const (
	// StreamMsgTypeText 文本类型
	StreamMsgTypeText   StreamMsgType = "text"   //正文内容
	StreamMsgTypeTool   StreamMsgType = "tool"   //工具内容
	StreamMsgTypeReason StreamMsgType = "reason" //结束标志
)

// StreamChunk 流式输出的数据块
type StreamChunk struct {
	Content string // 文本内容
	IsDone  bool   // 是否结束
	Err     error  // 错误信息
	MsgType StreamMsgType
}
