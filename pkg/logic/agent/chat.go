package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"msa/pkg/utils"
	"regexp"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"

	"msa/pkg/config"
	msamessage "msa/pkg/logic/message"
	"msa/pkg/logic/skills"
	tools2 "msa/pkg/logic/tools"
	"msa/pkg/model"
	msamodel "msa/pkg/model"

	log "github.com/sirupsen/logrus"
)

var agentCache *react.Agent

// ResetCache 清除 agent 缓存，用于模型切换时重新初始化
func ResetCache() {
	agentCache = nil
}

// maxHistoryRounds 最大保留的对话轮数（一轮 = 一条 User + 一条 Assistant）
const maxHistoryRounds = 10

func GetChatModel(ctx context.Context) (*react.Agent, error) {
	if agentCache != nil {
		return agentCache, nil
	}
	cfg := config.GetLocalStoreConfig()
	if cfg == nil {
		return nil, fmt.Errorf("config is nil, please run `msa config` first")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("openai api key is empty, please run `msa config` first")
	}
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("openai base url is empty, please run `msa config` first")
	}
	if cfg.Model == "" {
		return nil, fmt.Errorf("openai model is empty, please run `/models choose model` first")
	}

	modelConfig := &openai.ChatModelConfig{
		Model:   cfg.Model,
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
	}
	if strings.HasPrefix(cfg.Model, "deepseek-") {
		modelConfig.ExtraFields = map[string]any{
			"enable_thinking": true,
		}
	} else {
		modelConfig.ReasoningEffort = openai.ReasoningEffortLevelMedium
	}

	chatModel, err := openai.NewChatModel(ctx, modelConfig)
	if err != nil {
		return nil, err
	}

	allTools := tools2.GetAllTools()
	log.Infof("tools: %v", len(allTools))
	for _, tool := range allTools {
		tInfo, err := tool.Info(ctx)
		if err != nil {
			log.Errorf("tool info error: %v", err)
			continue
		}
		log.Infof("tool: %v", tInfo.Name)
	}

	tools := compose.ToolsNodeConfig{
		Tools: allTools,
	}

	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel:      chatModel,
		ToolsConfig:           tools,
		MaxStep:               100,
		StreamToolCallChecker: toolCallChecker,
	})
	if err != nil {
		return nil, err
	}
	agentCache = agent
	return agent, nil
}

// dsmlStartTag 是 DeepSeek DSML 格式工具调用的开始标记（｜ 为全角竖线 U+FF5C）
const dsmlStartTag = "<\uff5cDSML\uff5cfunction_calls>"

// dsmlStartPattern 用于检测 DSML 开始标记（大小写不敏感）
var dsmlStartPattern = regexp.MustCompile(`(?i)<[｜|]DSML[｜|]function_calls>`)

// pipeToolCallPattern 用于检测 <|tool_calls_section_begin|> 格式的工具调用
// 同时兼容半角竖线 | (U+007C) 和全角竖线 ｜ (U+FF5C)，部分模型会输出全角竖线
var pipeToolCallPattern = regexp.MustCompile(`<[|｜]tool_calls_section_begin[|｜]>`)

