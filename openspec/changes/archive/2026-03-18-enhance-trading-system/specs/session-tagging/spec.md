## ADDED Requirements

### Requirement: 自动时间标签

Session 初始化时 MUST 根据当前时间自动添加标签，不再依赖手动调用。

#### Scenario: 早盘时段自动标签
- **WHEN** Session 在 9:30-11:30 之间初始化
- **THEN** 自动添加 `morning-session` 标签

#### Scenario: 午盘时段自动标签
- **WHEN** Session 在 13:00-14:30 之间初始化
- **THEN** 自动添加 `afternoon-session` 标签

#### Scenario: 收盘时段自动标签
- **WHEN** Session 在 16:00 之后初始化
- **THEN** 自动添加 `close-session` 标签

#### Scenario: 非交易时段无标签
- **WHEN** Session 在其他时间初始化（如 11:30-13:00，14:30-16:00）
- **THEN** 不自动添加交易时段标签

---

### Requirement: 标签幂等性

同一标签 MUST NOT 重复添加。

#### Scenario: 重复添加同一标签
- **WHEN** 尝试添加已存在的标签
- **THEN** 忽略该操作，不产生重复标签

---

### Requirement: 手动标签保留

自动标签 SHALL NOT 影响手动添加其他标签的能力。

#### Scenario: 手动添加自定义标签
- **WHEN** 调用 add_session_tag 添加自定义标签
- **THEN** 标签成功添加到 Session

---

### Requirement: SKILL 文档更新

相关 SKILL MUST 移除手动添加标签的说明。

#### Scenario: morning-analysis SKILL 更新
- **WHEN** 查阅 morning-analysis/SKILL.md
- **THEN** 不再包含"为当前Session添加标签：morning-session"的手动说明

#### Scenario: afternoon-trade SKILL 更新
- **WHEN** 查阅 afternoon-trade/SKILL.md
- **THEN** 不再包含"为当前Session添加标签：afternoon-session"的手动说明

#### Scenario: market-close-summary SKILL 更新
- **WHEN** 查阅 market-close-summary/SKILL.md
- **THEN** 不再包含"为当前Session添加标签：close-session"的手动说明
