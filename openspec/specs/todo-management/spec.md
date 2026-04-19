# 规格：todo-management

> 📚 **知识上下文**：参考模式 #003 错误处理和重试策略，应用于 TODO 步骤失败处理场景

## Purpose

系统 SHALL 提供完整的 TODO 管理能力，MUST 支持 Agent 按需创建 TODO 文件、更新步骤状态、验证完成情况和填写执行总结，以结构化方式追踪 Skill 的执行进度。

## Requirements

### Requirement: 检查 Skill 是否需要 TODO

Agent 在调用 get_skill_content 后，MUST 检查该 Skill 是否需要创建 TODO 列表。

#### Scenario: 检查标记为需要 TODO 的 Skill
- **当** Skill 元数据中 `requires_todo` 为 `true`
- **那么** check_skill_todo 工具返回 `requires_todo: true`

#### Scenario: 检查不需要 TODO 的 Skill
- **当** Skill 元数据中 `requires_todo` 为 `false` 或不存在
- **那么** check_skill_todo 工具返回 `requires_todo: false`

---

### Requirement: 创建 TODO 文件

当 Skill 需要 TODO 时，系统 MUST 创建 TODO 文件记录执行步骤。

#### Scenario: 使用 Skill 自定义模板创建 TODO
- **当** Skill 的 references 目录存在 todo-template.md
- **那么** 使用该模板创建 TODO 文件
- **并且** TODO 文件路径为 `~/.msa/todos/{session-id}/{skill-name}.md`

#### Scenario: 使用默认模板创建 TODO
- **当** Skill 的 references 目录不存在 todo-template.md
- **那么** 使用默认模板创建 TODO 文件
- **并且** 默认模板位于 `pkg/logic/tools/todo/default-template.md`

#### Scenario: TODO 文件内容包含元数据
- **当** TODO 文件被创建
- **那么** 文件头部包含创建时间和会话 ID
- **并且** 包含"执行状态"部分用于标记整体完成情况

---

### Requirement: 更新步骤状态

Agent 每完成一个步骤，MUST 更新 TODO 文件中的步骤状态。

#### Scenario: 标记步骤为成功
- **当** 调用 update_todo_step(status="done")
- **那么** 对应步骤标记为 `[x]`

#### Scenario: 标记步骤为失败
- **当** 调用 update_todo_step(status="failed")
- **那么** 对应步骤标记为 `[!]`
- **并且** 可以附加失败原因备注

#### Scenario: 标记失败已处理
- **当** 调用 update_todo_step(status="handled")
- **那么** 对应步骤标记为 `[-]`
- **并且** 表示失败处理逻辑已执行

#### Scenario: 标记步骤为跳过
- **当** 调用 update_todo_step(status="skipped")
- **那么** 对应步骤标记为 `[~]`

---

### Requirement: 验证完成情况

所有步骤执行完成后，系统 MUST 验证 TODO 完成情况。

#### Scenario: 所有步骤成功完成
- **当** 所有步骤状态为 `[x]`
- **那么** verify_todo_completion 返回 `all_complete: true`
- **并且** `has_pending: false`
- **并且** `has_unhandled_fail: false`

#### Scenario: 有待执行步骤
- **当** 存在状态为 `[ ]` 的步骤
- **那么** verify_todo_completion 返回 `has_pending: true`
- **并且** `pending_steps` 列出待执行步骤 ID

#### Scenario: 有未处理的失败步骤
- **当** 存在状态为 `[!]` 的步骤
- **那么** verify_todo_completion 返回 `has_unhandled_fail: true`
- **并且** `failed_steps` 列出失败步骤 ID
- **并且** `suggestion` 提示执行失败处理

---

### Requirement: 填写执行总结

验证通过后，Agent MUST 填写 TODO 文件的执行总结部分。

#### Scenario: 填写总结内容
- **当** 调用 fill_todo_summary
- **那么** 更新 TODO 文件的"执行总结"部分
- **并且** 包含完成情况统计（总步骤数、成功数、失败数、跳过数）
- **并且** 包含最终结论

---

### Requirement: 步骤标识格式

TODO 文件中的步骤 MUST 使用统一的数字标识格式。

#### Scenario: 步骤 ID 格式
- **当** 定义 TODO 步骤
- **那么** 使用 `X.Y` 格式，其中 X 为任务编号，Y 为步骤编号
- **例如** `1.1`, `1.2`, `2.1`, `2.2`

#### Scenario: 失败处理关联
- **当** 定义失败处理逻辑
- **那么** 使用 `如果 X.Y 失败：[处理方式]` 格式
