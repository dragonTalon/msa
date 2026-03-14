## ADDED Requirements

### Requirement: CLI 命令目录结构

CLI 命令必须按功能域组织到子目录中，每个子目录使用独立的包。

#### Scenario: 目录结构规范

- **WHEN** 创建新的命令域
- **THEN** 必须在 `cmd/` 下创建子目录
- **AND** 子目录名使用 kebab-case 格式
- **AND** 子目录包名使用 `cmd_<domain>` 格式

#### Scenario: 根目录文件限制

- **WHEN** 查看 `cmd/` 根目录
- **THEN** 只包含 `root.go` 和 `register.go`
- **AND** 不包含任何具体命令实现文件

### Requirement: 命令注册机制

命令必须通过显式注册函数添加到根命令。

#### Scenario: 注册函数定义

- **WHEN** 定义命令注册机制
- **THEN** `cmd` 包必须暴露 `AddCommand(cmd *cobra.Command)` 函数
- **AND** 该函数将命令添加到 `rootCmd`

#### Scenario: 子包注册命令

- **WHEN** 子包需要注册命令
- **THEN** 子包必须暴露 `NewCommand() *cobra.Command` 构造函数
- **AND** 在 `cmd/root.go` 的 `init()` 中调用 `AddCommand(XXX.NewCommand())`

### Requirement: 包命名约定

子目录包名必须遵循 `cmd_*` 前缀约定。

#### Scenario: 包名格式

- **WHEN** 定义子目录包名
- **THEN** 使用 `cmd_<domain>` 格式
- **AND** 例如 `cmd/skill/` → `package cmd_skill`
- **AND** 例如 `cmd/config/` → `package cmd_config`

#### Scenario: Import 别名

- **WHEN** 导入命令子包
- **THEN** 包自动获得 `cmd_<domain>` 别名
- **AND** 无需显式指定 import 别名

### Requirement: 命令行为不变

重构后所有 CLI 命令的使用方式必须保持不变。

#### Scenario: 命令调用方式

- **WHEN** 用户执行 CLI 命令
- **THEN** 命令名、参数、输出格式与重构前完全一致
- **AND** 例如 `msa config` 仍然打开配置 TUI
- **AND** 例如 `msa skills list` 仍然列出所有 skills

#### Scenario: 命令返回值

- **WHEN** 命令执行完成
- **THEN** 返回值（成功/失败）与重构前一致
- **AND** 错误消息格式不变