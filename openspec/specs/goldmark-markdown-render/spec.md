# goldmark-markdown-render 规格说明

## Purpose

使用 goldmark AST 解析器 + lipgloss 样式映射替换 glamour，实现无语法标记符残留的 Markdown 渲染。

## Requirements

### Requirement: goldmark-as-parser

Markdown 渲染器使用 `yuin/goldmark` 解析文本为 AST，不再使用 glamour。

#### Scenario: goldmark 解析 Markdown 文本

- **GIVEN** 一段包含 Markdown 格式的文本（标题、加粗、斜体、列表、代码块等）
- **WHEN** 调用 `RenderMarkdown(content)`
- **THEN** 使用 goldmark 将文本解析为 AST 节点树
- **AND** 根据节点类型应用对应的 lipgloss 样式
- **AND** 返回带有 ANSI 转义码的格式化字符串

#### Scenario: 解析失败时降级

- **GIVEN** goldmark 解析失败
- **WHEN** 调用 `RenderMarkdown(content)`
- **THEN** 回退为原始文本
- **AND** 使用 `ChatNormalMsgStyle` 渲染原始文本
- **AND** 记录 logrus 警告日志

---

### Requirement: no-syntax-markers

渲染后的输出中不保留 Markdown 语法标记符（`####`、`**`、`*`、`` ` ``、`-` 等）。

#### Scenario: 标题不保留 # 标记

- **GIVEN** 输入 `#### 股票分析报告`
- **WHEN** 渲染后
- **THEN** 输出中不包含 `####`
- **AND** "股票分析报告" 以 bold + 主题色渲染（如 `PrimaryColor`）

#### Scenario: 加粗不保留 ** 标记

- **GIVEN** 输入 `**强烈买入**`
- **WHEN** 渲染后
- **THEN** 输出中不包含 `**`
- **AND** "强烈买入" 以 bold 样式渲染

#### Scenario: 斜体不保留 * 标记

- **GIVEN** 输入 `*风险提示*`
- **WHEN** 渲染后
- **THEN** 输出中不包含 `*`
- **AND** "风险提示" 以 italic 样式渲染

#### Scenario: 行内代码不保留反引号

- **GIVEN** 输入 `` `AAPL` ``
- **WHEN** 渲染后
- **THEN** 输出中不包含 `` ` ``
- **AND** "AAPL" 以代码样式渲染（灰底或异色）

#### Scenario: 无序列表转换子弹

- **GIVEN** 输入 `- 买入信号`
- **WHEN** 渲染后
- **THEN** 输出中 `- ` 替换为 `• `
- **AND** "买入信号" 保持正文样式

---

### Requirement: supported-markdown-elements

支持以下 Markdown 元素的渲染：标题(H1-H6)、加粗、斜体、行内代码、围栏代码块、无序列表、有序列表、引用块、链接、分割线、表格。

#### Scenario: 标题 H1-H6

- **GIVEN** 输入各级标题（`#` 到 `######`）
- **WHEN** 渲染后
- **THEN** H1-H6 均去除 `#` 标记符
- **AND** 标题文本以 bold 渲染
- **AND** 标题级别通过颜色深浅或前缀空格区分

#### Scenario: 围栏代码块语法高亮

- **GIVEN** 输入 `` ```go\nfunc main() {}\n``` ``
- **WHEN** 渲染后
- **THEN** 代码块以缩进展示
- **AND** 使用 chroma 进行 Go 语法高亮（chroma 为 go.mod 已有依赖）
- **AND** 不显示 `` ``` `` 标记

#### Scenario: 引用块

- **GIVEN** 输入 `> 重要提示内容`
- **WHEN** 渲染后
- **THEN** 输出中以竖线或颜色前缀标识引用
- **AND** 不显示 `> ` 标记符

#### Scenario: 链接仅显示文本

- **GIVEN** 输入 `[点击这里](https://example.com)`
- **WHEN** 渲染后
- **THEN** 输出中仅显示 "点击这里"（链接文本）
- **AND** 不显示 URL 和方括号

#### Scenario: 表格渲染

- **GIVEN** 输入 GFM 表格：
  ```markdown
  | 项目 | 数值 |
  |------|------|
  | 初始资金 | **500,000.00 元** |
  ```
- **WHEN** 渲染后
- **THEN** 表格以结构化格式显示，列之间以竖线分隔
- **AND** 表头行以 bold 渲染
- **AND** 数据行以正文样式渲染
- **AND** 不显示对齐分隔符行（`|------|------|`）
- **AND** 单元格内的 `**bold**` 正确渲染为 bold 样式

---

### Requirement: lipgloss-style-mapping

每种 Markdown 元素映射到预定义的 lipgloss 样式。

#### Scenario: 使用项目已有样式变量

- **GIVEN** `style/style.go` 中已有的样式变量（`ChatNormalMsgStyle`, `ChatToolMsgStyle` 等）
- **WHEN** 定义 MD 元素样式
- **THEN** 新样式基于现有颜色体系（`PrimaryColor`, `ToolColor`, `SecondaryColor` 等）
- **AND** 通过新增专用变量（如 `MDHeadingStyle`, `MDBoldStyle`, `MDCodeStyle`）暴露，便于统一管理

#### Scenario: 样式可在外部覆盖

- **GIVEN** MD 样式定义为 `var` 变量
- **WHEN** 需要调整样式
- **THEN** 可在运行时修改这些变量（如根据终端主题切换）
