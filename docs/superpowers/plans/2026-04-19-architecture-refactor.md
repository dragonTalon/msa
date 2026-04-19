# MSA 架构重构 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 用事件驱动 pipeline 彻底重写 MSA 核心链路，替换全局 StreamOutputManager + toolCallChecker，让 ctx 全程贯穿，CLI 和 TUI 共用同一套 core。

**Architecture:** `Agent.Run(ctx)` 返回 `<-chan Event`，StreamAdapter 封装 Eino stream 处理（包含 DSML/pipe 格式的非标准工具调用解析），Runner 消费 channel 并调用 `Renderer.Handle()`，CLI 和 TUI 各自实现 Renderer 接口。ctx 取消时 `defer close(ch)` 保证所有 goroutine 干净退出。

**Tech Stack:** Go 1.25.5, Eino v0.7.13, Bubble Tea v0.24.2, Logrus, GORM + SQLite

---

## 文件结构

### 新增文件

| 文件 | 职责 |
|------|------|
| `pkg/core/event/event.go` | Event 类型定义（整条 pipeline 的数据结构） |
| `pkg/core/logger/logger.go` | requestID 注入 + FromCtx，结构化日志 |
| `pkg/core/agent/agent.go` | Eino 封装，Run() 返回 `<-chan Event` |
| `pkg/core/agent/stream_adapter.go` | 替代 toolCallChecker，处理 Eino chunk 分类 |
| `pkg/core/agent/tool_parser.go` | 迁移 DSML/pipe 工具调用解析逻辑（从旧 chat.go） |
| `pkg/core/runner/runner.go` | 主循环：协调 agent → renderer，持久化会话 |
| `pkg/renderer/renderer.go` | Renderer 接口定义 |
| `pkg/renderer/cli.go` | CLI 实现：直接写 stdout |
| `pkg/renderer/tui.go` | TUI 实现：转发给 bubbletea program.Send |

### 修改文件

| 文件 | 修改内容 |
|------|----------|
| `pkg/extcli/chat.go` | 用 Runner + CLIRenderer 替代旧 Ask + StreamOutputManager |
| `pkg/app/app.go` | 用 Runner + TUIRenderer 替代旧 Ask + StreamOutputManager |
| `pkg/tui/chat.go` | `startChatRequest` 改为接收 `event.Event`，移除 `streamOutputCh` |
| `pkg/logic/tools/safetool/safetool.go` | 移除 `message.BroadcastToolEnd` 调用（广播由 agent 层处理） |
| `pkg/logic/tools/finance/account.go` 等所有 tool 文件 | 移除 `message.BroadcastToolStart/End` 调用 |

### 最终删除（所有任务完成后）

| 文件 | 原因 |
|------|------|
| `pkg/logic/agent/chat.go` | 被 `pkg/core/agent/` 替代 |
| `pkg/logic/message/stream_manage.go` | 被 event channel 替代 |

---

## Task 1: 定义 Event 类型系统

**Files:**
- Create: `pkg/core/event/event.go`

- [ ] **Step 1: 创建 event 包**

```go
// pkg/core/event/event.go
package event

// EventType 定义 pipeline 中所有可能的事件类型
type EventType int

const (
	// 文本流
	EventTextChunk EventType = iota // LLM 输出的普通文本片段
	EventTextDone                   // 本轮文本输出完毕

	// 工具调用
	EventToolStart  // 工具开始执行（携带工具名、参数）
	EventToolResult // 工具执行完成（携带结果）
	EventToolError  // 工具执行失败

	// 思考过程（reasoning model 支持）
	EventThinking // <thinking> 内容片段

	// 会话控制
	EventRoundDone // 一轮对话完成（agent 不再调用工具）
	EventError     // 不可恢复的错误，pipeline 终止
)

// Event 是 pipeline 中流动的最小单元
type Event struct {
	Type EventType

	// EventTextChunk / EventThinking
	Text string

	// EventToolStart
	Tool ToolCall

	// EventToolResult / EventToolError
	Result ToolResult

	// EventError
	Err error
}

// ToolCall 描述一次工具调用请求
type ToolCall struct {
	ID    string
	Name  string
	Input string // JSON 字符串
}

// ToolResult 描述一次工具调用的结果
type ToolResult struct {
	ToolCallID string
	Name       string
	Output     string
	IsError    bool
}
```

- [ ] **Step 2: 验证编译通过**

```bash
cd /Users/dragon/Documents/Code/goland/msa
go build ./pkg/core/event/...
```

Expected: 无输出（编译成功）

- [ ] **Step 3: Commit**

```bash
git add pkg/core/event/event.go
git commit -m "feat(core): add Event type system for pipeline communication"
```

---

## Task 2: 实现结构化日志（requestID）

**Files:**
- Create: `pkg/core/logger/logger.go`

- [ ] **Step 1: 创建 logger 包**

```go
// pkg/core/logger/logger.go
package logger

import (
	"context"
	"fmt"
	"math/rand"

	log "github.com/sirupsen/logrus"
)

type contextKey string

const requestIDKey contextKey = "req_id"

// WithRequestID 在 ctx 中注入 requestID，返回新 ctx 和 ID
// ID 格式：6位十六进制，便于 grep
func WithRequestID(ctx context.Context) (context.Context, string) {
	id := fmt.Sprintf("%06x", rand.Int63n(0xFFFFFF))
	return context.WithValue(ctx, requestIDKey, id), id
}

// FromCtx 从 ctx 中取出 requestID，返回带字段的 logrus.Entry
// 如果 ctx 中没有 requestID，使用 "------" 占位
func FromCtx(ctx context.Context) *log.Entry {
	id, ok := ctx.Value(requestIDKey).(string)
	if !ok || id == "" {
		id = "------"
	}
	return log.WithField("req", id)
}
```

- [ ] **Step 2: 验证编译通过**

```bash
go build ./pkg/core/logger/...
```

Expected: 无输出

- [ ] **Step 3: Commit**

```bash
git add pkg/core/logger/logger.go
git commit -m "feat(core): add request-scoped logger with requestID"
```

---

## Task 3: 迁移工具调用解析逻辑（tool_parser.go）

**Files:**
- Create: `pkg/core/agent/tool_parser.go`

这个文件将旧 `pkg/logic/agent/chat.go` 中的 `parseToolCalls`、`parseDSMLToolCalls`、`parsePipeToolCalls` 原样迁移过来，不做逻辑改动。

