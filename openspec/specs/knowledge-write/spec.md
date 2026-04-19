# 规格：knowledge-write

> 📚 **知识上下文**：基于模式 #003（错误处理和重试策略），写入操作必须验证成功并支持重试

## Purpose

write_knowledge 工具 SHALL 持久化交易知识（错误记录和每日总结），并提供写入验证和自动重试能力。

## Requirements

### Requirement: 写入错误记录

工具 MUST 能够追加错误记录到 errors.md。

#### Scenario: 首次写入错误
- **当** `~/.msa/errors.md` 文件不存在
- **那么** 创建文件并写入标题和第一条错误记录

#### Scenario: 追加错误记录
- **当** `~/.msa/errors.md` 文件已存在
- **那么** 追加新的错误记录，不覆盖已有内容

#### Scenario: 错误记录格式
- **当** 写入错误记录
- **那么** 包含以下字段：日期、标题、股票、错误描述、原因、教训

---

### Requirement: 写入每日总结

工具 MUST 能够写入每日总结文件。

#### Scenario: 创建总结文件
- **当** 调用 `write_knowledge(type="summary", date="2026-04-06", content=...)`
- **那么** 创建 `~/.msa/summaries/2026-04-06.md` 文件

#### Scenario: 覆盖已有总结
- **当** 同一日期的总结已存在
- **那么** 覆盖旧文件（每日总结只保留最新版本）

#### Scenario: 日期格式验证
- **当** date 参数格式不是 YYYY-MM-DD
- **那么** 返回参数错误

---

### Requirement: 写入验证

工具 MUST 验证写入成功。

> 📚 **历史问题参考**：基于模式 #003，文件写入失败可能导致数据丢失

#### Scenario: 验证内容完整性
- **当** 写入文件后
- **那么** 重新读取文件，验证内容已正确写入

#### Scenario: 验证失败时重试
- **当** 验证失败
- **那么** 自动重试，最多 3 次

#### Scenario: 返回验证状态
- **当** 写入完成
- **那么** 返回 `verified: true/false` 告知调用者

---

### Requirement: 参数验证

工具 MUST 验证输入参数。

#### Scenario: type 必填
- **当** 未提供 type 参数
- **那么** 返回错误 "type 是必填项"

#### Scenario: content 必填
- **当** 未提供 content 参数
- **那么** 返回错误 "content 是必填项"

#### Scenario: type 值验证
- **当** type 不是 "error" 或 "summary"
- **那么** 返回错误 "无效的 type，必须是 error 或 summary"

---

### Requirement: 目录自动创建

工具 MUST 自动创建必要的目录。

#### Scenario: .msa 目录不存在
- **当** `~/.msa/` 目录不存在
- **那么** 自动创建目录

#### Scenario: summaries 目录不存在
- **当** 写入总结时 `~/.msa/summaries/` 不存在
- **那么** 自动创建目录

---

### Requirement: 返回结构化数据

工具 MUST 返回结构化的 JSON 数据。

#### Scenario: 成功返回格式
- **当** 写入成功
- **那么** 返回包含以下字段：
  - `type`: 写入类型
  - `path`: 文件路径
  - `verified`: 是否验证成功
  - `content_len`: 写入内容长度
  - `retry_count`: 重试次数

#### Scenario: 失败返回格式
- **当** 写入失败（重试 3 次后仍失败）
- **那么** 返回 `success: false` 和错误信息

---

### Requirement: 工具注册

工具 MUST 注册到知识库工具组。

#### Scenario: 工具可被发现
- **当** 系统列出可用工具
- **那么** `write_knowledge` 出现在 knowledge 工具组中

#### Scenario: 工具描述
- **当** 查看工具描述
- **那么** 显示"写入交易知识库（错误记录 / 每日总结），带验证和重试"
