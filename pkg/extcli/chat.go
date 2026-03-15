package extcli

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	"msa/pkg/config"
	"msa/pkg/logic/agent"
	"msa/pkg/logic/message"
	"msa/pkg/model"
)

const (
	streamBufferSize = 100
)

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

	// 6. 注册流式输出
	streamCh, unregister := message.RegisterStreamOutput(streamBufferSize)
	defer unregister()

	// 7. 发起对话请求
	if err := agent.Ask(ctx, question, nil); err != nil {
		log.Errorf("对话请求失败: %v", err)
		return 1
	}

	// 8. 流式输出处理
	return processStreamOutput(streamCh)
}

// processStreamOutput 处理流式输出
func processStreamOutput(streamCh <-chan *model.StreamChunk) int {
	for chunk := range streamCh {
		if chunk.Err != nil {
			log.Errorf("流式输出错误: %v", chunk.Err)
			continue
		}

		if chunk.IsDone {
			// 流式输出完成
			log.Info("CLI 单轮对话完成")
			return 0
		}

		// 根据消息类型输出
		switch chunk.MsgType {
		case model.StreamMsgTypeText:
			fmt.Print(chunk.Content)
		case model.StreamMsgTypeReason:
			fmt.Print(chunk.Content)
		case model.StreamMsgTypeTool:
			fmt.Print(chunk.Content)
		default:
			// 忽略其他类型
		}
	}

	// channel 关闭，正常退出
	log.Info("流式通道关闭，正常退出")
	return 0
}
