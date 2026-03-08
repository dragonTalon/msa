# 提案：add-config-management

为 MSA 项目添加用户友好的配置管理能力，支持交互式 TUI 配置界面、命令行参数和环境变量。

## 知识上下文

### 相关历史问题

知识库中未找到与配置管理直接相关的问题。已记录的问题主要集中在：
- 浏览器自动化配置 (chromedp)
- Git 工作流程
- 搜索引擎集成

### 需避免的反模式

未识别到特定反模式。

### 推荐模式

基于知识库中的 `003` 错误处理和重试策略模式：
- 配置加载应该有清晰的错误处理
- 用户输入验证应该提供友好的错误提示
- 配置保存失败时应该有明确的反馈

## 为什么

当前 MSA 项目虽然拥有完整的配置系统后端（`pkg/config/local_config.go`），但缺乏用户友好的配置接口。用户只能通过：
1. 手动编辑 `~/.msa/msa_config.json` 文件
2. 修改代码中的默认值

这导致了以下问题：
- 新用户上手困难，不知道如何配置 API Key
- 切换 Provider 需要手动编辑 JSON 文件
- 无法快速查看当前配置
- 缺少配置验证机制

添加配置管理能力可以：
- 降低使用门槛，提供友好的交互式配置界面
- 支持多种配置方式（TUI、命令行参数、环境变量）
- 提供配置验证和错误提示
- 方便查看和管理当前配置

## 变更内容

### 新增功能

1. **TUI 交互式配置界面**
   - `msa config` 命令进入配置界面
   - 支持编辑所有配置项（Provider、API Key、Base URL、日志配置）
   - Provider 选择器（下拉菜单，显示友好名称）
   - API Key 隐藏显示和格式验证
   - 配置保存和重置功能

2. **命令行参数支持**
   - `--config` 支持配置文件路径
   - `--config` 支持 key=value 格式设置配置项
   - 可混合使用多种配置方式

3. **环境变量支持**
   - `MSA_PROVIDER` - LLM 提供商
   - `MSA_API_KEY` - API 密钥
   - `MSA_BASE_URL` - API 基础 URL
   - `MSA_LOG_LEVEL` - 日志级别
   - `MSA_LOG_FILE` - 日志文件路径

4. **Provider 友好名称和验证**
   - Provider 注册表，包含友好名称、描述、默认 Base URL
   - API Key 格式验证（前缀检查）
   - 切换 Provider 时清空 API Key 和 Base URL

5. **配置优先级系统**
   - 命令行参数 > 环境变量 > 配置文件 > 默认值

### 修改内容

- `pkg/logic/provider/api.go` - 添加 `GetRegisteredProviders()` 函数
- `pkg/model/comment.go` - 添加 `ProviderInfo` 和 `ProviderRegistry`
- `cmd/root.go` - 添加 `--config` 参数支持

## 能力

### 新增能力

- `cli-config-management`：命令行配置管理能力
  - TUI 配置界面
  - 命令行参数配置
  - 环境变量配置
  - 配置持久化

- `provider-registry`：Provider 注册表能力
  - Provider 友好名称映射
  - Provider 元信息（描述、默认 URL、Key 前缀）
  - API Key 格式验证

- `config-validation`：配置验证能力
  - API Key 格式验证
  - URL 格式验证
  - 日志级别枚举验证
  - 路径可写性检查

### 修改能力

- `local-config`：本地配置存储能力（已有）
  - 添加环境变量加载
  - 添加配置优先级合并
  - 添加配置验证流程

## 影响

### 代码影响

**新增文件**：
- `pkg/config/env.go` - 环境变量加载
- `pkg/config/validator.go` - 配置验证
- `cmd/config.go` - config 子命令
- `pkg/tui/config/` - TUI 配置界面目录

**修改文件**：
- `pkg/logic/provider/api.go` - 添加 Provider 列表导出
- `pkg/model/comment.go` - 添加 Provider 注册表
- `cmd/root.go` - 添加 --config 参数

### 用户体验影响

- **正面**：大幅降低配置门槛，提供友好的交互方式
- **兼容性**：保持向后兼容，现有配置文件无需修改
- **学习曲线**：新用户可以通过 `msa config` 快速了解所有可配置项

### 依赖影响

- TUI 界面使用现有的 `bubbletea` 框架（项目已依赖）
- 不引入新的外部依赖

---

**知识备注**：
- 此提案使用 `openspec/knowledge` 中的知识生成
- 最后知识同步：2026-03-08
- 知识库总问题数：5
- 总模式数：3
