# 实现任务：add-todo-list-skill

## ⚠️ 风险警报

基于历史问题识别的风险：

### 高：Agent 可能忽略 TODO 规则

- **相关问题**：#001 AI 助手不应自动执行 git commit
- **缓解措施**：在 prompt 中使用强语气"MUST"，明确步骤顺序，添加验证检查

### 中：TODO 文件格式被破坏

- **相关问题**：无直接历史问题
- **缓解措施**：update_todo_step 等工具做格式验证，解析失败时返回明确错误

---

## 1. 基础设施 - TODO 工具层

### 1.1 创建工具目录和常量

- [x] 1.1.1 创建 `pkg/logic/tools/todo/` 目录
- [x] 1.1.2 在 `pkg/model/constant.go` 增加 `TodoToolGroup` 常量

> 📚 **知识上下文**：参考模式 #002 工具接口设计模式，使用 Eino 框架的 tool 接口

### 1.2 实现 TODO 文件解析器

- [x] 1.2.1 创建 `pkg/logic/tools/todo/parser.go`
- [x] 1.2.2 实现 `ParseTodoFile(path)` 解析 TODO 文件
- [x] 1.2.3 实现 `UpdateStepStatus(path, stepId, status)` 更新步骤状态
- [x] 1.2.4 实现 `CountSteps(content)` 统计各状态步骤数

验收标准：
- [x] 解析支持 5 种状态：`[ ]`, `[x]`, `[!]`, `[-]`, `[~]`
- [x] 步骤 ID 格式验证（X.Y 格式）
- [x] 解析失败时返回明确错误信息

### 1.3 创建默认 TODO 模板

- [x] 1.3.1 创建 `pkg/logic/tools/todo/default-template.md`

### 1.4 实现 check_skill_todo 工具

- [x] 1.4.1 创建 `pkg/logic/tools/todo/check.go`
- [x] 1.4.2 实现 `CheckSkillTodo(ctx, param)` 函数
- [x] 1.4.3 返回 `requires_todo`, `has_custom_template`, `template_source`

### 1.5 实现 create_todo 工具

- [x] 1.5.1 创建 `pkg/logic/tools/todo/create.go`
- [x] 1.5.2 获取当前 Session ID
- [x] 1.5.3 优先使用 Skill 自定义模板，否则使用默认模板
- [x] 1.5.4 创建 `~/.msa/todos/{session-id}/{skill-name}.md` 文件
- [x] 1.5.5 返回 `todo_path`, `content`, `total_steps`

> 📚 **知识上下文**：避免反模式 - 不要硬编码模板路径，优先使用 Skill 自定义模板

### 1.6 实现 update_todo_step 工具

- [x] 1.6.1 创建 `pkg/logic/tools/todo/update.go`
- [x] 1.6.2 实现 `UpdateTodoStep(ctx, param)` 函数
- [x] 1.6.3 支持状态：done, failed, handled, skipped
- [x] 1.6.4 更新文件中对应步骤的状态标记

### 1.7 实现 verify_todo_completion 工具

- [x] 1.7.1 创建 `pkg/logic/tools/todo/verify.go`
- [x] 1.7.2 解析 TODO 文件，统计各状态步骤
- [x] 1.7.3 返回 `all_complete`, `has_pending`, `has_unhandled_fail`
- [x] 1.7.4 返回 `pending_steps`, `failed_steps`, `handled_steps`, `skipped_steps`
- [x] 1.7.5 返回 `suggestion` 提示下一步操作

### 1.8 实现 fill_todo_summary 工具

- [x] 1.8.1 创建 `pkg/logic/tools/todo/summary.go`
- [x] 1.8.2 实现 `FillTodoSummary(ctx, param)` 函数
- [x] 1.8.3 更新 TODO 文件的"执行总结"部分

### 1.9 注册 TODO 工具

- [x] 1.9.1 创建 `pkg/logic/tools/todo/register.go`
- [x] 1.9.2 在 `pkg/logic/tools/register.go` 中注册 TODO 工具

验收标准：
- [x] 运行 `go vet ./...` 无错误
- [x] 5 个工具均可通过 Agent 调用

---

## 2. Skill 系统扩展

### 2.1 扩展 Skill 结构体

- [x] 2.1.1 在 `pkg/logic/skills/skill.go` 的 `SkillMetadata` 增加 `RequiresTodo bool`
- [x] 2.1.2 在 `Skill` 结构体增加 `HasTodoTemplate()` 方法
- [x] 2.1.3 在 `Skill` 结构体增加 `GetTodoTemplate()` 方法

