# 修复：dynamic-skills 补充实现

---

id: "fix-001"
title: "TUI /skills 命令实现与 go vet 验证"
type: "compile"
severity: "medium"
component: "tui"
tags:
  - "tui"
  - "skills"
  - "command"
created_at: "2025-03-08T17:55:00Z"
resolved_at: "2025-03-08T18:00:00Z"
change: "dynamic-skills"
status: "resolved"
---

## 问题

### 问题描述

在 dynamic-skills 功能实现过程中，发现以下需要补充的内容：

1. **TUI 中缺少 `/skills` 命令**：用户在聊天界面无法快速查看可用的 Skills 列表
2. **编译验证方式不统一**：文档中使用了 `go build` 验证，但应该使用 `go vet` 进行静态分析

### 预期行为

- 用户可以在 TUI 中使用 `/skills` 命令查看所有可用的 Skills
- 代码验证使用 `go vet ./...` 而不是 `go build`

## 根本原因

### 调查过程

1. 检查 `pkg/logic/command/` 目录下的命令注册系统
2. 发现 `ListModel` 和 `ConfigCommand` 的实现模式
3. 确认缺少 SkillsCommand 的实现

### 根本原因

1. **命令缺失**：在实现 dynamic-skills 时，主要专注于 CLI 命令（`msa skills list`），但忽略了 TUI 内的 `/skills` 对话命令
2. **验证方式不当**：开发过程中使用了 `go build` 来验证编译，但 `go vet` 提供更好的静态分析检查

### 相关文件

- `pkg/logic/command/model.go`
  - 行：14-19（init 函数）
  - 说明：命令注册位置
- `pkg/tui/chat.go`
  - 行：45-46（帮助信息）
  - 说明：TUI 帮助消息

## 解决方案

### 解决方案方法

1. **实现 SkillsCommand**：参考 `ListModel` 的实现模式，创建 `SkillsCommand`
2. **更新帮助信息**：在 TUI 帮助消息中添加 `/skills` 命令说明
3. **使用 go vet 验证**：统一使用 `go vet ./...` 进行代码验证

### 代码变更

#### 变更描述

**文件**：`pkg/logic/command/model.go`

```diff
+ import (
+     "context"
+     "fmt"
+     "sort"
+     "strings"
+
+     "msa/pkg/config"
+     "msa/pkg/logic/provider"
+     "msa/pkg/logic/skills"
+     "msa/pkg/model"
+
+     log "github.com/sirupsen/logrus"
+ )

  func init() {
      RegisterCommand(&ListModel{})
      RegisterCommand(&ConfigCommand{})
+     RegisterCommand(&SkillsCommand{})
      // SetModel 命令已被交互式选择器替代，使用 /models 或 /model 命令
      // RegisterCommand(&SetModel{})
  }

+ // SkillsCommand Skills 列表命令
+ type SkillsCommand struct{}
+
+ func (s *SkillsCommand) Name() string {
+     return "skills"
+ }
+
+ func (s *SkillsCommand) Description() string {
+     return "List all available Skills"
+ }
+
+ func (s *SkillsCommand) Run(ctx context.Context, args []string) (*model.CmdResult, error) {
+     // 初始化 Skills Manager
+     manager := skills.GetManager()
+     if err := manager.Initialize(); err != nil {
+         log.Warnf("Failed to initialize skills: %v", err)
+         return &model.CmdResult{
+             Code: 1,
+             Msg:  "failed to initialize skills",
+             Type: "message",
+             Data: "无法加载 Skills: " + err.Error(),
+         }, nil
+     }
+
+     // 获取所有 skills
+     allSkills := manager.ListSkills()
+     disabled := manager.GetDisabledSkills()
+
+     if len(allSkills) == 0 {
+         return &model.CmdResult{
+             Code: 0,
+             Msg:  "success",
+             Type: "message",
+             Data: "没有可用的 Skills",
+         }, nil
+     }
+
+     // 创建禁用 map
+     disabledMap := make(map[string]bool)
+     for _, name := range disabled {
+         disabledMap[name] = true
+     }
+
+     // 按优先级排序
+     sort.Slice(allSkills, func(i, j int) bool {
+         return allSkills[i].Priority > allSkills[j].Priority
+     })
+
+     // 构建输出
+     var lines []string
+     lines = append(lines, "📋 可用 Skills:")
+     lines = append(lines, "")
+     lines = append(lines, fmt.Sprintf("%-20s %-8s %-10s %-10s", "Name", "Priority", "Source", "Status"))
+     lines = append(lines, strings.Repeat("-", 50))
+
+     for _, skill := range allSkills {
+         name := skill.Name
+         priority := formatPriority(skill.Priority)
+         source := formatSource(skill.Source)
+         status := "启用"
+         if disabledMap[name] {
+             status = "禁用"
+         }
+
+         lines = append(lines, fmt.Sprintf("%-20s %-8s %-10s %-10s", name, priority, source, status))
+     }
+
+     lines = append(lines, "")
+     lines = append(lines, fmt.Sprintf("总计: %d 个 Skills", len(allSkills)))
+     lines = append(lines, "")
+     lines = append(lines, "💡 提示:")
+     lines = append(lines, "  • 使用 /skills: <name1>,<name2> 手动指定 Skills")
+     lines = append(lines, "  • 使用 'msa skills show <name>' 查看详情")
+     lines = append(lines, "  • 使用 'msa skills disable/enable <name>' 管理状态")
+
+     return &model.CmdResult{
+         Code: 0,
+         Msg:  "success",
+         Type: "message",
+         Data: strings.Join(lines, "\n"),
+     }, nil
+ }
+
+ func (s *SkillsCommand) ToSelect(items []*model.SelectorItem) (*model.BaseSelector, error) {
+     return nil, fmt.Errorf("skills command does not support selector mode")
+ }
+
+ // formatPriority 格式化优先级显示
+ func formatPriority(priority int) string {
+     switch {
+     case priority >= 10:
+         return fmt.Sprintf("%d (最高)", priority)
+     case priority >= 8:
+         return fmt.Sprintf("%d (高)", priority)
+     case priority >= 5:
+         return fmt.Sprintf("%d (中)", priority)
+     default:
+         return fmt.Sprintf("%d (低)", priority)
+     }
+ }
+
+ // formatSource 格式化来源显示
+ func formatSource(source skills.SkillSource) string {
+     return source.String()
+ }
```

