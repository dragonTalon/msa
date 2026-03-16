## Context

当前项目中 tool 函数的错误处理存在两种模式：

1. **Search/Fetcher 模块**（已改进）：返回结构化 JSON，包含 `success`、`error_msg` 字段，永远不返回 Go error
2. **Finance/Stock/Knowledge/Skill 模块**（旧模式）：返回 `(string, error)`，错误时返回 Go error

问题在于 eino React Agent 在 tool 返回 Go error 时会直接中断执行流程（参见 eino 源码 `tool_node.go` 第 842-846 行），导致 Agent 无法理解错误并做出后续决策。

## Goals / Non-Goals

**Goals:**
- 统一所有 tool 的响应格式为 `{success, data, message, error_msg}`
- 确保 tool 永远返回 `(string, nil)`，不返回 Go error
- 提供简洁的辅助函数方便生成标准响应
- LLM 能够理解成功/失败状态并做出决策

**Non-Goals:**
- 不改变 tool 的参数结构和注册机制
- 不修改 eino 框架层代码
- 不为每个 tool 定义专门的 Data 类型
- 不保留错误码（error_code）字段

## Decisions

### 1. 响应结构设计

**决定**：采用嵌套 `data` 的格式

```go
type ToolResult struct {
    Success  bool        `json:"success"`
    Data     interface{} `json:"data,omitempty"`
    Message  string      `json:"message,omitempty"`   // 成功时的描述
    ErrorMsg string      `json:"error_msg,omitempty"` // 失败时的错误信息
}
```

**理由**：
- 结构清晰，`data` 嵌套让业务数据与元数据分离
- 不需要 `error_code`，错误信息足够 LLM 理解
- 与现有的 `FetchPageResponse` 格式相近，迁移成本低

**替代方案**：平铺字段（如现有 `FetchPageResponse`）
- 缺点：每个 tool 需要定义专门的响应结构，增加维护成本

### 2. 辅助函数设计

**决定**：提供两个核心辅助函数

```go
// 成功响应
func NewSuccessResult(data interface{}, message string) string

// 错误响应
func NewErrorResult(message string) string
```

**理由**：
- 最小化 API，覆盖 99% 的使用场景
- 自动处理 JSON 序列化
- `data` 使用 `interface{}` 保持灵活性

### 3. 现有响应类型的处理

**决定**：修改 `WebSearchResponse` 和 `FetchPageResponse` 使用统一格式

**理由**：
- 保持整个项目的一致性
- 现有结构（如 `SearchResultItem`）作为 `data` 的内容

### 4. 改造策略

**决定**：渐进式改造，按模块顺序进行

顺序：Search → Fetcher → Finance → Stock → Knowledge → Skill

**理由**：
- Search/Fetcher 已经接近目标格式，改动最小，可验证方案
- Finance/Stock 是核心业务，需要仔细处理
- Knowledge/Skill 相对独立

## Risks / Trade-offs

**[风险] LLM 适配成本**
- 新格式需要 LLM 理解 `success` 字段
- 缓解：格式简单直观，主流 LLM 都能理解 JSON 布尔值

**[风险] Data 字段缺乏类型安全**
- 使用 `interface{}` 丧失编译时类型检查
- 缓解：这是工具层，不需要严格的类型约束；后续如需要可增加泛型版本

**[风险] 现有代码中的错误处理逻辑**
- 部分 tool 有复杂的错误处理链
- 缓解：保持 `BroadcastToolEnd` 调用不变，只改变返回值格式

**[权衡] 不保留错误码**
- 优点：简化结构，LLM 理解更容易
- 缺点：无法程序化区分错误类型
- 缓解：错误信息已经足够 LLM 做决策，如后续需要可扩展
