# TUI Output Flush Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** TUI 输出类型切换时即时刷新到控制台，带 emoji 前缀和分割线分段显示。

**Architecture:** 利用 Bubble Tea 事件驱动机制，修改 `handleStreamContent` 在类型切换时返回 `Flush()` tea.Cmd，实现追加式输出。分割线作为 `RoleDivider` 独立消息项，在类型切换时插入。

**Tech Stack:** Go 1.21+, Bubble Tea, Lipgloss

---

## File Structure

```
pkg/model/comment.go     ← 新增 RoleDivider 常量
pkg/tui/style/style.go   ← 新增分割线样式和 emoji 前缀常量
pkg/tui/chat.go          ← 核心逻辑变更：handleStreamContent 签名、类型切换 flush、renderPendingMessages
```

---

## Task 1: 数据模型扩展

**Files:**
- Modify: `pkg/model/comment.go:90-95`

- [ ] **Step 1: 添加 RoleDivider 常量**

```go
// 在 MessageRole 常量定义区添加（第 91-95 行附近）
const (
	RoleLogo      MessageRole = "logo"      // Logo 显示
	RoleUser      MessageRole = "user"      // 用户消息
	RoleSystem    MessageRole = "system"    // 系统消息
	RoleAssistant MessageRole = "assistant" // AI 助手消息
	RoleDivider   MessageRole = "divider"   // 分割线（仅 UI，不持久化）
)
```

- [ ] **Step 2: 验证编译**

Run: `go vet ./pkg/model/...`
Expected: 无错误

- [ ] **Step 3: Commit**

```bash
git add pkg/model/comment.go
git commit -m "feat(model): add RoleDivider for UI segment separation"
```

---

## Task 2: 样式定义

**Files:**
- Modify: `pkg/tui/style/style.go`

- [ ] **Step 1: 添加分割线样式和 emoji 前缀常量**

在文件末尾（第 129 行后）添加：

```go
// ==================== 流式输出样式 ====================
var (
	// ChatDividerStyle 分割线样式
	ChatDividerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#555555"))

	// 流式输出 emoji 前缀
	ChatThinkingPrefix   = "💭 思考: "
	ChatToolCallPrefix   = "⚙️ 工具: "
	ChatToolResultPrefix = "✅ 工具: "
	ChatToolErrorPrefix  = "❌ 工具: "
	ChatTextPrefix       = "💬 正文: "
)

// DividerLine 分割线内容
const DividerLine = "────────────────────────────"
```

- [ ] **Step 2: 验证编译**

Run: `go vet ./pkg/tui/style/...`
Expected: 无错误

- [ ] **Step 3: Commit**

```bash
git add pkg/tui/style/style.go
git commit -m "feat(style): add divider style and emoji prefix constants"
```

---

## Task 3: 修改 handleStreamContent 签名

**Files:**
- Modify: `pkg/tui/chat.go:452`

- [ ] **Step 1: 修改函数签名，增加 prefix 参数**

将 `handleStreamContent` 函数签名从：

```go
func (c *Chat) handleStreamContent(content string, msgType model.StreamMsgType) (tea.Model, tea.Cmd) {
```

改为：

```go
func (c *Chat) handleStreamContent(content string, msgType model.StreamMsgType, prefix string) (tea.Model, tea.Cmd) {
```

- [ ] **Step 2: 更新函数内 prefix 使用**

将第 491 行：

```go
c.streamingMsg = style.ChatSystemMsgStyle.Render("🤖 MSA: ") +
```

改为：

```go
c.streamingMsg = style.ChatSystemMsgStyle.Render(prefix) +
```

- [ ] **Step 3: 验证编译**

Run: `go vet ./pkg/tui/...`
Expected: 编译错误（缺少参数）—— 这是预期的，下一步修复

---

## Task 4: 更新 handleEvent 调用方

**Files:**
- Modify: `pkg/tui/chat.go:418-444`

- [ ] **Step 1: 更新所有 handleStreamContent 调用，传入正确的前缀**

修改第 422-440 行：

