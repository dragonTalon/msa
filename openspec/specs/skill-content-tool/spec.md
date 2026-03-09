# 规格：skill-content-tool

agent 按需查询 skill 完整内容的工具能力。

## Purpose

提供 `get_skill_content` tool，供 react.Agent 在推理阶段根据 skill 名称懒加载完整的 skill 处理流程内容，避免在 system prompt 中预先载入所有 skill 的完整内容。

## Requirements

### Requirement: get_skill_content tool

系统 SHALL 提供 `get_skill_content` tool，接收 skill name，返回对应 skill 的完整 Markdown body 内容。

#### Scenario: 查询已存在的 skill
- **WHEN** agent 调用 `get_skill_content` 并传入有效的 skill name
- **THEN** 系统 SHALL 返回该 skill 的完整 body 内容（已去掉 YAML frontmatter）
- **AND** 内容通过 `skill.GetContent()` 懒加载获取

#### Scenario: 查询不存在的 skill
- **WHEN** agent 调用 `get_skill_content` 并传入不存在的 skill name
- **THEN** 系统 SHALL 返回明确的错误信息，说明 skill 未找到
- **AND** 不返回空字符串

#### Scenario: skill 内容加载失败
- **WHEN** skill 文件读取失败
- **THEN** 系统 SHALL 返回错误信息
- **AND** 记录 ERROR 级别日志

#### Scenario: tool 注册
- **WHEN** 应用启动时
- **THEN** `get_skill_content` tool SHALL 被注册到全局 tool 列表
- **AND** tool name 使用 snake_case：`get_skill_content`
- **AND** tool group 为 skill 相关分组
