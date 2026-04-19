# MSA 架构重构设计文档

**日期：** 2026-04-19  
**状态：** 已批准，待实现  
**作者：** 架构讨论（用户 + Claude）

---

## 背景与目标

### 当前痛点

1. **Eino stream 处理缺陷** — Eino 框架对 tool call 的 chunk 类型判断不稳定，第一个 chunk 有时不是 tool 类型，导致工具调用失败。为此引入了 `toolCallChecker` 作为补丁，但它反而成为了新的复杂度来源。

2. **ctx 未向下传递** — DB 操作、工具调用全程没有传递 context，无法超时、无法取消。

3. **全局 StreamOutputManager** — 广播模式导致通信链路不透明，难以追踪消息来源。

4. **goroutine 生命周期无管理** — `Ask()` 内部开 goroutine 没有结束信号，出错也无法感知。

5. **整体链路"绕"** — 加功能时需要同时改多处，难以维护。

### 设计目标

- **彻底重写核心链路**，摆脱历史债务
- **ctx 全程贯穿**，从入口到 DB 每一层都传递
- **通信链路清晰**，用 channel + Event 替代全局广播
- **CLI 和 TUI 共用同一套 core**，差异只在最后的 Renderer 层
- **日志可观测**，一条 grep 看完整条链路

---

## 使用场景

系统支持两种**互斥**的运行模式（不会同时存在）：

- **CLI 单次模式：** `msa -q '问题'` — 单次调用，输出结果后退出
- **TUI 多轮模式：** `msa` — 进入终端 UI，支持多轮对话

两种模式共用同一套 core 逻辑，差异由 `Renderer` 接口吸收。

---

## 整体架构：事件驱动 Pipeline

```
用户输入
  ↓
Runner.Ask(ctx, input)
  ↓ 构造 messages（注入 requestID）
Agent.Run(ctx, messages) → <-chan Event
  ↓ StreamAdapter 处理 Eino 原始 stream
  ↓ ReAct 循环：LLM → tool → LLM → ...
  ↓ 每步发射结构化 Event
Runner 消费 eventCh
  ↓
Renderer.Handle(ctx, event)
  ↓
CLI: fmt.Print          TUI: program.Send(event)
```

---

## 一、目录结构

```
msa/
├── cmd/
│   ├── root.go          # cobra 入口，构造 ctx（signal handling）
│   ├── tui.go           # `msa` → 启动 TUI 模式
│   └── query.go         # `msa -q` → 启动 CLI 单次模式
│
├── internal/
│   ├── core/
│   │   ├── agent/
│   │   │   ├── agent.go          # Eino 封装，Run() 返回 <-chan Event
│   │   │   └── stream_adapter.go # 替代 toolCallChecker，处理 chunk 分类
│   │   ├── event/
│   │   │   └── event.go          # Event 类型定义（整条 pipeline 的数据结构）
│   │   ├── runner/
│   │   │   └── runner.go         # 主循环：协调 agent → renderer，持久化会话
│   │   ├── tools/
│   │   │   └── ...               # 工具实现，所有方法接收 ctx
│   │   └── session/
│   │       └── ...               # 会话持久化
│   │
│   ├── renderer/
│   │   ├── renderer.go           # interface Renderer
│   │   ├── cli.go                # CLI 实现：直接写 stdout
│   │   └── tui.go                # TUI 实现：写入 bubbletea program.Send
│   │
│   └── infra/
│       ├── db/                   # SQLite，所有方法接收 ctx
│       ├── config/
│       └── provider/             # LLM provider 配置
│
└── tui/                          # Bubble Tea UI 组件（纯 UI，无业务逻辑）
    ├── chat.go
    └── ...
```

**分层原则：**
- `cmd/` 只负责"组装"——构造依赖、注入、启动
- `internal/core/` 是业务核心，**不依赖任何 UI 框架**
- `tui/` 是纯 UI 组件，只知道如何渲染，不知道 Agent 是什么
- `internal/renderer/` 是连接 core 和 UI 的适配层

---

## 二、Event 类型系统

整条 pipeline 通过 `Event` 流动，所有通信用结构化类型表达，不再依赖字符串解析或全局广播。

