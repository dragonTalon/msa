## REMOVED Requirements

### Requirement: LLM 自动选择 Skill
**Reason**: skill 选择职责转移给 react.Agent 本身，通过 system prompt 中的 skill 元数据列表自主决策，不再需要前置独立 LLM 调用。
**Migration**: 删除 `skills.Selector`、`selectSkillsViaSubAgent`、`buildSkillPromptText`、`validateAndEnsureBase` 相关代码。

### Requirement: Base Skill 保证
**Reason**: 不再有独立 selector，base skill 概念随选择器一起移除。agent 通过 system prompt 的指令自行判断流程。
**Migration**: 无需迁移，新架构中 agent 自主决策。

### Requirement: LLM 调用失败降级
**Reason**: 前置 LLM 调用整体移除，降级逻辑随之消失。
**Migration**: 无。

### Requirement: Skill 元数据传递
**Reason**: 元数据改为静态注入 system prompt，不再通过 LLM 调用传递。
**Migration**: 参见 agent-chat spec 的新需求。
