# 规格：dynamic-prompts

根据选中的 Skills 动态构建 System Prompt 的能力。

## Purpose

提供灵活的 Prompt 构建机制，支持多 Skill 组合、用户手动覆盖，并保持与现有 Prompt 框架的兼容性。

## Requirements

### Requirement: 动态 Prompt 构建

系统必须能够根据选中的 Skills 动态构建完整的 System Prompt。

#### Scenario: 构建 System Prompt
- **当**收到选中的 Skills 列表时
- **那么**系统必须按顺序合并各部分内容
- **并且**顺序为：基础角色定义 + 系统配置 + Skills 内容 + 基础规则
- **并且**每部分之间用双换行符分隔

#### Scenario: 合并多个 Skills 内容
- **当**选择了多个 Skills 时
- **那么**系统必须按 priority 排序后的顺序合并内容
- **并且**每个 Skill 内容前添加标题（如 `## Skill: base`）
- **并且**保持各 Skill 的原始内容格式

---

### Requirement: 基础 Prompt 保留

系统必须保留现有的基础 Prompt 框架。

#### Scenario: 保留角色定义
- **当**构建 System Prompt 时
- **那么**系统必须包含基础角色定义（role、style 变量）
- **并且**这些变量在运行时填充

#### Scenario: 保留系统配置
- **当**构建 System Prompt 时
- **那么**系统必须包含系统配置（时间、日期等变量）
- **并且**这些变量在运行时填充

#### Scenario: 保留基础工具规则
- **当**构建 System Prompt 时
- **那么**系统必须在末尾添加基础工具调用规则
- **并且**这些规则作为兜底，确保基本功能正常

---

### Requirement: 用户手动覆盖

系统必须支持用户手动指定要使用的 Skills。

#### Scenario: CLI 参数指定 Skills
- **当**用户通过 `--skills` 参数指定 Skills 时
- **那么**系统必须跳过 LLM 自动选择
- **并且**使用用户指定的 Skills 列表
- **并且**确保 base skill 总是包含在列表中

#### Scenario: 对话命令指定 Skills
- **当**用户在对话中使用 `/skills: skill1,skill2` 命令时
- **那么**系统必须解析命令中的 Skills 列表
- **并且**跳过 LLM 自动选择
- **并且**使用指定的 Skills 构建下一轮对话的 Prompt

---

### Requirement: Skill 内容提取

系统必须正确提取 SKILL.md 文件的 body 内容（跳过 frontmatter）。

#### Scenario: 提取标准格式的 Skill 内容
- **当**Skill 文件以 `---` 开头的 YAML frontmatter 时
- **那么**系统必须找到第二个 `---` 标记
- **并且**返回第二个 `---` 之后的所有内容
- **并且**去除开头的换行符

#### Scenario: 处理无 frontmatter 的 Skill
- **当**Skill 文件不以 `---` 开头时
- **那么**系统必须返回整个文件内容
- **并且**不进行 frontmatter 解析

#### Scenario: 处理只有 frontmatter 的 Skill
- **当**Skill 文件只有 frontmatter 没有 body 时
- **那么**系统必须返回空字符串
- **并且**记录警告日志

---

### Requirement: Prompt 变量填充

系统必须在运行时填充 Prompt 中的变量。

#### Scenario: 填充时间相关变量
- **当**构建 Prompt 时
- **那么**系统必须填充 `{{.time}}` 为当前时间
- **并且**填充 `{{.weekday}}` 为当前日期和星期

#### Scenario: 填充角色变量
- **当**构建 Prompt 时
- **那么**系统必须填充 `{{.role}}` 为"专业股票分析助手"
- **并且**填充 `{{.style}}` 为"理性、专业、客观且严谨"

#### Scenario: 填充用户问题变量
- **当**构建 Prompt 时
- **那么**系统必须填充 `{{.question}}` 为用户的实际输入

---

### Requirement: 错误处理

系统必须处理 Prompt 构建过程中的各种错误。

#### Scenario: Skill 文件不存在
- **当**指定的 Skill 文件不存在时
- **那么**系统必须记录错误日志
- **并且**跳过该 Skill
- **并且**继续处理其他 Skills

#### Scenario: Skill 内容为空
- **当**Skill 文件 body 内容为空时
- **那么**系统必须记录警告日志
- **并且**在 Prompt 中为该 Skill 保留标题但内容为空

#### Scenario: 变量填充失败
- **当**Prompt 变量填充失败时
- **那么**系统必须记录错误日志
- **并且**使用默认值或保留未填充的变量
