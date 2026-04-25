# 规格：tui-segment-display

TUI 输出按类型分段显示，带 emoji 前缀和分割线

## 概述

为不同类型的输出内容提供视觉区分，包括 emoji 前缀、颜色样式和分割线，让用户能快速识别当前阶段。

## 行为规格

### 类型样式定义

| 类型 | Emoji | 颜色 | 样式属性 |
|-----|-------|------|---------|
| 思考 (reason) | 💭 | 灰色 (#9ca3af) | 斜体 |
| 工具 (tool) | ⚙️ (请求) / ✅ (结果) | 黄色 (#FFED4E) | 正常 |
| 正文 (text) | 💬 | 白色 (#ffffff) | Markdown 渲染 |

### 工具输出格式

**工具请求** (`EventToolStart`)：
```
⚙️ 工具: 调用 {tool_name}
参数: {tool_input}
```

**工具结果** (`EventToolResult`)：
```
✅ 工具: {tool_name} 执行完成
```

**工具错误** (`EventToolError`)：
```
❌ 工具: {tool_name} 执行失败: {error_message}
```

### 分割线规格

| 属性 | 值 |
|-----|-----|
| 内容 | `────────────────────────────` |
| 颜色 | 灰色 (#555555) |
| 插入时机 | 类型切换时，在 flush 的内容之后 |
| 长度 | 固定 28 字符或根据终端宽度自适应 |

### 输出示例

```
💭 思考: 正在分析股票走势，需要获取实时数据...
────────────────────────────
⚙️ 工具: 调用 get_stock_info
参数: {"code": "000001"}
────────────────────────────
✅ 工具: get_stock_info 执行完成
────────────────────────────
💬 正文: 根据分析结果，当前股价处于上升通道...
```

## 样式实现

### 新增样式 (`pkg/tui/style/style.go`)

```go
// 分割线样式
ChatDividerStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#555555"))

// 各类型前缀样式
ChatThinkingPrefix  = "💭 思考: "
ChatToolCallPrefix  = "⚙️ 工具: "
ChatToolResultPrefix = "✅ 工具: "
ChatToolErrorPrefix = "❌ 工具: "
ChatTextPrefix      = "💬 正文: "
```

### 渲染逻辑 (`pkg/tui/chat.go`)

`renderPendingMessages()` 根据消息类型选择：
- 前缀 + 样式渲染
- Markdown 渲染（仅正文类型）

## 约束

- **分割线不侵入 Markdown**：分割线仅在消息之间，不影响正文内容的 Markdown 渲染
- **emoji 在渲染层添加**：emoji 前缀在 `renderPendingMessages()` 中添加，不修改消息内容本身
- **工具请求和结果共用类型**：`StreamMsgTypeTool` 不拆分，通过事件类型区分 emoji

## 验收标准

- [ ] 思考内容显示 💭 前缀和灰色斜体样式
- [ ] 工具请求显示 ⚙️ 前缀和黄色样式
- [ ] 工具结果显示 ✅ 前缀和黄色样式
- [ ] 正文内容显示 💬 前缀，并正确渲染 Markdown
- [ ] 类型切换时自动插入分割线
