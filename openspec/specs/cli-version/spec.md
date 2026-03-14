# cli-version 规范

## Purpose

定义 CLI 版本查询命令的功能和输出格式。

## Requirements

### Requirement: version 子命令

CLI 必须提供 `msa version` 命令显示版本信息。

#### Scenario: 执行 version 命令

- **WHEN** 用户执行 `msa version`
- **THEN** 系统显示版本号、commit hash 和构建时间
- **AND** 输出格式为易读的文本格式

#### Scenario: 开发版本显示

- **WHEN** 用户执行未发布的开发版本
- **THEN** 版本号显示为 `dev`
- **AND** commit 显示为 `none` 或实际 commit hash
- **AND** 日期显示为 `unknown` 或实际构建时间

### Requirement: 命令输出格式

version 命令的输出必须清晰易读。

#### Scenario: 标准输出格式

- **WHEN** 用户执行 `msa version`
- **THEN** 输出包含以下信息：
  ```
  MSA 版本: x.x.x
  Commit: abc1234
  构建时间: 2024-01-01T00:00:00Z
  ```

### Requirement: 命令目录结构

version 命令必须遵循项目的命令目录结构规范。

#### Scenario: 目录结构规范

- **WHEN** 实现 version 命令
- **THEN** 必须放置在 `cmd/version/` 目录
- **AND** 包名必须为 `cmd_version`
- **AND** 必须暴露 `NewCommand() *cobra.Command` 构造函数