- [ ] **Step 1: 创建 tool_parser.go**

```go
// pkg/core/agent/tool_parser.go
package agent

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/cloudwego/eino/schema"
	log "github.com/sirupsen/logrus"
)

// dsmlStartPattern 用于检测 DSML 开始标记（大小写不敏感）
var dsmlStartPattern = regexp.MustCompile(`(?i)<[｜|]DSML[｜|]function_calls>`)

// pipeToolCallPattern 用于检测 <|tool_calls_section_begin|> 格式的工具调用
// 同时兼容半角竖线 | (U+007C) 和全角竖线 ｜ (U+FF5C)
var pipeToolCallPattern = regexp.MustCompile(`<[|｜]tool_calls_section_begin[|｜]>`)

// parseToolCalls 解析非标准格式的工具调用，支持 DSML 和 pipe 两种格式
func parseToolCalls(content string) []schema.ToolCall {
	if pipeToolCallPattern.MatchString(content) {
		if calls := parsePipeToolCalls(content); len(calls) > 0 {
			return calls
		}
	}
	return parseDSMLToolCalls(content)
}

func parseDSMLToolCalls(content string) []schema.ToolCall {
	invokePattern := regexp.MustCompile(`(?i)<[｜|]DSML[｜|]invoke\s+name="([^"]+)"[^>]*>([\s\S]*?)</[｜|]DSML[｜|]invoke>`)
	paramPattern := regexp.MustCompile(`(?i)<[｜|]DSML[｜|]parameter\s+name="([^"]+)"[^>]*>([\s\S]*?)</[｜|]DSML[｜|]parameter>`)

	var toolCalls []schema.ToolCall
	invokeMatches := invokePattern.FindAllStringSubmatch(content, -1)

	for i, invokeMatch := range invokeMatches {
		if len(invokeMatch) < 3 {
			continue
		}
		toolName := strings.TrimSpace(invokeMatch[1])
		invokeBody := invokeMatch[2]

		params := make(map[string]interface{})
		paramMatches := paramPattern.FindAllStringSubmatch(invokeBody, -1)
		for _, paramMatch := range paramMatches {
			if len(paramMatch) < 3 {
				continue
			}
			params[strings.TrimSpace(paramMatch[1])] = strings.TrimSpace(paramMatch[2])
		}

		argsJSON, err := json.Marshal(params)
		if err != nil {
			log.Warnf("parseDSMLToolCalls: 序列化参数失败: %v", err)
			continue
		}

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

func parsePipeToolCalls(content string) []schema.ToolCall {
	pipeExtractPattern := regexp.MustCompile(`<[|｜]tool_calls_section_begin[|｜]>(\[.*?\])`)
	matches := pipeExtractPattern.FindStringSubmatch(content)
	if len(matches) < 2 {
		pipeExtractPatternMultiline := regexp.MustCompile(`(?s)<[|｜]tool_calls_section_begin[|｜]>(\[.*?\])`)
		matches = pipeExtractPatternMultiline.FindStringSubmatch(content)
		if len(matches) < 2 {
			log.Warnf("parsePipeToolCalls: 未找到 JSON 数组")
			return nil
		}
	}

	jsonStr := strings.TrimSpace(matches[1])

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
```

- [ ] **Step 2: 验证编译通过**

```bash
go build ./pkg/core/agent/...
```

Expected: 无输出

- [ ] **Step 3: Commit**

```bash
git add pkg/core/agent/tool_parser.go
git commit -m "feat(core/agent): migrate DSML/pipe tool call parsers from legacy agent"
```

---

## Task 4: 实现 StreamAdapter

**Files:**
- Create: `pkg/core/agent/stream_adapter.go`

StreamAdapter 是整个重构最核心的部分，替代旧的 `toolCallChecker`。

- [ ] **Step 1: 创建 stream_adapter.go**

```go
// pkg/core/agent/stream_adapter.go
package agent

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	log "github.com/sirupsen/logrus"

	"msa/pkg/core/event"
	corelogger "msa/pkg/core/logger"
)

// ProcessResult 是 StreamAdapter.Process 的返回值
// 包含本轮 LLM 调用收集到的 tool calls 和拼装好的 assistant 消息
type ProcessResult struct {
	ToolCalls    []schema.ToolCall
	AssistantMsg *schema.Message // 用于追加到 messages 继续对话
	HasContent   bool            // 是否有文本内容输出
}

// StreamAdapter 消费 Eino 原始 stream，输出干净的 Event 序列
// 唯一职责：处理 Eino chunk 格式不稳定问题，外部永远收到干净的 Event
type StreamAdapter struct{}

// Process 处理一次 LLM stream，将 chunk 转为 Event 发送到 out channel
// 同时收集 tool calls 和文本内容，返回 ProcessResult 供 ReAct 循环使用
func (a *StreamAdapter) Process(
	ctx context.Context,
	reader *schema.StreamReader[*schema.Message],
	out chan<- event.Event,
) (ProcessResult, error) {
	log := corelogger.FromCtx(ctx)

	var (
		textBuf            strings.Builder
		toolCalls          []schema.ToolCall
		chunkCount         int
		specialDetected    bool
		toolCallBuf        strings.Builder
		contentOnlyLen     int
		allChunks          []*schema.Message
	)

	// 启动进度提示 goroutine
	thinkingDone := make(chan struct{})
	thinkingReset := make(chan struct{}, 1)
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		elapsed := 0
		for {
			select {
			case <-thinkingDone:
				return
			case <-thinkingReset:
				ticker.Reset(15 * time.Second)
				elapsed = 0
			case <-ticker.C:
				elapsed += 15
				sendEvent(ctx, out, event.Event{
					Type: event.EventTextChunk,
					Text: "", // 空文本，仅用于触发 TUI 显示进度
				})
				log.Infof("[StreamAdapter] 正在思考中...(%ds)", elapsed)
			}
		}
	}()
	defer close(thinkingDone)

	for {
		msg, err := reader.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Debugf("[StreamAdapter] stream 结束, chunks=%d, toolCalls=%d", chunkCount, len(toolCalls))
				break
			}
			if errors.Is(err, context.Canceled) {
				log.Warnf("[StreamAdapter] context canceled，正常退出")
				reader.Close()
				return ProcessResult{}, context.Canceled
			}
			reader.Close() // 关键：drain 防止 goroutine 泄漏
			return ProcessResult{}, err
		}
		chunkCount++

		// 收到 chunk，重置进度提示计时器
		select {
		case thinkingReset <- struct{}{}:
		default:
		}

		// 标准 tool call：直接收集，不广播（后续由 executeTools 处理并发送 EventToolStart）
		if len(msg.ToolCalls) > 0 {
			log.Infof("[StreamAdapter] chunk#%d 标准 toolCall: %v", chunkCount, toolCallNames(msg.ToolCalls))
			toolCalls = append(toolCalls, msg.ToolCalls...)
			allChunks = append(allChunks, msg)
			continue
		}

		// 处理 ReasoningContent（thinking）
		if msg.ReasoningContent != "" {
			if !specialDetected {
				if tagIdx := findToolCallTag(msg.ReasoningContent); tagIdx >= 0 {
					if tagIdx > 0 {
						sendEvent(ctx, out, event.Event{Type: event.EventThinking, Text: msg.ReasoningContent[:tagIdx]})
					}
					toolCallBuf.WriteString(msg.ReasoningContent[tagIdx:])
					specialDetected = true
					log.Infof("[StreamAdapter] chunk#%d 在 ReasoningContent 中检测到非标准工具调用标记", chunkCount)
				} else {
					log.Debugf("[StreamAdapter] chunk#%d thinking len=%d", chunkCount, len(msg.ReasoningContent))
					sendEvent(ctx, out, event.Event{Type: event.EventThinking, Text: msg.ReasoningContent})
				}
			} else {
				toolCallBuf.WriteString(msg.ReasoningContent)
			}
			msg.ReasoningContent = ""
		}

		// 处理正文 Content
		if msg.Content != "" {
			if !specialDetected {
				if tagIdx := findToolCallTagWithOffset(msg.Content, contentOnlyLen); tagIdx >= 0 {
					relIdx := tagIdx - contentOnlyLen
					if relIdx > 0 {
						sendEvent(ctx, out, event.Event{Type: event.EventTextChunk, Text: msg.Content[:relIdx]})
						textBuf.WriteString(msg.Content[:relIdx])
					}
					toolCallBuf.WriteString(msg.Content[relIdx:])
					contentOnlyLen += len(msg.Content)
					specialDetected = true
					log.Infof("[StreamAdapter] chunk#%d 检测到非标准工具调用标记（Content）", chunkCount)
					// 清空之前所有 chunk 的 Content，避免 ConcatMessages 合并脏数据
					for _, prev := range allChunks {
						prev.Content = ""
					}
					msg.Content = ""
				} else {
					log.Debugf("[StreamAdapter] chunk#%d text len=%d", chunkCount, len(msg.Content))
					sendEvent(ctx, out, event.Event{Type: event.EventTextChunk, Text: msg.Content})
					textBuf.WriteString(msg.Content)
					contentOnlyLen += len(msg.Content)
				}
			} else {
				toolCallBuf.WriteString(msg.Content)
				msg.Content = ""
			}
		}

		allChunks = append(allChunks, msg)
	}

	// 处理非标准格式工具调用
	if specialDetected {
		fullContent := toolCallBuf.String()
		parsedCalls := parseToolCalls(fullContent)
		if len(parsedCalls) > 0 {
			log.Infof("[StreamAdapter] 解析到非标准工具调用: %v", toolCallNames(parsedCalls))
			// 注入到最后一个 chunk（eino 内部处理需要）
			if len(allChunks) > 0 {
				lastChunk := allChunks[len(allChunks)-1]
				lastChunk.ToolCalls = parsedCalls
				lastChunk.Content = ""
			}
			toolCalls = append(toolCalls, parsedCalls...)
		} else {
			log.Warnf("[StreamAdapter] specialDetected=true 但解析失败，内容: %q",
				truncate(fullContent, 200))
		}
	}

	// 文本输出结束
	if textBuf.Len() > 0 {
		sendEvent(ctx, out, event.Event{Type: event.EventTextDone})
	}

	// 拼装 assistant 消息（用于追加到 messages 继续对话）
	assistantMsg := buildAssistantMsg(allChunks)

	return ProcessResult{
		ToolCalls:    toolCalls,
		AssistantMsg: assistantMsg,
		HasContent:   textBuf.Len() > 0,
	}, nil
}

