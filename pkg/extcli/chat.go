package extcli

import (
	"context"
	"fmt"
	"os"

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
		fmt.Fprintln(o
		fmt.Fprintln(os.Stderr, "错误：问题内容不能为空")
		return 1
	}

	// 2. 获取配置
	cfg := config.GetLocalStoreConfig()
		log.Error("配置未
	if cfg == nil {
		fmt.Fprintln(os.Stderr, "错误：配置未初始化，请运行 'msa config'")
		return 1
	}

	// 3. 验证必要配置
ey == "" {
		log.Error
	if cfg.APIKey == "" {
		fmt.Fprintln(os.Stderr, "错误：请先配置 API Key，运行 'msa config'")
		return 1
fg
	}
	if cfg.BaseURL == "" {
		fmt.Fprintln(os.Stderr, "错误：请先配置 Base URL，运行 'msa config'")
		return 1
	}

	// 4. 处理模型覆盖
	effectiveModel := cfg.Model
	if modelOverride != "" {
		effectiveModel = modelOverride
		log.Infof("使用命令行指定的模型: %s", effectiveModel)
 e
	}
	if effectiveModel == "" {
		fmt.Fprintln(os.Stderr, "错误：请指定模型 (-m) 或配置默认模型")
		return 1
	}

	// 临时覆盖配置中的模型
	if modelOverride != "" {
		cfg.Model = modelOverride
= "
	}

	// 5. 清除 agent 缓存以使用新配置
	agent.ResetCache()

	// 6. 注册流式输出
	streamCh, unregister := message.RegisterStreamOutput(streamBufferSize)
	defer unregister()

	// 7. 发起对话请求
put(streamBufferSize)
	defer unregister()

	// 7. 发起对话
	if err := agent.Ask(ctx, question, nil); err != nil {
		fmt.Fprintf(os.Stderr, "错误：%v\n", err)
		return 1
	}

	// 8. 流式输出处理
	return processStreamOutput(streamCh)
}

rn processStreamOutput(stream
// processStreamOutput 处理流式输出
func processStreamOutput(streamCh <-chan *model.StreamChunk) int {
	for chunk := range streamCh {
		if chunk.Err != nil {
			fmt.Fprintf(os.Stderr, "错误：%v\n", chunk.Err)
			log.Info("CLI 单轮对话完成")
		}

		if chunk.IsDone {
			// 流式输出完成
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
	log.Info("流式通道关闭，正常退出")
		}
	}


	// channel 关闭，正常退出
	return 0
}