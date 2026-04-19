# 规格：session-persist

会话持久化能力，自动保存对话历史到 Markdown 文件。

> 📚 **知识上下文**：参考 `browser-manager-singleton` 单例模式管理全局 SessionManager，参考 `error-handling-strategy` 处理文件操作错误。

## Purpose

系统 SHALL 提供会话持久化能力，在每次对话过程中自动将消息保存到本地 Markdown 文件，并 MUST 按日期和 UUID 组织目录结构，确保会话可追溯和可恢复。

## Requirements

### Requirement: 会话标识生成

系统 SHALL 在会话开始时生成唯一标识符（UUID），用于标识和追踪会话。

#### Scenario: TUI 模式启动时生成 UUID
- **当** 用户启动 TUI 模式（`msa`）
- **那么** 系统生成一个 UUID 作为当前会话标识
- **那么** 欢迎信息显示会话 ID（UUID 前 8 位）

#### Scenario: CLI 模式执行时生成 UUID
- **当** 用户执行 CLI 模式（`msa -q "问题"`）
- **那么** 系统生成一个 UUID 作为当前会话标识
- **那么** 回复结束后输出完整会话 ID

### Requirement: 会话文件存储

系统 SHALL 将会话保存为 Markdown 文件，按日期组织目录结构。

#### Scenario: 文件路径生成
- **当** 创建新会话时
- **那么** 文件路径为 `~/.msa/memory/YYYY/MM/DD/{uuid}.md`
- **那么** 自动创建所需目录

#### Scenario: 文件格式
- **当** 创建会话文件时
- **那么** 文件包含 YAML Frontmatter（uuid, created_at, updated_at, mode）
- **那么** 文件正文按 Markdown 格式记录对话

### Requirement: 消息持久化

系统 SHALL 在对话过程中实时追加消息到会话文件。

#### Scenario: 用户消息持久化
- **当** 用户发送消息时
- **那么** 系统追加 `## 👤 用户\n{消息内容}\n\n` 到会话文件
- **那么** 更新 frontmatter 中的 `updated_at`

#### Scenario: 助手消息持久化
- **当** 助手回复完成（收到 `event.EventRoundDone`）时
- **那么** 系统追加 `## 🤖 MSA\n{回复内容}\n\n` 到会话文件
- **那么** 更新 frontmatter 中的 `updated_at`

#### Scenario: 追加写入
- **当** 持久化消息时
- **那么** 使用文件追加模式，不重写整个文件

### Requirement: 会话重置

系统 SHALL 允许用户通过 `clear` 命令重置会话，开始新的对话。

#### Scenario: clear 命令重置 UUID
- **当** 用户在 TUI 中输入 `clear` 命令
- **那么** 清空对话历史
- **那么** 生成新的 UUID
- **那么** 后续消息保存到新会话文件

### Requirement: 错误处理

系统 MUST 在文件操作失败时提供清晰的错误信息，且不影响主流程。

#### Scenario: 目录创建失败
- **当** 创建会话目录失败时
- **那么** 记录警告日志
- **那么** 继续执行对话（降级处理）

#### Scenario: 文件写入失败
- **当** 追加消息到文件失败时
- **那么** 记录错误日志
- **那么** 继续执行对话（降级处理）
