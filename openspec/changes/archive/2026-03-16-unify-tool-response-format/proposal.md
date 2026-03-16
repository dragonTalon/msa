## Why

当前项目中 tool 函数在发生错误时返回 Go error，这会导致 eino React Agent 直接中断整个执行流程，Agent 无法继续处理。同时，不同模块的响应格式不统一：
- Search/Fetcher 模块已使用结构化 JSON（含 success/error_msg 字段）
- Finance/Stock/Knowledge/Skill 模块仍返回文本 + Go error

需要统一所有 tool 的响应格式，确保 Agent 在遇到业务错误时不会中断，而是能够理解错误信息并做出决策。

## What Changes

- **新增** `pkg/model/tool_result.go`：定义统一的 `ToolResult` 结构体和辅助函数
- **修改** 所有 tool 函数：不再返回 Go error，改为返回结构化 JSON
- **修改** `pkg/model/search.go`：`WebSearchResponse` 和 `FetchPageResponse` 改用统一格式
- **修改** 所有 tool 实现文件（finance、stock、knowledge、skill、search）

统一响应格式：
```json
{
  "success": true/false,
  "data": { ... } 或 null,
  "message": "成功描述" 或 null,
  "error_msg": "错误描述" 或 null
}
```

## Capabilities

### New Capabilities

- `tool-response-format`：定义所有 tool 的统一响应结构，包含 success、data、message、error_msg 字段，以及辅助函数

### Modified Capabilities

无（这是新增能力，不修改现有规格要求）

## Impact

- **受影响文件**：
  - `pkg/model/tool_result.go`（新建）
  - `pkg/model/search.go`
  - `pkg/logic/tools/search/search.go`
  - `pkg/logic/tools/search/fetcher.go`
  - `pkg/logic/tools/finance/account.go`
  - `pkg/logic/tools/finance/trade.go`
  - `pkg/logic/tools/finance/position.go`
  - `pkg/logic/tools/stock/company.go`
  - `pkg/logic/tools/stock/company_info.go`
  - `pkg/logic/tools/stock/company_k.go`
  - `pkg/logic/tools/knowledge/read.go`
  - `pkg/logic/tools/knowledge/write.go`
  - `pkg/logic/tools/knowledge/list.go`
  - `pkg/logic/tools/knowledge/session.go`
  - `pkg/logic/tools/skill/skill.go`

- **API 变化**：所有 tool 函数签名从 `(string, error)` 改为 `(string, nil)`，错误信息封装在 JSON 中
- **向后兼容**：LLM 需要适配新的响应格式，但不会破坏现有功能

## Non-Goals

- 不改变 tool 的参数结构
- 不改变 tool 的注册机制
- 不修改 eino 框架层的任何代码
- 不为每个 tool 定义专门的 Data 类型（使用 `interface{}`）
