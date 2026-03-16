## ADDED Requirements

### Requirement: 统一的 Tool 响应结构

所有 tool 函数 SHALL 返回统一的 JSON 响应结构，包含以下字段：
- `success`：布尔值，表示操作是否成功
- `data`：操作成功时返回的业务数据，失败时为 `null`
- `message`：成功时的描述信息
- `error_msg`：失败时的错误信息

#### Scenario: 成功响应格式

- **WHEN** tool 执行成功
- **THEN** 返回 JSON 包含 `success: true`、`data` 字段和可选的 `message` 字段
- **AND** 不包含 `error_msg` 字段

#### Scenario: 失败响应格式

- **WHEN** tool 执行失败
- **THEN** 返回 JSON 包含 `success: false`、`data: null` 和 `error_msg` 字段
- **AND** 不包含 `message` 字段

### Requirement: Tool 函数不返回 Go error

所有 tool 函数 SHALL 返回 `(string, nil)`，即永远不返回 Go error。

#### Scenario: 错误时返回 JSON 而非 error

- **WHEN** tool 遇到业务错误（如数据库错误、参数错误）
- **THEN** 函数返回包含 `success: false` 和 `error_msg` 的 JSON 字符串
- **AND** 第二个返回值为 `nil`（不是 Go error）

### Requirement: 辅助函数

系统 SHALL 提供辅助函数生成标准响应：
- `NewSuccessResult(data interface{}, message string) string`：生成成功响应
- `NewErrorResult(message string) string`：生成失败响应

#### Scenario: 使用 NewSuccessResult

- **WHEN** 调用 `NewSuccessResult(data, "操作成功")`
- **THEN** 返回有效的 JSON 字符串
- **AND** JSON 可被正确解析为 `ToolResult` 结构

#### Scenario: 使用 NewErrorResult

- **WHEN** 调用 `NewErrorResult("参数错误")`
- **THEN** 返回有效的 JSON 字符串
- **AND** `success` 为 `false`
- **AND** `error_msg` 为 "参数错误"

### Requirement: 现有响应类型迁移

`WebSearchResponse` 和 `FetchPageResponse` SHALL 改为使用 `ToolResult` 结构。

#### Scenario: WebSearchResponse 迁移

- **WHEN** `web_search` tool 返回结果
- **THEN** 返回格式为 `{success, data: {query, results, ...}, message}` 或 `{success: false, data: null, error_msg}`

#### Scenario: FetchPageResponse 迁移

- **WHEN** `fetch_page_content` tool 返回结果
- **THEN** 返回格式为 `{success, data: {url, title, content, ...}, message}` 或 `{success: false, data: null, error_msg}`
