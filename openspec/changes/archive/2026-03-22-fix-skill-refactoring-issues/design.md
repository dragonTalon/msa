## Context

在之前的 Skill 系统重构中，我们基于 Google ADK 五种设计模式重构了 Skill 系统，新增了：
- `SkillPattern` 枚举（tool-wrapper, generator, reviewer, inversion, pipeline）
- `SkillTrigger` 触发条件结构
- `SkillMetadata` 扩展元数据（pattern, triggers, tools, dependencies）
- `references/` 和 `assets/` 目录支持
- `GetReference()`, `GetAsset()`, `GetDirPath()`, `HasReferences()`, `HasAssets()` 方法
- `path` 字段重命名为 `dirPath` 并改为私有字段

但在重构过程中遗漏了以下问题：

### 1. 测试代码未更新

#### loader_test.go
- 使用旧的 `skillMetadata` 类型名（应为 `skillMetadataYAML`）

#### manager_test.go
- 使用旧的 `path` 字段名
- `path` 已重命名为 `dirPath` 且变为私有字段

#### skill_test.go
- 使用 `path` 字段（私有字段）
- 使用 `GetPath()` 方法（应为 `GetDirPath()`）

#### selector_test.go
- 使用 `path` 字段（私有字段）
- 使用 `SkillMetadata` 类型名（应为 `SkillInfo`）
- 注意：selector.go 中将原来的 `SkillMetadata` 重命名为 `SkillInfo`，避免与 skill.go 中新增的 `SkillMetadata` 结构冲突

#### registry_test.go
- 使用 `path` 字段（私有字段）

### 2. 缺少 references/assets 加载工具
- `GetReference()` 和 `GetAsset()` 方法已在 `skill.go` 中实现
- Manager 中也有 `GetSkillReference()` 和 `GetSkillAsset()` 方法
- 但没有对应的工具供 Agent 调用

### 3. SkillContentTool 未扩展
- 返回内容未告知 Agent 该 skill 是否有 references/assets 可按需加载

### 4. SkillInfo 结构体未更新
- 当前只有 Name, Description, Priority
- 缺少 pattern, hasReferences, hasAssets 等新元数据

### 5. System Prompt 未利用新元数据
- 当前只展示 Name, Priority, Description
- 未展示 pattern, tools, hasReferences, hasAssets

## Goals / Non-Goals

**Goals:**
1. 修复所有测试代码，确保测试通过
2. 新增 `get_skill_reference` 和 `get_skill_asset` 工具
3. 扩展 `SkillContentTool` 返回更多元数据
4. 扩展 `SkillInfo` 结构体增加新元数据
5. 优化 System Prompt 展示更多元数据信息

**Non-Goals:**
1. 不修改 Selector 的选择逻辑（暂不使用 triggers 自动触发）
2. 不修改 SKILL.md 文件内容
3. 不添加新的 Skill 类型

## Decisions

### 决策 1: 测试修复策略

**选择**: 调整测试策略，使用临时目录创建真实 skill 文件

**理由**:
- `dirPath` 是私有字段，测试不能直接设置
- 使用反射设置私有字段会增加测试复杂度
- 创建真实文件结构让 Loader 正常加载更接近实际使用场景

### 决策 2: 新工具命名

**选择**: `get_skill_reference` 和 `get_skill_asset`

**理由**:
- 与现有 `get_skill_content` 保持一致的命名风格
- 使用 snake_case 符合工具命名规范

### 决策 3: 工具参数设计

**选择**: 使用 `skill_name` + `file_name` 双参数

```go
type SkillReferenceParam struct {
    SkillName string `json:"skill_name"`
    FileName  string `json:"file_name"`
}
```

**理由**:
- 与现有 `SkillContentParam` 结构一致
- 明确指定要加载的文件名

### 决策 4: SkillContentTool 扩展

**选择**: 在返回数据中增加元数据字段

```go
type SkillContentData struct {
    SkillName     string `json:"skill_name"`
    Content       string `json:"content"`
    Length        int    `json:"length"`
    HasReferences bool   `json:"has_references"`
    HasAssets     bool   `json:"has_assets"`
}
```

**理由**:
- 向后兼容现有返回结构
- Agent 可以知道是否需要加载 references/assets

### 决策 5: SkillInfo 扩展

**选择**: 在 selector.go 中扩展 `SkillInfo` 结构

```go
type SkillInfo struct {
    Name          string `json:"name"`
    Description   string `json:"description"`
    Priority      int    `json:"priority"`
    Pattern       string `json:"pattern"`
    HasReferences bool   `json:"has_references"`
    HasAssets     bool   `json:"has_assets"`
}
```

**理由**:
- 向后兼容现有结构
- Selector 选择时可参考更多信息

### 决策 6: System Prompt 增强

**选择**: 在现有模板中增加元数据展示

```
- **{{.Name}}** (模式: {{.Pattern}}, 优先级: {{.Priority}}): {{.Description}}
  {{if .HasReferences}}📖 参考资料可用 | {{end}}{{if .HasAssets}}📄 模板可用{{end}}
```

**理由**:
- 向后兼容现有模板结构
- Agent 可以看到更多信息来决定是否需要加载 references/assets

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|----------|
| 新工具增加 token 消耗 | 工具描述说明按需加载，不是每次都加载 |
| 测试修复可能引入新问题 | 先运行现有测试确认失败原因，再逐个修复 |
| Prompt 变更可能影响现有行为 | 增量添加，不删除现有信息 |
| 测试策略调整增加测试代码量 | 使用辅助函数简化重复的临时目录创建 |
