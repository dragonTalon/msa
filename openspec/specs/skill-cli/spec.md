# 规格：skill-cli

Skills 管理的 CLI 命令界面。

## Purpose

提供完整的 CLI 命令来管理 Skills，包括列表、查看、启用、禁用等功能。

## Requirements

### Requirement: Skills 列表命令

系统必须提供 `msa skills list` 命令来列出所有可用的 Skills。

#### Scenario: 列出所有 Skills
- **当**用户执行 `msa skills list` 时
- **那么**系统必须显示所有可用 Skills
- **并且**按 priority 降序排序显示
- **并且**每个 Skill 显示名称、描述、优先级、来源（builtin/user）
- **并且**标记已禁用的 Skills

#### Scenario: 列出内置 Skills
- **当**用户执行 `msa skills list --builtin` 时
- **那么**系统只显示内置 Skills

#### Scenario: 列出用户 Skills
- **当**用户执行 `msa skills list --user` 时
- **那么**系统只显示用户自定义 Skills

---

### Requirement: Skill 详情命令

系统必须提供 `msa skills show <name>` 命令来查看 Skill 的详细信息。

#### Scenario: 显示 Skill 详情
- **当**用户执行 `msa skills show <name>` 时
- **那么**系统必须显示 Skill 的完整元数据
- **并且**显示 Skill 的 body 内容（如果可用）
- **并且**显示 Skill 的来源目录
- **并且**显示当前启用/禁用状态

#### Scenario: Skill 不存在
- **当**用户查看不存在的 Skill 时
- **那么**系统必须显示错误信息
- **并且**提示可用的 Skills 名称

---

### Requirement: Skill 禁用命令

系统必须提供 `msa skills disable <name>` 命令来禁用指定的 Skill。

#### Scenario: 禁用单个 Skill
- **当**用户执行 `msa skills disable <name>` 时
- **那么**系统必须将该 Skill 名称添加到 `~/.msa/config.json` 的 `disable_skills` 数组
- **并且**如果该 Skill 已在列表中则不重复添加
- **并且**显示确认信息

#### Scenario: 禁用 base Skill
- **当**用户尝试禁用 `base` Skill 时
- **那么**系统必须显示错误信息
- **并且**说明 base Skill 不能被禁用
- **并且**不修改配置文件

#### Scenario: 禁用不存在的 Skill
- **当**用户禁用不存在的 Skill 时
- **那么**系统必须显示警告信息
- **并且**仍然将名称添加到禁用列表（预留）

---

### Requirement: Skill 启用命令

系统必须提供 `msa skills enable <name>` 命令来启用已禁用的 Skill。

#### Scenario: 启用已禁用的 Skill
- **当**用户执行 `msa skills enable <name>` 时
- **那么**系统必须从 `~/.msa/config.json` 的 `disable_skills` 数组中移除该名称
- **并且**显示确认信息

#### Scenario: 启用已启用的 Skill
- **当**用户启用未在禁用列表中的 Skill 时
- **那么**系统必须显示信息说明该 Skill 已启用
- **并且**不修改配置文件

---

### Requirement: 配置文件管理

系统必须正确管理 `~/.msa/config.json` 文件中的 Skill 配置。

#### Scenario: 配置文件不存在时创建
- **当**配置文件不存在时执行 disable/enable 操作
- **那么**系统必须创建配置文件
- **并且**添加 `disable_skills` 字段（初始为空数组）

#### Scenario: 配置文件存在但无 disable_skills 字段
- **当**配置文件存在但没有 `disable_skills` 字段时
- **那么**系统必须添加该字段到配置文件
- **并且**保留其他现有配置

#### Scenario: 禁用列表为空时删除字段
- **当**最后一个 Skill 被启用后 `disable_skills` 数组为空时
- **那么**系统可以选择保留空数组或删除该字段
- **并且**保持配置文件一致性

---

### Requirement: 输出格式

CLI 命令必须提供清晰易读的输出格式。

#### Scenario: 列表命令表格输出
- **当**执行 `msa skills list` 时
- **那么**系统必须使用表格格式输出
- **并且**包含列：Name、Priority、Source、Status
- **并且**使用合适的列宽和对齐方式

#### Scenario: 详情命令格式化输出
- **当**执行 `msa skills show` 时
- **那么**系统必须使用清晰的分节格式
- **并且**使用标题分隔不同部分（元数据、内容、状态等）

#### Scenario: JSON 输出选项
- **当**用户添加 `--json` 参数时
- **那么**系统必须输出 JSON 格式
- **并且**包含所有相关信息