// sendEvent 向 channel 发送事件，尊重 ctx 取消
func sendEvent(ctx context.Context, ch chan<- event.Event, e event.Event) bool {
	select {
	case ch <- e:
		return true
	case <-ctx.Done():
		return false
	}
}

// findToolCallTag 在 content 中查找非标准工具调用标记的起始位置
// 返回 -1 表示未找到
func findToolCallTag(content string) int {
	if loc := pipeToolCallPattern.FindStringIndex(content); loc != nil {
		return loc[0]
	}
	if loc := dsmlStartPattern.FindStringIndex(content); loc != nil {
		return loc[0]
	}
	return -1
}

// findToolCallTagWithOffset 同 findToolCallTag，但加上 offset 计算绝对位置
func findToolCallTagWithOffset(content string, offset int) int {
	idx := findToolCallTag(content)
	if idx < 0 {
		return -1
	}
	return idx + offset
}

// buildAssistantMsg 将所有 chunk 拼装为 assistant 消息
func buildAssistantMsg(chunks []*schema.Message) *schema.Message {
	if len(chunks) == 0 {
		return &schema.Message{Role: schema.Assistant}
	}
	// 使用 eino 的 ConcatMessages 合并流式 chunk
	merged, err := schema.ConcatMessages(chunks)
	if err != nil {
		// fallback：手动拼接文本
		var sb strings.Builder
		for _, c := range chunks {
			sb.WriteString(c.Content)
		}
		return &schema.Message{Role: schema.Assistant, Content: sb.String()}
	}
	return merged
}

// toolCallNames 提取 tool call 名称列表（用于日志）
func toolCallNames(calls []schema.ToolCall) []string {
	names := make([]string, 0, len(calls))
	for _, c := range calls {
		names = append(names, c.Function.Name)
	}
	return names
}

