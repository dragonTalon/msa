# 规格：installation

## Purpose

install.sh 安装脚本 SHALL 提供纯二进制下载安装能力，不依赖 go 和 gh，失败时提供清晰的手动下载指引。

> 📚 **知识上下文**：基于模式 error-handling-strategy，错误消息应包含问题描述、可能原因、建议解决方案。

## Requirements

### Requirement: 纯二进制下载安装

系统 SHALL 通过预编译二进制文件安装 MSA，不依赖源码构建。

#### Scenario: 成功安装
- **WHEN** 用户运行 `curl -fsSL https://raw.githubusercontent.com/dragonTalon/msa/main/install.sh | sh`
- **AND** GitHub Releases API 可访问
- **THEN** 脚本检测操作系统和架构
- **AND** 脚本下载对应的预编译二进制文件
- **AND** 脚本解压并安装到 ~/.local/bin 或 /usr/local/bin
- **AND** 脚本显示成功消息

#### Scenario: 本地项目目录安装
- **WHEN** 用户在 msa 项目根目录运行 `sh install.sh`
- **AND** 项目目录不包含 go 环境
- **THEN** 脚本检测操作系统和架构
- **AND** 脚本从 GitHub Releases 下载预编译二进制
- **AND** 安装成功完成

### Requirement: 版本获取失败处理

系统 SHALL 在无法获取版本信息时提供手动下载链接并退出。

#### Scenario: GitHub API 不可达
- **WHEN** 脚本尝试访问 GitHub Releases API
- **AND** 网络请求失败或超时（30秒）
- **THEN** 脚本显示错误消息
- **AND** 脚本显示手动下载链接 `https://github.com/dragonTalon/msa/releases`
- **AND** 脚本以非零状态退出

### Requirement: 下载失败处理

系统 SHALL 在下载失败时提供手动下载链接并退出。

#### Scenario: 二进制下载失败
- **WHEN** 脚本尝试下载预编译二进制
- **AND** 下载失败或超时（60秒）
- **THEN** 脚本显示错误消息
- **AND** 脚本显示手动下载链接 `https://github.com/dragonTalon/msa/releases`
- **AND** 脚本以非零状态退出

### Requirement: 不依赖 go 和 gh

安装脚本 SHALL NOT 依赖 go 编译环境或 gh CLI 工具。

#### Scenario: 无 go 环境安装
- **WHEN** 用户系统未安装 go
- **AND** 用户运行 install.sh
- **THEN** 脚本正常执行（不检测 go）
- **AND** 安装通过下载预编译二进制完成

#### Scenario: 无 gh CLI 安装
- **WHEN** 用户系统未安装 gh CLI
- **AND** 用户运行 install.sh
- **THEN** 脚本正常执行（不依赖 gh 认证）
- **AND** 安装通过公开 GitHub API 完成

### Requirement: 清晰的错误消息

系统 SHALL 提供包含解决方案的错误消息。

> 📚 **知识上下文**：基于模式 error-handling-strategy，用户友好的错误消息应包含：错误描述、可能原因、建议解决方案。

#### Scenario: 错误消息格式
- **WHEN** 安装过程中发生错误
- **THEN** 错误消息包含：
  1. 清晰的错误描述（如"无法获取版本信息"）
  2. 可能原因（如"网络问题或 GitHub 访问限制"）
  3. 建议方案（如"请检查网络连接后重试"）
  4. 手动下载链接

---

**知识引用**：
- 模式：error-handling-strategy - 错误处理和重试策略
- 问题：issue-004 - chromedp 超时配置（超时时间设置的参考）