```go
// internal/core/event/event.go

package event

type EventType int

const (
    // 文本流
    EventTextChunk EventType = iota // LLM 输出的普通文本片段
    EventTextDone                   // 本轮文本输出完毕

    // 工具调用
    EventToolStart                  // 工具开始执行（携带工具名、参数）
    EventToolResult                 // 工具执行完成（携带结果）
    EventToolError                  // 工具执行失败

    // 思考过程（reasoning model 支持）
    EventThinking                   // <thinking> 内容片段

    // 会话控制
    EventRoundDone                  // 一轮对话完成（agent 不再调用工具）
    EventError                      // 不可恢复的错误，pipeline 终止
)

type Event struct {
    Type   EventType

    // EventTextChunk / EventThinking
    Text   string

    // EventToolStart
    Tool   ToolCall

    // EventToolResult / EventToolError
    Result ToolResult

    // EventError
    Err    error
}

type ToolCall struct {
    ID    string
    Name  string
    Input string // JSON
}

type ToolResult struct {
    ToolCallID string
    Name       string
    Output     string
    IsError    bool
}
```

**设计要点：**
- `EventToolStart` + `EventToolResult` 配对，TUI 可以显示工具执行进度
- `EventRoundDone` 明确终止信号，替代现在 `IsDone` 消息丢失的问题
- `EventError` 携带 err，错误从 pipeline 内部自然流出
- `EventThinking` 独立类型，TUI 可折叠显示，CLI 可忽略

---

## 三、Agent + StreamAdapter

### agent.go

```go
// internal/core/agent/agent.go

type Agent struct {
    eino     *einoReActAgent
    tools    *tools.Registry
    provider provider.LLM
    adapter  *StreamAdapter
}

// Run 是唯一的对外接口
// ctx 取消时，channel 自动关闭，所有 goroutine 自动退出
func (a *Agent) Run(ctx context.Context, messages []model.Message) <-chan event.Event {
    ch := make(chan event.Event, 32)
    go a.run(ctx, messages, ch)
    return ch
}

func (a *Agent) run(ctx context.Context, messages []model.Message, ch chan<- event.Event) {
    defer close(ch) // channel close = 终止信号，消费方 range 自动退出
    log := logger.FromCtx(ctx)

    round := 0
    for {
        round++
        log.Infof("[Agent] 第 %d 轮 LLM 调用开始", round)

        streamReader, err := a.eino.Stream(ctx, messages)
        if err != nil {
            log.Errorf("[Agent] 第 %d 轮 LLM 调用失败: %v", round, err)
            send(ctx, ch, event.Event{Type: event.EventError, Err: err})
            return
        }

        result, err := a.adapter.Process(ctx, streamReader, ch)
        if err != nil {
            send(ctx, ch, event.Event{Type: event.EventError, Err: err})
            return
        }

        if len(result.ToolCalls) == 0 {
            log.Infof("[Agent] 第 %d 轮无工具调用，对话结束", round)
            send(ctx, ch, event.Event{Type: event.EventRoundDone})
            return
        }

        log.Infof("[Agent] 第 %d 轮需要调用工具: %v", round, toolNames(result.ToolCalls))
        toolMsgs := a.executeTools(ctx, result.ToolCalls, ch)
        messages = append(messages, result.AssistantMsg)
        messages = append(messages, toolMsgs...)
    }
}
```

### stream_adapter.go

