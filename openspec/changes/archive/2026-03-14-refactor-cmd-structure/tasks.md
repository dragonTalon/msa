# 实现任务：重构 cmd 目录结构

## 1. 创建注册机制

- [x] 1.1 创建 `cmd/register.go`
  - 暴露 `AddCommand(cmd *cobra.Command)` 函数
  - 将命令添加到 `rootCmd`

## 2. 迁移 config 命令

- [x] 2.1 创建 `cmd/config/` 目录
- [x] 2.2 创建 `cmd/config/command.go`
  - 重命名包为 `cmd_config`
  - 定义 `NewCommand() *cobra.Command` 构造函数
  - 将 `ConfigCmd` 改为局部变量
- [x] 2.3 删除原 `cmd/config.go`

## 3. 迁移 skill 命令

- [x] 3.1 创建 `cmd/skill/` 目录
- [x] 3.2 创建 `cmd/skill/command.go`
  - 重命名包为 `cmd_skill`
  - 定义 `NewCommand() *cobra.Command` 构造函数
  - 将子命令注册移到 `NewCommand()` 内部
- [x] 3.3 创建 `cmd/skill/list.go`
  - 迁移 `skillsListCmd` 和 `runSkillsList`
- [x] 3.4 创建 `cmd/skill/show.go`
  - 迁移 `skillsShowCmd` 和 `runSkillsShow`
- [x] 3.5 创建 `cmd/skill/enable.go`
  - 迁移 `skillsEnableCmd` 和 `runSkillsEnable`
- [x] 3.6 创建 `cmd/skill/disable.go`
  - 迁移 `skillsDisableCmd` 和 `runSkillsDisable`
- [x] 3.7 创建 `cmd/skill/format.go`
  - 迁移 `formatPriority`, `formatSource`, `formatStatus` 辅助函数
- [x] 3.8 删除原 `cmd/skills*.go` 文件

## 4. 更新根命令

- [x] 4.1 修改 `cmd/root.go`
  - 添加 import `msa/cmd/config` 和 `msa/cmd/skill`
  - 在 `init()` 中调用 `AddCommand(cmd_config.NewCommand())`
  - 在 `init()` 中调用 `AddCommand(cmd_skill.NewCommand())`
  - 移除原有的 `rootCmd.AddCommand(ConfigCmd)`

## 5. 验证和测试

- [x] 5.1 运行 `go vet ./...` 验证编译
- [x] 5.2 手动测试 CLI 命令
  - `msa config` 正常打开配置 TUI
  - `msa skills` 显示技能列表
  - `msa skills list` 列出所有技能
  - `msa skills show <name>` 显示技能详情
  - `msa skills enable <name>` 启用技能
  - `msa skills disable <name>` 禁用技能