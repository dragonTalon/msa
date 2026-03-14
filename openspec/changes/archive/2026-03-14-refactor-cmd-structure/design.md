## Context

当前 `cmd/` 目录结构：

```
cmd/                    # 单一 package cmd
├── root.go            (195行)
├── config.go          (17行)
├── skills.go          (63行)
├── skills_list.go     (131行)
├── skills_show.go     (92行)
├── skills_enable.go   (44行)
└── skills_disable.go  (44行)
```

所有文件共享 `package cmd`，通过 `init()` 自动注册到 `rootCmd`。这种模式在命令数量少时简单有效，但随着 `memory`、`trade` 等新命令域的加入，文件数量将显著增长，维护性下降。

## Goals / Non-Goals

**Goals:**
- 建立可扩展的命令组织模式，支持未来新增命令域
- 按功能域分组命令文件，提升代码可读性
- 遵循 Go 包命名惯例，避免命名冲突
- 保持 CLI 使用方式完全不变

**Non-Goals:**
- 不改变任何 CLI 用户界面（命令名、参数、输出）
- 不影响 TUI 内的斜杠命令（`pkg/logic/command/`）
- 不添加新的 CLI 命令功能

## Decisions

### 1. 包命名约定：`cmd_*`

**决定：** 子目录包名使用 `cmd_<domain>` 格式

```
cmd/           → package cmd
cmd/skill/     → package cmd_skill
cmd/config/    → package cmd_config
cmd/memory/    → package cmd_memory
```

**理由：**
- Go 标准库惯例（如 `net/http`, `archive/zip`）
- 自动获得语义化别名，无需显式重命名 import
- 避免与 `pkg/logic/skills` 等业务包冲突

**备选方案：**
- `<domain>cmd`（如 `skillcmd`）：不符合 Go 惯例，容易与业务包混淆
- 单一 `cmd` 包（平铺）：不可行，Go 不允许父子目录同名包

### 2. 命令注册机制

**决定：** 在 `cmd` 包暴露 `AddCommand()` 函数

```go
// cmd/register.go
package cmd

func AddCommand(cmd *cobra.Command) {
    rootCmd.AddCommand(cmd)
}
```

```go
// cmd/root.go
package cmd

import (
    "msa/cmd/config"   // 自动成为 cmd_config
    "msa/cmd/skill"    // 自动成为 cmd_skill
)

func init() {
    AddCommand(cmd_config.NewCommand())
    AddCommand(cmd_skill.NewCommand())
}
```

**理由：**
- 简单直接，无循环依赖风险
- 子包无需访问 `rootCmd` 变量
- 符合 Cobra 官方推荐模式

**备选方案：**
- 暴露 `GetRootCmd()` 让子包调用：增加耦合，不推荐
- 使用全局 `cobra` 注册：失去显式依赖关系

### 3. 命令组合模式

**决定：** 每个命令域暴露 `NewCommand()` 构造函数

```go
// cmd/skill/command.go
package cmd_skill

func NewCommand() *cobra.Command {
    cmd := &cobra.Command{Use: "skills", ...}
    cmd.AddCommand(newListCmd(), newShowCmd(), ...)
    return cmd
}
```

**理由：**
- 延迟初始化，避免 `init()` 副作用
- 便于单元测试（可独立构造命令）
- 明确的依赖关系

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|----------|
| Import 路径变更导致其他模块编译失败 | 全局搜索替换 import 路径 |
| 遗漏某些文件移动 | 编译验证 `go vet ./...` |
| `init()` 顺序变化 | 移除对 `init()` 顺序的依赖，显式注册 |

## Migration Plan

1. 创建 `cmd/register.go`
2. 创建 `cmd/skill/` 目录，移动文件并重命名包
3. 创建 `cmd/config/` 目录，移动文件并重命名包
4. 更新 `cmd/root.go` 导入和注册逻辑
5. 运行 `go vet ./...` 验证编译
6. 手动测试 CLI 命令功能