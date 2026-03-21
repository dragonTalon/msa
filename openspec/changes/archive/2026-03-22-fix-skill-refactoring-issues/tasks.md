## 1. 测试代码修复

### 1.1 loader_test.go
- [x] 1.1.1 修复第 15、28、45 行：`skillMetadata` → `skillMetadataYAML`

### 1.2 manager_test.go
- [x] 1.2.1 修复第 66、89-91、142-143、284、393 行：移除 `path` 字段（`dirPath` 是私有字段，测试不能直接设置）
- [x] 1.2.2 需要调整测试策略：使用临时目录创建真实 skill 文件，让 Loader 正常加载

### 1.3 skill_test.go
- [x] 1.3.1 修复第 34、98、164、205 行：移除 `path` 字段，改用 `dirPath` 所在目录
- [x] 1.3.2 修复第 211 行：`GetPath()` → `GetDirPath()`
- [x] 1.3.3 调整测试策略：创建临时目录结构，设置 `dirPath` 私有字段

### 1.4 selector_test.go
- [x] 1.4.1 修复第 19-22、229-231 行：移除 `path` 字段
- [x] 1.4.2 修复第 73、318、485、500 行：`SkillMetadata` → `SkillInfo`（selector.go 中已重命名）
- [x] 1.4.3 修复 `formatSkillsForLog` 函数调用：参数类型从 `[]SkillMetadata` 改为 `[]SkillInfo`

### 1.5 registry_test.go
- [x] 1.5.1 修复第 17、24、117-120、170、177 行：移除 `path` 字段

### 1.6 验证测试
- [x] 1.6.1 运行 `go test ./pkg/logic/skills/... -v` 验证所有测试通过

## 2. 新增 Skill Reference 工具

- [x] 2.1 在 `pkg/logic/tools/skill/skill.go` 中添加 `SkillReferenceParam` 结构体
- [x] 2.2 实现 `GetSkillReference` 函数
- [x] 2.3 添加 `SkillReferenceTool` 结构体及其方法（GetToolInfo, GetName, GetDescription, GetToolGroup）
- [x] 2.4 在 `pkg/logic/tools/register.go` 中注册新工具

## 3. 新增 Skill Asset 工具

- [x] 3.1 在 `pkg/logic/tools/skill/skill.go` 中添加 `SkillAssetParam` 结构体
- [x] 3.2 实现 `GetSkillAsset` 函数
- [x] 3.3 添加 `SkillAssetTool` 结构体及其方法
- [x] 3.4 在 `pkg/logic/tools/register.go` 中注册新工具

## 4. 扩展 Skill Content 工具

- [x] 4.1 修改 `SkillContentData` 结构体，添加 `HasReferences` 和 `HasAssets` 字段
- [x] 4.2 在 `GetSkillContent` 函数中调用 `sk.HasReferences()` 和 `sk.HasAssets()`
- [x] 4.3 更新工具描述，提示 Agent 可按需加载 references/assets

## 5. 优化 Skill Selector 元数据

- [x] 5.1 在 `pkg/logic/skills/selector.go` 中扩展 `SkillInfo` 结构体
- [x] 5.2 添加 `Pattern`, `HasReferences`, `HasAssets` 字段
- [x] 5.3 更新 `SelectSkills()` 中的 `metadataList` 构建逻辑

## 6. 优化 System Prompt

- [x] 6.1 修改 `pkg/logic/agent/prompt.go` 中的 `BasePromptTemplate`
- [x] 6.2 在 Skill 列表中添加 pattern、hasReferences、hasAssets 展示
- [x] 6.3 添加提示：告知 Agent 如有 references/assets 可按需加载

## 7. 验证与清理

- [x] 7.1 运行 `go vet ./...` 验证编译
- [x] 7.2 运行 `go test ./...` 验证所有测试通过
- [x] 7.3 将探索记录 `openspec/explore/2026-03-21-skill-refactoring.md` 移动到归档目录
