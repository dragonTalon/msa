# 规格：skill-asset-tool

## Purpose

系统 SHALL 提供工具让 Agent 按需加载指定 Skill 的 assets 目录下的文件。

## Requirements

### Requirement: 加载 Skill Asset 文件

系统 SHALL 提供工具让 Agent 按需加载指定 Skill 的 assets 目录下的文件。

#### Scenario: 成功加载 asset 文件
- **WHEN** Agent 调用 `get_skill_asset` 工具，指定 skill_name 和 file_name
- **AND** 该 skill 存在且 assets 目录下存在指定文件
- **THEN** 系统返回文件内容

#### Scenario: Skill 不存在
- **WHEN** Agent 调用 `get_skill_asset` 工具
- **AND** 指定的 skill_name 不存在
- **THEN** 系统返回错误信息 "skill not found"

#### Scenario: Asset 文件不存在
- **WHEN** Agent 调用 `get_skill_asset` 工具
- **AND** skill 存在但 assets 目录下不存在指定文件
- **THEN** 系统返回错误信息 "asset file not found"

#### Scenario: Skill 无 assets 目录
- **WHEN** Agent 调用 `get_skill_asset` 工具
- **AND** 该 skill 没有 assets 目录
- **THEN** 系统返回错误信息 "skill has no assets directory"
