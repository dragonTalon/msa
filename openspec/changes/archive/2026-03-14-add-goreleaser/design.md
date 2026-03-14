## Context

当前项目没有自动化发布流程，发布新版本需要：
1. 手动编译多平台二进制文件
2. 手动创建 GitHub Release
3. 手动上传编译产物

这种方式效率低、容易出错，且用户安装门槛高。项目使用纯 Go 实现的 SQLite 驱动（glebarez/sqlite），支持 `CGO_ENABLED=0` 交叉编译。

## Goals / Non-Goals

**Goals:**
- 创建 Git Tag 时自动触发发布流程
- 自动构建 Linux/macOS/Windows 多平台二进制文件
- 自动创建 GitHub Release 并上传编译产物
- 提供版本查询命令 `msa version`
- 编译时注入版本信息

**Non-Goals:**
- 不实现 Homebrew/Docker/Snap 等额外发布渠道
- 不改变现有 CI 测试流程
- 不支持 32 位架构

## Decisions

### 1. 发布工具：GoReleaser

**决定：** 使用 GoReleaser 作为发布工具

**理由：**
- Go 项目标准发布工具，社区成熟
- 配置简单，一个 YAML 文件搞定所有
- 自动生成 Changelog 和校验和
- 未来可轻松扩展 Homebrew、Docker 等发布渠道

**备选方案：**
- 手写 GitHub Actions 矩阵：需要更多配置，功能有限
- GoReleaser Pro：功能更多但需要付费，社区版已足够

### 2. 版本注入方式

**决定：** 在 `main.go` 声明版本变量，通过 ldflags 注入

```go
// main.go
var (
    Version = "dev"
    Commit  = "none"
    Date    = "unknown"
)
```

**理由：**
- Go 标准做法，简单可靠
- 无需额外依赖
- 开发时显示 `dev`，发布时注入真实值

### 3. version 命令位置

**决定：** 创建 `cmd/version/` 子目录，遵循新的命令目录结构规范

```
cmd/version/
└── command.go    # package cmd_version
```

**理由：**
- 符合 `cmd-structure` 规范
- 与 config、skill 命令保持一致
- 便于独立维护

### 4. 目标平台矩阵

**决定：** 支持 6 个平台组合

| OS | Arch | 说明 |
|----|------|------|
| linux | amd64 | 服务器主流 |
| linux | arm64 | ARM 服务器 |
| darwin | amd64 | Intel Mac |
| darwin | arm64 | M1/M2 Mac |
| windows | amd64 | Windows 主流 |
| windows | arm64 | ARM Windows |

**理由：**
- 覆盖主流使用场景
- 不支持 386（现代系统基本不需要）

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|----------|
| GoReleaser 版本更新导致配置不兼容 | 锁定主版本号 `~> v2` |
| 首次发布配置错误 | 使用 `goreleaser release --snapshot` 本地测试 |
| Tag 格式不符合预期 | 明确文档说明使用 `v*` 格式（如 `v1.0.0`） |

## Migration Plan

1. 添加 `main.go` 版本变量声明
2. 创建 `cmd/version/command.go`
3. 在 `cmd/root.go` 注册 version 命令
4. 创建 `.goreleaser.yml` 配置
5. 创建 `.github/workflows/release.yml` 工作流
6. 本地测试：`goreleaser release --snapshot --clean`
7. 创建 Tag 验证完整流程

## Open Questions

无。方案明确，可直接实施。