package session

import (
	"os"
	"path/filepath"
	"sync"

	log "github.com/sirupsen/logrus"
)

// Manager 会话管理器
type Manager struct {
	mu        sync.RWMutex
	current   *Session
	homeDir   string
	memoryDir string
}

var (
	globalManager *Manager
	once          sync.Once
)

// GetManager 获取全局 SessionManager 单例
func GetManager() *Manager {
	once.Do(func() {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Warnf("获取用户目录失败: %v", err)
			homeDir = "."
		}
		globalManager = &Manager{
			homeDir:   homeDir,
			memoryDir: filepath.Join(homeDir, ".msa", "memory"),
		}
	})
	return globalManager
}

// Current 获取当前会话
func (m *Manager) Current() *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current
}

// SetCurrent 设置当前会话
func (m *Manager) SetCurrent(session *Session) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.current = session
}

// Clear 清除当前会话，开始新会话
func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.current = nil
}

// GetMemoryDir 获取 memory 目录路径
func (m *Manager) GetMemoryDir() string {
	return m.memoryDir
}
