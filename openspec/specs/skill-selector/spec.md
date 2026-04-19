# 规格：skill-selector

Skill 选择相关的规格（排序、过滤、日志）。

## Purpose

定义 skill 列表的排序和过滤规则，以及相关日志要求。skill 选择职责已转移给 react.Agent 通过 system prompt 自主决策。

## Requirements

### Requirement: Priority 排序

选中的 Skills MUST 按 priority 降序排序。

#### Scenario: 排序多个 Skills
- **当**选择了多个 Skills 时
- **那么**系统必须按 priority 从高到低排序
- **并且**高 priority 的 Skill 排在前面

#### Scenario: 相同 Priority 的处理
- **当**多个 Skills 具有相同 priority 时
- **那么**系统必须保持其原始相对顺序

---

### Requirement: 禁用 Skills 过滤

系统 MUST 在构建 skill 列表时过滤掉配置中禁用的 Skills。

#### Scenario: 过滤禁用的 Skills
- **当**构建可用 Skills 列表时
- **那么**系统必须从 `~/.msa/config.json` 读取 `disable_skills` 配置
- **并且**从列表中移除禁用的 Skills
- **并且**不将这些 Skills 注入 system prompt

#### Scenario: 所有 Skills 被禁用
- **当**所有 Skills 都被禁用时
- **那么**系统必须返回空列表
- **并且**记录警告日志

---

### Requirement: 选择结果日志

系统 MUST 记录详细的 skill 构建过程日志。

#### Scenario: 记录选择输入
- **当**构建 skill 元数据列表时
- **那么**系统必须记录 INFO 级别日志，包含可用 skill 数量
- **并且**记录 DEBUG 级别日志，包含所有可用 Skills 的名称和 priority