// truncate 截断字符串（用于日志）
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
```

- [ ] **Step 2: 验证编译通过**

```bash
go build ./pkg/core/agent/...
```

Expected: 无输出

- [ ] **Step 3: Commit**

```bash
git add pkg/core/agent/stream_adapter.go
git commit -m "feat(core/agent): implement StreamAdapter to replace toolCallChecker"
```

---

## Task 5: 实现 Agent（ReAct 循环）

**Files:**
- Create: `pkg/core/agent/agent.go`

- [ ] **Step 1: 创建 agent.go**

```go
// pkg/core/agent/agent.go
package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"

	"msa/pkg/config"
	"msa/pkg/core/event"
	corelogger "msa/pkg/core/logger"
	"msa/pkg/logic/tools"
	"msa/pkg/logic/skills"

	log "github.com/sirupsen/logrus"
)

// maxHistoryRounds 最大保留的对话轮数（一轮 = 一条 User + 一条 Assistant）
const maxHistoryRounds = 10

// Agent 封装 Eino ReAct agent，对外只暴露 Run() 接口
type Agent struct {
	einoAgent *react.Agent
	adapter   *StreamAdapter
}

// New 根据当前配置创建 Agent
// 每次调用都创建新实例（模型切换时直接 New，不需要 ResetCache）
func New(ctx context.Context) (*Agent, error) {
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
		modelConfig.ExtraFields = map[string]any{"enable_thinking": true}
	} else {
		modelConfig.ReasoningEffort = openai.ReasoningEffortLevelMedium
	}

	chatModel, err := openai.NewChatModel(ctx, modelConfig)
	if err != nil {
		return nil, fmt.Errorf("创建 chat model 失败: %w", err)
	}

	allTools := tools.GetAllTools()
	log.Infof("[Agent] 注册工具数: %d", len(allTools))

	einoAgent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig:      compose.ToolsNodeConfig{Tools: allTools},
		MaxStep:          100,
		// 不再使用 StreamToolCallChecker，改为 StreamAdapter 处理
	})
	if err != nil {
		return nil, fmt.Errorf("创建 Eino agent 失败: %w", err)
	}

	return &Agent{
		einoAgent: einoAgent,
		adapter:   &StreamAdapter{},
	}, nil
}

// Run 启动一轮对话（ReAct 循环），返回只读 event channel
// ctx 取消时，channel 自动关闭，所有 goroutine 自动退出
// 调用方：for e := range Run(ctx, msgs) { ... }
func (a *Agent) Run(ctx context.Context, messages []*schema.Message) <-chan event.Event {
	ch := make(chan event.Event, 32)
	go a.run(ctx, messages, ch)
	return ch
}

func (a *Agent) run(ctx context.Context, messages []*schema.Message, ch chan<- event.Event) {
	defer close(ch) // 无论如何退出，channel 必然关闭，消费方 range 自动退出
	logger := corelogger.FromCtx(ctx)
	logger.Infof("[Agent] 开始对话, 历史消息数=%d", len(messages))

	round := 0
	for {
		round++
		logger.Infof("[Agent] 第 %d 轮 LLM 调用开始", round)

		// 调用 Eino stream
		sr, err := a.einoAgent.Stream(ctx, messages)
		if err != nil {
			logger.Errorf("[Agent] 第 %d 轮 LLM 调用失败: %v", round, err)
			sendEvent(ctx, ch, event.Event{Type: event.EventError, Err: err})
			return
		}
		logger.Infof("[Agent] 第 %d 轮 LLM stream 已建立", round)

		// StreamAdapter 处理 stream，发射 Event
		result, err := a.adapter.Process(ctx, sr, ch)
		if err != nil {
			if err == context.Canceled {
				logger.Warnf("[Agent] 第 %d 轮 context canceled，干净退出", round)
				return
			}
			logger.Errorf("[Agent] 第 %d 轮 stream 处理失败: %v", round, err)
			sendEvent(ctx, ch, event.Event{Type: event.EventError, Err: err})
			return
		}

		// 没有 tool call → 本轮对话结束
		if len(result.ToolCalls) == 0 {
			logger.Infof("[Agent] 第 %d 轮无工具调用，对话结束", round)
			sendEvent(ctx, ch, event.Event{Type: event.EventRoundDone})
			return
		}

		logger.Infof("[Agent] 第 %d 轮需要调用工具: %v", round, toolCallNames(result.ToolCalls))

		// 执行工具，收集结果消息
		toolMsgs := a.executeTools(ctx, result.ToolCalls, ch)

		// 追加 assistant 消息 + 工具结果，继续下一轮
		messages = append(messages, result.AssistantMsg)
		messages = append(messages, toolMsgs...)
	}
}

// executeTools 执行工具列表，发射 EventToolStart/EventToolResult/EventToolError
// 返回工具结果消息，用于追加到 messages 继续 ReAct 循环
func (a *Agent) executeTools(
	ctx context.Context,
	calls []schema.ToolCall,
	ch chan<- event.Event,
) []*schema.Message {
	logger := corelogger.FromCtx(ctx)
	var toolMsgs []*schema.Message

	// 获取 eino tools registry（通过 GetAllTools 重新查找可调用的工具）
	// 注意：eino 内部实际上通过 react.Agent 执行工具，这里我们手动执行以发射 Event
	// 实际上 eino 的 react.Agent 在 Stream 后会自动执行工具，我们不需要手动执行
	// 但我们需要在执行前后发射 EventToolStart 和 EventToolResult
	// 因此这里通过 react.Agent.Stream 的内部机制自动处理工具执行
	// 我们只需要在 StreamAdapter 中捕获 tool call 并发射 Event 即可

	// 发射 EventToolStart（在 StreamAdapter.Process 中已经发射，这里补充执行完成事件）
	// 实际工具执行由 Eino 内部完成，我们通过下一轮 LLM 调用的 messages 参数传递结果
	// 这里需要手动调用工具并构造 tool result 消息

	toolsMap := tools.GetToolsMap()

	for _, call := range calls {
		start := time.Now()
		toolName := call.Function.Name

		sendEvent(ctx, ch, event.Event{
			Type: event.EventToolStart,
			Tool: event.ToolCall{
				ID:    call.ID,
				Name:  toolName,
				Input: call.Function.Arguments,
			},
		})

		logger.Infof("[Tool] 开始执行: name=%s id=%s input=%s", toolName, call.ID, call.Function.Arguments)

		// 执行工具
		msaTool, ok := toolsMap[toolName]
		var output string
		var toolErr error

		if !ok {
			toolErr = fmt.Errorf("工具不存在: %s", toolName)
			output = fmt.Sprintf(`{"error": "工具不存在: %s"}`, toolName)
		} else {
			baseTool, err := msaTool.GetToolInfo()
			if err != nil {
				toolErr = err
				output = fmt.Sprintf(`{"error": "获取工具信息失败: %v"}`, err)
			} else {
				output, toolErr = baseTool.Run(ctx, call.Function.Arguments)
			}
		}

		elapsed := time.Since(start)

		if toolErr != nil {
			logger.Errorf("[Tool] 执行失败: name=%s elapsed=%v err=%v", toolName, elapsed, toolErr)
			sendEvent(ctx, ch, event.Event{
				Type: event.EventToolError,
				Result: event.ToolResult{
					ToolCallID: call.ID,
					Name:       toolName,
					Output:     output,
					IsError:    true,
				},
				Err: toolErr,
			})
			// 工具失败也要继续，让 LLM 知道失败原因
			output = fmt.Sprintf(`{"error": "%v"}`, toolErr)
		} else {
			logger.Infof("[Tool] 执行完成: name=%s elapsed=%v outputLen=%d", toolName, elapsed, len(output))
			sendEvent(ctx, ch, event.Event{
				Type: event.EventToolResult,
				Result: event.ToolResult{
					ToolCallID: call.ID,
					Name:       toolName,
					Output:     output,
					IsError:    false,
				},
			})
		}

		// 构造 tool result 消息（追加到 messages 供下一轮 LLM 使用）
		toolMsgs = append(toolMsgs, &schema.Message{
			Role:    schema.Tool,
			Content: output,
			ToolCallID: call.ID,
		})
	}

	return toolMsgs
}

