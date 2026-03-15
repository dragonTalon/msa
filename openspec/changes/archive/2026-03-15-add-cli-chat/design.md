# 设计：add-cli-chat

## 上下文

当前 MSA 只支持 TUI 交互模式，用户必须进入 Bubble Tea 界面才能进行对话。本设计添加 CLI 单轮对话模式，复用现有流式架构，直接输出到终端。

**当前架构**：

```
cmd/root.go ──▶ app.Run(ctx) ──▶ TUI (bubbletea)
                                        │
                                        ▼
                               agent.Ask() ──▶ StreamManager.Broadcast()
                                                      │
                                                      ▼
                                               TUI 订阅者渲染
```

**目标架构**：

```
cmd/root.go ──▶ if -q: cli.Run(ctx, question, model)
               │              │
               │              ▼
               │        agent.Ask() ──▶ StreamManager.Broadcast()
               │                                    │
               │                                    ▼
               │                             CLI 订阅者 ──▶ stdout
               │
               └── else: app.Run(ctx) ──▶ TUI (现有流程)
```

## 目标 / 非目标

**目标：**

- 添加 `-q/--question` 参数支持单轮对话
- 添加 `-m/--model` 参数支持模型覆盖
- 复用现有 `agent.Ask()` 和 `StreamOutputManager` 架构
- 支持三种消息类型的流式输出（正文、思考、工具调用）
- 提供清晰的错误信息和退出码

**非目标：**

- 不支持多轮对话历史上下文
- 不支持会话恢复（`--resume` 与 `-q` 互斥）
- 不支持自定义输出格式（JSON/Markdown）
- 不修改 TUI 现有行为
- 不支持流式中断（Ctrl+C 处理由系统默认）

## 推荐模式

### 分类错误处理模式

**模式ID**：error-handling-strategy
**来源问题**：知识库模式

**描述**：
根据错误类型采取不同的处理策略：参数错误立即返回，永久错误不重试，临时错误可重试。

**适用场景**：
- 参数验证（空问题、缺少配置）
- API 调用失败处理
- 流式输出错误处理

**实现**：

```go
// 参数错误 - 立即返回
if question == "" {
    fmt.Fprintln(os.Stderr, "错误：问题内容不能为空")
    return 1
}

// 永久错误 - 不重试
if cfg.APIKey == "" {
    fmt.Fprintln(os.Stderr, "错误：请先配置 API Key，运行 'msa config'")
    return 1
}

// 流式错误 - 记录并退出
if chunk.Err != nil {
    fmt.Fprintf(os.Stderr, "错误：%v\n", chunk.Err)
    return 1
}
```

### 配置优先级模式

**模式ID**：config-auto-fill-on-provider-change
**来源问题**：知识库模式

**描述**：
命令行参数具有最高优先级，覆盖配置文件中的值。

**适用场景**：
- `-m` 参数覆盖 Model 配置
- 保持现有配置优先级链：CLI > ENV > File > Default

**实现**：

```go
// 在 runRoot 中处理 -m 参数
if modelOverride != "" {
    // 临时覆盖模型配置
    cfg.Model = modelOverride
    config.SetTempModel(modelOverride)
}
```

## 避免的反模式

### 阻塞通道操作

**反模式ID**：blocking-channel
**来源问题**：知识库问题 #004 (浏览器超时)

**问题所在**：
在流式输出中使用阻塞通道操作可能导致死锁。

**影响**：
程序无响应，无法正常退出。

**不要这样做**：

```go
// 不要这样做 - 阻塞读取
for {
    chunk := <-streamCh  // 如果没有数据会永久阻塞
    process(chunk)
}
```

**应该这样做**：

```go
// 应该这样做 - 使用 for-range 或 select
for chunk := range streamCh {
    process(chunk)
}

// 或者使用 select 处理超时
select {
case chunk := <-streamCh:
    process(chunk)
case <-time.After(timeout):
    return
}
```

### 忽略错误上下文

**反模式ID**：ignore-error-context
**来源问题**：知识库模式 error-handling-strategy

**问题所在**：
错误信息不包含具体原因，用户无法定位问题。

**不要这样做**：

```go
// 不要这样做
if err != nil {
    fmt.Println("出错了")
    return 1
}
```

**应该这样做**：

```go
// 应该这样做
if err != nil {
    fmt.Fprintf(os.Stderr, "错误：%v\n", err)
    return 1
}
```

## 设计决策

### 决策 1：复用现有流式架构

**选择**：复用 `StreamOutputManager` 订阅模式
**理由**：
- 现有 TUI 已使用此架构，成熟稳定
- CLI 模式只是另一个订阅者，代码复用度高
- 避免重复实现流式处理逻辑

**替代方案**：创建同步版本的 `AskSync()`
- 优点：逻辑更简单
- 缺点：需要新增代码，无法复用现有测试

### 决策 2：模块位置

**选择**：新建 `pkg/extcli/chat.go`
**理由**：
- 与 `pkg/tui/chat.go` 对称，职责清晰
- `extcli` 表示扩展 CLI 能力，用于快速处理命令
- 方便后续扩展 TUI 等能力的快捷命令
- CLI 是独立入口，不依赖 Bubble Tea

### 决策 3：模型参数处理方式

**选择**：在 `runRoot` 中临时覆盖配置
**理由**：
- 不修改配置文件，保持配置纯净
- 遵循现有配置优先级链
- 简单直接，无副作用

### 决策 4：参数互斥处理

**选择**：忽略 `--resume` 参数，输出警告
**理由**：
- 用户意图明确（单轮对话）
- 不中断用户操作
- 警告提示避免混淆

## 风险 / 权衡

| 风险 | 缓解措施 |
|------|----------|
| CLI 输出可能与 TUI 不一致 | 复用相同的消息类型定义，共享格式逻辑 |
| 流式输出中断处理 | 依赖 Go 默认信号处理，确保资源清理 |
| 模型覆盖可能影响其他配置 | 只覆盖 Model 字段，保留其他配置 |

## 架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                        cmd/root.go                              │
├─────────────────────────────────────────────────────────────────┤
│  -q, -m 参数解析                                                 │
│       │                                                         │
│       ├── if question != "" ──────────────────────┐             │
│       │                                           │             │
│       └── else ──▶ app.Run(ctx) ──▶ TUI          │             │
│                                                   │             │
│                                                   ▼             │
│                                          pkg/extcli/chat.go    │
│                                                   │             │
│                                                   ▼             │
│                                          RegisterStreamOutput() │
│                                                   │             │
│                                                   ▼             │
│                                          agent.Ask()            │
│                                                   │             │
│                                                   ▼             │
│                                   StreamManager.Broadcast()      │
│                                                   │             │
│                                                   ▼             │
│                                          流式输出到 stdout       │
└─────────────────────────────────────────────────────────────────┘
```

## 文件变更

| 文件 | 变更类型 | 描述 |
|------|----------|------|
| `cmd/root.go` | 修改 | 添加 `-q`、`-m` 参数，路由逻辑 |
| `pkg/extcli/chat.go` | 新增 | CLI 流式输出模块 |

## 迁移计划

无需迁移，纯新增功能，向后兼容。