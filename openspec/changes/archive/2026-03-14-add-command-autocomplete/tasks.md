# 实现任务：命令自动补全

## 1. 数据结构定义

- [x] 1.1 在 `pkg/logic/command/cmd.go` 中定义 `CommandSuggestion` 结构体
  - 包含 `Name string` 和 `Description string` 字段
- [x] 1.2 实现 `GetCommandSuggestions(prefix string) []CommandSuggestion` 函数
  - 复用现有 `commandMap` 获取命令及其描述
  - 支持前缀过滤，按名称排序返回

## 2. Chat 结构体扩展

- [x] 2.1 在 `pkg/tui/chat.go` 中扩展 `Chat` 结构体
  - 添加 `commandSelectedIndex int` 字段
  - 添加 `commandSuggestions []command.CommandSuggestion` 字段
- [x] 2.2 修改 `NewChat()` 初始化新字段
  - `commandSelectedIndex` 初始化为 0
  - `commandSuggestions` 初始化为空切片

## 3. 按键处理逻辑

- [x] 3.1 在 `Update()` 方法中添加 `tea.KeyTab` 处理
  - 检查是否在命令模式且建议列表非空
  - 调用补全函数填充输入框
  - 退出命令模式
- [x] 3.2 在 `Update()` 方法中添加 `tea.KeyUp` 处理
  - 仅在命令模式下响应
  - 减少 `commandSelectedIndex`（不小于 0）
- [x] 3.3 在 `Update()` 方法中添加 `tea.KeyDown` 处理
  - 仅在命令模式下响应
  - 增加 `commandSelectedIndex`（不超过列表长度-1）
- [x] 3.4 修改命令模式检测逻辑
  - 使用 `GetCommandSuggestions()` 替代 `GetLikeCommand()`
  - 更新 `commandSuggestions` 字段
  - 重置 `commandSelectedIndex` 为 0

## 4. 视图渲染

- [x] 4.1 修改 `View()` 方法中的命令建议渲染
  - 显示每个建议的名称和描述
  - 使用 `>` 前缀或颜色高亮选中项
  - 显示底部快捷键提示
- [x] 4.2 在 `pkg/tui/style/style.go` 中添加样式定义
  - `CommandSuggestionSelectedStyle` - 选中项样式
  - `CommandSuggestionNormalStyle` - 普通项样式
  - `CommandSuggestionDescStyle` - 描述文本样式

## 5. 验证和测试

- [x] 5.1 运行 `go vet ./...` 验证编译
- [x] 5.2 手动测试命令补全功能
  - 输入 `/` 显示所有命令
  - 输入 `/re` 过滤匹配命令
  - 使用 `↑/↓` 导航
  - 使用 Tab 补全
  - 验证补全后光标位置正确
- [x] 5.3 测试边界情况
  - 无匹配命令时不显示列表
  - 单个匹配时 Tab 直接补全
  - 非命令模式 Tab 无效