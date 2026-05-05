# markdown-table-render 规格说明

## Purpose

为 goldmark markdown 渲染器添加 GFM 表格渲染支持，使表格内容以结构化格式显示，而非丢失表格布局。

## Requirements

### Requirement: table-rendering

goldmark 渲染器 SHALL 处理 GFM 扩展产生的表格相关 AST 节点（KindTable、KindTableHeader、KindTableRow、KindTableCell），并输出为 lipgloss 样式格式化的表格。

#### Scenario: 渲染简单表格

- **GIVEN** LLM 输出了 Markdown 表格：
  ```markdown
  | 项目 | 数值 |
  |------|------|
  | 初始资金 | 500,000.00 元 |
  | 当前总资产 | 493,233.00 元 |
  ```
- **WHEN** 调用 `RenderMarkdown(content)`
- **THEN** 输出包含结构化的表格内容
- **AND** "项目" 和 "数值" 以表头样式（bold）渲染
- **AND** "500,000.00 元" 和 "493,233.00 元" 以正文样式渲染
- **AND** 列之间以固定分隔符（如 `|` 或空格）分隔

#### Scenario: 表格内联样式保留

- **GIVEN** 表格单元格包含内联 Markdown 格式（如 `**500,000.00 元**`、`*备注*`）
- **WHEN** 渲染表格
- **THEN** 单元格内的 `**bold**` 以 bold 样式渲染
- **AND** 单元格内的 `*italic*` 以 italic 样式渲染
- **AND** 不显示 `**` 和 `*` 语法标记

#### Scenario: 多列表格

- **GIVEN** 一个包含 5 列的表格
- **WHEN** 渲染表格
- **THEN** 所有列都被渲染
- **AND** 长内容不会被截断

#### Scenario: 空表格

- **GIVEN** 一个只有表头没有数据行的表格
- **WHEN** 渲染表格
- **THEN** 仅显示表头行
- **AND** 不发生 panic 或错误

---

### Requirement: table-style-definition

表格渲染使用专用 lipgloss 样式变量，与现有样式体系一致。

#### Scenario: 表头加粗

- **GIVEN** 表格包含表头行
- **WHEN** 渲染表头单元格
- **THEN** 表头文本以 `MDBoldStyle` 渲染
- **AND** 数据行以 `ChatNormalMsgStyle` 渲染

#### Scenario: 表格样式可配置

- **GIVEN** `MDTableHeaderStyle` 和 `MDTableCellStyle` 定义为 `var` 变量
- **WHEN** 需要在外部调整表格样式
- **THEN** 可在运行时修改这些变量
