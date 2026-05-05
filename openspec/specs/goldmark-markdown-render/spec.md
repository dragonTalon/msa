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

支持以下 Markdown 元素的渲染：标题(H1-H6，层级视觉区分)、加粗、斜体、删除线、行内代码、围栏代码块、无序列表（含任务列表）、有序列表、引用块、链接（含 URL 显示）、自动链接、分割线、表格。

#### Scenario: 标题 H1-H6（层级区分）

- **GIVEN** 输入各级标题（`#` 到 `######`）
- **WHEN** 渲染后
- **THEN** H1-H6 均去除 `#` 标记符
- **AND** 各级标题使用不同样式以区分层级（通过颜色深浅或前缀缩进）
- **AND** 标题文本以 bold 渲染

#### Scenario: 删除线

- **GIVEN** 输入 `~~已废弃的功能~~`
- **WHEN** 渲染后
- **THEN** 不包含 `~~` 语法标记
- **AND** "已废弃的功能" 以 strikethrough 样式渲染

#### Scenario: 任务列表

- **GIVEN** 输入 `- [ ] 待办` 或 `- [x] 已完成`
- **WHEN** 渲染后
- **THEN** 输出包含 `[ ]` 或 `[x]` 复选框标记
- **AND** 复选框跟在列表 bullet 之后

#### Scenario: 围栏代码块语法高亮

- **GIVEN** 输入 `` ```go\nfunc main() {}\n``` ``
- **WHEN** 渲染后
- **THEN** 代码块以缩进展示
- **AND** 使用 chroma 进行 Go 语法高亮
- **AND** 不显示 `` ``` `` 标记

#### Scenario: 引用块

- **GIVEN** 输入 `> 重要提示内容`
- **WHEN** 渲染后
- **THEN** 输出中以竖线或颜色前缀标识引用
- **AND** 不显示 `> ` 标记符

#### Scenario: 链接显示文本和 URL

- **GIVEN** 输入 `[点击这里](https://example.com)`
- **WHEN** 渲染后
- **THEN** 输出中显示链接文本 "点击这里"
- **AND** URL 以淡化样式在文本后显示
- **AND** 不显示原始 Markdown 方括号语法

#### Scenario: 自动链接

- **GIVEN** 输入包含裸 URL `https://example.com`
- **WHEN** 渲染后
- **THEN** URL 以链接样式渲染为可见文本

#### Scenario: 表格渲染

- **GIVEN** 输入 GFM 表格
- **WHEN** 渲染后
- **THEN** 表格以结构化格式显示
- **AND** 表头行以 bold 渲染
- **AND** 列宽按内容对齐

---

### Requirement: lipgloss-style-mapping

每种 Markdown 元素映射到预定义的 lipgloss 样式。包含 strikethrough 样式和 H1-H6 分级标题样式。

#### Scenario: 使用项目已有样式变量

- **GIVEN** `style/style.go` 中已有的样式变量
- **WHEN** 定义 MD 元素样式
- **THEN** 新样式基于现有颜色体系（`PrimaryColor`, `ToolColor`, `SecondaryColor` 等）
- **AND** 新增 `MDStrikethroughStyle`、`MDH1Style`~`MDH6Style`、`MDLinkURLStyle` 变量
- **AND** 通过 `var` 变量暴露，便于统一管理

#### Scenario: 样式可在外部覆盖

- **GIVEN** MD 样式定义为 `var` 变量
- **WHEN** 需要调整样式
- **THEN** 可在运行时修改这些变量（如根据终端主题切换）
