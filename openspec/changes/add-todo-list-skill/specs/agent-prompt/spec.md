# 规格：agent-prompt

## MODIFIED Requirements

### 需求：技能列表显示 TODO 标记

技能列表输出必须显示哪些技能需要创建 TODO。

#### 场景：显示 TODO 标记
- **当** Skill 的 `requires_todo` 为 `true`
- **那么** 技能列表中显示 `📋 需创建TODO` 标记

#### 场景：不显示 TODO 标记
- **当** Skill 的 `requires_todo` 为 `false` 或未设置
- **那么** 技能列表中不显示 TODO 标记

---

### 需求：技能使用规则增加 TODO 步骤

技能使用规则必须在第 2 步和第 4 步之间增加 TODO 检查和创建步骤。

#### 场景：规则流程
- **当** Agent 匹配到技能
- **步骤 1** 对照技能列表判断是否有匹配的技能模块
- **步骤 2** 调用 get_skill_content 获取技能完整内容
- **步骤 3** 检查是否需要创建 TODO 列表
  - 调用 check_skill_todo 检查
  - 如果需要，调用 create_todo 创建文件
- **步骤 4** 按需加载 references/assets
- **步骤 5** 严格按照技能内容执行
- **步骤 6** 每完成一步调用 update_todo_step 更新状态
- **步骤 7** 所有步骤完成后调用 verify_todo_completion 验证
- **步骤 8** 输出总结和结论

---

### 需求：失败处理规则

Agent 必须从 TODO 文件读取失败处理逻辑并执行。

#### 场景：步骤失败时读取失败处理
- **当** 某步骤执行失败
- **那么** 从 TODO 文件中查找对应的"失败处理"部分
- **并且** 读取并执行失败处理逻辑
- **并且** 调用 update_todo_step(status="handled")

#### 场景：无失败处理定义
- **当** 步骤失败但 TODO 中无对应失败处理定义
- **那么** 提示用户该步骤失败，等待用户指示

---

### 需求：验证不通过时的处理

验证结果不通过时，Agent 必须根据提示继续处理。

#### 场景：有待执行步骤
- **当** verify_todo_completion 返回 `has_pending: true`
- **那么** 继续执行待执行的步骤

#### 场景：有未处理的失败
- **当** verify_todo_completion 返回 `has_unhandled_fail: true`
- **那么** 读取失败处理逻辑并执行
- **然后** 再次调用 verify_todo_completion

#### 场景：全部完成
- **当** verify_todo_completion 返回 `all_complete: true`
- **那么** 调用 fill_todo_summary 填写总结
- **并且** 向用户输出最终结果
