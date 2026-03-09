## Context

当前 `Ask` 函数在每次调用时，先通过 `selectSkillsViaSubAgent` 单独发起一次 LLM Generate（非流式）调用，由 `skills.Selector` 根据用户输入选择 skill 名称列表，再加载完整内容注入 UserMessage，最后才调用 `react.Agent.Stream`。这导致：
- 每次对话多一次 LLM 同步调用（延迟 + 费用）
- 逻辑分散在 sub-agent + agent 两个层次
- `manualSkills` 参数透传从 TUI 到 agent 层，增加耦合

`skills.Manager` 已是全局单例，skill metadata（name/description/priority）体积极小，完全适合在 system prompt 中一次性静态注入。

## Goals / Non-Goals

**Goals:**
- 将 skill 元数据列表静态注入 system prompt，让 agent 自主判断匹配的 skill
- 提供 `get_skill_content` tool，agent 按需懒加载 skill 完整内容
- 移除前置 sub-agent 的独立 LLM 调用，简化 `Ask` 函数签名
- 清理 TUI 层 `manualSkills`/`oneTimeSkills`/`sessionSkills` 逻辑

**Non-Goals:**
- 不修改 skill 加载、存储、启用/禁用逻辑
- 不改变流式输出管道
- 不重新设计 skill 文件格式

## Decisions

### 决策 1：skill 列表放 SystemMessage 而非 UserMessage

**选择**：注入 `BasePromptTemplate`（SystemMessage）

**理由**：skill 元数据是全局角色定义的一部分，与"当前问题是什么"无关；放在 system 层级语义更清晰，且不会随每轮 user 消息重复。

**备选**：继续放 UserMessage（现有做法）—— 否决，因为 user message 应只含当前问题，不含系统级规则。

---

### 决策 2：get_skill_content 直接调用全局单例

**选择**：tool 内部直接调用 `skills.GetManager()`

**理由**：`Manager` 本身已是 singleton 设计，其他 tool（如 stock、finance）也是无状态直调外部 API，风格一致；不引入注入机制可减少改动范围。

**备选**：构造时注入 `*skills.Manager` —— 否决，改动 `register.go` 初始化逻辑，复杂度收益比低。

---

### 决策 3：移除 manualSkills 参数而非保留兼容

**选择**：彻底删除 `manualSkills []string` 参数及 TUI 层的 `/skills:` 命令支持

**理由**：新架构中 agent 自主选择 skill，手动指定的意义消失；保留反而增加迷惑性。

**备选**：保留参数但不使用 —— 否决，死代码产生维护负担。

---

### 决策 4：get_skill_content 返回格式

**选择**：直接返回 SKILL.md 的 body 文本（已去掉 frontmatter）

**理由**：agent 需要的是可直接阅读的 Markdown 内容；不需要额外包装 JSON，减少 token 解析开销。

## Risks / Trade-offs

- **Skill 列表过长时 system prompt 膨胀** → 缓解：只注入 name + description + priority（metadata），不注入 body 内容；body 通过 tool 按需获取
- **agent 未调用 get_skill_content 直接回答** → 缓解：system prompt 中明确指令"有匹配 skill 时必须先调用 get_skill_content"；这是 prompt engineering 问题，可持续调整

## Migration Plan

1. 新增 `pkg/logic/tools/skill/skill.go`，实现 `get_skill_content`
2. 修改 `register.go` 注册新 tool
3. 修改 `prompt.go`：BasePromptTemplate 增加 skill 列表区块；BaseUserPrompt 去掉 `{{.skill_prompt}}`
4. 修改 `tempate.go`（BuildQueryMessages）：注入 skills 元数据，移除 skillPrompt 参数
5. 修改 `chat.go`（Ask）：删除 sub-agent 调用链，移除 manualSkills 参数
6. 修改 `pkg/tui/chat.go`：移除 sessionSkills/oneTimeSkills/startChatRequest 中的 manualSkills 传参
7. 修改 `cmd/root.go`：移除 manualSkillsKey 和 --skills 参数注入
8. `go vet ./...` 验证编译

回滚：git revert，无数据迁移需求。

## Open Questions

无
