# 设计：add-todo-list-skill

## 上下文

当前 MSA 系统已有完善的 Skill 系统和 Session 管理。Agent 执行交易相关 Skill 时需要更严格的步骤追踪和失败处理机制。

**现有组件**：
- `pkg/logic/skills/` - Skill 管理系统，支持元数据、references、assets
- `pkg/logic/tools/skill/` - Skill 相关工具（get_skill_content, get_skill_reference, get_skill_asset）
- `pkg/session/` - Session 管理，提供 UUID、SessionID
- `pkg/logic/agent/prompt.go` - Agent prompt 模板

**新增组件**：
- `pkg/logic/tools/todo/` - TODO 工具层

## 目标 / 非目标

**目标：**
- 为 morning-analysis、afternoon-trade、market-close-summary 三个 Skill 添加 TODO 列表功能
- 实现 5 个 TODO 工具：check_skill_todo, create_todo, update_todo_step, verify_todo_completion, fill_todo_summary
- 支持 Skill 自定义 TODO 模板
- 修改 Agent prompt 增加 TODO 创建规则

**非目标：**
- 不修改现有 Skill 的业务逻辑
- 不改变 Session 管理机制
- 不添加数据库存储（TODO 文件本身就是持久化）
- 不支持 TODO 步骤的并行执行

## 推荐模式

### 错误处理和重试策略

**模式ID**：p003
**来源**：openspec/knowledge/index.yaml

**描述**：
TODO 步骤失败时，Agent 从 TODO 文件读取失败处理逻辑并执行。失败处理逻辑由 Skill 作者在 todo-template.md 中预定义。

**适用场景**：
- 工具调用失败（如 API 超时、网络错误）
- 数据验证失败
- 业务规则违反

**实现**：
```
步骤执行流程：
1. 执行步骤
2. 如果失败：
   a. 调用 update_todo_step(status="failed", note="原因")
   b. 从 TODO 内容读取 "如果 X.Y 失败：[处理方式]"
   c. 执行失败处理
   d. 调用 update_todo_step(status="handled")
3. 如果成功：
   a. 调用 update_todo_step(status="done")
```

### 工具接口设计模式

**模式ID**：p002
**来源**：openspec/knowledge/index.yaml

**描述**：
TODO 工具遵循现有工具接口规范，使用 Eino 框架的 tool 接口，返回统一的 Result 格式。

**适用场景**：
- 所有 TODO 工具

**实现**：
```go
// 工具结构
type TodoTool struct{}

func (t *TodoTool) GetToolInfo() (tool.BaseTool, error) {
    return utils.InferTool(t.GetName(), t.GetDescription(), t.Execute)
}

func (t *TodoTool) GetToolGroup() model.ToolGroup {
    return model.TodoToolGroup // 新增工具组
}

// 返回格式
return model.NewSuccessResult(data, "操作成功描述")
return model.NewErrorResult("错误描述")
```

## 避免的反模式

### 硬编码模板路径

**问题所在**：
在代码中硬编码模板路径会导致灵活性降低，无法支持 Skill 自定义模板。

**不要这样做**：
```go
templatePath := "pkg/logic/tools/todo/default-template.md"
```

**应该这样做**：
```go
// 优先使用 Skill 自定义模板
if skill.HasTodoTemplate() {
    return skill.GetTodoTemplate()
}
// 回退到默认模板
return getDefaultTemplate()
```

### 在 Agent 层面持久化状态

**问题所在**：
依赖 Agent 记住执行状态会导致不可靠，Agent 可能忘记或混淆。

**不要这样做**：
期望 Agent 在内存中记住哪些步骤已完成。

**应该这样做**：
每次状态变更都写入 TODO 文件，Agent 通过工具读取文件获取状态。

## 设计决策

### D1: TODO 文件存储位置

**决策**：存储在 `~/.msa/todos/{session-id}/{skill-name}.md`

**理由**：
- 与现有 Session 文件位置 `~/.msa/memory/` 一致
- 按会话分组，方便管理
- 用户可以手动查看和编辑

### D2: 步骤状态标记符号

**决策**：使用特定符号标记状态

| 状态 | 符号 | 说明 |
|------|------|------|
| 待执行 | `[ ]` | 未开始 |
| 成功 | `[x]` | 已完成 |
| 失败 | `[!]` | 未处理 |
| 已处理 | `[-]` | 失败后已执行处理 |
| 跳过 | `[~]` | 被跳过 |

