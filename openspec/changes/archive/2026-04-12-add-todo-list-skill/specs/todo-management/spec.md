# 规格：todo-management

> 📚 **知识上下文**：参考模式 #003 错误处理和重试策略，应用于 TODO 步骤失败处理场景

## ADDED Requirements

### 需求：检查 Skill 是否需要 TODO

Agent 在调用 get_skill_content 后，必须检查该 Skill 是否需要创建 TODO 列表。

#### 场景：检查标记为需要 TODO 的 Skill
- **当** Skill 元数据中 `requires_todo` 为 `true`
- **那么** check_skill_todo 工具返回 `requires_todo: true`

#### 场景：检查不需要 TODO 的 Skill
- **当** Skill 元数据中 `requires_todo` 为 `false` 或不存在
- **那么** check_skill_todo 工具返回 `requires_todo: false`

---

### 需求：创建 TODO 文件

当 Skill 需要 TODO 时，必须创建 TODO 文件记录执行步骤。

#### 场景：使用 Skill 自定义模板创建 TODO
- **当** Skill 的 references 目录存在 todo-template.md
- **那么** 使用该模板创建 TODO 文件
- **并且** TODO 文件路径为 `~/.msa/todos/{session-id}/{skill-name}.md`

#### 场景：使用默认模板创建 TODO
- **当** Skill 的 references 目录不存在 todo-template.md
- **那么** 使用默认模板创建 TODO 文件
- **并且** 默认模板位于 `pkg/logic/tools/todo/default-template.md`

#### 场景：TODO 文件内容包含元数据
- **当** TODO 文件被创建
- **那么** 文件头部包含创建时间和会话 ID
- **并且** 包含"执行状态"部分用于标记整体完成情况

---

### 需求：更新步骤状态

Agent 每完成一个步骤，必须更新 TODO 文件中的步骤状态。

#### 场景：标记步骤为成功
- **当** 调用 update_todo_step(status="done")
- **那么** 对应步骤标记为 `[x]`

#### 场景：标记步骤为失败
- **当** 调用 update_todo_step(status="failed")
- **那么** 对应步骤标记为 `[!]`
- **并且** 可以附加失败原因备注

#### 场景：标记失败已处理
- **当** 调用 update_todo_step(status="handled")
- **那么** 对应步骤标记为 `[-]`
- **并且** 表示失败处理逻辑已执行

#### 场景：标记步骤为跳过
- **当** 调用 update_todo_step(status="skipped")
- **那么** 对应步骤标记为 `[~]`

---

### 需求：验证完成情况

所有步骤执行完成后，必须验证 TODO 完成情况。

#### 场景：所有步骤成功完成
- **当** 所有步骤状态为 `[x]`
- **那么** verify_todo_completion 返回 `all_complete: true`
- **并且** `has_pending: false`
- **并且** `has_unhandled_fail: false`

#### 场景：有待执行步骤
- **当** 存在状态为 `[ ]` 的步骤
- **那么** verify_todo_completion 返回 `has_pending: true`
- **并且** `pending_steps` 列出待执行步骤 ID

#### 场景：有未处理的失败步骤
- **当** 存在状态为 `[!]` 的步骤
- **那么** verify_todo_completion 返回 `has_unhandled_fail: true`
- **并且** `failed_steps` 列出失败步骤 ID
- **并且** `suggestion` 提示执行失败处理

---

### 需求：填写执行总结

验证通过后，必须填写 TODO 文件的执行总结部分。

#### 场景：填写总结内容
- **当** 调用 fill_todo_summary
- **那么** 更新 TODO 文件的"执行总结"部分
- **并且** 包含完成情况统计（总步骤数、成功数、失败数、跳过数）
- **并且** 包含最终结论

---

### 需求：步骤标识格式

TODO 文件中的步骤必须使用统一的数字标识格式。

#### 场景：步骤 ID 格式
- **当** 定义 TODO 步骤
- **那么** 使用 `X.Y` 格式，其中 X 为任务编号，Y 为步骤编号
- **例如** `1.1`, `1.2`, `2.1`, `2.2`

#### 场景：失败处理关联
- **当** 定义失败处理逻辑
- **那么** 使用 `如果 X.Y 失败：[处理方式]` 格式
