package message

import (
	"fmt"
	msamodel "msa/pkg/model"
	"sync"

	log "github.com/sirupsen/logrus"
)

// StreamOutputManager 管理流式输出的订阅者
type StreamOutputManager struct {
	mu          sync.RWMutex
	subscribers []chan *msamodel.StreamChunk
}

// 全局流式输出管理器
var globalStreamManager = &StreamOutputManager{}

// GetStreamManager 获取全局流式输出管理器
func GetStreamManager() *StreamOutputManager {
	return globalStreamManager
}

// RegisterStreamOutput 注册一个流式输出的 channel
// 返回 channel 和取消注册的函数
func RegisterStreamOutput(bufferSize int) (<-chan *msamodel.StreamChunk, func()) {
	return globalStreamManager.Register(bufferSize)
}

// Register 注册一个流式输出的 channel
// 返回 channel 和取消注册的函数
func (m *StreamOutputManager) Register(bufferSize int) (<-chan *msamodel.StreamChunk, func()) {
	ch := make(chan *msamodel.StreamChunk, bufferSize)

	m.mu.Lock()
	m.subscribers = append(m.subscribers, ch)
	m.mu.Unlock()

	// 返回取消注册的函数
	unregister := func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		for i, sub := range m.subscribers {
			if sub == ch {
				m.subscribers = append(m.subscribers[:i], m.subscribers[i+1:]...)
				close(ch)
				break
			}
		}
	}

	return ch, unregister
}

// Broadcast 广播消息给所有订阅者
func (m *StreamOutputManager) Broadcast(chunk *msamodel.StreamChunk) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, ch := range m.subscribers {
		// 非阻塞发送，防止慢消费者阻塞整个流程
		select {
		case ch <- chunk:
		default:
			// channel 满了就跳过，避免阻塞
			log.Warnf("stream output channel is full, skipping chunk")
		}
	}
}

// HasSubscribers 检查是否有订阅者
func (m *StreamOutputManager) HasSubscribers() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.subscribers) > 0
}

// SubscriberCount 返回当前订阅者数量
func (m *StreamOutputManager) SubscriberCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.subscribers)
}

// BroadcastToolStart 广播工具调用开始消息
// toolName: 工具名称
// params: 工具参数描述
func BroadcastToolStart(toolName string, params string) {
	globalStreamManager.Broadcast(&msamodel.StreamChunk{
		Content: fmt.Sprintf("\n正在调用工具: %s, 参数: %s", toolName, params),
		MsgType: msamodel.StreamMsgTypeTool,
	})
	log.Debugf("Tool [%s] started with params: %s", toolName, params)
}

// BroadcastToolEnd 广播工具调用结束消息
// toolName: 工具名称
// result: 工具执行结果描述
// err: 错误信息（如果有）
func BroadcastToolEnd(toolName string, result string, err error) {
	if err != nil {
		globalStreamManager.Broadcast(&msamodel.StreamChunk{
			Content: fmt.Sprintf("工具 %s 执行失败: %v\n", toolName, err),
			MsgType: msamodel.StreamMsgTypeTool,
			Err:     err,
		})
		log.Errorf("Tool [%s] failed: %v", toolName, err)
	} else {
		globalStreamManager.Broadcast(&msamodel.StreamChunk{
			Content: fmt.Sprintf("工具 %s 执行完成: %s\n", toolName, result),
			MsgType: msamodel.StreamMsgTypeTool,
		})
		log.Debugf("Tool [%s] completed: %s", toolName, result)
	}
}
