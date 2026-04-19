# 规格：skill-reference-tool

## Purpose

系统 SHALL 提供工具让 Agent 按需加载指定 Skill 的 references 目录下的文件，MUST 支持通过 `get_skill_reference` 工具以 skill 名称和文件名精确获取 reference 文件内容。

## Requirements

### Requirement: 加载 Skill Reference 文件

系统 MUST 提供工具让 Agent 按需加载指定 Skill 的 references 目录下的文件。

#### Scenario: 成功加载 reference 文件
- **WHEN** Agent 调用 `get_skill_reference` 工具，指定 skill_name 和 file_name
- **AND** 该 skill 存在且 references 目录下存在指定文件
- **THEN** 系统返回文件内容

#### Scenario: Skill 不存在
- **WHEN** Agent 调用 `get_skill_reference` 工具
- **AND** 指定的 skill_name 不存在
- **THEN** 系统返回错误信息 "skill not found"

#### Scenario: Reference 文件不存在
- **WHEN** Agent 调用 `get_skill_reference` 工具
- **AND** skill 存在但 references 目录下不存在指定文件
- **THEN** 系统返回错误信息 "reference file not found"

#### Scenario: Skill 无 references 目录
- **WHEN** Agent 调用 `get_skill_reference` 工具
- **AND** 该 skill 没有 references 目录
- **THEN** 系统返回错误信息 "skill has no references directory"
