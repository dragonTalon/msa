## 1. 基础设施

- [x] 1.1 添加 `Masterminds/semver/v3` 依赖到 go.mod
- [x] 1.2 创建 `pkg/updater/updater.go` 文件，定义 Updater 结构体

## 2. 核心更新逻辑

- [x] 2.1 实现 `CheckLatestVersion()` 方法，调用 GitHub Releases API 获取最新版本信息
- [x] 2.2 实现 `CompareVersions()` 方法，使用 semver 比较当前版本与最新版本
- [x] 2.3 实现 `buildDownloadURL()` 方法，根据 OS/arch 构建下载 URL
- [x] 2.4 实现 `downloadFile()` 方法，下载文件到临时目录
- [x] 2.5 实现 `verifyChecksum()` 方法，校验下载文件的 SHA256
- [x] 2.6 实现 `replaceBinary()` 方法，备份、替换、清理二进制文件
- [x] 2.7 实现 `Update()` 方法，整合上述步骤完成更新流程

## 3. CLI 命令

- [x] 3.1 创建 `cmd/update/command.go`，定义 update 子命令
- [x] 3.2 实现 `--check` 参数，仅检查版本不更新
- [x] 3.3 在 `cmd/root.go` 中注册 update 子命令

## 4. 安装脚本

- [x] 4.1 创建 `install.sh` 一键安装脚本
- [x] 4.2 实现 OS 和架构检测逻辑
- [x] 4.3 实现下载最新 release 二进制
- [x] 4.4 实现安装路径选择逻辑（~/.local/bin 优先）
- [x] 4.5 实现 PATH 检测和提示

## 5. 文档

- [x] 5.1 更新 `README.md`，添加安装说明（install.sh 使用方法）
- [x] 5.2 更新 `README.md`，添加更新说明（msa update 使用方法）

## 6. 验证

- [x] 6.1 运行 `go vet ./...` 验证编译
- [x] 6.2 手动测试 `msa update --check` 功能
- [x] 6.3 手动测试 `msa update` 功能（在测试环境）
- [x] 6.4 测试 `install.sh` 脚本
