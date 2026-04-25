# 规格：tui-realtime-flush

TUI 输出类型切换时即时刷新到控制台

## 概述

在 TUI 模式下，当输出内容类型（思考/工具/正文）发生变化时，立即将已累积的内容刷新到控制台，而不是等待整个会话结束。

## 行为规格

### 输入

| 输入 | 类型 | 描述 |
|-----|------|------|
| `event.Event` | Event | 来自 Agent 的事件流 |
| `currentSegmentType` | StreamMsgType | 当前累积内容的类型 |

### 输出

| 输出 | 类型 | 描述 |
|-----|------|------|
| Flush 操作 | tea.Cmd | 将内容写入控制台的命令 |

### 触发条件

当收到的事件类型与 `currentSegmentType` 不同时触发 flush：

```
┌───────────────────────────────────────────────────────────┐
│  当前类型      │  新事件类型      │  动作                  │
├───────────────────────────────────────────────────────────┤
│  reason        │  tool            │  flush reason → 显示   │
│  tool          │  text            │  flush tool → 显示     │
│  text          │  reason          │  flush text → 显示     │
│  text          │  tool            │  flush text → 显示     │
│  (空)          │  any             │  开始累积，不 flush     │
│  same          │  same            │  继续累积，不 flush     │
└───────────────────────────────────────────────────────────┘
```

### 事件类型映射

| Event 类型 | StreamMsgType |
|-----------|---------------|
| `EventThinking` | `reason` |
| `EventToolStart` | `tool` |
| `EventToolResult` | `tool` |
| `EventTextChunk` | `text` |

### 流程

```
事件到达
    │
    ▼
判断事件类型是否等于 currentSegmentType
    │
    ├─ 是 ─▶ 继续累积到 fullStreamContent
    │
    └─ 否 ─▶
             ├─ 将 fullStreamContent 内容加入 pendingMsgs
             ├─ 调用 Flush() 输出到控制台
             ├─ 重置 fullStreamContent
             ├─ 更新 currentSegmentType
             └─ 开始累积新类型内容
```

## 约束

- **不阻塞事件流**：Flush 操作必须是异步的（通过 `tea.Cmd`），不能阻塞事件消费
- **不修改历史累积**：`allStreamContent`（跨 segment 累积）保持不变，用于最终历史记录
- **保持流式状态**：`isStreaming` 标志在整个会话期间保持 true，仅在 `EventRoundDone` 时重置

## 验收标准

- [ ] 类型切换时，已累积内容立即显示在控制台
- [ ] 同类型内容连续到达时，不会频繁 flush（避免闪烁）
- [ ] `EventRoundDone` 时仍然执行最终 flush 和历史记录更新
- [ ] 用户输入框在整个会话期间保持禁用状态
