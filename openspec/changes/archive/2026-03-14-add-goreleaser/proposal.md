## Why

当前项目每次发布需要手动编译多平台二进制文件，效率低且容易出错。用户安装需要从源码编译，门槛较高。通过引入 GoReleaser，可以在创建 Git Tag 时自动构建 Linux/macOS/Windows 多平台二进制文件并发布到 GitHub Release，降低用户使用门槛。

## What Changes

- 新增 `.goreleaser.yml` 配置文件
- 新增 `.github/workflows/release.yml` 发布工作流
- 新增 `version` 子命令显示版本信息
- 在编译时注入版本号、commit hash、构建时间

## Capabilities

### New Capabilities

- `release-automation`: 基于 GoReleaser 的自动化发布流程，支持多平台二进制构建和 GitHub Release 发布
- `cli-version`: CLI 版本查询命令，显示当前版本、commit 和构建信息

### Modified Capabilities

无。此变更仅影响发布流程，不改变任何用户功能。

## Impact

**新增文件：**
- `.goreleaser.yml` - GoReleaser 配置
- `.github/workflows/release.yml` - 发布工作流

**修改文件：**
- `cmd/root.go` 或新建 `cmd/version.go` - 添加 version 子命令
- `main.go` - 添加版本变量声明（如需要）

**发布产物：**
- `msa_<version>_linux_amd64.tar.gz`
- `msa_<version>_linux_arm64.tar.gz`
- `msa_<version>_darwin_amd64.tar.gz`
- `msa_<version>_darwin_arm64.tar.gz`
- `msa_<version>_windows_amd64.zip`
- `msa_<version>_windows_arm64.zip`
- `checksums.txt`

**兼容性：**
- 向后兼容，现有功能不受影响
- 需要创建 Git Tag（格式 `v*`）才会触发发布流程

## Non-Goals

- 不实现 Homebrew formula 发布（后续可扩展）
- 不实现 Docker 镜像发布（后续可扩展）
- 不改变现有的 CI 测试流程（`.github/workflows/go.yml`）
- 不支持 386 架构（现代系统基本不需要）