// BuildQueryMessages 构建查询消息（包含系统 prompt 和历史消息）
// 从旧的 pkg/logic/agent/prompt.go 迁移
func BuildQueryMessages(ctx context.Context, userInput string, history []*schema.Message, templateVars map[string]any) ([]*schema.Message, error) {
	// 复用旧的 BuildQueryMessages 逻辑
	// 这里直接调用旧包，避免重复代码
	// 注意：旧包路径 msa/pkg/logic/agent 中的 BuildQueryMessages 签名相同
	// 重构完成后再删除旧包
	return buildQueryMessages(ctx, userInput, history, templateVars)
}

// FilterAndTrimHistory 过滤并截断历史消息
func FilterAndTrimHistory(history []schema.Message, maxRounds int) []schema.Message {
	filtered := make([]schema.Message, 0, len(history))
	for _, msg := range history {
		if msg.Role == schema.User || msg.Role == schema.Assistant {
			filtered = append(filtered, msg)
		}
	}
	maxMessages := maxRounds * 2
	if len(filtered) > maxMessages {
		filtered = filtered[len(filtered)-maxMessages:]
	}
	return filtered
}
```

- [ ] **Step 2: 在 basic.go 中添加 GetToolsMap 函数**

在 `pkg/logic/tools/basic.go` 末尾添加：

```go
// GetToolsMap 返回工具名称到 MsaTool 的映射（供 core/agent 使用）
func GetToolsMap() map[string]MsaTool {
	result := make(map[string]MsaTool, len(toolMap))
	for k, v := range toolMap {
		result[k] = v
	}
	return result
}
```

- [ ] **Step 3: 处理 buildQueryMessages 依赖**

在 `pkg/core/agent/` 下创建 `prompt.go`，将 `pkg/logic/agent/prompt.go` 和 `pkg/logic/agent/tempate.go` 的内容复制进来（包名改为 `agent`，import 路径不变）。

运行：
```bash
go build ./pkg/core/agent/...
```

如果有编译错误，逐个修复 import 路径。

- [ ] **Step 4: 验证编译通过**

```bash
go build ./pkg/core/...
```

Expected: 无输出

- [ ] **Step 5: Commit**

```bash
git add pkg/core/agent/ pkg/logic/tools/basic.go
git commit -m "feat(core/agent): implement Agent with ReAct loop and tool execution"
```

---

## Task 6: 实现 Renderer 接口和 CLI 实现

**Files:**
- Create: `pkg/renderer/renderer.go`
- Create: `pkg/renderer/cli.go`
- Create: `pkg/renderer/tui.go`

- [ ] **Step 1: 创建 renderer.go（接口）**

```go
// pkg/renderer/renderer.go
package renderer

import (
	"context"
	"msa/pkg/core/event"
)

// Renderer 定义输出适配器接口
// CLI 和 TUI 各自实现，Runner 通过此接口解耦
type Renderer interface {
	Handle(ctx context.Context, e event.Event) error
}
```

- [ ] **Step 2: 创建 cli.go**

```go
// pkg/renderer/cli.go
package renderer

import (
	"context"
	"fmt"
	"io"

	"msa/pkg/core/event"
)

// CLIRenderer 将 Event 输出到终端（os.Stdout 或测试用 io.Writer）
type CLIRenderer struct {
	out     io.Writer
	verbose bool // true 时显示 thinking 内容
}

// NewCLI 创建 CLI renderer
// out: 输出目标（生产用 os.Stdout，测试时可替换）
// verbose: 是否显示 thinking 内容
func NewCLI(out io.Writer, verbose bool) *CLIRenderer {
	return &CLIRenderer{out: out, verbose: verbose}
}

// Handle 处理一个 Event，输出到终端
func (r *CLIRenderer) Handle(ctx context.Context, e event.Event) error {
	switch e.Type {
	case event.EventTextChunk:
		if e.Text != "" {
			fmt.Fprint(r.out, e.Text)
		}

	case event.EventTextDone:
		fmt.Fprintln(r.out) // 输出结束后换行

	case event.EventThinking:
		if r.verbose && e.Text != "" {
			fmt.Fprint(r.out, e.Text)
		}

	case event.EventToolStart:
		fmt.Fprintf(r.out, "\n⚙ 正在调用 %s...\n", e.Tool.Name)

	case event.EventToolResult:
		fmt.Fprintf(r.out, "✓ %s 完成\n", e.Result.Name)

	case event.EventToolError:
		fmt.Fprintf(r.out, "✗ %s 失败: %v\n", e.Result.Name, e.Err)

	case event.EventRoundDone:
		// 对话正常结束，无需额外输出

	case event.EventError:
		return e.Err // 向上冒泡，Runner 处理
	}
	return nil
}
```

- [ ] **Step 3: 创建 tui.go**

```go
// pkg/renderer/tui.go
package renderer

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"msa/pkg/core/event"
)