```go
// internal/core/agent/stream_adapter.go
// 唯一职责：消费 Eino 原始 stream，输出干净的 Event，收集 tool calls

type StreamAdapter struct{}

type ProcessResult struct {
    ToolCalls    []event.ToolCall
    AssistantMsg model.Message
}

func (a *StreamAdapter) Process(
    ctx context.Context,
    reader *einoStreamReader,
    out chan<- event.Event,
) (ProcessResult, error) {
    log := logger.FromCtx(ctx)
    var (
        textBuf   strings.Builder
        toolCalls []event.ToolCall
        chunkCount int
    )

    for {
        chunk, err := reader.Recv()
        if errors.Is(err, io.EOF) {
            log.Debugf("[StreamAdapter] stream 结束, chunks=%d, toolCalls=%d",
                chunkCount, len(toolCalls))
            break
        }
        if err != nil {
            reader.Close() // 关键：drain 防止 goroutine 泄漏
            return ProcessResult{}, err
        }
        chunkCount++

        // Eino chunk 类型不稳定的问题全部在这里处理
        // 外部永远收到干净的 Event
        switch classifyChunk(chunk) {
        case chunkText:
            text := extractText(chunk)
            textBuf.WriteString(text)
            log.Debugf("[StreamAdapter] chunk#%d text len=%d", chunkCount, len(text))
            send(ctx, out, event.Event{Type: event.EventTextChunk, Text: text})

        case chunkThinking:
            text := extractThinking(chunk)
            log.Debugf("[StreamAdapter] chunk#%d thinking", chunkCount)
            send(ctx, out, event.Event{Type: event.EventThinking, Text: text})

        case chunkToolCall:
            tc := extractToolCall(chunk)
            log.Infof("[StreamAdapter] chunk#%d toolCall: name=%s id=%s", chunkCount, tc.Name, tc.ID)
            toolCalls = append(toolCalls, tc)
            send(ctx, out, event.Event{Type: event.EventToolStart, Tool: tc})
        }
    }

    if textBuf.Len() > 0 {
        send(ctx, out, event.Event{Type: event.EventTextDone})
    }

    return ProcessResult{
        ToolCalls:    toolCalls,
        AssistantMsg: buildAssistantMsg(textBuf.String(), toolCalls),
    }, nil
}
```

### send 工具函数

```go
// 尊重 ctx 取消，不阻塞
func send(ctx context.Context, ch chan<- event.Event, e event.Event) bool {
    select {
    case ch <- e:
        return true
    case <-ctx.Done():
        return false
    }
}
```

---

## 四、日志设计

### requestID 串联整条链路

```go
// internal/core/logger/logger.go

func WithRequestID(ctx context.Context) (context.Context, string) {
    id := newShortID() // 短 ID，便于 grep
    return context.WithValue(ctx, requestIDKey, id), id
}

func FromCtx(ctx context.Context) *logrus.Entry {
    id, _ := ctx.Value(requestIDKey).(string)
    return logrus.WithField("req", id)
}
```

### 日志级别约定

| 级别 | 用途 | 例子 |
|------|------|------|
| `ERROR` | 需要关注的失败 | 工具执行失败、LLM 调用失败、DB 错误 |
| `INFO` | 关键节点 | 对话开始/结束、每轮 LLM 调用、工具调用开始/完成 |
| `DEBUG` | 排查细节 | 每个 chunk、DB 查询、stream 状态 |

### 一次完整对话的日志示例

```
[req=01ARZ3] [Runner] 收到输入, len=12
[req=01ARZ3] [Agent] 开始对话, 历史消息数=4
[req=01ARZ3] [Agent] 第 1 轮 LLM 调用开始
[req=01ARZ3] [Agent] 第 1 轮 LLM stream 已建立
[req=01ARZ3] [StreamAdapter] chunk#1 text len=8
[req=01ARZ3] [StreamAdapter] chunk#8 toolCall: name=get_stock_price id=call_abc
[req=01ARZ3] [StreamAdapter] stream 结束, chunks=8, toolCalls=1
[req=01ARZ3] [Tool] 开始执行: name=get_stock_price id=call_abc input={"symbol":"AAPL"}
[req=01ARZ3] [DB] StockRepo.GetBySymbol: symbol=AAPL
[req=01ARZ3] [Tool] 执行完成: name=get_stock_price elapsed=230ms outputLen=142
[req=01ARZ3] [Agent] 第 1 轮需要调用工具: [get_stock_price]
[req=01ARZ3] [Agent] 第 2 轮 LLM 调用开始
[req=01ARZ3] [Agent] 第 2 轮无工具调用，对话结束
[req=01ARZ3] [Runner] 本轮对话完成, replyLen=380
```

---

## 五、Runner + Renderer

### runner.go

```go
// internal/core/runner/runner.go

type Runner struct {
    agent    *agent.Agent
    session  *session.Manager
    renderer renderer.Renderer
}

func New(agent *agent.Agent, session *session.Manager, r renderer.Renderer) *Runner {
    return &Runner{agent: agent, session: session, renderer: r}
}

func (r *Runner) Ask(ctx context.Context, input string) error {
    ctx, reqID := logger.WithRequestID(ctx)
    log := logger.FromCtx(ctx)
    log.Infof("[Runner] 收到输入, reqID=%s, len=%d", reqID, len(input))

    messages := r.session.BuildMessages(input)
    eventCh := r.agent.Run(ctx, messages)

    var assistantReply strings.Builder
    for e := range eventCh {
        if err := r.renderer.Handle(ctx, e); err != nil {
            return err
        }
        if e.Type == event.EventTextChunk {
            assistantReply.WriteString(e.Text)
        }
    }

    r.session.Append(input, assistantReply.String())
    log.Infof("[Runner] 本轮对话完成, replyLen=%d", assistantReply.Len())
    return nil
}
```

