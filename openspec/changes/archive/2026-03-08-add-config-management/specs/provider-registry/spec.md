# 规格：provider-registry

Provider 注册表能力，管理 LLM 提供商的元信息和验证规则。

## 增加需求

### 需求：Provider 注册表

维护所有可用 Provider 的元信息注册表。

#### 场景：注册 Provider
- **当** 新的 Provider 被添加到系统
- **那么** 应在 `ProviderRegistry` 中添加对应的 `ProviderInfo`
- **并且** 包含以下信息：
  - `ID` - Provider 唯一标识符
  - `DisplayName` - 友好显示名称
  - `Description` - Provider 描述
  - `DefaultBaseURL` - 默认 API 基础 URL
  - `KeyPrefix` - API Key 前缀（用于验证）

#### 场景：获取友好名称
- **当** 需要显示 Provider 名称
- **那么** 应使用 `GetDisplayName()` 方法获取友好名称
- **并且** 如果未注册，返回原始 ID 字符串

#### 场景：获取默认 Base URL
- **当** 需要获取 Provider 的默认 Base URL
- **那么** 应使用 `GetDefaultBaseURL()` 方法
- **并且** 如果未注册，返回空字符串

#### 场景：列出已注册的 Provider
- **当** TUI 需要显示 Provider 选择器
- **那么** 应从 `ProviderRegistry` 获取所有已注册的 Provider
- **并且** 按注册顺序或名称排序显示

### 需求：Provider 信息结构

定义 Provider 元信息的标准结构。

#### 场景：ProviderInfo 结构
- **当** 定义 Provider 信息
- **那么** 应包含以下字段：
  ```
  type ProviderInfo struct {
      ID            LlmProvider
      DisplayName   string
      Description   string
      DefaultBaseURL string
      KeyPrefix     string
  }
  ```

#### 场景：默认 Provider 信息
- **当** 系统初始化
- **那么** `ProviderRegistry` 应至少包含 Siliconflow：
  ```
  Siliconflow: {
      ID:            "siliconflow",
      DisplayName:   "SiliconFlow (硅基流动)",
      Description:   "国内 LLM API 提供商，兼容 OpenAI 格式",
      DefaultBaseURL: "https://api.siliconflow.cn/v1",
      KeyPrefix:     "sk-",
  }
  ```

### 需求：API Key 验证

为每个 Provider 提供 API Key 格式验证。

#### 场景：验证 API Key 前缀
- **当** 用户输入 API Key
- **并且** Provider 定义了 `KeyPrefix`
- **那么** 系统应验证 API Key 是否以该前缀开头
- **并且** 如果不匹配，显示错误信息

#### 场景：跳过验证
- **当** API Key 验证失败
- **那么** 系统应提供选项让用户强制保存
- **并且** 显示警告提示风险

#### 场景：无前缀要求
- **当** Provider 的 `KeyPrefix` 为空
- **那么** 系统应跳过前缀验证
- **并且** 接受任何非空 API Key

### 需求：Provider 扩展性

支持动态添加新的 Provider。

#### 场景：添加新 Provider
- **当** 开发者添加新的 LLM Provider
- **那么** 只需在 `ProviderRegistry` 中添加新的 `ProviderInfo`
- **并且** TUI 会自动显示在 Provider 选择器中
- **并且** 无需修改 TUI 代码

#### 场景：Provider 选择器更新
- **当** `ProviderRegistry` 中添加新的 Provider
- **那么** 下次打开 TUI 配置界面时
- **并且** 新 Provider 应自动出现在选择列表中

### 需求：与现有 Provider 系统集成

与现有的 `providerMap` 系统协同工作。

#### 场景：导出 Provider 列表
- **当** TUI 需要获取可用 Provider 列表
- **那么** 应提供 `GetRegisteredProviders()` 函数
- **并且** 返回 `ProviderRegistry` 中的所有 Provider ID

#### 场景：同步状态
- **当** Provider 通过 `RegisterProvider()` 注册到 `providerMap`
- **那么** 对应的 `ProviderInfo` 应存在于 `ProviderRegistry`
- **并且** 确保 `providerMap` 和 `ProviderRegistry` 保持同步