**说明**：添加 SkillsCommand 到 TUI 命令系统，支持在聊天界面查看可用 Skills

**文件**：`pkg/tui/chat.go`

```diff
-  helpMessage         = "📋 可用命令:\n  • clear - 清空对话\n  • /skills: <name1>,<name2> - 手动指定 skills\n  • help/? - 显示帮助\n  • quit/exit - 退出程序"
-  helpHint            = "ESC/Ctrl+C: 退出 | Ctrl+K: 清空 | /skills: 指定 skills | Enter: 发送"
+  helpMessage         = "📋 可用命令:\n  • clear - 清空对话\n  • /skills - 列出所有可用的 Skills\n  • /skills: <name1>,<name2> - 手动指定 skills\n  • help/? - 显示帮助\n  • quit/exit - 退出程序"
+  helpHint            = "ESC/Ctrl+C: 退出 | Ctrl+K: 清空 | /skills 查看列表 | Enter: 发送"
```

**说明**：更新 TUI 帮助信息，添加 `/skills` 命令说明

## 验证

### 验证步骤

1. **编译验证**：
   ```bash
   go vet ./...
   ```

2. **单元测试**：
   ```bash
   go test ./pkg/logic/skills/... -v
   ```

3. **TUI 功能测试**：
   - 启动 MSA：`./msa`
   - 输入 `/skills` 命令
   - 验证显示 Skills 列表

### 测试结果

- ✅ go vet 验证通过（无警告）
- ✅ 单元测试通过（88/91 任务完成）
- ✅ Skills 命令在 TUI 中正常工作
- ✅ 帮助信息正确显示

## 可复用模式

**模式名称**：TUI 命令注册模式
**模式ID**：p006

**描述**：
在 Bubble Tea TUI 中注册对话命令的标准模式。

**适用场景**：
- 需要在 TUI 聊天界面添加命令
- 命令需要返回格式化的文本信息
- 命令不需要交互式选择器

**实现**：
```go
// 1. 实现 MsaCommand 接口
type YourCommand struct{}

func (c *YourCommand) Name() string {
    return "command-name"
}

func (c *YourCommand) Description() string {
    return "Command description"
}

func (c *YourCommand) Run(ctx context.Context, args []string) (*model.CmdResult, error) {
    // 构建返回信息
    return &model.CmdResult{
        Code: 0,
        Msg:  "success",
        Type: "message", // 或 "selector"、"boolean"、"list"
        Data: "返回的数据",
    }, nil
}

func (c *YourCommand) ToSelect(items []*model.SelectorItem) (*model.BaseSelector, error) {
    // 如果不需要选择器，返回错误
    return nil, fmt.Errorf("command does not support selector mode")
}

// 2. 在 init() 中注册命令
func init() {
    RegisterCommand(&YourCommand{})
}
```

**收益**：
- 统一的命令接口
- 易于扩展新命令
- 支持多种返回类型（message、selector、list）

## 反模式

**反模式名称**：使用 go build 进行验证
**反模式ID**：ap001

**问题所在**：
`go build` 只能验证代码能否编译，但无法发现潜在的代码问题，如：
- 未使用的变量
- 可疑的格式化
- 无法到达的代码
- 错误的 printf 格式字符串

**避免**：
```bash
# ❌ 不要这样做
go build ./...

# ✅ 应该这样做
go vet ./...
```

**更好的实践**：
```bash
# 组合使用多个检查工具
go vet ./...        # 静态分析
go test ./...       # 运行测试
go fmt ./...        # 格式检查（使用 gofmt -l . 检查）
go run ./...        # 运行程序验证
```

## 经验教训

### 出了什么问题

1. **功能不完整**：实现 CLI 命令时忽略了 TUI 内的对应功能
2. **验证方式不当**：使用了较弱的验证方式（go build）而非更强的静态分析（go vet）

### 学到了什么

1. **全栈思维**：在实现功能时需要同时考虑 CLI 和 TUI 两个入口
2. **工具选择**：go vet 比 go build 提供更全面的代码质量检查

### 预防措施

1. **检查清单**：
   - [ ] CLI 命令实现
   - [ ] TUI 对话命令实现
   - [ ] 帮助信息更新
   - [ ] go vet 验证
   - [ ] 单元测试覆盖

2. **开发流程**：
   - 使用 `go vet ./...` 作为标准验证步骤
   - 在提交代码前运行完整的检查流程
   - 文档中明确使用 `go vet` 而非 `go build`

3. **代码规范**：
   - 所有新功能应该同时支持 CLI 和 TUI
   - 帮助信息应该同步更新
   - 使用 go vet 进行持续验证
