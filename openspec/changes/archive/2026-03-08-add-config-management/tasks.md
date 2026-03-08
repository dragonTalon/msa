# 实现任务：add-config-management

> 📚 **知识上下文**：基于知识库模式 #003（错误处理和重试策略）
> - 配置加载应该有清晰的错误处理
> - 用户输入验证应该提供友好的错误提示
> - 配置保存失败时应该有明确的反馈

---

## 1. 阶段 1：基础设施（核心代码）

### 1.1 Provider 注册表

- [x] 1.1.1 在 `pkg/model/comment.go` 中添加 `ProviderInfo` 结构体
  - 包含字段：ID, DisplayName, Description, DefaultBaseURL, KeyPrefix

- [x] 1.1.2 在 `pkg/model/comment.go` 中创建 `ProviderRegistry` 全局 map
  - 初始化 Siliconflow Provider 信息
  - DisplayName: "SiliconFlow (硅基流动)"
  - Description: "国内 LLM API 提供商，兼容 OpenAI 格式"
  - DefaultBaseURL: "https://api.siliconflow.cn/v1"
  - KeyPrefix: "sk-"

- [x] 1.1.3 在 `pkg/model/comment.go` 中添加 `LlmProvider.GetDisplayName()` 方法
  - 如果 Provider 在注册表中，返回友好名称
  - 否则返回原始 ID 字符串

- [x] 1.1.4 在 `pkg/model/comment.go` 中添加 `LlmProvider.GetDefaultBaseURL()` 方法
  - 如果 Provider 在注册表中，返回默认 Base URL
  - 否则返回空字符串

- [x] 1.1.5 在 `pkg/model/comment.go` 中添加 `LlmProvider.ValidateAPIKey()` 方法
  - 如果 Provider 定义了 KeyPrefix，验证前缀
  - 如果前缀不匹配，返回错误
  - 如果 KeyPrefix 为空，跳过验证

### 1.2 环境变量加载

- [x] 1.2.1 创建 `pkg/config/env.go` 文件
  - 实现 `LoadFromEnv() *LocalStoreConfig` 函数

- [x] 1.2.2 在 `LoadFromEnv()` 中读取环境变量
  - MSA_PROVIDER
  - MSA_API_KEY
  - MSA_BASE_URL
  - MSA_LOG_LEVEL
  - MSA_LOG_FILE

- [x] 1.2.3 处理空环境变量
  - 空字符串视为未设置
  - 返回 nil 字段而非空字符串

- [x] 1.2.4 日志级别大小写转换
  - 将环境变量值转换为小写
  - 验证是否为有效日志级别

> 📚 **知识上下文**：基于知识库模式 #003（错误处理和重试策略）
> - 环境变量加载失败不应阻止程序启动
> - 记录警告日志但继续使用默认配置

### 1.3 配置验证

- [x] 1.3.1 创建 `pkg/config/validator.go` 文件
  - 定义 `ValidationError` 结构体
  - 定义 `ValidationSeverity` 枚举（Error, Warning）

- [x] 1.3.2 实现 `ValidateAPIKey(key, provider string) *ValidationError`
  - 非空验证
  - 前缀验证（使用 Provider.ValidateAPIKey()）
  - 长度验证（警告少于 10 字符）

- [x] 1.3.3 实现 `ValidateBaseURL(url string) *ValidationError`
  - URL 格式验证
  - 协议验证（必须 http:// 或 https://）

- [x] 1.3.4 实现 `ValidateLogLevel(level string) *ValidationError`
  - 枚举值验证
  - 大小写不敏感

- [x] 1.3.5 实现 `ValidateLogPath(path string) *ValidationError`
  - 路径展开（~ 和 $HOME）
  - 目录可写性检查
  - 尝试创建不存在的目录

- [x] 1.3.6 实现 `ValidateProvider(provider string) *ValidationError`
  - 检查是否在 ProviderRegistry 中注册

- [x] 1.3.7 实现 `ValidateConfig(cfg *LocalStoreConfig) []*ValidationError`
  - 运行所有验证
  - 按严重程度排序

