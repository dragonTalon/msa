# 规格：agent-chat

> 📚 **知识上下文**：此能力扩展现有对话功能，需保持向后兼容性。

## MODIFIED Requirements

### Requirement: 集成 Skill 选择流程

现有对话能力必须集成 Skill 选择流程，在处理用户输入前先选择合适的 Skills。

#### Scenario: 正常对话流程集成
- **当**收到用户输入时
- **那么**系统必须先调用 Skill Selector 选择合适的 Skills
- **并且**加载选中的 Skills 内容
- **并且**构建包含 Skills 的 System Prompt
- **并且**继续原有的对话处理流程

#### Scenario: 用户手动指定 Skills
- **当**用户通过 CLI 参数或对话命令指定 Skills 时
- **那么**系统必须跳过 Skill Selector
- **并且**使用用户指定的 Skills
- **并且**记录日志说明使用了手动指定的 Skills

---

### Requirement: GetDefaultTemplate 函数签名变更

`GetDefaultTemplate` 函数必须增加 `selectedSkills` 参数以支持动态 Prompt 构建。

#### Scenario: 调用新签名
- **当**调用 `GetDefaultTemplate` 时
- **那么**必须传入 `selectedSkills` 参数
- **并且**参数类型为 `[]string`（Skill 名称列表）
- **并且**函数内部根据 Skills 构建 Prompt

#### Scenario: 向后兼容空 Skills
- **当**传入空的 `selectedSkills` 时
- **那么**函数必须使用基础 Prompt（不含额外 Skills）
- **并且**保持与现有行为一致

---

### Requirement: 保持向后兼容

现有对话功能必须保持向后兼容，不影响现有用户。

#### Scenario: 现有代码无需修改
- **当**其他代码调用 `GetDefaultTemplate` 时
- **那么**新签名必须兼容（通过默认参数或函数重载）
- **或者**提供新的函数名（如 `GetDefaultTemplateWithSkills`）

#### Scenario: 配置文件无 disable_skills 时
- **当**用户的配置文件中没有 `disable_skills` 字段时
- **那么**系统必须正常工作
- **并且**假设没有 Skills 被禁用

---

### Requirement: 错误处理

Skill 选择和加载失败时必须优雅降级，不影响对话功能。

#### Scenario: Skill 选择失败
- **当**Skill Selector 返回错误时
- **那么**系统必须记录错误日志
- **并且**使用 base skill 继续对话
- **并且**不中断用户的使用

#### Scenario: Skill 内容加载失败
- **当**Skill 内容加载失败时
- **那么**系统必须记录错误日志
- **并且**跳过该 Skill
- **并且**使用其他可用 Skills 继续对话

#### Scenario: 所有 Skills 加载失败
- **当**所有 Skills（包括 base）加载失败时
- **那么**系统必须回退到原有的硬编码 Prompt
- **并且**记录错误日志
- **并且**保证基本对话功能可用

---

### Requirement: 日志记录

对话流程中的 Skill 相关操作必须记录适当的日志。

#### Scenario: 记录 Skill 选择结果
- **当**Skill 选择完成时
- **那么**系统必须记录 INFO 日志
- **并且**包含选中的 Skills 列表
- **以及**用户输入摘要

#### Scenario: 记录 Skill 加载状态
- **当**加载 Skill 内容时
- **那么**系统必须记录 DEBUG 日志
- **并且**显示哪些 Skills 从缓存获取
- **以及**哪些 Skills 从文件加载