### renderer.go — 接口

```go
// internal/renderer/renderer.go

type Renderer interface {
    Handle(ctx context.Context, e event.Event) error
}
```

### cli.go — CLI 实现

```go
// internal/renderer/cli.go

type CLIRenderer struct {
    out     io.Writer // 默认 os.Stdout，测试时可替换
    verbose bool      // true 时显示 thinking 内容
}

func (r *CLIRenderer) Handle(ctx context.Context, e event.Event) error {
    switch e.Type {
    case event.EventTextChunk:
        fmt.Fprint(r.out, e.Text)
    case event.EventTextDone:
        fmt.Fprintln(r.out)
    case event.EventThinking:
        if r.verbose {
            fmt.Fprintf(r.out, "\n💭 %s\n", e.Text)
        }
    case event.EventToolStart:
        fmt.Fprintf(r.out, "\n⚙ 正在调用 %s...\n", e.Tool.Name)
    case event.EventToolResult:
        fmt.Fprintf(r.out, "✓ %s 完成\n", e.Result.Name)
    case event.EventToolError:
        fmt.Fprintf(r.out, "✗ %s 失败: %v\n", e.Result.Name, e.Err)
    case event.EventError:
        return e.Err
    }
    return nil
}
```

### tui.go — TUI 实现

```go
// internal/renderer/tui.go

type TUIRenderer struct {
    send func(tea.Msg) // bubbletea program.Send
}

func (r *TUIRenderer) Handle(ctx context.Context, e event.Event) error {
    r.send(e) // 直接把 Event 转发给 Bubble Tea，TUI 自己决定如何渲染
    return nil
}
```

### 两个入口组装

```go
// cmd/query.go — CLI 单次模式
func runQuery(ctx context.Context, question string) error {
    ag, _ := agent.New(cfg)
    sess := session.NewSingleShot()
    r := runner.New(ag, sess, renderer.NewCLI(os.Stdout, false))
    return r.Ask(ctx, question)
}

// cmd/tui.go — TUI 多轮模式
func runTUI(ctx context.Context) error {
    ag, _ := agent.New(cfg)
    sess, _ := session.Load(cfg.SessionID)
    p := tea.NewProgram(tui.NewChat(...))
    r := runner.New(ag, sess, renderer.NewTUI(p.Send))
    return p.Run()
}
```

---

## 六、ctx 取消退出路径

```
ctx.Done() 触发
  → agent.run() 中 send() select ctx.Done()，退出循环
  → defer close(ch) 执行
  → Runner 的 for range eventCh 退出
  → Runner.Ask() 返回
  → 整条链路干净退出，无 goroutine 泄漏
```

---

## 七、DB 层 ctx 传递规范

所有 DB 操作统一接收 ctx，使用 `WithContext` 确保可取消：

```go
func (r *AccountRepo) GetByID(ctx context.Context, id int64) (*model.Account, error) {
    log := logger.FromCtx(ctx)
    log.Debugf("[DB] AccountRepo.GetByID: id=%d", id)

    var account model.Account
    if err := r.db.WithContext(ctx).First(&account, id).Error; err != nil {
        log.Errorf("[DB] AccountRepo.GetByID 失败: id=%d, err=%v", id, err)
        return nil, err
    }
    return &account, nil
}
```

**所有工具实现同理，第一个参数必须是 `ctx context.Context`。**

---

## 实现顺序建议

1. **event/** — 定义 Event 类型（无依赖，先建好地基）
2. **logger/** — requestID 注入和 FromCtx（被所有层使用）
3. **infra/db/** — 所有 Repo 加 ctx 参数
4. **core/agent/stream_adapter.go** — 替代 toolCallChecker（最核心）
5. **core/agent/agent.go** — 重写 Run() + ReAct 循环
6. **renderer/** — CLI 和 TUI 两个实现
7. **core/runner/** — 组装层
8. **cmd/** — 更新两个入口
9. **tui/** — 适配新的 Event 类型渲染
