# 规格：agent-chat

对话能力与 Skill 系统的集成规格。

## Purpose

定义 agent 对话功能如何与 Skill 系统集成：通过 system prompt 注入 skill 元数据，由 agent 自主判断并按需调用 `get_skill_content` 获取完整流程。

## Requirements

### Requirement: Skill 元数据注入 System Prompt

系统 SHALL 在每次构建对话消息时，将所有可用 skill 的元数据（name、description、priority）静态注入 BasePromptTemplate（SystemMessage）。

#### Scenario: 正常构建消息
- **WHEN** `BuildQueryMessages` 被调用
- **THEN** SystemMessage SHALL 包含所有可用 skill 的名称、描述和优先级列表
- **AND** skill 列表通过 `skills.GetManager().ListSkills()` 获取（已过滤禁用的）
- **AND** skill 按 priority 降序排列（registry 已保证顺序）

#### Scenario: system prompt 包含 skill 选择指令
- **WHEN** agent 收到 system prompt
- **THEN** prompt SHALL 包含明确指令：识别匹配的 skill 后须调用 `get_skill_content` 获取完整流程
- **AND** prompt SHALL 说明"先选 skill，再按 skill 流程处理用户问题"

#### Scenario: 无可用 skill 时
- **WHEN** 所有 skill 均被禁用或 skill 目录为空
- **THEN** 系统 SHALL 正常工作，system prompt 中 skill 列表为空
- **AND** agent 继续使用基础 prompt 处理问题

---

### Requirement: 对话入口统一为 Runner.Ask

对话发起 SHALL 统一通过 `Runner.Ask(ctx, input, history)` 完成，不再直接调用 `agent.Ask`。

#### Scenario: TUI 发起对话
- **WHEN** TUI 层用户按 Enter 提交输入
- **THEN** 调用 `runner.Runner.Ask(ctx, input, history []model.Message)`
- **AND** Runner 内部创建 `core/agent.Agent`，调用 `Agent.Run(ctx, messages)` 返回 `<-chan event.Event`
- **AND** Runner 消费 event channel，通过注入的 `Renderer` 接口输出

#### Scenario: CLI 发起对话
- **WHEN** CLI 模式（`msa -q "..."`)
- **THEN** 同样调用 `runner.Runner.Ask`，注入 `CLIRenderer`（写 stdout）
- **AND** 单次对话，无历史

#### Scenario: BaseUserPrompt 简化
- **WHEN** 构建 user message
- **THEN** BaseUserPrompt SHALL 只包含用户问题占位符 `{{.question}}`
- **AND** 不包含 `{{.skill_prompt}}` 占位符

---

### Requirement: 错误处理

Skill 内容加载失败时 MUST 优雅降级，不影响对话功能。

#### Scenario: Skill 内容加载失败
- **WHEN** `get_skill_content` tool 调用失败时
- **THEN** 系统必须返回错误信息给 agent
- **AND** agent 应继续使用通用规范处理用户问题

#### Scenario: 所有 Skills 加载失败
- **WHEN** skills 目录初始化失败时
- **THEN** 系统必须记录警告日志
- **AND** 保证基本对话功能可用（system prompt 中 skill 列表为空）

---

### Requirement: 日志记录

对话流程中的关键节点 MUST 记录结构化日志，支持 requestID 链路追踪。

#### Scenario: requestID 注入
- **WHEN** `Runner.Ask` 被调用
- **THEN** 系统必须通过 `corelogger.WithRequestID(ctx)` 生成 6 位十六进制 requestID 注入 ctx
- **AND** 整条链路（Runner → Agent → StreamAdapter → Tool）所有日志均携带 `req` 字段

#### Scenario: 记录关键节点
- **WHEN** 一轮对话执行中
- **THEN** 系统必须记录 INFO 日志，包含：Runner 收到输入、每轮 LLM 调用开始/结束、工具执行开始/完成
- **AND** 工具执行记录包含 name、elapsed、outputLen

#### Scenario: 记录消息构建
- **WHEN** `BuildQueryMessages` 被调用
- **THEN** 系统必须记录 INFO 日志，包含消息总数

---

### Requirement: 会话初始化增强

聊天会话初始化时 MUST 同时初始化记忆管理器。

#### Scenario: 启动时初始化记忆系统

- **WHEN** 创建新的 Chat 实例
- **THEN** 初始化 MemoryManager（单例）
- **AND** 生成唯一的会话 ID（UUID）
- **AND** 记录会话开始时间
- **AND** 创建空的短期内存（最多 10 条消息）
- **AND** 如果记忆系统初始化失败，记录警告但不阻止聊天

> 📚 **推荐模式**：使用单例模式（pattern 004）
> MemoryManager 应使用 sync.Once 确保全局唯一实例

### Requirement: 消息记录增强

每条消息 MUST 同时记录到历史和记忆系统。

#### Scenario: 用户输入消息

- **WHEN** 用户输入消息并按 Enter
- **THEN** 将消息添加到 Chat.history（现有行为）
- **AND** 调用 memoryManager.AddMessage() 记录到记忆
- **AND** 更新短期内存（保持最近 10 条）
- **AND** 持久化短期内存到 `~/.msa/remember/short-term/current.json`
- **AND** 如果记录失败，不影响聊天继续进行

#### Scenario: AI 回复消息

