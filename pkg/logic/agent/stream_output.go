package agent

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

// StreamChunk 流式输出的数据块
type StreamChunk struct {
	Content string // 文本内容
	IsDone  bool   // 是否结束
	Err     error  // 错误信息
}

// StreamOutputManager 管理流式输出的订阅者
type StreamOutputManager struct {
	mu          sync.RWMutex
	subscribers []chan *StreamChunk
}

// 全局流式输出管理器
var globalStreamManager = &StreamOutputManager{}

// GetStreamManager 获取全局流式输出管理器
func GetStreamManager() *StreamOutputManager {
	return globalStreamManager
}

// RegisterStreamOutput 注册一个流式输出的 channel
// 返回 channel 和取消注册的函数
func RegisterStreamOutput(bufferSize int) (<-chan *StreamChunk, func()) {
	return globalStreamManager.Register(bufferSize)
}

// Register 注册一个流式输出的 channel
// 返回 channel 和取消注册的函数
func (m *StreamOutputManager) Register(bufferSize int) (<-chan *StreamChunk, func()) {
	ch := make(chan *StreamChunk, bufferSize)

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
func (m *StreamOutputManager) Broadcast(chunk *StreamChunk) {
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
