## 1. 新增 get_skill_content Tool

- [x] 1.1 创建 `pkg/logic/tools/skill/skill.go`，实现 `SkillContentTool` 结构体，包含 `GetToolInfo`、`GetName`、`GetDescription`、`GetToolGroup` 方法
- [x] 1.2 实现 `GetSkillContent` 函数：接收 `SkillContentParam{SkillName string}`，调用 `skills.GetManager().GetSkill(name)` 获取 skill，再调用 `skill.GetContent()` 返回 body 内容
- [x] 1.3 处理 skill 不存在和内容加载失败两种错误情况，返回明确错误信息并记录日志
- [x] 1.4 在 `pkg/logic/tools/register.go` 中添加 `registerSkill()` 函数，注册 `SkillContentTool`，并在 `init()` 中调用

## 2. 修改 System Prompt 结构

- [x] 2.1 修改 `pkg/logic/agent/prompt.go`：在 `BasePromptTemplate` 末尾添加 skill 列表区块，使用 `{{range .skills}}` 模板变量渲染 name/priority/description，并添加"先识别 skill 再调用 get_skill_content"的处理指令
- [x] 2.2 修改 `BaseUserPrompt`：删除 `{{.skill_prompt}}` 区块及相关说明，只保留 `{{.question}}` 占位符

## 3. 修改 BuildQueryMessages

- [x] 3.1 修改 `pkg/logic/agent/tempate.go` 中的 `BuildQueryMessages`：移除 `skillPrompt string` 参数，新增 skills 元数据注入（从 `skills.GetManager().ListSkills()` 获取，转为模板可用的轻量结构）
- [x] 3.2 更新 `templateVars`：删除 `skill_prompt` 键，新增 `skills` 键（`[]SkillMeta{Name, Description, Priority}`）

## 4. 简化 Ask 函数及删除 Sub-Agent 调用链

- [x] 4.1 修改 `pkg/logic/agent/chat.go`：`Ask` 函数签名移除 `manualSkills []string` 参数
- [x] 4.2 删除 `selectSkillsViaSubAgent`、`buildSkillPromptText`、`validateAndEnsureBase` 三个函数
- [x] 4.3 简化 `Ask` 函数体：初始化 skillManager（确保 skills 已加载）→ 过滤历史 → 调用 `BuildQueryMessages` → 异步调用 `agent.Stream`
- [x] 4.4 删除 `GetChatModelOnly` 函数（不再需要给 selector 使用）及 `chatModelCache` 变量（如无其他引用）

## 5. 清理 TUI 层和 CMD 层

- [x] 5.1 修改 `pkg/tui/chat.go`：删除 `sessionSkills`、`oneTimeSkills` 字段，删除 `handleSkillsCommand` 函数，删除 `startChatRequest` 中的 skills 相关逻辑，更新 `agent.Ask` 调用（移除第四个参数）
- [x] 5.2 修改 `pkg/tui/chat.go`：删除 `manualSkillsKey` 常量和 `NewChat` 中读取 context skills 的逻辑，删除启动时显示 skills 的提示消息
- [x] 5.3 修改 `cmd/root.go`：删除 `manualSkillsKey` 常量和向 context 注入 skills 的逻辑，删除 `--skills` CLI 参数（如有）

## 6. 验证

- [x] 6.1 运行 `go vet ./...` 确认无编译错误
- [x] 6.2 检查 `skills.GetManager()` 在 tool 被调用前已完成 `Initialize()`（在 `Ask` 开头调用 `skillManager.Initialize()` 保证）
