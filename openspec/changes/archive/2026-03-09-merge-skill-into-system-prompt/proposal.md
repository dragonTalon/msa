## Why

当前架构在每次用户提问前，通过独立的前置 sub-agent（Selector）额外发起一次 LLM Generate 调用来选择 skill，增加了延迟和成本。实际上，agent 本身完全有能力在 system prompt 的引导下自主判断适用的 skill，并通过 tool 按需获取 skill 完整内容，无需两步拆分。

## What Changes

- **移除** `selectSkillsViaSubAgent`、`buildSkillPromptText`、`validateAndEnsureBase` 函数及前置 sub-agent 的整个调用链
- **移除** `Ask` 函数的 `manualSkills []string` 参数，以及 TUI 层相关的 `sessionSkills`/`oneTimeSkills`/`/skills:` 命令逻辑
- **修改** `BasePromptTemplate`：在 system prompt 中注入所有可用 skill 的名称、描述与优先级列表，并添加"先识别匹配 skill，再调用 get_skill_content 获取完整流程"的处理指令
- **修改** `BuildQueryMessages`：注入 skill 元数据列表，去掉 `skill_prompt` 参数
- **新增** `pkg/logic/tools/skill/skill.go`：实现 `get_skill_content` tool，接收 skill name，返回对应 SKILL.md 完整内容
- **修改** `pkg/logic/tools/register.go`：注册新 tool

## Non-Goals

- 不修改 skill 的加载、存储、启用/禁用逻辑
- 不修改 `skills.Manager` 单例以外的 skills 包接口
- 不改变流式输出管道和 TUI 渲染逻辑

## Capabilities

### New Capabilities
- `skill-content-tool`: 提供 `get_skill_content` tool，供 agent 在推理阶段按需查询 skill 完整内容

### Modified Capabilities
- `skill-selector`: 原有 LLM 自动选择器流程被移除，skill 选择职责转移给 agent 通过 system prompt 自主决策
- `agent-chat`: Ask 函数签名变更，system prompt 结构调整，去掉 manualSkills 参数

## Impact

- `pkg/logic/agent/chat.go`：Ask 签名变更（移除 manualSkills），删除 sub-agent 调用链
- `pkg/logic/agent/prompt.go`：BasePromptTemplate 增加 skill 列表区块，BaseUserPrompt 去掉 skill_prompt 占位
- `pkg/logic/agent/tempate.go`：BuildQueryMessages 参数变更
- `pkg/logic/tools/skill/skill.go`：新文件
- `pkg/logic/tools/register.go`：新增注册
- `pkg/tui/chat.go`：移除 manualSkills/oneTimeSkills/sessionSkills 相关逻辑
- `cmd/root.go`：移除 manualSkillsKey 和 --skills 参数注入逻辑