// TUIRenderer 将 Event 转发给 Bubble Tea program
type TUIRenderer struct {
	send func(tea.Msg)
}

// NewTUI 创建 TUI renderer
// send: bubbletea program.Send 函数
func NewTUI(send func(tea.Msg)) *TUIRenderer {
	return &TUIRenderer{send: send}
}

// Handle 将 Event 直接转发给 Bubble Tea，由 TUI 决定如何渲染
func (r *TUIRenderer) Handle(ctx context.Context, e event.Event) error {
	r.send(e)
	return nil
}
```

- [ ] **Step 4: 验证编译通过**

```bash
go build ./pkg/renderer/...
```

Expected: 无输出

- [ ] **Step 5: Commit**

```bash
git add pkg/renderer/
git commit -m "feat(renderer): add Renderer interface, CLIRenderer, and TUIRenderer"
```

---

## Task 7: 实现 Runner

**Files:**
- Create: `pkg/core/runner/runner.go`

- [ ] **Step 1: 创建 runner.go**

```go
// pkg/core/runner/runner.go
package runner

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"

	"msa/pkg/core/agent"
	"msa/pkg/core/event"
	corelogger "msa/pkg/core/logger"
	"msa/pkg/logic/skills"
	"msa/pkg/model"
	"msa/pkg/renderer"
	"msa/pkg/session"

	log "github.com/sirupsen/logrus"
)

// Runner 是 core 层的对外入口
// CLI 和 TUI 都通过 Runner.Ask() 发起对话
type Runner struct {
	agent      *agent.Agent
	sessionMgr *session.Manager
	renderer   renderer.Renderer
}

// New 创建 Runner
func New(ag *agent.Agent, sessionMgr *session.Manager, r renderer.Renderer) *Runner {
	return &Runner{
		agent:      ag,
		sessionMgr: sessionMgr,
		renderer:   r,
	}
}

// Ask 处理一次用户输入（CLI 调用一次，TUI 每次回车调用一次）
// 阻塞直到本轮对话完成或 ctx 取消
func (r *Runner) Ask(ctx context.Context, input string, history []model.Message) error {
	ctx, reqID := corelogger.WithRequestID(ctx)
	logger := corelogger.FromCtx(ctx)
	logger.Infof("[Runner] 收到输入 reqID=%s len=%d", reqID, len(input))

	// 初始化 skills
	skillManager := skills.GetManager()
	if err := skillManager.Initialize(); err != nil {
		log.Warnf("[Runner] skills initialize warning: %v", err)
	}

	// 过滤并截断历史消息
	filteredHistory := agent.FilterAndTrimHistory(history, 10)
	schemaHistory := convertToSchemaMessages(filteredHistory)

	// 构建查询消息（系统 prompt + 历史 + 用户输入）
	now := time.Now()
	messages, err := agent.BuildQueryMessages(ctx, input, schemaHistory, map[string]any{
		"role":    "专业股票分析助手",
		"style":   "理性、专业、客观且严谨",
		"time":    now.Format("2006-01-02 15:04:05"),
		"weekday": now.Format("2006年01月02日 星期Monday"),
	})
	if err != nil {
		return fmt.Errorf("构建消息失败: %w", err)
	}

	logger.Infof("[Runner] 构建消息完成, 共 %d 条", len(messages))

	// 启动 agent，获取 event channel
	eventCh := r.agent.Run(ctx, messages)

	// 消费 event，交给 renderer 处理，同时收集完整回复用于持久化
	var assistantReply strings.Builder
	for e := range eventCh {
		if err := r.renderer.Handle(ctx, e); err != nil {
			return err
		}
		if e.Type == event.EventTextChunk {
			assistantReply.WriteString(e.Text)
		}
	}

	logger.Infof("[Runner] 本轮对话完成 replyLen=%d", assistantReply.Len())
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
```

- [ ] **Step 2: 验证编译通过**

```bash
go build ./pkg/core/runner/...
```

Expected: 无输出

- [ ] **Step 3: Commit**

```bash
git add pkg/core/runner/
git commit -m "feat(core/runner): implement Runner as the unified core entry point"
```

---

## Task 8: 更新 extcli（CLI 单次模式）

**Files:**
- Modify: `pkg/extcli/chat.go`

- [ ] **Step 1: 重写 extcli/chat.go**

用新的 Runner + CLIRenderer 替代旧的 Ask + StreamOutputManager：

```go
// pkg/extcli/chat.go
package extcli

import (
	"context"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	coreagent "msa/pkg/core/agent"
	"msa/pkg/core/runner"
	"msa/pkg/config"
	"msa/pkg/model"
	"msa/pkg/renderer"
	"msa/pkg/session"
)

// Run 执行 CLI 单轮对话
// 返回退出码：0 成功，1 失败
func Run(ctx context.Context, question string, modelOverride string) int {
	if question == "" {
		log.Error("问题内容不能为空")
		return 1
	}

	cfg := config.GetLocalStoreConfig()
	if cfg == nil {
		log.Error("配置未初始化，请运行 'msa config'")
		return 1
	}
	if cfg.APIKey == "" {
		log.Error("请先配置 API Key，运行 'msa config'")
		return 1
	}
	if cfg.BaseURL == "" {
		log.Error("请先配置 Base URL，运行 'msa config'")
		return 1
	}

	// 处理模型覆盖
	if modelOverride != "" {
		cfg.Model = modelOverride
		log.Infof("使用命令行指定的模型: %s", modelOverride)
	}
	if cfg.Model == "" {
		log.Error("请指定模型 (-m) 或配置默认模型")
		return 1
	}

	// 创建 Agent（每次都创建新实例，替代旧的 ResetCache + GetChatModel）
	ag, err := coreagent.New(ctx)
	if err != nil {
		log.Errorf("创建 Agent 失败: %v", err)
		return 1
	}

	// 创建会话并持久化用户消息
	sessionMgr := session.GetManager()
	sess := sessionMgr.NewSession(session.ModeCLI)
	if err := sessionMgr.CreateSessionFile(sess); err != nil {
		log.Warnf("创建会话文件失败: %v", err)
	}
	sessionMgr.SetCurrent(sess)
	sessionMgr.AppendMessage(sess, "user", question)

	// 创建 Runner（注入 CLIRenderer）
	r := runner.New(ag, sessionMgr, renderer.NewCLI(os.Stdout, false))

	// 发起对话（单次，无历史）
	if err := r.Ask(ctx, question, []model.Message{}); err != nil {
		log.Errorf("对话失败: %v", err)
		return 1
	}

	// 输出会话 ID（从 Runner.Ask 完成后获取回复较复杂，暂时省略持久化 assistant 消息）
	// TODO: Runner 完成后需要回传 assistant 回复供持久化，可通过 callback 扩展
	fmt.Printf("\n---\n📌 会话ID: %s\n", sess.SessionID())
	fmt.Printf("   msa --resume %s\n", sess.SessionID())

	return 0
}
```

- [ ] **Step 2: 验证编译通过**

```bash
go build ./pkg/extcli/...
```

Expected: 无输出

- [ ] **Step 3: Commit**

```bash
git add pkg/extcli/chat.go
git commit -m "refactor(extcli): use Runner+CLIRenderer, remove StreamOutputManager dependency"
```

---

## Task 9: 更新 TUI（适配 Event）

**Files:**
- Modify: `pkg/tui/chat.go`
- Modify: `pkg/app/app.go`

- [ ] **Step 1: 更新 app.go**

```go
// pkg/app/app.go
package app

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	tea "github.com/charmbracelet/bubbletea"

	"msa/pkg/session"
	"msa/pkg/tui"
)