> 📚 **知识上下文**：基于知识库模式 #003（错误处理和重试策略）
> - 用户输入验证应该提供友好的错误提示
> - 验证错误应分组显示，按严重程度排序

### 1.4 配置合并逻辑

- [x] 1.4.1 在 `pkg/config/local_config.go` 中创建 `ConfigBuilder` 结构体
  - 包含字段：defaultCfg, fileCfg, envCfg, cliCfg

- [x] 1.4.2 实现 `ConfigBuilder.merge(base, override *LocalStoreConfig)` 方法
  - 只覆盖非零值
  - 处理嵌套的 LogConfig

- [x] 1.4.3 实现 `ConfigBuilder.Build() *LocalStoreConfig` 方法
  - 按优先级合并：默认 → 文件 → 环境变量 → 命令行

- [x] 1.4.4 修改 `InitConfig()` 函数
  - 使用 ConfigBuilder 进行配置合并
  - 集成环境变量加载

- [x] 1.4.5 修改 `InitConfig()` 的错误处理
  - 配置文件读取失败时记录警告，不阻塞启动
  - JSON 解析失败时记录警告，使用默认配置
  - 友好的错误提示引导用户运行 `msa config`

> 📚 **知识上下文**：基于知识库模式 #003（错误处理和重试策略）
> - **避免**：静默失败反模式（a001）
> - 配置加载失败不应阻止程序启动
> - 应显示友好的错误消息并引导用户修复

### 1.5 命令行参数解析

- [x] 1.5.1 在 `pkg/config/local_config.go` 中实现 `ParseConfigArg(arg string) (*LocalStoreConfig, error)`
  - 检测是否包含 `=` 符号
  - 解析 key=value 格式
  - 支持的 key：provider, apikey, baseurl, loglevel, logfile

- [x] 1.5.2 在 `ParseConfigArg()` 中处理文件路径
  - 如果不含 `=` 且是有效文件路径
  - 加载该配置文件

- [x] 1.5.3 在 `ParseConfigArg()` 中处理无效格式
  - 返回错误和用法说明

### 1.6 导出 Provider 列表

- [x] 1.6.1 在 `pkg/logic/provider/api.go` 中添加 `GetRegisteredProviders() []model.LlmProvider` 函数
  - 遍历 ProviderRegistry
  - 返回所有已注册的 Provider ID 列表

- [x] 1.6.2 添加单元测试 `TestGetRegisteredProviders()`
  - 验证 Siliconflow 在返回列表中

---

## 2. 阶段 2：命令行接口

### 2.1 --config 参数

- [x] 2.1.1 在 `cmd/root.go` 中添加 `--config` 字符串数组参数
  - 使用 `StringArrayVar`
  - 允许多个 --config 参数

- [x] 2.1.2 在 `ExecuteWithSignal()` 中解析 --config 参数
  - 处理所有 --config 参数
  - 先加载文件路径，后应用 key=value

- [x] 2.1.3 将解析的配置传递给配置系统
  - 合并到 ConfigBuilder.cliCfg

- [ ] 2.1.4 添加集成测试 `TestRootConfigArg()`
  - 测试文件路径格式
  - 测试 key=value 格式
  - 测试混合使用

### 2.2 config 子命令

- [x] 2.2.1 创建 `cmd/config.go` 文件
  - 添加 `configCmd` 子命令
  - 设置 Short 和 Description

- [x] 2.2.2 在 `cmd/root.go` 中注册 config 子命令
  - `rootCmd.AddCommand(configCmd)`

- [x] 2.2.3 实现 configCmd 的 `RunE` 函数
  - 调用 TUI 配置界面
  - 处理 TUI 返回的错误

- [ ] 2.2.4 添加单元测试 `TestConfigCommand()`
  - 验证命令可以正确执行

### 2.3 TUI 斜杠命令

- [x] 2.3.1 在 `pkg/logic/command/model.go` 中添加 `ConfigCommand` 结构体
  - 实现 `Name()` 方法返回 "config"
  - 实现 `Description()` 方法返回命令描述
  - 实现 `Run()` 方法返回当前配置信息

