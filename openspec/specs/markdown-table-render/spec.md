# markdown-table-render 规格说明

## Purpose

为 goldmark markdown 渲染器添加 GFM 表格渲染支持，使表格内容以结构化格式显示。实现两遍扫描列宽计算算法，使各列垂直对齐。

## Requirements

### Requirement: table-rendering

goldmark 渲染器 SHALL 处理 GFM 表格 AST 节点，并输出为 lipgloss 样式格式化的表格。列宽 SHALL 基于各列内容最大宽度计算，使各列垂直对齐。

#### Scenario: 渲染简单表格（列对齐）

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
- **AND** 列之间以固定分隔符（`│`）分隔
- **AND** 各列垂直对齐，"数值"列下各数字在同一列对齐

#### Scenario: 不等宽列对齐

- **GIVEN** 表格不同行的同列内容长度不同：
  ```markdown
  | ID | 描述 |
  |----|------|
  | 1 | 短 |
  | 2 | 这是一个非常长的描述文本 |
  ```
- **WHEN** 渲染表格
- **THEN** "描述" 列宽度 = 最长单元格内容宽度
- **AND** 短内容在列内按对齐方式填充空格
- **AND** 所有行在该列宽度一致

#### Scenario: 对齐方式支持

- **GIVEN** 表格包含列对齐定义（`:---` 左对齐、`:---:` 居中、`---:` 右对齐）
- **WHEN** 渲染表格
- **THEN** 表头和数据单元格内容根据对齐方式左对齐/居中/右对齐

#### Scenario: 表格内联样式保留

- **GIVEN** 表格单元格包含内联 Markdown 格式（如 `**500,000.00 元**`、`*备注*`、`~~已废弃~~`）
- **WHEN** 渲染表格
- **THEN** 单元格内的 `**bold**` 以 bold 样式渲染
- **AND** 单元格内的 `*italic*` 以 italic 样式渲染
- **AND** 单元格内的 `~~strikethrough~~` 以 strikethrough 样式渲染
- **AND** 不显示语法标记

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

---

### Requirement: table-column-width-calculation

表格渲染器 SHALL 实现两遍扫描算法：第一遍收集所有单元格内容并计算每列最大显示宽度，第二遍按计算出的列宽渲染各行。

#### Scenario: 两遍扫描

- **GIVEN** 任意 GFM 表格
- **WHEN** 渲染表格
- **THEN** 第一遍扫描所有行收集原始单元格文本
- **AND** 计算每列最大宽度（基于 ANSI 剥离后的纯文本长度）
- **AND** 第二遍使用 lipgloss `Width()` 或 `Padding()` 按统一列宽渲染
- **AND** 所有输出行垂直对齐

#### Scenario: ANSI 转义码不影响宽度计算

- **GIVEN** 表格单元格内容包含 bold 样式（带 ANSI 转义码前缀）
- **WHEN** 计算列宽
- **THEN** ANSI 转义码长度不计入列宽
- **AND** 仅计算可见文本的长度
