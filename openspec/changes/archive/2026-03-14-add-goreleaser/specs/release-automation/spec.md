## ADDED Requirements

### Requirement: Tag 触发自动发布

系统必须在创建符合格式的 Git Tag 时自动触发发布流程。

#### Scenario: 创建 v 前缀 Tag 触发发布

- **WHEN** 用户推送格式为 `v*` 的 Tag（如 `v1.0.0`）
- **THEN** GitHub Actions 自动触发发布工作流
- **AND** 构建所有目标平台的二进制文件
- **AND** 创建 GitHub Release 并上传产物

#### Scenario: 非 v 前缀 Tag 不触发发布

- **WHEN** 用户推送格式不为 `v*` 的 Tag（如 `release-1.0`）
- **THEN** 发布工作流不被触发

### Requirement: 多平台二进制构建

系统必须构建 Linux、macOS、Windows 平台的二进制文件。

#### Scenario: 构建所有目标平台

- **WHEN** 发布流程执行
- **THEN** 生成以下二进制文件：
  - linux/amd64
  - linux/arm64
  - darwin/amd64
  - darwin/arm64
  - windows/amd64
  - windows/arm64
- **AND** 每个平台生成对应的压缩包（Linux/macOS 用 tar.gz，Windows 用 zip）

### Requirement: 版本信息注入

编译时必须注入版本号、commit hash 和构建时间。

#### Scenario: 发布版本注入真实信息

- **WHEN** 发布流程编译二进制文件
- **THEN** Version 变量被设置为 Tag 名称（去掉 v 前缀）
- **AND** Commit 变量被设置为当前 commit hash
- **AND** Date 变量被设置为构建时间

### Requirement: 校验和生成

系统必须为所有发布产物生成 SHA256 校验和文件。

#### Scenario: 生成校验和文件

- **WHEN** 所有二进制文件打包完成
- **THEN** 生成 `checksums.txt` 文件
- **AND** 包含所有发布产物的 SHA256 哈希值

### Requirement: Changelog 自动生成

系统必须从 git commits 自动生成发布说明。

#### Scenario: 生成 Changelog

- **WHEN** 创建 GitHub Release
- **THEN** 自动生成包含 commit 历史的 Changelog
- **AND** 排除 `docs:`、`test:`、`ci:` 前缀的 commit