- [x] 2.3.2 在 `init()` 中注册 ConfigCommand
  - 调用 `RegisterCommand(&ConfigCommand{})`

- [x] 2.3.3 在 `pkg/tui/chat.go` 中添加 "message" 类型处理
  - 在 `analyzeResult()` 函数中添加 case "message"

---

## 3. 阶段 3：TUI 界面

### 3.1 TUI 基础结构

- [x] 3.1.1 创建 `pkg/tui/config/` 目录

- [x] 3.1.2 创建 `pkg/tui/config/styles.go` 文件
  - 定义通用样式
  - 复用 `pkg/tui/style` 中的现有样式

- [x] 3.1.3 创建 `pkg/tui/config/config.go` 文件
  - 定义主模型 `configModel`
  - 实现 bubbletea.Model 接口

### 3.2 TUI 主界面

- [x] 3.2.1 实现 `configModel.Init()` 函数
  - 加载当前配置
  - 初始化状态

- [x] 3.2.2 实现 `configModel.View()` 函数
  - 显示配置分组（API 配置、日志配置）
  - 显示当前配置值
  - API Key 隐藏显示（maskAPIKey）
  - Provider 显示友好名称

- [x] 3.2.3 实现 `configModel.Update()` 函数
  - 处理按键导航
  - 处理编辑、保存、重置、退出操作

- [x] 3.2.4 实现配置项高亮选择
  - 上下键移动选择
  - Enter 进入编辑

### 3.3 编辑界面

- [x] 3.3.1 创建 `pkg/tui/config/editor.go` 文件
  - 定义编辑模型 `editorModel`
  - 支持不同类型的编辑（文本、选择、密码）

- [x] 3.3.2 实现 `editorModel.View()` 函数
  - 显示当前值
  - 显示输入框
  - 显示验证状态（实时验证反馈）

- [x] 3.3.3 实现 `editorModel.Update()` 函数
  - 处理文本输入
  - 处理确认和取消

- [x] 3.3.4 实现延迟验证
  - 用户停止输入 500ms 后验证
  - 在输入框下方显示验证状态
  - 通过显示绿色 "✓ 有效"
  - 失败显示红色 "✗ 无效" 及错误原因

> 📚 **知识上下文**：基于知识库模式 #003（错误处理和重试策略）
> - 用户输入验证应该提供友好的错误提示
> - 避免在输入过程中显示错误（延迟验证）

### 3.4 Provider 选择器

- [x] 3.4.1 创建 `pkg/tui/config/provider_selector.go` 文件
  - 定义选择器模型 `providerSelectorModel`
  - 从 `GetRegisteredProviders()` 获取列表

- [x] 3.4.2 实现 `providerSelectorModel.View()` 函数
  - 显示 Provider 下拉列表
  - 显示友好名称和描述
  - 标记当前选中的 Provider

- [x] 3.4.3 实现 `providerSelectorModel.Update()` 函数
  - 处理上下选择
  - 处理确认选择

- [x] 3.4.4 实现切换 Provider 逻辑
  - 切换时清空 API Key 和 Base URL
  - 自动填充新 Provider 的默认 Base URL
  - 显示提示消息

### 3.5 保存和验证

- [x] 3.5.1 实现保存前验证
  - 运行 `ValidateConfig()`
  - 收集所有验证错误

- [x] 3.5.2 实现多错误显示
  - 列出所有错误
  - 高亮显示失败的配置项
  - 按严重程度排序

- [x] 3.5.3 实现保存功能
  - 调用 `SaveConfig()`
  - 处理保存失败的错误

- [x] 3.5.4 实现保存失败处理
  - 保留用户输入
  - 显示友好的错误信息
  - 允许用户重试或取消

> 📚 **知识上下文**：基于知识库模式 #003（错误处理和重试策略）
> - 保存失败时应保留用户输入，允许重试
> - 错误消息应包含上下文和修复建议

### 3.6 重置功能

- [x] 3.6.1 实现重置确认对话框
  - 提示用户确认
  - 显示将要重置的配置项

