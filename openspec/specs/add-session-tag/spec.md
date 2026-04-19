# 规格：add-session-tag

## Purpose

系统 SHALL 提供工具来为 Session 添加标签，用于区分早盘、午盘、收盘等不同类型的会话。

## Requirements

### Requirement: 为 Session 添加标签

系统 SHALL 提供工具来为 Session 添加标签，用于区分早盘、午盘、收盘等不同类型的会话。

#### Scenario: 添加新标签
- **WHEN** 调用工具为当前 Session 添加标签 `morning-session`
- **THEN** 标签成功添加到 Session

#### Scenario: 添加多个标签
- **WHEN** 调用工具添加多个标签 `["morning-session", "trading"]`
- **THEN** 所有标签成功添加

#### Scenario: 添加已存在的标签
- **WHEN** 添加的标签已存在于 Session
- **THEN** 返回成功（幂等操作，不报错）

---

### Requirement: Session 识别

系统 SHALL 支持多种方式识别目标 Session。

#### Scenario: 为当前 Session 添加标签
- **WHEN** 不指定 Session ID
- **THEN** 为当前正在进行的 Session 添加标签

#### Scenario: 为指定 Session 添加标签
- **WHEN** 指定 Session ID
- **THEN** 为该 Session 添加标签

---

### Requirement: 标签格式校验

标签 MUST 符合格式要求。

#### Scenario: 拒绝无效标签
- **WHEN** 标签包含特殊字符或空格
- **THEN** 返回错误 "invalid tag format"

#### Scenario: 标准化标签
- **WHEN** 标签包含大写字母
- **THEN** 自动转换为小写

---

### Requirement: 预设标签支持

系统 SHALL 预设特定用途的标签。

#### Scenario: 预设标签列表
- **WHEN** 添加以下预设标签时应有特殊含义：
  - `morning-session`：早盘会话
  - `afternoon-session`：午盘会话
  - `close-session`：收盘会话
- **THEN** 系统应识别并正确处理
