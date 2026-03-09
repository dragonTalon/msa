## REMOVED Requirements

### Requirement: 集成 Skill 选择流程
**Reason**: 前置 sub-agent 选择流程整体移除，skill 选择由 agent 自主完成。
**Migration**: 删除 `selectSkillsViaSubAgent` 调用及相关函数。

### Requirement: 用户手动指定 Skills（在集成 Skill 选择流程的 Scenario 中）
**Reason**: manualSkills 机制随前置选择器一起移除。
**Migration**: 删除 TUI 层的 sessionSkills/oneTimeSkills/`/skills:` 命令及 cmd/root.go 的 --skills 参数。

## ADDED Requirements

### Requirement: Skill 元数据注入 System Prompt
系统 SHALL 在每次构建对话消息时，将所有可用 skill 的元数据（name、description、priority）静态注入 BasePromptTemplate（SystemMessage）。

#### Scenario: 正常构建消息
- **WHEN** `BuildQueryMessages` 被调用
- **THEN** SystemMessage SHALL 包含所有可用 skill 的名称、描述和优先级列表
- **AND** skill 列表通过 `skills.GetManager().ListSkills()` 获取（已过滤禁用的）
- **AND** skill 按 priority 降序排列（registry 已保证顺序）

#### Scenario: system prompt 包含 skill 选择指令
- **WHEN** agent 收到 system prompt
- **THEN** prompt SHALL 包含明确指令：识别匹配的 skill 后须调用 `get_skill_content` 获取完整流程
- **AND** prompt SHALL 说明"先选 skill，再按 skill 流程处理用户问题"

#### Scenario: 无可用 skill 时
- **WHEN** 所有 skill 均被禁用或 skill 目录为空
- **THEN** 系统 SHALL 正常工作，system prompt 中 skill 列表为空
- **AND** agent 继续使用基础 prompt 处理问题

### Requirement: Ask 函数签名简化
`Ask` 函数 SHALL 移除 `manualSkills []string` 参数。

#### Scenario: 调用 Ask
- **WHEN** TUI 层调用 `agent.Ask`
- **THEN** 签名为 `Ask(ctx context.Context, messages string, history []model.Message) error`
- **AND** 不再需要传入 skills 列表

#### Scenario: BaseUserPrompt 简化
- **WHEN** 构建 user message
- **THEN** BaseUserPrompt SHALL 只包含用户问题占位符 `{{.question}}`
- **AND** 不包含 `{{.skill_prompt}}` 占位符
