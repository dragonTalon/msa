# tui-event-handling 规格说明

## Requirement: event-loop-continuation

当 `handleStreamContent()` 检测到消息段类型切换（如 text→tool、text→reason）时，
必须在 flush 当前段内容的同时继续读取下一个事件，确保事件循环不中断。

### Scenario: 文本段切换到工具段

- **WHEN** 用户输入触发工具调用（如 "查询茅台股价"）
- **AND** TUI 先收到多个 `EventTextChunk`，然后收到 `EventToolStart`
- **THEN** `handleStreamContent()` 必须使用 `tea.Batch(c.Flush(), c.receiveNextChunk())` 同时 flush 和继续读取
- **AND** 后续所有事件（`EventToolResult`、新的 `EventTextChunk` 等）均被正确处理

### Scenario: 思考段切换到文本段

- **WHEN** AI 模型先输出 thinking 内容，再输出正文
- **AND** TUI 先收到多个 `EventThinking`，然后收到 `EventTextChunk`
- **THEN** 段切换时 flush 思考内容后事件循环继续
- **AND** 正文内容被完整渲染

### Scenario: 工具段切换到文本段

- **WHEN** 工具执行完成后 AI 继续输出文本
- **AND** TUI 先收到 `EventToolResult`，然后收到 `EventTextChunk`
- **THEN** 段切换时 flush 工具结果后事件循环继续
- **AND** 后续文本内容完整显示

---

## Requirement: no-duplicate-tool-start-event

`EventToolStart` 事件必须仅由 `Agent.executeTools()` 发送一次，且携带完整的 `Tool` 字段。
`StreamAdapter.Process()` 不得发送 `EventToolStart` 事件。

### Scenario: 标准工具调用事件序列

- **WHEN** LLM 返回的 chunk 包含 `ToolCalls`
- **AND** `StreamAdapter.Process()` 处理该 chunk
- **THEN** StreamAdapter 仅收集 tool calls 到 `toolCalls` 切片，不发送任何 `EventToolStart` 事件
- **AND** 后续 `Agent.executeTools()` 发送一次完整的 `EventToolStart`（包含 `Tool.ID`、`Tool.Name`、`Tool.Input`）

### Scenario: TUI 接收 EventToolStart 时 Tool.Name 有值

- **WHEN** TUI 的 `handleEvent()` 收到 `EventToolStart` 事件
- **THEN** `e.Tool.Name` 必须为非空字符串
- **AND** 工具消息正确渲染为 "调用 {tool_name}"

---

## Requirement: error-event-channel-drain

当收到 `EventError` 时，必须在显示错误消息后继续消费 channel 中的剩余事件，
直到 channel 关闭（收到 `EventRoundDone` 或 channel 被 close），确保无 goroutine 泄漏。

### Scenario: 错误发生后 channel 正确关闭

- **WHEN** Agent 发送 `EventError` 到 event channel
- **AND** channel 中可能还有未发送的事件
- **THEN** TUI 显示错误消息后继续读取 channel
- **AND** channel 最终被正确关闭和清理
- **AND** `clearStreamState()` 在 `EventRoundDone` 时被调用

### Scenario: 错误后用户可继续对话

- **WHEN** TUI 处理完 `EventError` 和后续事件
- **THEN** `isStreaming` 被重置为 false
- **AND** 输入框重新聚焦
- **AND** 用户可以输入新问题

---

## 验收标准

- [ ] 段切换后事件循环不中断（`tea.Batch` 修复）
- [ ] `EventToolStart` 仅发送一次，且 `Tool.Name` 非空
- [ ] 错误事件后 channel 被完整消费
- [ ] 纯文本对话完整显示
- [ ] 思考+文本对话完整显示
- [ ] 文本+工具调用完整显示
- [ ] 多工具连续调用完整显示