func toolCallChecker(ctx context.Context, sr *schema.StreamReader[*schema.Message]) (bool, error) {
	defer sr.Close()

	streamManager := msamessage.GetStreamManager()

	// 收集所有 chunk，用于检测非标准格式工具调用（DSML / pipe 格式）
	var allChunks []*schema.Message
	var toolCallBuilder strings.Builder // 只收集工具调用标记及其后面的内容
	var contentOnlyLen int              // 已写入 toolCallBuilder 之前的 Content 总长度（用于 prevLen 计算）
	var specialFormatDetected bool      // 是否已检测到非标准格式工具调用开始标记

	// 启动进度提示 goroutine：当 LLM 长时间没有输出时，定期广播"正在思考..."提示
	thinkingDone := make(chan struct{})
	thinkingReset := make(chan struct{}, 1) // 收到 LLM chunk 时重置计时器
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		elapsed := 0
		for {
			select {
			case <-thinkingDone:
				return
			case <-thinkingReset:
				// 收到 LLM 输出，重置计时器
				ticker.Reset(15 * time.Second)
				elapsed = 0
			case <-ticker.C:
				elapsed += 15
				streamManager.Broadcast(&msamodel.StreamChunk{
					Content: fmt.Sprintf("⏳ 正在思考中...(%ds)\n", elapsed),
					MsgType: msamodel.StreamMsgTypeText,
				})
			}
		}
	}()
	defer close(thinkingDone)

	for {
		msg, err := sr.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			// context.Canceled 通常是因为进程退出或 LLM 请求被中断（非用户主动取消）
			// 不广播带 Err 的 IsDone，避免 CLI 以错误码退出
			if errors.Is(err, context.Canceled) {
				log.Warnf("toolCallChecker: sr.Recv() context canceled，广播完成信号")
				streamManager.Broadcast(&msamodel.StreamChunk{IsDone: true})
				return false, nil
			}
			streamManager.Broadcast(&msamodel.StreamChunk{
				Err:    err,
				IsDone: true,
			})
			return false, err
		}

		// 标准工具调用：直接返回 true（不广播 IsDone，工具执行完后还会继续输出）
		if len(msg.ToolCalls) > 0 {
			log.Infof("tool call: %v", msg.ToolCalls)
			return true, nil
		}

		// 收到 LLM chunk，重置进度提示计时器（非阻塞）
		select {
		case thinkingReset <- struct{}{}:
		default:
		}

		// 广播思考内容（同时检测是否包含非标准格式工具调用标记）
		if msg.ReasoningContent != "" {
			if !specialFormatDetected {
				// 检测 ReasoningContent 里是否有工具调用标记
				var tagIdx int = -1
				if loc := pipeToolCallPattern.FindStringIndex(msg.ReasoningContent); loc != nil {
					tagIdx = loc[0]
				} else if loc := dsmlStartPattern.FindStringIndex(msg.ReasoningContent); loc != nil {
					tagIdx = loc[0]
				}
				if tagIdx >= 0 {
					// 标记之前的部分是普通思考内容，广播给用户
					if tagIdx > 0 {
						streamManager.Broadcast(&msamodel.StreamChunk{
							Content: msg.ReasoningContent[:tagIdx],
							MsgType: model.StreamMsgTypeReason,
						})
					}
					// 标记及之后的部分加入 toolCallBuilder 进行解析
					log.Infof("toolCallChecker: 在 ReasoningContent 中检测到工具调用标记，加入 toolCallBuilder")
					toolCallBuilder.WriteString(msg.ReasoningContent[tagIdx:])
					specialFormatDetected = true
					msg.ReasoningContent = ""
				} else {
					// 没有工具调用标记，正常广播
					streamManager.Broadcast(&msamodel.StreamChunk{
						Content: msg.ReasoningContent,
						MsgType: model.StreamMsgTypeReason,
					})
				}
			} else {
				// 已检测到工具调用标记，ReasoningContent 也是工具调用的一部分
				toolCallBuilder.WriteString(msg.ReasoningContent)
				msg.ReasoningContent = ""
			}
		}

		// 处理正文内容
		if msg.Content != "" {
			// 检测是否出现非标准格式工具调用开始标记（DSML 或 pipe 格式）
			if !specialFormatDetected {
				var tagIdx int = -1
				if loc := dsmlStartPattern.FindStringIndex(msg.Content); loc != nil {
					tagIdx = loc[0] + contentOnlyLen
					log.Infof("toolCallChecker: 检测到 DSML 格式标记（Content），tagIdx=%d", tagIdx)
				} else if loc := pipeToolCallPattern.FindStringIndex(msg.Content); loc != nil {
					tagIdx = loc[0] + contentOnlyLen
					log.Infof("toolCallChecker: 检测到 pipe 格式标记（Content），tagIdx=%d", tagIdx)
				}

				if tagIdx >= 0 {
					// 发现工具调用标记，广播标记之前的内容
					relIdx := tagIdx - contentOnlyLen // 标记在当前 chunk 里的相对位置
					if relIdx > 0 {
						streamManager.Broadcast(&msamodel.StreamChunk{
							Content: msg.Content[:relIdx],
							MsgType: model.StreamMsgTypeText,
						})
					}
					// 标记及之后的部分加入 toolCallBuilder
					toolCallBuilder.WriteString(msg.Content[relIdx:])
					contentOnlyLen += len(msg.Content)
					specialFormatDetected = true
					// 清空当前 chunk 的 Content，避免工具调用文本被广播给用户
					// 同时清空之前所有 chunk 的 Content（避免 ConcatMessages 合并脏数据）
					msg.Content = ""
					for _, prev := range allChunks {
						prev.Content = ""
					}
				} else {
					// 还没有工具调用标记，正常广播
					streamManager.Broadcast(&msamodel.StreamChunk{
						Content: msg.Content,
						MsgType: model.StreamMsgTypeText,
					})
					contentOnlyLen += len(msg.Content)
				}
			} else {
				// 已检测到工具调用标记，Content 也是工具调用的一部分
				toolCallBuilder.WriteString(msg.Content)
				msg.Content = ""
			}
		}

		allChunks = append(allChunks, msg)
	}

	// 检查完整内容是否包含非标准格式工具调用（DSML 或 pipe 格式）
	if specialFormatDetected {
		fullContent := toolCallBuilder.String()
		toolCalls := parseToolCalls(fullContent)
		if len(toolCalls) > 0 {
			log.Infof("检测到非标准格式工具调用，解析结果: %v", toolCalls)
			// 将解析后的 ToolCalls 注入到最后一个 chunk 中
			// 由于 tee 流共享同一个 *schema.Message 指针，修改会影响 eino 内部的处理
			// 注意：不广播 IsDone，工具执行完后还会继续输出
			if len(allChunks) > 0 {
				lastChunk := allChunks[len(allChunks)-1]
				lastChunk.ToolCalls = toolCalls
				lastChunk.Content = ""
			}
			return true, nil
		}
		// specialFormatDetected=true 但解析失败（如模型只输出了标记头，JSON 被截断）
		// 这种情况说明模型意图调用工具但输出异常，不应结束对话
		// 返回 true 让 eino 继续下一轮（eino 会把当前消息作为 assistant 消息加入历史，LLM 会重试）
		log.Warnf("toolCallChecker: specialFormatDetected=true 但解析失败，toolCallBuilder 内容: %q，返回 true 让 eino 继续", func() string {
			if len(fullContent) > 200 {
				return fullContent[:200]
			}
			return fullContent
		}())
		// 注入一个空的 ToolCall 占位，防止 eino 认为没有工具调用而终止
		// 实际上这里直接广播 IsDone=false 并返回 false，让 eino 重新发起一轮
		// 但 eino 的 StreamToolCallChecker 返回 false 时会结束 agent 循环
		// 所以这里广播 IsDone: true，但不算错误，让用户看到当前输出后重新发起
		streamManager.Broadcast(&msamodel.StreamChunk{IsDone: true})
		return false, nil
	}

	// 没有工具调用，广播完成信号
	log.Infof("toolCallChecker: 没有工具调用，广播完成信号")
	streamManager.Broadcast(&msamodel.StreamChunk{IsDone: true})
	return false, nil
}

