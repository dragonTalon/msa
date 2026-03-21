## Why

在 Skill 系统重构（基于 Google ADK 五种设计模式）完成后，进行自检发现存在多处遗漏：

### 测试代码问题

#### loader_test.go
- 第 15、28、45 行使用旧的 `skillMetadata` 类型名（应为 `skillMetadataYAML`）

#### manager_test.go
- 多处使用旧的 `path` 字段名（第 66、89-91、142-143、284、393 行）
- `path` 已重命名为 `dirPath` 且变为私有字段，测试不能直接设置

#### skill_test.go
- 第 34、98、164、205 行使用 `path` 字段（私有字段问题）
- 第 211 行使用 `GetPath()` 方法（应为 `GetDirPath()`）

#### selector_test.go
- 第 19-22、229-231 行使用 `path` 字段（私有字段问题）
- 第 73、318、485、500 行使用 `SkillMetadata` 类型名
- 但 `SkillMetadata` 在 selector.go 中已重命名为 `SkillInfo`（避免与 skill.go 中的新结构冲突）

#### registry_test.go
- 第 17、24、117-120、170、177 行使用 `path` 字段（私有字段问题）

### 工具缺失
1. `GetReference()` 和 `GetAsset()` 方法已在 `skill.go` 中实现
2. 但没有对应的 `get_skill_reference` 和 `get_skill_asset` 工具供 Agent 调用
3. `SkillContentTool` 未告知 Agent 该 skill 是否有 references/assets 可按需加载

### Selector 问题
1. `SkillInfo` 结构体只包含 Name, Description, Priority
2. 缺少 pattern, hasReferences, hasAssets 等新元数据

### System Prompt 问题
1. 当前只展示 Name, Priority, Description
2. 未展示 pattern, tools, hasReferences, hasAssets 等新元数据

## What Changes

- **修复测试代码**：更新 4 个测试文件中的类型名、字段名、方法名
- **新增工具**：添加 `get_skill_reference` 和 `get_skill_asset` 工具
- **扩展现有工具**：`SkillContentTool` 返回更多元数据
- **优化 Selector**：`SkillInfo` 增加新元数据字段
- **优化 Prompt**：在 System Prompt 中展示 pattern、tools、hasReferences、hasAssets 等新元数据
- **归档探索记录**：将探索记录移动到正确的归档位置

## Capabilities

### New Capabilities
- `skill-reference-tool`: 加载指定 skill 的 reference 文件的工具
- `skill-asset-tool`: 加载指定 skill 的 asset 文件的工具

### Modified Capabilities
- `skill-content-tool`: 扩展工具返回内容，告知 Agent 该 skill 是否有 references/assets
- `skill-selector`: `SkillInfo` 增加新元数据字段
- `agent-chat`: System Prompt 展示更多 Skill 元数据信息

## Impact

- `pkg/logic/skills/loader_test.go` - 类型名修复
- `pkg/logic/skills/manager_test.go` - 字段名修复 + 测试策略调整
- `pkg/logic/skills/skill_test.go` - 字段名 + 方法名修复
- `pkg/logic/skills/selector_test.go` - 字段名 + 类型名修复
- `pkg/logic/skills/registry_test.go` - 字段名修复
- `pkg/logic/tools/skill/skill.go` - 新增工具 + 扩展 SkillContentTool
- `pkg/logic/tools/register.go` - 注册新工具
- `pkg/logic/skills/selector.go` - 扩展 SkillInfo 结构
- `pkg/logic/agent/prompt.go` - 优化 Skill 元数据展示
