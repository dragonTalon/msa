# 提案：add-cli-chat

添加命令行单轮对话功能，支持 `-q` 和 `-m` 参数快速查询。

## 知识上下文

### 相关历史问题

知识库中未找到与 CLI 命令直接相关的问题。此为新功能开发。

### 需避免的反模式

未识别到特定反模式。

### 推荐模式

✅ **推荐**：考虑使用以下模式：

- **error-handling-strategy**：分类错误处理
  - 参数错误（如 -q 为空）立即返回
  - 永久错误（如配置缺失）不重试
  - 临时错误（如网络问题）可重试

- **config-auto-fill-on-provider-change**：配置优先级
  - `-m` 参数应具有最高优先级覆盖模型配置
  - 遵循 CLI > ENV > File > Default 的配置优先级原则

## 为什么

当前 MSA 必须通过 TUI 界面进行对话，适合交互式使用场景。但用户经常需要快速执行单次查询：

- 快速查询股票信息
- 脚本集成自动化
- CI/CD 流程调用

添加 CLI 单轮对话模式可以显著提升工具的可用性和集成能力。

## 变更内容

- 在 `cmd/root.go` 添加 `-q/--question` 和 `-m/--model` 参数
- 新建 `pkg/extcli/chat.go` 实现 CLI 流式输出
- 复用现有 `agent.Ask()` 和 `StreamOutputManager` 架构
- 支持三种消息类型输出：正文、思考过程、工具调用

## 能力

### 新增能力

- `cli-chat`：命令行单轮对话，支持流式输出

### 修改能力

无（这是纯新增功能，不修改现有能力需求）

## 影响

### 代码变更

- `cmd/root.go`：添加参数定义和路由逻辑
- `pkg/extcli/chat.go`：新增 CLI 对话处理模块
- `pkg/config/local_config.go`：可能需要支持临时模型覆盖

### 兼容性

- 完全向后兼容
- 无 `-q` 参数时行为不变，启动 TUI

### 依赖

- 复用现有 `pkg/logic/agent` 模块
- 复用现有 `pkg/logic/message/stream_manage.go`

## 非目标

- 不支持多轮对话历史上下文
- 不支持会话恢复（`--resume` 与 `-q` 互斥）
- 不支持自定义输出格式（JSON/Markdown）
- 不修改 TUI 现有行为

---

**知识备注**：
- 此提案使用 `openspec/knowledge/` 中的知识生成
- 相关模式：error-handling-strategy, config-auto-fill-on-provider-change