# 规格：query-sessions-by-date

## Purpose

系统 SHALL 提供工具来按日期查询 Session 记录，用于 SKILL 获取早盘、午盘、收盘等历史会话信息。

## Requirements

### Requirement: 按日期查询 Session

系统 SHALL 提供工具来按日期查询 Session 记录，用于 SKILL 获取早盘、午盘、收盘等历史会话信息。

#### Scenario: 查询今日所有 Session
- **WHEN** 调用工具查询日期为今天
- **THEN** 返回当日所有 Session 记录

#### Scenario: 查询指定日期的 Session
- **WHEN** 调用工具查询日期 `2026-03-14`
- **THEN** 返回该日期的所有 Session 记录

#### Scenario: 查询无 Session 的日期
- **WHEN** 查询的日期没有任何 Session
- **THEN** 返回空列表

---

### Requirement: Session 信息返回

返回的 Session 信息 SHALL 包含关键内容。

#### Scenario: 返回 Session 基本信息
- **WHEN** 查询成功
- **THEN** 返回每个 Session 的 ID、创建时间、标签列表

#### Scenario: 返回 Session 摘要
- **WHEN** 查询成功
- **THEN** 返回每个 Session 的消息摘要（首尾消息内容）

---

### Requirement: 日期格式支持

系统 SHALL 支持多种日期格式输入。

#### Scenario: 支持 ISO 日期格式
- **WHEN** 日期参数为 `2026-03-15`
- **THEN** 正确解析并查询

#### Scenario: 支持相对日期
- **WHEN** 日期参数为 `today` 或 `yesterday`
- **THEN** 自动转换为对应日期

---

### Requirement: 标签过滤

系统 SHALL 支持按标签过滤 Session。

#### Scenario: 过滤特定标签的 Session
- **WHEN** 提供标签参数 `morning-session`
- **THEN** 只返回带有该标签的 Session
