# deepseek-provider 规格说明

## Purpose

DeepSeek 作为 LLM 提供商集成到 MSA 中，支持官方 API 的模型列表获取、思考模式参数配置及多轮工具调用中的 reasoning_content 回传。

## Requirements

### Requirement: provider-registration

DeepSeek 作为 LLM 提供商注册到 `ProviderRegistry` 和 `providerMap`，可通过 TUI 配置界面选择。

#### Scenario: Provider 常量定义

- **GIVEN** 需要支持 DeepSeek 官方 API
- **WHEN** 定义 `model.Deepseek` 常量
- **THEN** `LlmProvider` 类型值为 `"deepseek"`
- **AND** `ProviderRegistry` 包含 DeepSeek 条目，DisplayName 为 `"DeepSeek (深度求索)"`
- **AND** `DefaultBaseURL` 为 `"https://api.deepseek.com"`
- **AND** `KeyPrefix` 为 `"sk-"`

#### Scenario: TUI Provider 切换

- **GIVEN** 用户在 TUI 配置界面选择 DeepSeek provider
- **WHEN** 切换完成
- **THEN** API Key 自动清空
- **AND** Base URL 自动更新为 `"https://api.deepseek.com"`
- **AND** 提示用户重新配置 API Key

#### Scenario: 环境变量配置

- **GIVEN** 设置了环境变量 `MSA_PROVIDER=deepseek` 和 `MSA_API_KEY=sk-xxx`
- **WHEN** 应用启动
- **THEN** 配置中的 Provider 为 `"deepseek"`
- **AND** API Key 为 `"sk-xxx"`
- **AND** `InitConfig()` 正确合并环境变量配置

---

### Requirement: model-list

实现 `LLMProvider` 接口的 `ListModels()` 方法，从 DeepSeek API 获取可用模型列表。

#### Scenario: 获取模型列表成功

- **GIVEN** 配置了有效的 DeepSeek API Key
- **WHEN** 调用 `DeepseekProvider.ListModels(ctx)`
- **THEN** 向 `GET https://api.deepseek.com/models` 发送请求（使用配置的 BaseURL）
- **AND** 携带 `Authorization: Bearer {APIKey}` header
- **AND** 解析响应的 `data` 数组，每项包含 `id` 和 `object` 字段
- **AND** 返回 `[]*model.LLMModel` 切片

#### Scenario: 模型列表请求失败

- **GIVEN** API Key 无效或网络不可达
- **WHEN** 调用 `ListModels(ctx)`
- **THEN** 返回 error 包含 HTTP 状态码和响应体信息
- **AND** 不 panic

---

### Requirement: thinking-mode-parameter

对于模型名以 `deepseek-` 开头的请求，正确设置 DeepSeek 思考模式参数。

#### Scenario: deepseek 模型启用思考模式

- **GIVEN** 配置的 Model 以 `"deepseek-"` 开头（如 `"deepseek-chat"`）
- **WHEN** `agent.New()` 创建 ChatModelConfig
- **THEN** `ExtraFields` 包含 `"thinking": {"type": "enabled"}`
- **AND** 不使用已废弃的 `"enable_thinking": true` 格式

#### Scenario: 非 deepseek 模型不设置 thinking

- **GIVEN** 配置的 Model 不以 `"deepseek-"` 开头
- **WHEN** `agent.New()` 创建 ChatModelConfig
- **THEN** `ExtraFields` 不包含 `"thinking"` 字段
- **AND** 使用 `ReasoningEffort = "medium"`（现有行为）

---

### Requirement: reasoning-content-preservation

多轮工具调用中，`reasoning_content` 在后续请求中正确回传给 DeepSeek API。

#### Scenario: StreamAdapter 接收 reasoning_content

- **GIVEN** DeepSeek API 返回的流式 chunk 包含 `delta.reasoning_content`
- **WHEN** Eino openai client 解析 chunk
- **THEN** `schema.Message.ReasoningContent` 被正确填充
- **AND** `StreamAdapter` 将其作为 `EventThinking` 事件发送到 UI

#### Scenario: ConcatMessages 合并 reasoning_content

- **GIVEN** 多个流式 chunk 包含分散的 `reasoning_content`
- **WHEN** `schema.ConcatMessages(chunks)` 合并消息
- **THEN** 合并后的 `Message.ReasoningContent` 是所有 chunk 的拼接
- **AND** 不丢失任何 reasoning 内容

#### Scenario: 下一轮请求回传 reasoning_content

- **GIVEN** 前一轮 LLM 响应包含 `reasoning_content` 和 `tool_calls`
- **WHEN** ReAct agent 发起下一轮 LLM 请求
- **THEN** 请求 messages 中包含 assistant 消息的完整 `reasoning_content`
- **AND** 不返回 400 错误

---

### Requirement: api-key-validation

DeepSeek API Key 格式验证。

#### Scenario: API Key 前缀验证

- **GIVEN** Provider 为 `"deepseek"`
- **WHEN** 验证 API Key
- **THEN** Key 必须以 `"sk-"` 开头
- **AND** Key 长度不少于 10 个字符
- **AND** 不符合时返回验证错误
