# streaming-markdown-render 规格说明

## Requirement: markdown-render-on-flush

正文内容（StreamMsgTypeText）在 flush 时必须使用 Markdown 渲染器渲染。

### Scenario: 正文 Markdown 正确渲染

- **WHEN** AI 输出包含 Markdown 格式的正文（如标题、列表、代码块）
- **AND** 流式输出结束，执行 flush
- **THEN** 正文内容使用 `RenderMarkdown()` 渲染
- **AND** 标题、列表、代码块等格式正确显示

### Scenario: 流式期间使用简单渲染

- **WHEN** AI 正在流式输出正文内容
- **AND** 内容尚未完整（流式进行中）
- **THEN** 使用带宽度限制的简单渲染显示
- **AND** 用户可以实时看到内容更新

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