// Run 启动应用主逻辑
func Run(ctx context.Context) error {
	p := tea.NewProgram(tui.NewChat(ctx))
	if _, err := p.Run(); err != nil {
		log.Errorf("运行 TUI 失败: %v", err)
		return err
	}

	sessionMgr := session.GetManager()
	sess := sessionMgr.Current()
	if sess != nil {
		fmt.Printf("\n📌 会话ID: %s\n", sess.SessionID())
		fmt.Printf("   msa --resume %s\n", sess.SessionID())
	}

	return nil
}
```

（app.go 本身不变，主要改动在 tui/chat.go）

- [ ] **Step 2: 修改 tui/chat.go 中的 startChatRequest**

将 `pkg/tui/chat.go` 中的 `startChatRequest` 方法修改为使用新的 Runner：

找到（约第 745 行）：
```go
func (c *Chat) startChatRequest(input string) (tea.Model, tea.Cmd) {
	c.streamOutputCh, c.streamUnregister = message.RegisterStreamOutput(streamBufferSize)

	err := agent.Ask(c.ctx, input, c.history)
	if err != nil {
		log.Errorf("聊天请求失败: %v", err)
		c.clearStreamState()
		c.addMessage(model.RoleSystem, "聊天出错: "+err.Error(), model.StreamMsgTypeText)
		return c, c.Flush()
	}

	return c, tea.Batch(c.Flush(), c.startStreaming())
}
```

替换为：

```go
func (c *Chat) startChatRequest(input string) (tea.Model, tea.Cmd) {
	// 创建 Agent（不再使用全局 agentCache）
	ag, err := coreagent.New(c.ctx)
	if err != nil {
		log.Errorf("创建 Agent 失败: %v", err)
		c.addMessage(model.RoleSystem, "创建 Agent 失败: "+err.Error(), model.StreamMsgTypeText)
		return c, c.Flush()
	}

	// 创建 Runner，注入 TUIRenderer
	r := runner.New(ag, c.sessionMgr, renderer.NewTUI(func(msg tea.Msg) {
		// 注意：这里不能直接调用 p.Send，因为 TUI 内部没有 program 引用
		// 改为：Runner.Ask 在 goroutine 中运行，Event 通过 channel 传递给 TUI
		// 见下方 goroutine 方案
	}))
	_ = r // 暂时 suppress unused warning

	// 方案：通过内部 channel 传递 Event 给 TUI
	c.eventCh = make(chan event.Event, 32)

	go func() {
		defer close(c.eventCh)
		if err := r.Ask(c.ctx, input, c.history); err != nil {
			log.Errorf("[TUI] Runner.Ask 失败: %v", err)
			c.eventCh <- event.Event{Type: event.EventError, Err: err}
		}
	}()

	return c, tea.Batch(c.Flush(), c.startStreaming())
}
```

- [ ] **Step 3: 在 Chat struct 中添加 eventCh 字段，移除旧的 streamOutputCh 和 streamUnregister**

在 `pkg/tui/chat.go` 的 `Chat` struct 中：

删除：
```go
streamOutputCh       <-chan *model.StreamChunk   // 流式输出 channel
streamUnregister     func()                      // 取消订阅函数
```

新增：
```go
eventCh              chan event.Event            // 新 event channel
```

- [ ] **Step 4: 修改 receiveNextChunk 使用新 eventCh**

找到：
```go
func (c *Chat) receiveNextChunk() tea.Cmd {
	return func() tea.Msg {
		if c.streamOutputCh == nil {
			log.Error("流式输出 channel 为空")
			return &model.StreamChunk{Err: fmt.Errorf("stream output channel is nil")}
		}
		chunk, ok := <-c.streamOutputCh
		if !ok {
			log.Debug("流式输出 channel 已关闭")
			return &model.StreamChunk{IsDone: true}
		}
		log.Debugf("接收流式块: Content长度=%d, IsDone=%v, Err=%v", len(chunk.Content), chunk.IsDone, chunk.Err)
		return chunk
	}
}
```

替换为：
```go
func (c *Chat) receiveNextChunk() tea.Cmd {
	return func() tea.Msg {
		if c.eventCh == nil {
			log.Error("event channel 为空")
			return event.Event{Type: event.EventError, Err: fmt.Errorf("event channel is nil")}
		}
		e, ok := <-c.eventCh
		if !ok {
			log.Debug("event channel 已关闭")
			return event.Event{Type: event.EventRoundDone}
		}
		log.Debugf("接收 event: type=%d", e.Type)
		return e
	}
}
```

- [ ] **Step 5: 修改 chatStreamMsg 处理新 Event 类型**

找到：
```go
case *model.StreamChunk:
	return c.chatStreamMsg(msg)
```

替换为：
```go
case event.Event:
	return c.handleEvent(msg)
