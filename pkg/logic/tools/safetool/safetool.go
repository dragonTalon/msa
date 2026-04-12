package safetool

import (
	"fmt"
	"runtime/debug"

	log "github.com/sirupsen/logrus"
	"msa/pkg/logic/message"
	"msa/pkg/model"
)

// SafeExecute 提供统一的 panic recovery 保护
// 用于包装工具入口函数，防止 panic 导致会话中断
func SafeExecute(toolName string, params string, fn func() (string, error)) (result string, err error) {
	defer func() {
		if r := recover(); r != nil {
			// 记录完整日志（包含堆栈信息）
			log.Errorf("Tool [%s] panic recovered: %v\n%s", toolName, r, debug.Stack())

			// 广播工具结束消息
			errMsg := fmt.Errorf("工具内部错误: %v", r)
			message.BroadcastToolEnd(toolName, "", errMsg)

			// 返回错误 JSON（不返回 error，让 LLM 继续处理）
			result = model.NewErrorResult(fmt.Sprintf("工具内部错误: %v", r))
			err = nil
		}
	}()

	return fn()
}
