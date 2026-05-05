# tasklist-render 规格说明

## Purpose

为 goldmark markdown 渲染器添加 GFM 任务列表渲染支持，使 `- [ ]` / `- [x]` 语法正确渲染为带复选框标记的列表。

## Requirements

### Requirement: task-list-rendering

goldmark 渲染器 SHALL 处理 GFM 扩展产生的任务列表节点（`TaskCheckBox`），在列表项前显示 `[ ]` 或 `[x]` 复选框标记。

#### Scenario: 未完成任务

- **GIVEN** 输入 `- [ ] 买入 AAPL`
- **WHEN** 调用 `RenderMarkdown(content)`
- **THEN** 输出包含 `[ ]` 复选框标记
- **AND** "买入 AAPL" 以正文样式渲染
- **AND** 不显示 `- [ ] ` 原始语法

#### Scenario: 已完成任务

- **GIVEN** 输入 `- [x] 卖出 GOOGL`
- **WHEN** 渲染后
- **THEN** 输出包含 `[x]` 复选框标记
- **AND** "卖出 GOOGL" 以正文样式渲染

#### Scenario: 混合任务列表

- **GIVEN** 输入包含多个任务项：
  ```
  - [ ] 待办事项1
  - [x] 已完成事项
  - [ ] 待办事项2
  ```
- **WHEN** 渲染后
- **THEN** 每个列表项以 `• ` 开头
- **AND** 每项显示对应的复选框状态 `[ ]` 或 `[x]`
- **AND** 复选框紧跟在 bullet 之后
