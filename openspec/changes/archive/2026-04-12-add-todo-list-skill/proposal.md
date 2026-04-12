# 提案：add-todo-list-skill

为交易相关 Skill 添加 TODO 列表功能，确保 Agent 在执行股票分析时按步骤完成所有操作，防止遗漏和执行不确定性。

## 知识上下文

### 相关历史问题

知识库中未找到与本变更直接相关的历史问题。这是新领域，如遇问题请创建 fix.md 来构建知识库。

### 需避免的反模式

未识别到特定反模式。

### 推荐模式

✅ **推荐**：考虑使用以下模式：

- **#003 错误处理和重试策略**：适用于 TODO 步骤失败处理场景
- **#002 搜索引擎接口设计模式**：适用于工具接口设计

## 为什么

当前 Agent 执行 `morning-analysis`、`afternoon-trade`、`market-close-summary` 三个交易 Skill 时，虽然 SKILL.md 定义了步骤，但：
1. Agent 可能跳过某些步骤
2. 步骤执行失败时缺乏明确的处理逻辑
3. 无法追踪执行进度和结果

为了防止股票分析中执行的不确定性，需要在 Agent 读取目标 Skill 后先创建 TODO 列表，按步骤执行并记录结果。

## 变更内容

- 新增 TODO 工具层（5 个工具）：check_skill_todo, create_todo, update_todo_step, verify_todo_completion, fill_todo_summary
- 修改 Skill 元数据，增加 `requires_todo` 字段
- 为三个交易 Skill 添加 TODO 模板文件
- 修改 Agent prompt，增加 TODO 创建规则
- 新增默认 TODO 模板

## 能力

### 新增能力

- `todo-management`：TODO 列表管理能力，包括创建、更新、验证和总结

### 修改能力

- `skill-system`：Skill 元数据增加 requires_todo 字段，references 支持 todo-template.md
- `agent-prompt`：技能使用规则增加 TODO 创建步骤

## 非目标

- 不修改现有 Skill 的业务逻辑
- 不改变 Session 管理机制
- 不添加持久化存储（TODO 文件本身就是持久化）

## 影响

| 影响范围 | 说明 |
|----------|------|
| `pkg/logic/tools/todo/` | 新增目录，5 个工具实现 |
| `pkg/logic/skills/skill.go` | 增加 RequiresTodo 字段 |
| `pkg/logic/skills/loader.go` | 加载 requires_todo 元数据 |
| `pkg/logic/agent/prompt.go` | 修改技能使用规则 |
| `pkg/logic/skills/plugs/morning-analysis/` | 增加 requires_todo 和 todo-template.md |
| `pkg/logic/skills/plugs/afternoon-trade/` | 增加 requires_todo 和 todo-template.md |
| `pkg/logic/skills/plugs/market-close-summary/` | 增加 requires_todo 和 todo-template.md |

---

**知识备注**：
- 此提案使用 `openspec/knowledge/index.yaml` 中的知识生成
- 知识库总问题数：5
- 总模式数：6
