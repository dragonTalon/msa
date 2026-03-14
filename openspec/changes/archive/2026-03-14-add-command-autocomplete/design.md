# 设计：命令自动补全

## Context

当前 MSA 的 TUI 聊天界面使用 Bubble Tea 框架的 `textinput` 组件处理用户输入。当用户输入 `/` 开头的命令时，系统会调用 `command.GetLikeCommand()` 获取匹配的命令列表，并在 `View()` 中显示。

现有实现的问题：
- 命令列表只读，用户无法交互选择
- 缺少 Tab 补全功能
- 缺少键盘导航支持

## Goals / Non-Goals

**Goals:**
- 实现 Tab 键自动补全选中的命令
- 实现 `↑/↓` 键在命令列表中导航
- 显示命令名称和描述信息
- 补全后自动添加空格，方便输入参数

**Non-Goals:**
- 不实现参数补全
- 不实现历史命令记录
- 不实现命令别名

## Decisions

### 决策 1：扩展 Chat 结构体 vs 新建组件

**选择**：扩展 Chat 结构体

**理由**：
- 改动最小，逻辑集中
- 复用现有的 `isCommandMode` 和 `commandList` 字段
- 避免引入新的组件间通信复杂度

**备选方案**：创建独立的 `AutocompleteInput` 组件
- 优点：可复用，关注点分离
- 缺点：需要重构 Chat，改动较大

### 决策 2：命令建议数据结构

**选择**：新增 `CommandSuggestion` 结构体

```go
type CommandSuggestion struct {
    Name        string
    Description string
}
```

**理由**：现有 `GetLikeCommand()` 只返回 `[]string`，无法携带描述信息

### 决策 3：补全行为

**选择**：补全后添加空格，光标移至末尾

**理由**：方便用户继续输入命令参数

### 决策 4：选中项高亮样式

**选择**：使用 `>` 前缀 + 不同颜色

```text
> remember    打开记忆浏览器
  models      List all models
```

## Risks / Trade-offs

### 风险 1：中文输入兼容性

**风险**：`↑/↓` 键可能与输入法冲突

**缓解**：
- 使用 `tea.KeyUp`/`tea.KeyDown` 类型检测，而非字符检测
- 参考记忆系统中已修复的中文输入处理模式

### 风险 2：状态管理复杂度

**风险**：Chat 结构体变大，状态增多

**缓解**：
- 状态变化仅限于命令模式激活时
- 清晰的状态转换：正常模式 → 命令模式 → 补全/执行 → 正常模式