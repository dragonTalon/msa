# streaming-text-wrap 规格说明

## Requirement: auto-width-wrap

流式输出内容必须根据终端宽度自动换行，避免内容溢出屏幕。

### Scenario: 长文本自动换行

- **WHEN** AI 输出超过终端宽度的长文本（如思考过程）
- **AND** TUI 正在流式显示该内容
- **THEN** 内容自动在终端宽度处换行
- **AND** 用户可以看到完整内容，无溢出

### Scenario: 终端宽度动态变化

- **WHEN** 用户调整终端窗口大小
- **AND** 流式输出正在进行
- **THEN** 新输出内容使用新的终端宽度
- **AND** 已显示内容保持不变（不重新渲染）

---

## Requirement: width-calculation

流式输出的可用宽度必须预留边距，确保内容不与输入框重叠。

### Scenario: 宽度计算包含边距

- **WHEN** 计算流式输出的渲染宽度
- **THEN** 使用 `terminalWidth - margin` 作为有效宽度
- **AND** margin 至少为 10 个字符
- **AND** 最小有效宽度为 20 个字符

---

## 验收标准

- [ ] 长文本（超过 100 字符）在流式显示时自动换行
- [ ] 调整终端窗口后，新内容使用新宽度
- [ ] 内容不与输入框重叠
