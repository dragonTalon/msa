# strikethrough-render 规格说明

## Purpose

为 goldmark markdown 渲染器添加 GFM 删除线渲染支持，使 `~~text~~` 语法正确渲染为带删除线样式的终端输出。

## Requirements

### Requirement: strikethrough-rendering

goldmark 渲染器 SHALL 处理 GFM 扩展产生的 `KindStrikethrough` AST 节点，将其渲染为带删除线样式的终端输出。

#### Scenario: 基本删除线渲染

- **GIVEN** 输入包含 `~~删除的文本~~`
- **WHEN** 调用 `RenderMarkdown(content)`
- **THEN** 输出中不包含 `~~` 语法标记
- **AND** "删除的文本" 以删除线样式（strikethrough）渲染
- **AND** 周围文本正常渲染

#### Scenario: 多段删除线

- **GIVEN** 输入包含 `这是**加粗**和~~删除~~混合`
- **WHEN** 渲染后
- **THEN** "加粗" 以 bold 样式渲染
- **AND** "删除" 以 strikethrough 样式渲染
- **AND** 格式标记均被移除

#### Scenario: 表格内的删除线

- **GIVEN** 表格单元格包含 `~~已废弃~~`
- **WHEN** 渲染表格
- **THEN** 单元格内 "已废弃" 以 strikethrough 样式渲染
- **AND** `~~` 标记被移除