- [x] 3.6.2 实现重置逻辑
  - 将配置重置为默认值
  - Provider 重置为 Siliconflow
  - API Key 和 Base URL 清空
  - 日志配置重置为默认值

### 3.7 TUI 测试

- [x] 3.7.1 添加 `configModel` 单元测试
  - 测试 Init()
  - 测试 View() 渲染
  - 测试 Update() 状态转换

- [x] 3.7.2 添加 `editorModel` 单元测试
  - 测试输入处理
  - 测试验证逻辑

- [x] 3.7.3 添加 `providerSelectorModel` 单元测试
  - 测试 Provider 列表获取
  - 测试选择逻辑

- [X] 3.7.4 添加集成测试
  - 测试完整的 TUI 交互流程
  - 测试保存流程
  - 测试验证流程

---

## 4. 阶段 4：文档和发布

### 4.1 文档更新

- [x] 4.1.1 更新 README.md
  - 添加 `msa config` 命令说明
  - 添加环境变量配置说明
  - 添加 `--config` 参数说明

- [x] 4.1.2 添加配置安全注意事项
  - 说明 API Key 明文存储
  - 建议设置合适的文件权限

- [x] 4.1.3 添加 Provider 扩展指南
  - 如何添加新的 Provider
  - 如何在 ProviderRegistry 中注册

### 4.2 发布准备

- [x] 4.2.1 运行所有测试
  - 单元测试
  - 集成测试

- [x] 4.2.2 更新 CHANGELOG
  - 记录新功能
  - 记录破坏性变更（如有）

- [ ] 4.2.3 确认向后兼容性
  - 现有配置文件无需修改
  - 旧版本配置可以正常加载

---

## 基于知识的测试

> 📚 **知识来源**：知识库模式 #003（错误处理和重试策略）

### 模式 #003 回归测试：错误处理和重试策略

**历史问题**：
配置相关操作缺乏清晰的错误处理，导致用户体验差和难以排查问题。

**测试场景**：

#### 配置加载失败测试
1. 模拟配置文件读取失败（权限不足）
2. 验证程序能够启动，不崩溃
3. 验证记录了警告日志
4. 验证使用了默认配置
5. 验证显示友好的错误消息

#### 配置文件损坏测试
1. 创建包含无效 JSON 的配置文件
2. 启动程序
3. 验证程序能够启动，不崩溃
4. 验证记录了 JSON 解析失败的警告
5. 验证使用了默认配置
6. 验证提示用户运行 `msa config` 修复

#### 配置保存失败测试
1. 模拟磁盘空间不足或权限不足
2. 在 TUI 中尝试保存配置
3. 验证显示友好的错误信息
4. 验证用户输入的值被保留
5. 验证提供重试选项

#### 验证错误显示测试
1. 输入无效的 API Key（错误前缀）
2. 输入无效的 URL
3. 输入无效的日志级别
4. 尝试保存
5. 验证所有错误都被列出
6. 验证错误按严重程度排序
7. 验证提供友好的修复建议

**预期行为**：
- 所有配置操作都有明确的错误处理
- 错误消息包含上下文和修复建议
- 配置加载失败不应阻止程序启动
- 用户输入验证显示友好的错误提示
- 保存失败时保留用户输入并允许重试

---

## 验收标准

### 功能验收
- [x] `msa config` 命令能够启动 TUI 配置界面
- [x] TUI 能够显示和编辑所有配置项
- [x] Provider 选择器显示友好名称
- [x] 切换 Provider 时清空 API Key 和 Base URL
- [x] `--config` 参数支持文件路径和 key=value 格式
- [x] 环境变量正确覆盖配置文件值
- [x] 配置优先级正确：命令行 > 环境变量 > 配置文件 > 默认值
- [x] 所有验证规则正确工作
- [x] 保存失败时提供友好的错误处理

### 代码质量验收
- [x] 所有新代码有单元测试
- [ ] 测试覆盖率 >= 80%
- [ ] 代码通过 golangci-lint 检查
- [x] 避免了所有已知的反模式

### 文档验收
- [x] README 更新完整
- [x] 有安全注意事项说明
- [x] 有 Provider 扩展指南
