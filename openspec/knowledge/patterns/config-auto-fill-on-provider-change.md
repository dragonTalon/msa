# Provider 切换时自动填充配置模式

## 模式标识

- **ID**: config-auto-fill-on-provider-change
- **类别**: 用户体验模式
- **来源**: add-config-management 实施

## 问题描述

在配置管理界面中，当用户切换 Provider（服务提供商）时，相关的配置字段（如 Base URL）应该自动填充为该 Provider 的默认值，避免用户手动查找和输入。

## 解决方案

### 1. Provider 注册表

维护一个 Provider 注册表，包含每个 Provider 的默认配置：

```go
type ProviderInfo struct {
    ID             LlmProvider
    DisplayName    string
    Description    string
    DefaultBaseURL string
    KeyPrefix      string
}

var ProviderRegistry = map[LlmProvider]ProviderInfo{
    Siliconflow: {
        ID:             Siliconflow,
        DisplayName:    "SiliconFlow (硅基流动)",
        Description:    "国内 LLM API 提供商",
        DefaultBaseURL: "https://api.siliconflow.cn/v1",
        KeyPrefix:      "sk-",
    },
    OpenAI: {
        ID:             OpenAI,
        DisplayName:    "OpenAI",
        Description:    "OpenAI 官方 API",
        DefaultBaseURL: "https://api.openai.com/v1",
        KeyPrefix:      "sk-",
    },
}
```

### 2. 获取默认值方法

提供方法获取 Provider 的默认配置：

```go
func (p LlmProvider) GetDefaultBaseURL() string {
    if info, ok := ProviderRegistry[p]; ok {
        return info.DefaultBaseURL
    }
    return ""
}
```

### 3. 切换时自动更新

在 Provider 选择器完成时，自动更新相关字段：

```go
func (m *ConfigModel) updateEditMode(msg tea.Msg) (tea.Model, tea.Cmd) {
    if selector, ok := m.state.EditorModel.(*ProviderSelectorModel); ok {
        switch msg.(type) {
        case tea.QuitMsg:
            if !selector.IsAborted() {
                // 获取新选择的 Provider
                item := m.state.Items[m.state.SelectedIndex]
                newProvider := selector.GetSelectedProvider()
                item.Value = string(newProvider)

                // 自动更新相关配置
                for _, it := range m.state.Items {
                    if it.Key == "apikey" {
                        it.Value = ""  // 清空 API Key（安全考虑）
                    }
                    if it.Key == "baseurl" {
                        it.Value = newProvider.GetDefaultBaseURL()  // 自动填充
                    }
                }

                m.state.SuccessMessage = fmt.Sprintf(
                    "已切换到 %s，Base URL 已更新，请重新配置 API Key",
                    newProvider.GetDisplayName(),
                )
            }
            m.state.EditMode = false
            return m, nil
        }
    }
    return m, nil
}
```

## 优点

1. **减少用户输入**：Base URL 自动填充，无需手动查找
2. **降低错误率**：避免用户输入错误的 URL
3. **改善用户体验**：切换流程更流畅
4. **易于扩展**：添加新 Provider 时只需更新注册表

## 安全考虑

### 清空敏感信息

切换 Provider 时，应该清空 API Key 等敏感信息：

```go
if it.Key == "apikey" {
    it.Value = ""  // 出于安全考虑，清空 API Key
}
```

原因：
- 不同 Provider 的 API Key 格式可能不同
- 避免使用错误的 API Key
- 强制用户重新配置，确保安全

### 验证 API Key 格式

新 Provider 可能对 API Key 有特定格式要求：

```go
func (p LlmProvider) ValidateAPIKey(key string) error {
    if info, ok := ProviderRegistry[p]; ok && info.KeyPrefix != "" {
        if !strings.HasPrefix(key, info.KeyPrefix) {
            return fmt.Errorf("API Key 应该以 %s 开头", info.KeyPrefix)
        }
    }
    return nil
}
```

## 用户反馈

### 成功消息

```go
m.state.SuccessMessage = fmt.Sprintf(
    "已切换到 %s，Base URL 已更新，请重新配置 API Key",
    newProvider.GetDisplayName(),
)
```

### 日志记录

```go
log.Infof("切换 Provider: %q -> %q", oldProvider, newProvider)
log.Infof("清空 API Key: %q -> %q", oldAPIKey, "")
log.Infof("更新 Base URL: %q -> %q", oldBaseURL, newBaseURL)
```

## 应用场景

- API 客户端配置
- 数据库连接配置
- 服务端点配置
- 任何需要根据选项自动填充相关配置的场景

## 扩展模式

### 级联更新

切换一个字段时，更新多个相关字段：

```go
switch newProvider {
case Siliconflow:
    updateModel("deepseek-ai/DeepSeek-V3")
    updateMaxTokens(4000)
case OpenAI:
    updateModel("gpt-4")
    updateMaxTokens(8000)
}
```

### 条件更新

根据条件决定是否更新：

```go
if oldProvider != newProvider {
    // 只有 Provider 真正改变时才更新
    updateRelatedFields(newProvider)
}
```

## 相关知识

- [Provider 注册表模式](./provider-registry-pattern.md)
- [配置验证策略](./config-validation-strategy.md)
- [用户反馈模式](./user-feedback-pattern.md)
