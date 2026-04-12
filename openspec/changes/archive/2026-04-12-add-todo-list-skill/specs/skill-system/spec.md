# 规格：skill-system

> 📚 **知识上下文**：参考模式 #002 搜索引擎接口设计模式，应用于工具接口设计

## MODIFIED Requirements

### 需求：Skill 元数据支持 requires_todo

Skill 的 YAML frontmatter 必须支持 `requires_todo` 字段。

#### 场景：标记 Skill 需要 TODO
- **当** Skill 需要在执行前创建 TODO 列表
- **那么** YAML frontmatter 中设置 `requires_todo: true`

#### 场景：默认不需要 TODO
- **当** Skill 未设置 `requires_todo` 字段
- **那么** 默认值为 `false`

---

### 需求：Skill 支持 TODO 模板文件

Skill 的 references 目录可以包含 todo-template.md 文件。

#### 场景：使用自定义 TODO 模板
- **当** Skill 的 `references/todo-template.md` 存在
- **那么** create_todo 工具使用该模板创建 TODO 文件

#### 场景：无自定义模板时使用默认
- **当** Skill 的 `references/todo-template.md` 不存在
- **那么** create_todo 工具使用系统默认模板

---

### 需求：Skill 结构扩展

Skill 结构体必须支持 TODO 相关字段。

#### 场景：加载 requires_todo 元数据
- **当** loader 加载 Skill
- **那么** 解析 YAML frontmatter 中的 `requires_todo` 字段
- **并且** 设置到 Skill 结构体的 Metadata 中

#### 场景：检查 TODO 模板是否存在
- **当** 调用 Skill.HasTodoTemplate()
- **那么** 检查 references/todo-template.md 是否存在
- **并且** 返回布尔值

---

### 需求：Manager 支持 TODO 相关方法

Skill Manager 必须提供 TODO 相关的查询方法。

#### 场景：获取 TODO 模板内容
- **当** 调用 manager.GetTodoTemplate(skillName)
- **那么** 返回该 Skill 的 TODO 模板内容
- **如果** Skill 有自定义模板，返回自定义模板
- **否则** 返回默认模板

#### 场景：检查 Skill 是否需要 TODO
- **当** 调用 manager.RequiresTodo(skillName)
- **那么** 返回该 Skill 的 requires_todo 字段值