```go
case event.EventTextChunk:
	if e.Text == "" {
		return c, c.receiveNextChunk()
	}
	return c.handleStreamContent(e.Text, model.StreamMsgTypeText, style.ChatTextPrefix)

case event.EventThinking:
	if e.Text == "" {
		return c, c.receiveNextChunk()
	}
	return c.handleStreamContent(e.Text, model.StreamMsgTypeReason, style.ChatThinkingPrefix)

case event.EventToolStart:
	toolMsg := fmt.Sprintf("调用 %s", e.Tool.Name)
	if e.Tool.Input != "" {
		toolMsg += fmt.Sprintf("\n参数: %s", e.Tool.Input)
	}
	return c.handleStreamContent(toolMsg, model.StreamMsgTypeTool, style.ChatToolCallPrefix)

case event.EventToolResult:
	toolMsg := fmt.Sprintf("%s 执行完成", e.Result.Name)
	return c.handleStreamContent(toolMsg, model.StreamMsgTypeTool, style.ChatToolResultPrefix)

case event.EventToolError:
	toolMsg := fmt.Sprintf("%s 执行失败: %v", e.Result.Name, e.Err)
	return c.handleStreamContent(toolMsg, model.StreamMsgTypeTool, style.ChatToolErrorPrefix)
```

- [ ] **Step 2: 验证编译**

Run: `go vet ./pkg/tui/...`
Expected: 无错误

- [ ] **Step 3: Commit**

```bash
git add pkg/tui/chat.go
git commit -m "feat(tui): update handleStreamContent calls with emoji prefixes"
```

---

## Task 5: 实现类型切换时 flush

**Files:**
- Modify: `pkg/tui/chat.go:456-496`

> 📚 **知识上下文**：设计决策 1 — 在 `currentSegmentType` 变化时 flush，而非每个事件到达时 flush。同类型事件连续到达时不触发 flush，避免频繁 IO。

- [ ] **Step 1: 重构 handleStreamContent 类型切换逻辑**

将第 456-496 行的函数体替换为：

```go
func (c *Chat) handleStreamContent(content string, msgType model.StreamMsgType, prefix string) (tea.Model, tea.Cmd) {
	isFirst := c.fullStreamContent.Len() == 0
	segmentChanged := !isFirst && c.currentSegmentType != "" && c.currentSegmentType != msgType

	// If segment type changed, flush previous segment
	if segmentChanged {
		prevContent := c.fullStreamContent.String()
		prevMsgType := c.currentSegmentType

		// 累积到历史（非 tool 类型）
		if prevContent != "" && prevMsgType != model.StreamMsgTypeTool {
			if c.allStreamContent.Len() > 0 {
				c.allStreamContent.WriteString("\n")
			}
			c.allStreamContent.WriteString(prevContent)
		}

		// 重置当前累积
		c.fullStreamContent.Reset()
		c.streamingMsg = ""

		// 将前一段内容加入 pendingMsgs（带回传的前缀信息）
		if prevContent != "" {
			c.addMessage(model.RoleAssistant, prevContent, prevMsgType)
		}

		// 插入分割线
		c.addMessage(model.RoleDivider, style.DividerLine, "")

		// 立即 flush 到控制台
		return c, c.Flush()
	}

	// 累积当前内容
	c.fullStreamContent.WriteString(content)
	c.currentSegmentType = msgType

	// 渲染流式内容
	var contentStyle lipgloss.Style
	switch msgType {
	case model.StreamMsgTypeTool:
		contentStyle = style.ChatToolMsgStyle
	case model.StreamMsgTypeReason:
		contentStyle = style.ChatReasonMsgStyle
	default:
		contentStyle = style.ChatNormalMsgStyle
	}

	if isFirst {
		c.streamingMsg = style.ChatSystemMsgStyle.Render(prefix) +
			contentStyle.Render(content) + "\n"
	} else {
		c.streamingMsg += contentStyle.Render(content)
	}
	return c, c.receiveNextChunk()
}
```

- [ ] **Step 2: 验证编译**

Run: `go vet ./pkg/tui/...`
Expected: 无错误

- [ ] **Step 3: Commit**

```bash
git add pkg/tui/chat.go
git commit -m "feat(tui): flush content on segment type change"
```

---

## Task 6: 修改 renderPendingMessages 添加 emoji 前缀

**Files:**
- Modify: `pkg/tui/chat.go:192-236`

- [ ] **Step 1: 修改 renderPendingMessages，按消息类型添加 emoji 前缀**

将第 211-226 行修改为：

