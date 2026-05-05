# Provider Extension Guide

MSA supports extending new LLM Providers. Follow these steps to add one.

## 1. Register in ProviderRegistry

Edit `pkg/model/comment.go` and add the new Provider to `ProviderRegistry`:

```go
var ProviderRegistry = map[LlmProvider]ProviderInfo{
    Siliconflow: {
        ID:             Siliconflow,
        DisplayName:    "SiliconFlow",
        Description:    "LLM API provider, OpenAI compatible",
        DefaultBaseURL: "https://api.siliconflow.cn/v1",
        KeyPrefix:      "sk-",
    },
    // Add new Provider
    YourProvider: {
        ID:             YourProvider,
        DisplayName:    "Your Provider Name",
        Description:    "Provider description",
        DefaultBaseURL: "https://api.yourprovider.com/v1",
        KeyPrefix:      "custom-",
    },
}
```

## 2. Add Provider Constant

In `pkg/model/comment.go`:

```go
const (
    Siliconflow  LlmProvider = "siliconflow"
    YourProvider LlmProvider = "yourprovider"
)
```

## 3. Update DisplayName/BaseURL Methods (if needed)

If the new Provider has special display name or default URL logic, add handling in the corresponding methods.

## 4. Verify

```bash
go build
msa config
# Select new Provider and verify configuration is saved
```
