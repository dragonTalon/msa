## Why

用户目前需要手动从 GitHub Releases 下载新版本替换二进制文件，更新体验不友好。提供 `msa update` 自更新能力可以简化升级流程，提升用户体验。

## What Changes

- 新增 `msa update` 子命令，支持自动检查并更新到最新版本
- 新增 `msa update --check` 参数，仅检查是否有新版本（不执行更新）
- 新增 `install.sh` 安装脚本，支持一键安装最新版本
- 更新 README 文档，添加安装和更新说明

## Capabilities

### New Capabilities

- `self-update`: MSA 自更新能力，包括版本检查、下载、替换二进制文件

### Modified Capabilities

- 无

## Impact

- 新增 `cmd/update/` 子命令目录
- 新增 `pkg/updater/` 更新核心逻辑
- 新增 `install.sh` 根目录安装脚本
- 更新 `README.md` 文档