// parseToolCalls 解析非标准格式的工具调用，支持以下两种格式：
//
// 1. DeepSeek DSML 格式：
//
//	<｜DSML｜function_calls>
//	<｜DSML｜invoke name="tool_name">
//	<｜DSML｜parameter name="param_name" string="true">value</｜DSML｜parameter>
//	</｜DSML｜invoke>
//	</｜DSML｜function_calls>
//
// 2. Pipe 格式：
//
//	<|tool_calls_section_begin|>[{"name":"tool_name","parameters":{"key":"value"}}]
func parseToolCalls(content string) []schema.ToolCall {
	// 优先尝试解析 pipe 格式
	if pipeToolCallPattern.MatchString(content) {
		if calls := parsePipeToolCalls(content); len(calls) > 0 {
			return calls
		}
	}
	// 尝试解析 DSML 格式
	return parseDSMLToolCalls(content)
}

func parseDSMLToolCalls(content string) []schema.ToolCall {
	// 匹配 invoke 块：<｜DSML｜invoke name="...">...</｜DSML｜invoke>
	invokePattern := regexp.MustCompile(`(?i)<[｜|]DSML[｜|]invoke\s+name="([^"]+)"[^>]*>([\s\S]*?)</[｜|]DSML[｜|]invoke>`)
	// 匹配 parameter 块：<｜DSML｜parameter name="...">value</｜DSML｜parameter>
	paramPattern := regexp.MustCompile(`(?i)<[｜|]DSML[｜|]parameter\s+name="([^"]+)"[^>]*>([\s\S]*?)</[｜|]DSML[｜|]parameter>`)

	var toolCalls []schema.ToolCall
	invokeMatches := invokePattern.FindAllStringSubmatch(content, -1)

	for i, invokeMatch := range invokeMatches {
		if len(invokeMatch) < 3 {
			continue
		}
		toolName := strings.TrimSpace(invokeMatch[1])
		invokeBody := invokeMatch[2]

		// 解析参数
		params := make(map[string]interface{})
		paramMatches := paramPattern.FindAllStringSubmatch(invokeBody, -1)
		for _, paramMatch := range paramMatches {
			if len(paramMatch) < 3 {
				continue
			}
			paramName := strings.TrimSpace(paramMatch[1])
			paramValue := strings.TrimSpace(paramMatch[2])
			params[paramName] = paramValue
		}

		// 序列化参数为 JSON
		argsJSON, err := json.Marshal(params)
		if err != nil {
			log.Warnf("parseDSMLToolCalls: 序列化参数失败: %v", err)
			continue
		}

		// 生成唯一 ID
		idx := i
		toolCalls = append(toolCalls, schema.ToolCall{
			Index: &idx,
			ID:    fmt.Sprintf("dsml_call_%d", i),
			Type:  "function",
			Function: schema.FunctionCall{
				Name:      toolName,
				Arguments: string(argsJSON),
			},
		})
	}

	return toolCalls
}

