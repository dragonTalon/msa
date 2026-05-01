# streaming-markdown-render 规格说明

## Requirement: markdown-render-on-flush

正文内容（StreamMsgTypeText）在流式期间和 flush 时均使用 goldmark + lipgloss 渲染器渲染，消除 Markdown 语法标记符。

### Scenario: 正文 Markdown 正确渲染

- **WHEN** AI 输出包含 Markdown 格式的正文（如标题、列表、代码块）
- **AND** 流式输出结束，执行 flush
- **THEN** 正文内容使用 goldmark 渲染器渲染
- **AND** 标题、加粗、列表、代码块等格式正确显示
- **AND** 不保留 `####`、`**`、`*` 等语法标记符

### Scenario: 流式期间使用 Markdown 渲染

- **WHEN** AI 正在流式输出正文内容
- **AND** 收到每个 chunk
- **THEN** 将完整 buffer（累积内容）传入 goldmark 渲染器
- **AND** 渲染结果更新到 `streamingMsg` 显示
- **AND** 用户实时看到格式化后的内容，不看到 `####` 等原始标记

### Scenario: 流式期间不完整语法处理

- **WHEN** 流式 buffer 中包含未闭合的 Markdown 语法（如 `**加粗文字` 缺少闭合 `**`）
- **THEN** 未闭合的语法标记作为纯文本显示
- **AND** 当闭合标记到达后，内容自动转为正确的格式化样式
- **AND** 不出现渲染异常或崩溃

### Scenario: 换行边界渲染触发

- **WHEN** 当前 chunk 包含 `\n` 换行符
- **THEN** 触发一次完整的 buffer 重渲染
- **AND** 未包含换行符的 chunk 仅累积，不触发重渲染
- **AND** 减少渲染频率，优化性能

---

## Requirement: non-text-no-markdown

思考内容和工具内容不使用 Markdown 渲染。

### Scenario: 思考内容不渲染 Markdown

- **WHEN** AI 输出思考过程（StreamMsgTypeReason）
- **THEN** 使用 `ChatReasonMsgStyle.Render()` 渲染
- **AND** 不进行 Markdown 解析

### Scenario: 工具内容不渲染 Markdown

- **WHEN** 显示工具调用信息（StreamMsgTypeTool）
- **THEN** 使用 `ChatToolMsgStyle.Render()` 渲染
- **AND** 不进行 Markdown 解析

---

## 验收标准

- [ ] 正文内容（StreamMsgTypeText）flush 后正确渲染 Markdown
- [ ] 思考内容（StreamMsgTypeReason）不渲染 Markdown
- [ ] 工具内容（StreamMsgTypeTool）不渲染 Markdown
- [ ] 流式期间内容实时显示，flush 后格式正确
