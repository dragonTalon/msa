# 规格：skill-selector

LLM 自动选择 Skills 的能力，根据用户输入和 Skill 描述动态选择最合适的 Skills。

## Purpose

提供智能的 Skill 选择机制，通过 LLM 分析用户输入，自动选择和组合相关的 Skills，同时保证系统稳定性和可配置性。

## Requirements

### Requirement: LLM 自动选择 Skill

系统必须使用 LLM 根据用户输入自动选择合适的 Skills。

#### Scenario: 构建选择 Prompt
- **当**请求选择 Skills 时
- **那么**系统必须使用模板变量构建选择 Prompt
- **并且**包含所有可用 Skills 的 metadata（name, description, priority）
- **并且**包含用户输入内容
- **并且**使用 Go 模板语法填充变量

#### Scenario: 调用 LLM 进行选择
- **当**选择 Prompt 构建完成后
- **那么**系统必须调用 LLM 获取选择结果
- **并且**使用与主对话相同的模型
- **并且**记录选择请求日志（INFO 级别，包含用户输入摘要）
- **并且**记录可用 Skills 列表（DEBUG 级别）

#### Scenario: 解析 LLM 返回结果
- **当**LLM 返回 JSON 结果时
- **那么**系统必须解析 `selected_skills` 和 `reasoning` 字段
- **并且**如果解析失败则降级到 base skill
- **并且**记录选择结果日志（INFO 级别）

---

### Requirement: Base Skill 保证

`base` Skill 必须始终被选中，无论 LLM 返回什么结果。

#### Scenario: LLM 返回包含 base
- **当**LLM 返回的 `selected_skills` 已包含 `base` 时
- **那么**系统必须保持原样

#### Scenario: LLM 返回不包含 base
- **当**LLM 返回的 `selected_skills` 不包含 `base` 时
- **那么**系统必须自动添加 `base` 到结果中
- **并且**记录日志说明

---

### Requirement: Priority 排序

选中的 Skills 必须按 priority 降序排序。

#### Scenario: 排序多个 Skills
- **当**选择了多个 Skills 时
- **那么**系统必须按 priority 从高到低排序
- **并且**高 priority 的 Skill 排在前面

#### Scenario: 相同 Priority 的处理
- **当**多个 Skills 具有相同 priority 时
- **那么**系统必须保持其原始相对顺序

---

### Requirement: 禁用 Skills 过滤

系统必须在选择前过滤掉配置中禁用的 Skills。

#### Scenario: 过滤禁用的 Skills
- **当**构建可用 Skills 列表时
- **那么**系统必须从 `~/.msa/config.json` 读取 `disable_skills` 配置
- **并且**从列表中移除禁用的 Skills
- **并且**不将这些 Skills 发送给 LLM

#### Scenario: 所有 Skills 被禁用
- **当**所有非 base Skills 都被禁用时
- **那么**系统必须只返回 `base` skill
- **并且**记录警告日志

---

### Requirement: LLM 调用失败降级

> 📚 **知识上下文**：基于模式 #003（错误处理和重试策略），此能力需实现 LLM 调用失败时的降级策略。

当 LLM 调用失败时，系统必须降级到 base skill。

#### Scenario: LLM 网络错误
- **当**LLM 调用因网络问题失败时
- **那么**系统必须记录警告日志
- **并且**返回只包含 `base` skill 的结果
- **并且**在 reasoning 中说明降级原因

#### Scenario: LLM 超时
- **当**LLM 调用超时时
- **那么**系统必须记录警告日志
- **并且**返回只包含 `base` skill 的结果
- **并且**在 reasoning 中说明超时降级

#### Scenario: LLM 返回无效 JSON
- **当**LLM 返回的 JSON 无法解析时
- **那么**系统必须记录警告日志（包含原始响应）
- **并且**返回只包含 `base` skill 的结果
- **并且**在 reasoning 中说明解析失败

---

### Requirement: 选择结果日志

系统必须记录详细的选择过程日志。

#### Scenario: 记录选择输入
- **当**开始选择时
- **那么**系统必须记录 INFO 级别日志，包含用户输入的前 50 字符
- **并且**记录 DEBUG 级别日志，包含所有可用 Skills 的名称和 priority

#### Scenario: 记录选择输出
- **当**选择完成时
- **那么**系统必须记录 INFO 级别日志，包含选中的 Skills 列表
- **并且**记录 DEBUG 级别日志，包含 LLM 的 reasoning

---

### Requirement: Skill 元数据传递

系统只传递 Skill metadata 给 LLM，不加载完整内容。

#### Scenario: 构建轻量级 Skill 列表
- **当**构建发送给 LLM 的 Skill 列表时
- **那么**系统只包含 `name`、`description`、`priority` 字段
- **并且**不加载 Skill 的 body 内容
- **并且**不加载 Skill 的文件内容
