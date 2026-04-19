# 规格：knowledge-read

> 📚 **知识上下文**：基于模式 #003（错误处理和重试策略），读取操作应优雅处理文件不存在的情况

## Purpose

read_knowledge 工具 SHALL 提供对交易知识库（历史错误记录和昨日总结）的结构化读取能力，并优雅处理文件不存在的情况。

## Requirements

### Requirement: 读取历史错误

工具 MUST 能够读取历史错误记录。

#### Scenario: 文件存在时读取错误
- **当** `~/.msa/errors.md` 文件存在
- **那么** 解析文件内容，返回结构化的错误列表

#### Scenario: 文件不存在时返回空列表
- **当** `~/.msa/errors.md` 文件不存在
- **那么** 返回空列表，不报错

#### Scenario: 解析错误条目
- **当** errors.md 包含格式正确的错误记录
- **那么** 提取以下字段：date、title、stock、error、reason、lesson

---

### Requirement: 读取昨日总结

工具 MUST 能够读取最近一个交易日的总结。

#### Scenario: 存在多个总结文件
- **当** `~/.msa/summaries/` 目录下存在多个日期的总结文件
- **那么** 返回今天之前的最后一个工作日的总结

#### Scenario: 无总结文件
- **当** `~/.msa/summaries/` 目录不存在或为空
- **那么** 返回 `has_summary: false`，不报错

#### Scenario: 解析总结内容
- **当** 总结文件存在
- **那么** 返回以下信息：date、prev_asset、curr_asset、pnl、pnl_rate、raw（原始内容）

---

### Requirement: 按类型读取

工具 MUST 支持按类型读取。

#### Scenario: 只读错误
- **当** 调用 `read_knowledge(type="errors")`
- **那么** 只返回错误记录，不读取总结

#### Scenario: 只读总结
- **当** 调用 `read_knowledge(type="summary")`
- **那么** 只返回昨日总结，不读取错误

#### Scenario: 读取全部
- **当** 调用 `read_knowledge(type="all")`
- **那么** 同时返回错误记录和昨日总结

---

### Requirement: 返回结构化数据

工具 MUST 返回结构化的 JSON 数据。

#### Scenario: 返回格式
- **当** 读取成功
- **那么** 返回包含以下字段：
  - `errors`: 错误条目数组
  - `prev_summary`: 昨日总结对象
  - `has_errors`: 是否有错误
  - `has_summary`: 是否有总结
  - `errors_path`: 错误文件路径
  - `summary_path`: 总结文件路径

---

### Requirement: 工具注册

工具 MUST 注册到知识库工具组。

#### Scenario: 工具可被发现
- **当** 系统列出可用工具
- **那么** `read_knowledge` 出现在 knowledge 工具组中

#### Scenario: 工具描述
- **当** 查看工具描述
- **那么** 显示"读取交易知识库（历史错误 + 昨日总结）"
