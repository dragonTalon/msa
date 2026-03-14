package memory

import (
	"context"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"msa/pkg/model"
)

// ==================== Chat 与记忆系统集成 ====================

// InitChatMemory 初始化 Chat 的记忆系统
// 应该在 NewChat 时调用
func InitChatMemory(ctx context.Context) error {
	manager := GetManager()
	if manager == nil {
		return fmt.Errorf("记忆管理器未初始化")
	}

	// 初始化新会话
	if err := manager.Initialize(); err != nil {
		log.Warnf("初始化记忆会话失败: %v", err)
		// 记忆系统初始化失败不应阻止聊天
		return nil
	}

	log.Infof("记忆会话已初始化: %s", manager.GetSessionID())
	return nil
}

// AddChatMessage 添加聊天消息到记忆
// 应该在发送用户消息后调用
func AddChatMessage(role string, content string) error {
	manager := GetManager()
	if manager == nil || !manager.IsInitialized() {
		// 记忆系统未初始化，跳过
		return nil
	}

	// 转换角色
	var msgType model.MemoryMessageType
	switch role {
	case "user":
		msgType = model.MemoryMessageTypeUser
	case "assistant":
		msgType = model.MemoryMessageTypeAssistant
	case "system":
		msgType = model.MemoryMessageTypeSystem
	default:
		return fmt.Errorf("未知的消息角色: %s", role)
	}

	msg := model.MemoryMessage{
		ID:        generateMessageID(),
		Type:      msgType,
		Content:   content,
		Timestamp: time.Now(),
	}

	if err := manager.AddMessage(msg); err != nil {
		log.Warnf("添加消息到记忆失败: %v", err)
		// 不影响聊天继续进行
	}

	return nil
}

// EndChatMemory 结束 Chat 的记忆会话
// 应该在退出时调用
func EndChatMemory() (string, error) {
	manager := GetManager()
	if manager == nil || !manager.IsInitialized() {
		// 记忆系统未初始化，跳过
		return "", nil
	}

	if err := manager.EndSession(); err != nil {
		log.Warnf("结束记忆会话失败: %v", err)
		return "", err
	}

	// 获取会话ID
	sessionID := manager.GetLastSessionID()

	log.Infof("记忆会话已结束: %s", sessionID)
	return sessionID, nil
}

// GetLastSessionID 获取最后的会话ID（用于恢复提示）
func GetLastSessionID() string {
	manager := GetManager()
	if manager == nil {
		return ""
	}
	return manager.GetLastSessionID()
}

// GetKnowledgeForPrompt 获取知识用于注入 AI Prompt
// 应该在构建 AI 查询消息时调用
func GetKnowledgeForPrompt() string {
	manager := GetManager()
	if manager == nil {
		return ""
	}

	knowledge := manager.GetKnowledgeForPrompt()
	if knowledge != "" {
		log.Debugf("知识已加载，长度: %d 字符", len(knowledge))
	}

	return knowledge
}

// ==================== 辅助函数 ====================

// generateMessageID 生成消息 ID
func generateMessageID() string {
	return fmt.Sprintf("msg-%d", time.Now().UnixNano())
}

// ==================== 环境变量支持 ====================

const (
	// 记忆系统启用环境变量
	EnvMemoryEnabled = "MSA_MEMORY_ENABLED"
)

// IsMemoryEnabled 检查记忆系统是否启用
func IsMemoryEnabled() bool {
	// 检查环境变量
	if envValue := getEnv(EnvMemoryEnabled); envValue != "" {
		return envValue != "false" && envValue != "0"
	}

	// 默认启用
	return true
}

// getEnv 获取环境变量（可测试的包装）
func getEnv(key string) string {
	return os.Getenv(key)
}

// ==================== 统计信息 ====================

// GetMemoryStats 获取记忆系统统计信息
func GetMemoryStats() (*model.MemoryStats, error) {
	if !IsMemoryEnabled() {
		return nil, fmt.Errorf("记忆系统未启用")
	}

	manager := GetManager()
	if manager == nil {
		return nil, fmt.Errorf("记忆管理器未初始化")
	}

	return manager.store.GetStats()
}