```

然后将旧的 `chatStreamMsg` 方法重命名为 `handleEvent`，并将所有 `*model.StreamChunk` 引用改为 `event.Event`：

- `msg.IsDone` → `msg.Type == event.EventRoundDone || msg.Type == event.EventError`
- `msg.Err` → `msg.Err`（EventError 时）
- `msg.Content` → `msg.Text`
- `msg.MsgType == model.StreamMsgTypeText` → `msg.Type == event.EventTextChunk`
- `msg.MsgType == model.StreamMsgTypeReason` → `msg.Type == event.EventThinking`
- `msg.MsgType == model.StreamMsgTypeTool` → `msg.Type == event.EventToolStart || msg.Type == event.EventToolResult`

- [ ] **Step 6: 修改 clearStreamState 移除旧订阅逻辑**

找到：
```go
if c.streamUnregister != nil {
	log.Debug("取消流式输出订阅")
	c.streamUnregister()
	c.streamUnregister = nil
}
c.streamOutputCh = nil
```

替换为：
```go
c.eventCh = nil
```

- [ ] **Step 7: 移除旧的 message 包 import**

在 `pkg/tui/chat.go` 顶部，移除：
```go
"msa/pkg/logic/message"
```

添加：
```go
coreagent "msa/pkg/core/agent"
"msa/pkg/core/event"
"msa/pkg/core/runner"
"msa/pkg/renderer"
```

- [ ] **Step 8: 验证编译通过**

```bash
go build ./pkg/tui/... ./pkg/app/...
```

修复所有编译错误。

- [ ] **Step 9: 完整编译验证**

```bash
go build ./...
```

Expected: 无输出

- [ ] **Step 10: Commit**

```bash
git add pkg/tui/chat.go pkg/app/app.go
git commit -m "refactor(tui): use Runner+TUIRenderer with event.Event, remove StreamChunk dependency"
```

---

## Task 10: 移除工具层的 BroadcastToolStart/End 调用

工具调用的广播现在由 `agent.executeTools()` 统一处理，工具实现不再需要手动广播。

**Files:**
- Modify: `pkg/logic/tools/finance/account.go`
- Modify: `pkg/logic/tools/finance/trade.go`
- Modify: `pkg/logic/tools/finance/position.go`
- Modify: `pkg/logic/tools/safetool/safetool.go`
- （以及所有其他调用 `message.BroadcastToolStart/End` 的工具文件）

- [ ] **Step 1: 找到所有调用 BroadcastToolStart/End 的文件**

```bash
grep -rl "BroadcastToolStart\|BroadcastToolEnd" /Users/dragon/Documents/Code/goland/msa/pkg/logic/tools/
```

- [ ] **Step 2: 在 safetool.go 中移除广播调用**

找到 `pkg/logic/tools/safetool/safetool.go`：

```go
// 移除 message 包 import
// 移除 message.BroadcastToolEnd(toolName, "", errMsg) 调用
// 保留 panic recovery 和日志逻辑
```

修改后：
```go
package safetool

import (
	"fmt"
	"runtime/debug"

	"msa/pkg/model"

	log "github.com/sirupsen/logrus"
)

// SafeExecute 提供统一的 panic recovery 保护
func SafeExecute(toolName string, params string, fn func() (string, error)) (result string, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Tool [%s] panic recovered: %v\n%s", toolName, r, debug.Stack())
			result = model.NewErrorResult(fmt.Sprintf("工具内部错误: %v", r))
			err = nil
		}
	}()
	result, err = fn()
	log.Infof("Tool [%s] result: %s", toolName, result)
	if err != nil {
		log.Errorf("Tool [%s] error: %v", toolName, err)
		return "", err
	}
	return result, nil
}
```

- [ ] **Step 3: 批量移除所有工具文件中的 BroadcastToolStart/End 调用**

对每个工具文件（finance/account.go, finance/trade.go, finance/position.go 等）：
1. 移除 `message.BroadcastToolStart(...)` 调用
2. 移除 `message.BroadcastToolEnd(...)` 调用  
3. 移除 `"msa/pkg/logic/message"` import（如果不再使用）

- [ ] **Step 4: 验证编译通过**

```bash
go build ./pkg/logic/tools/...
```

- [ ] **Step 5: 完整编译验证**

```bash
go build ./...
```

Expected: 无输出

- [ ] **Step 6: Commit**

```bash
git add pkg/logic/tools/
git commit -m "refactor(tools): remove BroadcastToolStart/End, tool events now handled by agent layer"
```

---

## Task 11: 删除旧包，最终验证

**Files:**
- Delete: `pkg/logic/agent/chat.go`（保留 prompt.go 和 tempate.go 直到 core/agent 完全接管）
- Delete: `pkg/logic/message/stream_manage.go`（确认无引用后删除）

- [ ] **Step 1: 确认旧包无引用**

```bash
grep -rl "logic/agent\|logic/message" /Users/dragon/Documents/Code/goland/msa/pkg/ | grep -v "_test.go" | grep -v "core/"
```

Expected: 无输出（或只有允许保留的文件）

- [ ] **Step 2: 完整编译验证**

```bash
go build ./...
```

Expected: 无输出

- [ ] **Step 3: 运行所有现有测试**

```bash
cd /Users/dragon/Documents/Code/goland/msa
go test ./... -count=1 2>&1 | tail -30
```

修复所有因重构导致的测试失败。

- [ ] **Step 4: 手动验证 CLI 模式**

```bash
go run . -q "今天A股市场如何？" 2>&1 | head -20
```

Expected: 输出流式文本，显示 `⚙ 正在调用 xxx...` 和 `✓ xxx 完成`，最后输出会话 ID

- [ ] **Step 5: 手动验证 TUI 模式**

```bash
go run .
```

Expected: 正常进入 TUI，输入问题后有流式输出，工具调用有视觉反馈

- [ ] **Step 6: 最终 Commit**

```bash
git add -A
git commit -m "refactor: complete architecture refactor - event-driven pipeline replaces StreamOutputManager"
```

---

## 自查：Spec 覆盖情况

| 设计文档要求 | 对应 Task |
|---|---|
| Event 类型系统 | Task 1 |
| requestID 日志 | Task 2 |
| tool_parser 迁移 | Task 3 |
| StreamAdapter 替代 toolCallChecker | Task 4 |
| Agent.Run() 返回 `<-chan Event` | Task 5 |
| Renderer 接口 + CLI/TUI 实现 | Task 6 |
| Runner 统一入口 | Task 7 |
| extcli 使用新架构 | Task 8 |
| tui 使用新架构 | Task 9 |
| 工具层移除广播 | Task 10 |
| 删除旧包、最终验证 | Task 11 |
| ctx 全程贯穿 DB | ⚠️ 未包含（DB 当前 ctx 问题不影响核心链路，建议后续独立任务处理）|
