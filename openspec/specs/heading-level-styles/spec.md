# heading-level-styles 规格说明

## Purpose

为 goldmark markdown 渲染器添加标题层级视觉区分，使 H1-H6 各级标题通过不同样式呈现层级结构。

## Requirements

### Requirement: heading-level-distinction

goldmark 渲染器 SHALL 根据标题级别（H1-H6）应用不同的 lipgloss 样式，使层级结构可视化。

#### Scenario: H1 和 H2 视觉区分

- **GIVEN** 输入包含 `# 大标题` 和 `## 小标题`
- **WHEN** 调用 `RenderMarkdown(content)`
- **THEN** H1 "大标题" 和 H2 "小标题" 使用不同的样式渲染
- **AND** 用户可区分两个标题的层级

#### Scenario: 所有六级标题

- **GIVEN** 输入包含 `# H1` 到 `###### H6` 六个标题
- **WHEN** 渲染后
- **THEN** 六个标题级别有视觉差异
- **AND** 层级通过字号、颜色深浅或前缀缩进区分
- **AND** 所有标题均为 bold
- **AND** 不保留 `#` 标记

#### Scenario: 标题样式可配置

- **GIVEN** 各级标题样式定义为 `var` 变量（如 `MDH1Style` 到 `MDH6Style`）
- **WHEN** 需要在外部调整标题样式
- **THEN** 可在运行时修改这些变量