```go
case model.RoleAssistant:
	// 根据消息类型选择样式和渲染方式
	switch msg.MsgType {
	case model.StreamMsgTypeTool:
		// 工具消息 - 黄色，不渲染 Markdown
		sb.WriteString(style.ChatToolMsgStyle.Render(msg.Content))
	case model.StreamMsgTypeReason:
		// 思考消息 - 灰色斜体，不渲染 Markdown
		sb.WriteString(style.ChatReasonMsgStyle.Render(msg.Content))
	case model.StreamMsgTypeText:
		// 正文消息 - 渲染 Markdown
		sb.WriteString(style.RenderMarkdown(msg.Content))
	default:
		// 默认 - 普通 Markdown 渲染
		sb.WriteString(style.RenderMarkdown(msg.Content))
	}
	// 记录当前消息类型
	lastMsgType = msg.MsgType
case model.RoleDivider:
	// 分割线 - 使用分隔线样式
	sb.WriteString(style.ChatDividerStyle.Render(msg.Content))
```

- [ ] **Step 2: 验证编译**

Run: `go vet ./pkg/tui/...`
Expected: 无错误

- [ ] **Step 3: Commit**

```bash
git add pkg/tui/chat.go
git commit -m "feat(tui): render pending messages with type-specific styles"
```

---

## Task 7: 确保会话持久化忽略 RoleDivider

**Files:**
- Verify: `pkg/tui/chat.go:400-410`
- Verify: `pkg/session/manager.go`

- [ ] **Step 1: 验证 EventRoundDone 处理不持久化 RoleDivider**

检查 `handleEvent` 中 `EventRoundDone` 分支（第 386-416 行），确认：
- 只将 `RoleAssistant` 类型的消息写入 `history`
- `RoleDivider` 不会被持久化

当前代码第 402-406 行：
```go
if allContent != "" {
	c.history = append(c.history, model.Message{
		Role:    model.RoleAssistant,
		Content: allContent,
		MsgType: model.StreamMsgTypeText,
	})
```

已满足要求 — 只持久化 `RoleAssistant`。

- [ ] **Step 2: 验证 session.AppendMessage 不受影响**

检查 `pkg/session/manager.go` 中的 `AppendMessage` 方法，确认其不依赖消息类型过滤。

- [ ] **Step 3: 基于知识的测试**

> 📚 **来源**：设计风险 — 分割线在会话持久化中泄漏

**测试场景**：
1. 启动 TUI，触发多类型输出（思考 → 工具 → 正文）
2. 观察分割线是否正确显示
3. 检查会话文件内容，确认无 `divider` 角色消息

- [ ] **Step 4: Commit（如需修改）**

```bash
git add pkg/session/manager.go  # 如有修改
git commit -m "fix(session): ensure RoleDivider is not persisted"
```

---

## Task 8: 完整验证与测试

- [ ] **Step 1: 全量编译验证**

Run: `go vet ./...`
Expected: 无错误

- [ ] **Step 2: 手动测试 TUI 交互**

启动 TUI：
```bash
go run main.go
```

测试场景：
1. 输入问题：`分析一下 000001 这支股票`
2. 观察输出：
   - 思考内容显示 💭 前缀和灰色斜体样式
   - 工具请求显示 ⚙️ 前缀和黄色样式
   - 工具结果显示 ✅ 前缀和黄色样式
   - 正文内容显示 💬 前缀，并正确渲染 Markdown
3. 类型切换时自动插入分割线

- [ ] **Step 3: 验证 flush 时机**

观察输出是否在类型切换时立即显示，而非等到会话结束。

- [ ] **Step 4: 最终 Commit**

```bash
git add -A
git commit -m "feat(tui): implement real-time flush with segment display

- Add RoleDivider for UI segment separation
- Add emoji prefix constants and divider style
- Flush content on segment type change
- Type-specific rendering with emoji prefixes
- Insert divider lines between segments"
```

---

## Self-Review Checklist

**1. Spec coverage:**
- [x] `tui-realtime-flush`: 类型切换时 flush → Task 5
- [x] `tui-segment-display`: emoji 前缀和分割线 → Task 2, 6
- [x] 不修改 Event 类型系统 → 已满足
- [x] 不修改 StreamMsgType 枚举 → 已满足

**2. Placeholder scan:**
- [x] 无 "TBD", "TODO", "implement later"
- [x] 无 "add appropriate error handling"
- [x] 所有代码步骤都有完整实现

**3. Type consistency:**
- `RoleDivider` 在 comment.go 定义，在 chat.go 使用 ✓
- `handleStreamContent` 签名变更后所有调用方已更新 ✓
- 样式常量名一致 (`ChatDividerStyle`, `DividerLine`) ✓

---

**Plan complete and saved to `docs/superpowers/plans/2026-04-25-tui-output-flush.md`. Two execution options:**

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

**Which approach?**
