# 规格：cli-chat

> 📚 **知识上下文**：基于推荐模式 error-handling-strategy 和 config-auto-fill-on-provider-change

## Purpose

CLI chat 功能 SHALL 允许用户通过命令行参数直接发起单轮对话，无需进入 TUI 界面。

## Requirements

### Requirement: CLI 单轮对话参数

系统 SHALL 支持通过命令行参数快速发起单轮对话，无需进入 TUI 界面。

#### Scenario: 使用默认模型发起对话

- **当** 用户执行 `msa -q "茅台最近走势怎么样"`
- **那么** 系统使用配置文件中的默认模型发起对话
- **那么** 流式输出响应内容到终端
- **那么** 对话完成后程序退出，返回退出码 0

#### Scenario: 指定模型发起对话

- **当** 用户执行 `msa -q "茅台最近走势怎么样" -m "deepseek-r1"`
- **那么** 系统使用指定的模型发起对话
- **那么** 模型参数优先级高于配置文件

#### Scenario: 无 TUI 模式启动

- **当** 用户执行 `msa` 不带 `-q` 参数
- **那么** 系统启动 TUI 界面（现有行为保持不变）

---

### Requirement: 流式输出格式

CLI 模式 SHALL 通过 `Runner.Ask` + `CLIRenderer` 处理流式输出，直接写 stdout。

#### Scenario: 正文消息输出

- **当** 收到 `event.EventTextChunk` 类型事件
- **那么** 直接输出 `e.Text` 到 stdout（不加换行）
- **当** 收到 `event.EventTextDone`
- **那么** 输出换行

#### Scenario: 思考过程输出

- **当** 收到 `event.EventThinking` 类型事件
- **那么** 如果 verbose=true，输出思考内容到 stdout
- **那么** 默认（verbose=false）不输出思考内容

#### Scenario: 工具调用输出

- **当** 收到 `event.EventToolStart` 类型事件
- **那么** 输出 `⚙ 正在调用 <tool_name>...` 格式
- **当** 收到 `event.EventToolResult`
- **那么** 输出 `✓ <tool_name> 完成`
- **当** 收到 `event.EventToolError`
- **那么** 输出 `✗ <tool_name> 失败: <err>`

#### Scenario: 流式输出完成

- **当** 收到 `event.EventRoundDone` 或 event channel 关闭
- **那么** `Runner.Ask` 返回
- **那么** CLI 程序输出会话 ID 后退出，返回退出码 0

---

### Requirement: 错误处理

CLI 模式 SHALL 分类错误处理，提供清晰的错误信息和退出码。

> 📚 **历史问题参考**：基于模式 error-handling-strategy

#### Scenario: 参数错误

- **当** 用户执行 `msa -q ""` (空问题)
- **那么** 输出错误信息 "问题内容不能为空"
- **那么** 程序退出，返回退出码 1

#### Scenario: 配置缺失（永久错误）

- **当** 配置文件中 API Key 为空
- **那么** 输出错误信息 "请先配置 API Key，运行 'msa config'"
- **那么** 程序退出，返回退出码 1
- **那么** 不进行重试

#### Scenario: 模型未配置（永久错误）

- **当** 未指定 `-m` 参数且配置文件中 Model 为空
- **那么** 输出错误信息 "请指定模型 (-m) 或配置默认模型"
- **那么** 程序退出，返回退出码 1

#### Scenario: 流式请求失败

- **当** AI 请求过程中发生错误
- **那么** 输出错误信息到 stderr
- **那么** 程序退出，返回退出码 1

---

### Requirement: 模型参数优先级

CLI 模式 SHALL 按照特定优先级处理模型参数配置。

#### Scenario: -m 参数覆盖模型配置

- **当** 用户指定 `-m "deepseek-r1"`
- **那么** 使用指定模型，忽略配置文件中的 Model 字段
- **那么** 其他配置（Provider, APIKey, BaseURL）仍从配置文件读取

#### Scenario: -m 参数未指定

- **当** 用户未指定 `-m` 参数
- **那么** 使用配置文件中的 Model 字段
- **那么** 如果 Model 为空，返回错误

---

### Requirement: 参数互斥

系统 SHALL 处理互斥参数的冲突情况。

#### Scenario: -q 与 --resume 互斥

- **当** 用户同时指定 `-q` 和 `--resume` 参数
- **那么** 输出警告信息 "-q 和 --resume 不能同时使用"
- **那么** 忽略 --resume 参数，执行单轮对话

---

### Requirement: 退出码规范

CLI 模式 SHALL 遵循统一的退出码规范。

#### Scenario: 成功退出

- **当** 对话正常完成
- **那么** 返回退出码 0

#### Scenario: 错误退出

- **当** 任何错误发生
- **那么** 返回退出码 1
- **那么** 错误信息输出到 stderr
