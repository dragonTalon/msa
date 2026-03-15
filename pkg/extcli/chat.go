package extcli

import (
	"context"
	"fmt"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"msa/pkg/config"
	"msa/pkg/logic/agent"
	"msa/pkg/logic/memory"
	"msa/pkg/logic/message"
	"msa/pkg/model"
)

const (
	streamBufferSize = 100
)

// streamCollector 用于收集流式输出的完整内容
type streamCollector struct {
	mu            sync.Mutex
	textContent   strings.Builder // 正文内容
	reasonContent strings.Builder // 思考内容
	toolContent   strings.Builder // 工具调用内容
}

func (c *streamCollector) append(content string, msgType model.StreamMsgType) {
	c.mu.Lock()
	defer c.mu.Unlock()
	switch msgType {
	case model.StreamMsgTypeText:
		c.textContent.WriteString(content)
	case model.StreamMsgTypeReason:
		c.reasonContent.WriteString(content)
	case model.StreamMsgTypeTool:
		c.toolContent.WriteString(content)
	}
}

// getContent 获取完整回复内容（正文 + 思考内容）
func (c *streamCollector) getContent() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	var result strings.Builder

	// 先添加思考内容（如果有）
	if c.reasonContent.Len() > 0 {
		result.WriteString("【思考过程】\n")
		result.WriteString(c.reasonContent.String())
		result.WriteString("\n\n")
	}

	// 再添加正文内容
	result.WriteString(c.textContent.String())

	return result.String()
}

// hasContent 检查是否有实际内容
func (c *streamCollector) hasContent() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.textContent.Len() > 0 || c.reasonContent.Len() > 0
}

// Run 执行 CLI 单轮对话
// 返回退出码：0 成功，1 失败
func Run(ctx context.Context, question string, modelOverride string) int {
	// 1. 参数验证
	if question == "" {
		log.Error("问题内容不能为空")
		return 1
	}

	// 2. 获取配置
	cfg := config.GetLocalStoreConfig()
	if cfg == nil {
		log.Error("配置未初始化，请运行 'msa config'")
		return 1
	}

	// 3. 验证必要配置
	if cfg.APIKey == "" {
		log.Error("请先配置 API Key，运行 'msa config'")
		return 1
	}
	if cfg.BaseURL == "" {
		log.Error("请先配置 Base URL，运行 'msa config'")
		return 1
	}

	// 4. 处理模型覆盖
	effectiveModel := cfg.Model
	if modelOverride != "" {
		effectiveModel = modelOverride
		log.Infof("使用命令行指定的模型: %s", effectiveModel)
	}
	if effectiveModel == "" {
		log.Error("请指定模型 (-m) 或配置默认模型")
		return 1
	}

	// 临时覆盖配置中的模型
	if modelOverride != "" {
		cfg.Model = modelOverride
	}

	// 5. 清除 agent 缓存以使用新配置
	agent.ResetCache()

	// 6. 初始化记忆系统
	memoryInitialized := false
	if memory.IsMemoryEnabled() {
		if err := memory.InitChatMemory(ctx); err != nil {
			log.Warnf("初始化记忆系统失败: %v", err)
		} else {
			memoryInitialized = true
			log.Info("记忆系统已初始化")
		}
	}

	// 7. 记录用户消息到记忆系统
	if memoryInitialized {
		if err := memory.AddChatMessage("user", question); err != nil {
			log.Debugf("记录用户消息到记忆失败: %v", err)
		}
	}

	// 8. 注册流式输出
	streamCh, unregister := message.RegisterStreamOutput(streamBufferSize)
	defer unregister()

	// 创建流式内容收集器
	collector := &streamCollector{}

	// 9. 发起对话请求
	if err := agent.Ask(ctx, question, nil); err != nil {
		log.Errorf("对话请求失败: %v", err)
		return 1
	}

	// 10. 流式输出处理
	exitCode := processStreamOutputWithCollector(streamCh, collector)

	// 11. 记录助手回复到记忆系统
	if memoryInitialized && exitCode == 0 {
		assistantContent := collector.getContent()
		if assistantContent != "" {
			if err := memory.AddChatMessage("assistant", assistantContent); err != nil {
				log.Debugf("记录助手消息到记忆失败: %v", err)
			}
		}

		// 12. 结束记忆会话
		if sessionID, err := memory.EndChatMemory(); err != nil {
			log.Debugf("结束记忆会话失败: %v", err)
		} else if sessionID != "" {
			log.Infof("会话已保存: %s", sessionID)
		}
	}

	return exitCode
}

// processStreamOutputWithCollector 处理流式输出并收集内容
func processStreamOutputWithCollector(streamCh <-chan *model.StreamChunk, collector *streamCollector) int {
	for chunk := range streamCh {
		if chunk.Err != nil {
			log.Errorf("流式输出错误: %v", chunk.Err)
			// 如果错误消息标记为完成，退出循环
			if chunk.IsDone {
				fmt.Printf("\n❌ 对话出错: %v\n", chunk.Err)
				return 1
			}
			continue
		}

		if chunk.IsDone {
			// 流式输出完成
			log.Info("CLI 单轮对话完成")
			return 0
		}

		// 根据消息类型输出和收集
		switch chunk.MsgType {
		case model.StreamMsgTypeText:
			fmt.Print(chunk.Content)
			collector.append(chunk.Content, model.StreamMsgTypeText)
		case model.StreamMsgTypeReason:
			fmt.Print(chunk.Content)
			collector.append(chunk.Content, model.StreamMsgTypeReason)
		case model.StreamMsgTypeTool:
			fmt.Print(chunk.Content)
			collector.append(chunk.Content, model.StreamMsgTypeTool)
		default:
			// 忽略其他类型
		}
	}

	// channel 关闭，正常退出
	log.Info("流式通道关闭，正常退出")
	return 0
}