// parsePipeToolCalls 解析 <|tool_calls_section_begin|> 格式的工具调用
// 格式示例：
//
//	<|tool_calls_section_begin|>[{"name":"tool_name","parameters":{"key":"value"}}]
func parsePipeToolCalls(content string) []schema.ToolCall {
	// 提取 <|tool_calls_section_begin|> 之后的 JSON 数组
	// 同时兼容半角竖线 | (U+007C) 和全角竖线 ｜ (U+FF5C)
	pipeExtractPattern := regexp.MustCompile(`<[|｜]tool_calls_section_begin[|｜]>(\[.*?\])`)
	matches := pipeExtractPattern.FindStringSubmatch(content)
	if len(matches) < 2 {
		// 尝试贪婪匹配（JSON 可能跨行）
		pipeExtractPatternMultiline := regexp.MustCompile(`(?s)<[|｜]tool_calls_section_begin[|｜]>(\[.*?\])`)
		matches = pipeExtractPatternMultiline.FindStringSubmatch(content)
		if len(matches) < 2 {
			log.Warnf("parsePipeToolCalls: 未找到 JSON 数组")
			return nil
		}
	}

	jsonStr := strings.TrimSpace(matches[1])

	// 解析 JSON 数组
	type pipeToolCall struct {
		Name       string                 `json:"name"`
		Parameters map[string]interface{} `json:"parameters"`
	}
	var rawCalls []pipeToolCall
	if err := json.Unmarshal([]byte(jsonStr), &rawCalls); err != nil {
		log.Warnf("parsePipeToolCalls: JSON 解析失败: %v, json: %s", err, jsonStr)
		return nil
	}

	var toolCalls []schema.ToolCall
	for i, raw := range rawCalls {
		argsJSON, err := json.Marshal(raw.Parameters)
		if err != nil {
			log.Warnf("parsePipeToolCalls: 序列化参数失败: %v", err)
			continue
		}
		idx := i
		toolCalls = append(toolCalls, schema.ToolCall{
			Index: &idx,
			ID:    fmt.Sprintf("pipe_call_%d", i),
			Type:  "function",
			Function: schema.FunctionCall{
				Name:      strings.TrimSpace(raw.Name),
				Arguments: string(argsJSON),
			},
		})
	}
	return toolCalls
}

