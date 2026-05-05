# autolink-render 规格说明

## Purpose

为 goldmark markdown 渲染器添加 GFM 自动链接渲染支持，使裸 URL 和邮箱地址以链接样式可见渲染。

## Requirements

### Requirement: autolink-rendering

goldmark 渲染器 SHALL 处理 GFM 扩展产生的自动链接节点，将裸 URL 渲染为可见文本。

#### Scenario: 裸 URL 渲染

- **GIVEN** 输入包含 `https://example.com`
- **WHEN** 调用 `RenderMarkdown(content)`
- **THEN** URL 文本被渲染为可见输出
- **AND** URL 以可区分的样式显示（如下划线颜色）

#### Scenario: 邮箱自动链接

- **GIVEN** 输入包含 `user@example.com`
- **WHEN** 渲染后
- **THEN** 邮箱地址以链接样式渲染

#### Scenario: URL 在段落中

- **GIVEN** 输入 `请访问 https://example.com 了解更多`
- **WHEN** 渲染后
- **THEN** URL 部分以链接样式渲染
- **AND** 周围文本以正文样式渲染
