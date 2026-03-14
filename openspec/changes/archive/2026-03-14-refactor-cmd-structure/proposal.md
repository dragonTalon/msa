## Why

当前 `cmd/` 目录下所有命令文件平铺放置，随着项目发展将引入更多命令域（如 `memory`、`trade` 等），单一目录结构难以维护。需要采用子目录分包模式，提升代码组织性和可扩展性。

## What Changes

- 重构 `cmd/` 目录结构，采用子目录分包模式
- 引入命令组合模式（Command Composition Pattern）
- 使用 `cmd_*` 包命名约定（如 `cmd_skill`、`cmd_config`）
- 创建 `register.go` 暴露命令注册辅助函数
- 迁移现有 `skill` 相关文件到 `cmd/skill/` 子目录
- 迁移 `config.go` 到 `cmd/config/` 子目录

## Capabilities

### New Capabilities

- `cmd-structure`: 定义 CLI 命令的组织模式和扩展机制

### Modified Capabilities

无。此重构仅影响内部实现，不改变 CLI 用户界面行为。

## Impact

**文件变更：**
- 新增 `cmd/register.go`
- 移动 `cmd/config.go` → `cmd/config/command.go`
- 移动 `cmd/skills*.go` → `cmd/skill/` 目录
- 简化 `cmd/root.go`

**包结构：**
```
cmd/                    # package cmd
├── root.go
├── register.go
├── config/             # package cmd_config
│   └── command.go
└── skill/              # package cmd_skill
    ├── command.go
    ├── list.go
    ├── show.go
    ├── enable.go
    ├── disable.go
    └── format.go
```

**向后兼容：**
- CLI 使用方式完全不变（`msa config`、`msa skills list` 等）
- 不影响任何公开 API