- **WHEN** AI 生成回复消息
- **THEN** 将消息添加到 Chat.history（现有行为）
- **AND** 调用 memoryManager.AddMessage() 记录到记忆
- **AND** 更新会话上下文（如提取股票代码）

#### Scenario: 更新会话上下文

- **WHEN** 记录消息到记忆系统
- **THEN** 分析消息内容更新会话上下文：
  - 提取股票代码（如 00700.HK、600519）
  - 检测使用的工具
  - 检测交易操作
  - 生成标签候选
- **AND** 上下文信息用于后续的知识提取

### Requirement: 会话结束增强

会话结束时 MUST 保存完整会话并触发知识提取。

#### Scenario: 正常退出保存会话

- **WHEN** 用户按 Ctrl+C 或 ESC 退出
- **THEN** 调用 memoryManager.EndSession()
- **AND** 保存完整会话到 `~/.msa/remember/sessions/<YYYY-MM>/<session-id>.json`
- **AND** 异步触发 AI 知识提取
- **AND** 更新会话索引
- **AND** 清空短期内存文件
- **AND** 退出不应被保存操作阻塞

#### Scenario: 异常中断恢复

- **WHEN** 检测到未完整保存的会话（如短期内存文件存在但会话不存在）
- **THEN** 下次启动时提示用户
- **AND** 询问是否恢复会话
- **AND** 如果用户确认，将短期内存内容合并到新会话

> 📚 **推荐模式**：遵循 AI 助手工作流程（pattern 002）
> 知识提取是异步任务，不应阻塞用户退出或自动执行影响性操作

### Requirement: 短期内存管理

系统 MUST 维护最近 10 条消息的快速访问内存。

#### Scenario: 维护短期内存

- **WHEN** 添加新消息时
- **THEN** 将消息添加到短期内存
- **AND** 如果超过 10 条，移除最旧的
- **AND** 只保留用户和助手的消息（过滤系统消息）
- **AND** 持久化到 `short-term/current.json`

#### Scenario: 启动时恢复短期内存

- **WHEN** MSA 启动时
- **THEN** 检查是否存在短期内存文件
- **AND** 如果存在，询问用户是否继续上次会话
- **AND** 如果用户确认，加载短期内存内容到新会话
- **AND** 如果用户拒绝，删除短期内存文件

### Requirement: 记忆系统状态显示

用户 SHALL 能够看到记忆系统的状态。

#### Scenario: 显示记忆系统状态

- **WHEN** 用户输入 `/status` 或类似命令
- **THEN** 显示记忆系统状态：
  - 当前会话 ID
  - 会话时长
  - 消息数量
  - 记忆系统状态（启用/禁用）
  - 知识库大小
- **AND** 如果记忆系统被禁用，显示启用方法

#### Scenario: 记忆系统错误通知

- **WHEN** 记忆系统操作失败
- **THEN** 在下次发送消息时显示系统提示
- **AND** 说明错误类型（如"保存会话失败"）
- **AND** 不影响聊天继续进行
- **AND** 只通知一次（不重复）

### Requirement: 记忆系统可选禁用

用户 SHALL 能够选择不使用记忆系统。

#### Scenario: 禁用记忆系统

- **WHEN** 用户设置环境变量 `MSA_MEMORY_ENABLED=false`
- **THEN** 跳过记忆系统初始化
- **AND** Chat 行为与现有版本完全一致
- **AND** 不显示任何记忆相关提示

#### Scenario: 运行时禁用记忆

- **WHEN** 用户执行 `/remember disable` 命令
- **THEN** 停止记录消息到记忆
- **AND** 会话结束时不再保存
- **AND** 设置应持久化（下次启动仍然禁用）
- **AND** 可以通过 `/remember enable` 重新启用

---

### Requirement: 输入处理增强

聊天输入框 MUST 处理命令自动补全相关的按键事件。

#### Scenario: 处理 Tab 键补全

- **WHEN** 用户在命令模式下按下 Tab 键
- **THEN** 执行命令补全（如有匹配项）
- **AND** 更新输入框内容
- **AND** 退出命令模式

#### Scenario: 处理方向键导航

- **WHEN** 用户在命令模式下按下 `↑` 或 `↓` 键
- **THEN** 更新选中项索引
- **AND** 重新渲染建议列表
- **AND** 不影响输入框内容

#### Scenario: 处理 Enter 执行命令

- **WHEN** 用户在命令模式下按下 Enter 键
- **THEN** 执行当前选中的命令（如有选中）
- **OR** 执行输入框中的命令（如无选中但输入有效命令）

#### Scenario: 处理 Esc 取消

- **WHEN** 用户在命令模式下按下 Esc 键
- **THEN** 清空输入框
- **AND** 退出命令模式
- **AND** 关闭建议列表

### Requirement: 命令建议数据

Chat 模型 MUST 维护命令建议的状态数据。

#### Scenario: 维护建议状态

- **WHEN** Chat 初始化
- **THEN** 包含以下命令建议相关字段：
  - `commandSelectedIndex` - 当前选中项索引
  - `commandSuggestions` - 建议列表（包含名称和描述）

#### Scenario: 更新建议列表

- **WHEN** 输入内容匹配命令前缀
- **THEN** 调用 `command.GetCommandSuggestions()` 获取带描述的建议
- **AND** 更新 `commandSuggestions` 字段
- **AND** 重置 `commandSelectedIndex` 为 0