**理由**：
- 保持 Markdown 兼容性
- 易于解析和可视化
- 支持失败追踪

### D3: 模板优先级

**决策**：Skill 自定义模板 > 默认模板

**理由**：
- 允许 Skill 作者定义特定步骤和失败处理
- 默认模板提供兜底，确保所有 Skill 都能创建 TODO

### D4: 不使用数据库

**决策**：TODO 以 Markdown 文件形式存储

**理由**：
- 简单直接，无需额外数据库表
- 用户可以直接查看和编辑
- 与 Session 文件管理方式一致

## 风险 / 权衡

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| Agent 可能忽略 TODO 规则 | 高 | 在 prompt 中明确要求，使用强语气"MUST" |
| TODO 文件格式被破坏 | 中 | update_todo_step 等工具做格式验证 |
| 步骤 ID 格式不一致 | 低 | 在模板中预定义格式，工具验证格式 |
| 默认模板不适用于所有 Skill | 低 | Skill 可以提供自定义模板覆盖 |

## 架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                        Agent 执行流程                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   1. get_skill_content(skill_name)                              │
│      │                                                          │
│      ▼                                                          │
│   2. check_skill_todo(skill_name)  ◄── TodoTool                │
│      │                                                          │
│      ├── requires_todo = false → 跳过                           │
│      │                                                          │
│      ▼                                                          │
│   3. create_todo(skill_name)        ◄── TodoTool                │
│      │   - 读取 Skill todo-template.md                          │
│      │   - 或使用默认模板                                        │
│      │   - 创建 ~/.msa/todos/{session}/{skill}.md               │
│      │                                                          │
│      ▼                                                          │
│   4. 执行 Skill 步骤（循环）                                     │
│      │                                                          │
│      ├── 成功 → update_todo_step(status="done")                 │
│      │                                                          │
│      └── 失败 → update_todo_step(status="failed")               │
│                 读取失败处理逻辑                                  │
│                 执行失败处理                                      │
│                 update_todo_step(status="handled")              │
│      │                                                          │
│      ▼                                                          │
│   5. verify_todo_completion(path)   ◄── TodoTool                │
│      │                                                          │
│      ▼                                                          │
│   6. fill_todo_summary(path, ...)   ◄── TodoTool                │
│      │                                                          │
│      ▼                                                          │
│   7. 输出最终总结                                                │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## 文件结构

```
pkg/logic/tools/todo/
├── check.go           # check_skill_todo
├── create.go          # create_todo
├── update.go          # update_todo_step
├── verify.go          # verify_todo_completion
├── summary.go         # fill_todo_summary
├── parser.go          # TODO 文件解析器
├── default-template.md # 默认 TODO 模板
└── register.go        # 工具注册

pkg/logic/skills/
├── skill.go           # 增加 RequiresTodo, HasTodoTemplate
├── loader.go          # 加载 requires_todo 元数据
└── manager.go         # 增加 GetTodoTemplate, RequiresTodo

pkg/logic/skills/plugs/
├── morning-analysis/
│   ├── SKILL.md       # 增加 requires_todo: true
│   └── references/
│       └── todo-template.md  # 新增
├── afternoon-trade/
│   ├── SKILL.md
│   └── references/
│       └── todo-template.md
└── market-close-summary/
    ├── SKILL.md
    └── references/
        └── todo-template.md

pkg/model/
└── constant.go        # 增加 TodoToolGroup
```

## 迁移计划

1. **阶段一：基础设施**
   - 创建 `pkg/logic/tools/todo/` 目录和工具
   - 修改 `pkg/model/constant.go` 增加 TodoToolGroup
   - 注册工具到 `pkg/logic/tools/register.go`

2. **阶段二：Skill 系统扩展**
   - 修改 `pkg/logic/skills/skill.go` 增加 RequiresTodo 字段
   - 修改 `pkg/logic/skills/loader.go` 加载元数据
   - 修改 `pkg/logic/skills/manager.go` 增加方法

3. **阶段三：Agent Prompt 修改**
   - 修改 `pkg/logic/agent/prompt.go` 增加 TODO 规则

4. **阶段四：Skill 模板**
   - 为三个 Skill 添加 `requires_todo: true`
   - 创建各自的 `todo-template.md`

**无回滚需求**：此变更为增量添加，不影响现有功能。