### 2.2 修改 Skill Loader

- [x] 2.2.1 在 `pkg/logic/skills/loader.go` 解析 `requires_todo` 元数据
- [x] 2.2.2 设置到 Skill 的 Metadata.RequiresTodo 字段

### 2.3 扩展 Skill Manager

- [x] 2.3.1 在 `pkg/logic/skills/manager.go` 增加 `GetTodoTemplate(name)` 方法
- [x] 2.3.2 在 `pkg/logic/skills/manager.go` 增加 `RequiresTodo(name)` 方法

验收标准：
- [x] 运行 `go vet ./...` 无错误
- [x] 现有 Skill 加载不受影响
- [x] 新字段可正确读取

---

## 3. Agent Prompt 修改

### 3.1 修改技能列表显示

- [x] 3.1.1 在 `pkg/logic/agent/prompt.go` 技能列表模板增加 TODO 标记
- [x] 3.1.2 当 `RequiresTodo` 为 true 时显示 `📋 需创建TODO`

### 3.2 修改技能使用规则

- [x] 3.2.1 在技能使用规则第 2 步后增加 TODO 检查步骤
- [x] 3.2.2 明确说明：检查 → 创建 → 执行 → 更新 → 验证 → 总结

> 📚 **知识上下文**：来源问题 #001 - 使用"MUST"强语气确保 Agent 遵守规则

验收标准：
- [x] prompt 模板语法正确
- [x] 运行 `go vet ./...` 无错误

---

## 4. Skill 模板配置

### 4.1 morning-analysis Skill

- [x] 4.1.1 修改 `pkg/logic/skills/plugs/morning-analysis/SKILL.md` 增加 `requires_todo: true`
- [x] 4.1.2 创建 `pkg/logic/skills/plugs/morning-analysis/references/todo-template.md`

### 4.2 afternoon-trade Skill

- [x] 4.2.1 修改 `pkg/logic/skills/plugs/afternoon-trade/SKILL.md` 增加 `requires_todo: true`
- [x] 4.2.2 创建 `pkg/logic/skills/plugs/afternoon-trade/references/todo-template.md`

### 4.3 market-close-summary Skill

- [x] 4.3.1 修改 `pkg/logic/skills/plugs/market-close-summary/SKILL.md` 增加 `requires_todo: true`
- [x] 4.3.2 创建 `pkg/logic/skills/plugs/market-close-summary/references/todo-template.md`

验收标准：
- [x] 三个 Skill 的 SKILL.md 格式正确
- [x] 三个 todo-template.md 内容完整

---

## 5. 集成测试

### 5.1 单元测试

- [x] 5.1.1 创建 `pkg/logic/tools/todo/parser_test.go`
- [x] 5.1.2 测试 TODO 文件解析
- [x] 5.1.3 测试步骤状态更新
- [x] 5.1.4 测试完成情况验证

### 5.2 集成测试

- [x] 5.2.1 测试 check_skill_todo 工具
- [x] 5.2.2 测试 create_todo 工具
- [x] 5.2.3 测试 update_todo_step 工具
- [x] 5.2.4 测试 verify_todo_completion 工具
- [x] 5.2.5 测试 fill_todo_summary 工具

验收标准：
- [x] 运行 `go test ./pkg/logic/tools/todo/...` 全部通过
- [x] 运行 `go vet ./...` 无错误

---

## 基于知识的测试

### 问题 #001 回归测试

> 📚 **来源**：问题 #001 - AI 助手不应自动执行 git commit

**历史问题**：
AI 助手可能忽略规则，自动执行不应执行的操作。

**测试场景**：
1. Agent 执行需要 TODO 的 Skill
2. 验证 Agent 是否按顺序调用 TODO 工具
3. 验证 Agent 是否在 TODO 未完成时给出提示

**预期行为**：
- Agent 必须先调用 check_skill_todo
- Agent 必须在步骤完成后调用 update_todo_step
- Agent 必须在所有步骤完成后调用 verify_todo_completion

### 错误处理模式测试

> 📚 **来源**：模式 #003 错误处理和重试策略

**测试场景**：
1. 创建 TODO 文件
2. 模拟步骤失败（标记为 `[!]`）
3. 执行失败处理
4. 标记为已处理 `[-]`
5. 验证 verify_todo_completion 返回正确状态

**预期行为**：
- `has_unhandled_fail` 从 true 变为 false
- `handled_steps` 包含已处理步骤
- `all_complete` 最终为 true
