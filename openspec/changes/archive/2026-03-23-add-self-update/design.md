## Context

MSA 通过 GoReleaser 发布到 GitHub Releases，支持 Linux/macOS/Windows 多平台。当前用户需要手动下载、解压、替换二进制文件，体验不流畅。需要提供自动化更新能力。

### 当前发布架构

```
GitHub Releases URL 结构:
https://github.com/dragonTalon/msa/releases/download/{tag}/msa_{version}_{os}_{arch}.{ext}

文件名示例:
- msa_0.0.6-beta_darwin_arm64.tar.gz
- msa_0.0.6-beta_linux_amd64.tar.gz
- msa_0.0.6-beta_windows_amd64.zip
- checksums.txt
```

## Goals / Non-Goals

**Goals:**
- 实现 `msa update` 子命令，自动检查并更新到最新版本
- 实现 `msa update --check` 参数，仅检查版本不更新
- 提供 `install.sh` 一键安装脚本
- 支持 Linux/macOS 自动替换，Windows 手动指引

**Non-Goals:**
- 不支持指定版本安装（始终更新到 latest）
- 不支持版本回滚
- 不实现后台自动更新
- 不支持 Windows 自动替换（Windows 锁定运行中文件）

## Decisions

### 1. 版本检查方式

**决定**: 使用 GitHub Releases API (`/repos/{owner}/{repo}/releases/latest`)

**备选方案**:
- 直接解析 HTML 页面：脆弱，易受页面结构变化影响
- 自建版本服务器：需要额外基础设施

**理由**: GitHub API 稳定、免费、与现有发布流程无缝集成

### 2. 二进制替换策略

**决定**: 备份 → 下载 → 替换 → 清理

**流程**:
```
1. 备份: mv msa msa.bak
2. 下载: 下载 tar.gz/zip 到临时目录
3. 校验: SHA256 校验（使用 checksums.txt）
4. 解压: 解压到临时目录
5. 替换: mv /tmp/msa /path/to/msa
6. 清理: rm msa.bak, rm 临时文件
```

**失败回滚**: 如果步骤 4-5 失败，恢复 `msa.bak`

### 3. Windows 处理

**决定**: Windows 平台仅下载，提示手动替换

**理由**: Windows 不允许替换运行中的可执行文件

### 4. 安装脚本位置

**决定**: 放在项目根目录 `install.sh`

**安装路径优先级**:
1. `~/.local/bin` (用户级，无需 sudo)
2. `/usr/local/bin` (系统级，需要 sudo)

### 5. 依赖库

**决定**: 使用 `Masterminds/semver/v3` 进行版本比较

**理由**: 支持预发布版本（如 `-beta`）的正确比较

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|----------|
| GitHub API 频率限制 (60次/小时未认证) | 对于更新检查场景足够使用 |
| 下载大文件超时 | 设置合理超时时间，支持重试 |
| 权限不足无法替换 | 检测权限错误，提示用户使用 sudo |
| 网络不稳定 | 显示下载进度，支持断点续传 |