// Ask 聊天（异步处理流式数据，通过 StreamOutputManager 广播）
// 流程：
//  1. 初始化 skills，确保 get_skill_content tool 可正常使用
//  2. 根据 skill 元数据构建 system prompt，通过 BuildQueryMessages 拼接消息
//  3. 异步调用 agent.Stream 获取流式结果
func Ask(ctx context.Context, messages string, history []model.Message) error {
	// ===== 第一步：初始化 skills =====
	skillManager := skills.GetManager()
	if err := skillManager.Initialize(); err != nil {
		log.Warnf("skills initialize warning: %v", err)
	}

	// ===== 第二步：构建消息 =====
	filteredHistory := filterAndTrimHistory(history, maxHistoryRounds)
	historyMessages := convertToSchemaMessages(filteredHistory)

	now := time.Now()
	queryMsg, err := BuildQueryMessages(ctx, messages, historyMessages, map[string]any{
		"role":    "专业股票分析助手",
		"style":   "理性、专业、客观且严谨",
		"time":    now.Format("2006-01-02 15:04:05"),
		"weekday": now.Format("2006年01月02日 星期Monday"),
	})
	if err != nil {
		log.Errorf("Failed to build query messages: %v", err)
		return err
	}

	log.Infof("queryMsg count: %d", len(queryMsg))
	log.Infof("query  msg\n %s", utils.ToJSONString(queryMsg))

	// ===== 第三步：获取 Agent 并异步调用 Stream =====
	agent, err := GetChatModel(ctx)
	if err != nil {
		return err
	}

	go func() {
		sr, err := agent.Stream(ctx, queryMsg)
		if err != nil {
			log.Errorf("Failed to stream: %v", err)
			msamessage.GetStreamManager().Broadcast(&msamodel.StreamChunk{
				Err:    err,
				IsDone: true,
			})
			return
		}
		// 必须消费（drain）agent.Stream 返回的 StreamReader，否则 eino 内部 goroutine
		// 会阻塞在写入该流上，导致 toolCallChecker 无法正常完成，工具执行后对话中断。
		// 实际的流式内容已由 toolCallChecker 通过 StreamOutputManager 广播给订阅者，
		// 此处只需将流排空即可。
		defer sr.Close()
		for {
			_, recvErr := sr.Recv()
			if recvErr != nil {
				if !errors.Is(recvErr, io.EOF) {
					log.Errorf("stream recv error: %v", recvErr)
				}
				break
			}
		}
	}()

	return nil
}

// convertToSchemaMessages 将 model.Message 切片转为 eino schema.Message 切片
func convertToSchemaMessages(history []model.Message) []*schema.Message {
	msgs := make([]*schema.Message, 0, len(history))
	for _, msg := range history {
		var role schema.RoleType
		switch msg.Role {
		case model.RoleUser:
			role = schema.User
		case model.RoleAssistant:
			role = schema.Assistant
		default:
			continue
		}
		msgs = append(msgs, &schema.Message{
			Role:    role,
			Content: msg.Content,
		})
	}
	return msgs
}

// filterAndTrimHistory 过滤历史消息，只保留 User 和 Assistant 角色，并限制最大轮数
func filterAndTrimHistory(history []model.Message, maxRounds int) []model.Message {
	filtered := make([]model.Message, 0, len(history))
	for _, msg := range history {
		if msg.Role == model.RoleUser || msg.Role == model.RoleAssistant {
			filtered = append(filtered, msg)
		}
	}

	maxMessages := maxRounds * 2
	if len(filtered) > maxMessages {
		filtered = filtered[len(filtered)-maxMessages:]
	}

	return filtered
}
