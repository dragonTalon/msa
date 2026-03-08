# 规格：skill-management

Skills 的加载、注册、缓存管理能力。

## Purpose

提供灵活的 Skill 管理机制，支持内置和自定义 Skills，实现懒加载优化内存使用，并提供健壮的错误处理。

## Requirements

### Requirement: Skill 文件扫描

系统必须能够扫描指定目录下的所有 Skill 文件并解析其元数据。

#### Scenario: 扫描内置 Skills 目录
- **当**系统启动或请求 Skill 列表时
- **那么**系统必须扫描 `pkg/logic/skills/plugs/` 目录
- **并且**找到所有包含 `SKILL.md` 文件的子目录
- **并且**解析每个文件的 YAML frontmatter

#### Scenario: 扫描用户自定义 Skills 目录
- **当**系统启动或请求 Skill 列表时
- **那么**系统必须扫描 `~/.msa/skills/` 目录
- **并且**如果目录不存在则忽略（不报错）
- **并且**解析每个找到的 `SKILL.md` 文件

#### Scenario: 处理无效的 Skill 文件
- **当**扫描到格式错误的 `SKILL.md` 文件时
- **那么**系统必须记录警告日志并跳过该文件
- **并且**不影响其他有效 Skill 的加载

---

### Requirement: Skill 元数据解析

系统必须正确解析 SKILL.md 文件的 YAML frontmatter。

#### Scenario: 解析完整的元数据
- **当**Skill 文件包含所有必需字段时
- **那么**系统必须解析 `name`、`description`、`version`、`priority` 字段
- **并且**使用默认值 5 作为 priority（如果未指定）
- **并且**记录解析成功日志（INFO 级别）

#### Scenario: 解析缺少字段的 Skill
- **当**Skill 文件缺少 `name` 或 `description` 字段时
- **那么**系统必须记录错误日志并标记该 Skill 为无效
- **并且**不将该 Skill 加入注册表

---

### Requirement: Skill 懒加载

Skill 内容（SKILL.md body）必须按需加载，不在启动时全部加载到内存。

#### Scenario: 首次请求 Skill 内容
- **当**首次请求某个 Skill 的内容时
- **那么**系统必须从文件系统读取 `SKILL.md` 文件
- **并且**跳过 YAML frontmatter，只返回 body 内容
- **并且**将内容缓存到内存中
- **并且**记录加载日志（DEBUG 级别）

#### Scenario: 重复请求 Skill 内容
- **当**再次请求已加载的 Skill 内容时
- **那么**系统必须从缓存返回内容
- **并且**不重新读取文件

#### Scenario: 文件读取失败
- **当**读取 Skill 文件失败时
- **那么**系统必须返回错误
- **并且**记录错误日志
- **并且**不影响其他 Skill 的使用

---

### Requirement: Skill 注册表

系统必须维护一个 Skill 注册表，支持按名称查询和遍历。

#### Scenario: 按 name 查询 Skill
- **当**通过名称查询 Skill 时
- **那么**系统必须返回对应的 Skill 元数据
- **或者**如果不存在则返回 nil

#### Scenario: 获取所有可用 Skills
- **当**请求所有 Skills 列表时
- **那么**系统必须返回所有有效 Skill 的列表
- **并且**按 priority 降序排序

#### Scenario: 获取禁用的 Skills
- **当**请求配置中禁用的 Skills 时
- **那么**系统必须读取 `~/.msa/config.json` 中的 `disable_skills` 字段
- **并且**返回禁用列表

---

### Requirement: 错误处理

> 📚 **知识上下文**：基于模式 #003（错误处理和重试策略），此能力需实现健壮的错误处理机制。

系统必须实现健壮的错误处理和降级机制。

#### Scenario: 目录扫描失败
- **当**扫描目录因权限问题失败时
- **那么**系统必须记录错误日志
- **并且**继续扫描其他目录
- **并且**不影响已加载的 Skills

#### Scenario: YAML 解析失败
- **当**YAML frontmatter 解析失败时
- **那么**系统必须记录警告日志
- **并且**跳过该 Skill 文件
- **并且**不中断整个加载流程

---

### Requirement: Skill 来源标识

系统必须区分内置 Skills 和用户自定义 Skills。

#### Scenario: 标记内置 Skills
- **当**Skill 来自 `pkg/logic/skills/plugs/` 目录时
- **那么**系统必须标记其来源为 `builtin`

#### Scenario: 标记用户 Skills
- **当**Skill 来自 `~/.msa/skills/` 目录时
- **那么**系统必须标记其来源为 `user`
