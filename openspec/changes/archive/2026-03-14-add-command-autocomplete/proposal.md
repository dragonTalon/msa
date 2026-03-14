# 提案：命令自动补全

## Why

在 TUI 聊天界面中输入 `/` 命令时，用户需要完整输入命令名称，体验不够友好。现有实现虽然会显示匹配的命令列表，但该列表是只读的，用户无法通过键盘选择或自动补全，增加了输入负担。

## What Changes

- 新增 Tab 键自动补全功能：当用户输入 `/` 开头的命令前缀时，按 Tab 键可补全选中的命令
- 新增 `↑/↓` 键命令选择功能：在命令建议列表中上下移动选择
- 改进命令列表显示：每个命令显示名称和描述信息
- 改进补全行为：补全后自动添加空格，光标移至末尾，方便输入参数

## Capabilities

### New Capabilities

- `command-autocomplete`: 定义命令自动补全的交互行为、按键映射和显示格式

### Modified Capabilities

- `agent-chat`: 扩展聊天输入框以支持命令自动补全交互

## Impact

- `pkg/logic/command/cmd.go` - 新增 `GetCommandSuggestions()` 函数，返回带描述的命令建议
- `pkg/tui/chat.go` - 扩展 `Chat` 结构体，处理 Tab/↑/↓ 按键，渲染命令建议列表
- `pkg/tui/style/style.go` - 新增命令建议列表的样式定义

## Non-Goals

- 不实现参数补全（命令补全后可继续输入参数，但不会提示参数选项）
- 不实现历史命令记录
- 不实现命令别名