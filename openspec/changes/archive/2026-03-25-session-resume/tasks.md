# 实现任务：session-resume

## ⚠️ 风险警报

基于历史问题和设计分析识别的风险：

### 中等风险：文件并发写入冲突

- **相关问题**：无历史问题，但设计分析识别
- **缓解措施**：使用 sync.Mutex 保护文件操作

### 低风险：解析兼容性

- **相关问题**：无历史问题
- **缓解措施**：严格解析格式，错误时给出清晰提示

## 1. 创建 session 模块基础结构

> 📚 **知识上下文**：
> 来源：知识库 patterns/browser-manager-singleton.md
> 经验：使用单例模式 + sync.Once 管理全局资源

- [x] 1.1 创建 `pkg/session/session.go`，定义 Session 数据结构
- [x] 1.2 创建 `pkg/session/manager.go`，实现 SessionManager 单例
- [x] 1.3 创建 `pkg/session/store.go`，实现文件存储路径生成逻辑
- [x] 1.4 验证：`go vet ./pkg/session/...` 无错误

## 2. 实现会话文件写入

> 📚 **知识上下文**：
> 来源：知识库 patterns/error-handling-strategy.md
> 经验：错误分类处理，区分可恢复和不可恢复错误

- [x] 2.1 实现 `NewSession()` 创建新会话，生成 UUID 和文件路径
- [x] 2.2 实现 `CreateSessionFile()` 创建会话文件，写入 frontmatter
- [x] 2.3 实现 `AppendUserMessage()` 追加用户消息到文件
- [x] 2.4 实现 `AppendAssistantMessage()` 追加助手消息到文件
- [x] 2.5 使用 `os.O_APPEND` 模式，避免重写整个文件
- [x] 2.6 错误处理：文件操作失败记录日志，不影响主流程
- [x] 2.7 验证：手动测试创建会话文件并追加消息

## 3. 实现会话文件解析

- [x] 3.1 创建 `pkg/session/parser.go`，实现 Markdown 解析
- [x] 3.2 实现 `ParseFrontmatter()` 解析 YAML frontmatter
- [x] 3.3 实现 `ParseMessages()` 解析用户和助手消息
- [x] 3.4 实现 `LoadSession()` 加载会话文件并返回历史消息
- [x] 3.5 错误处理：解析失败时返回清晰错误信息
- [x] 3.6 验证：手动测试解析会话文件

## 4. 集成到 CLI 模式

- [x] 4.1 修改 `pkg/extcli/chat.go`，注入 SessionManager
- [x] 4.2 在 `Run()` 开始时创建新会话
- [x] 4.3 用户问题发送后追加用户消息
- [x] 4.4 回复完成后追加助手消息
- [x] 4.5 回复结束后输出会话 ID 到 stdout
- [x] 4.6 验证：`msa -q "测试问题"` 输出会话 ID

## 5. 集成到 TUI 模式

- [x] 5.1 修改 `pkg/tui/chat.go`，添加 SessionManager 字段
- [x] 5.2 在 `NewChat()` 时创建新会话
- [x] 5.3 在 `handleEnterKey()` 发送消息后追加用户消息
- [x] 5.4 在 `chatStreamMsg()` 的 `IsDone` 分支追加助手消息
- [x] 5.5 欢迎信息显示当前会话 ID（UUID 前 8 位）
- [x] 5.6 修改 `handleClearHistory()`，重置 UUID 开始新会话
- [x] 5.7 验证：TUI 对话生成会话文件，clear 后生成新会话

## 6. 实现 --resume 参数

- [x] 6.1 修改 `cmd/root.go`，添加 `--resume` 参数
- [x] 6.2 实现 `ParseResumeArg()` 解析参数为文件路径
- [x] 6.3 在 `runRoot()` 中处理 `--resume` 参数
- [x] 6.4 加载会话文件，解析历史消息
- [x] 6.5 将历史消息注入到 TUI 的 history
- [x] 6.6 使用原有 UUID 继续会话
- [x] 6.7 欢迎信息显示"已恢复会话：{uuid前8位}"
- [x] 6.8 验证：`msa --resume {id}` 恢复会话继续对话

## 7. 错误处理与边界情况

> 📚 **知识上下文**：
> 来源：知识库 patterns/error-handling-strategy.md
> 经验：提供清晰的错误信息

- [x] 7.1 `--resume` 参数缺失会话 ID 时提示正确格式
- [x] 7.2 会话文件不存在时输出错误并退出
- [x] 7.3 会话文件格式错误时输出错误并退出
- [x] 7.4 frontmatter 缺失时输出错误并退出
- [x] 7.5 目录创建失败时记录警告继续执行
- [x] 7.6 验证：测试各种错误场景

## 8. 最终验证

- [x] 8.1 运行 `go vet ./...` 无错误
- [x] 8.2 测试 CLI 模式：`msa -q "问题"` 生成会话文件
- [x] 8.3 测试 TUI 模式：对话生成会话文件
- [x] 8.4 测试恢复：`msa --resume {id}` 正确恢复
- [x] 8.5 测试 clear：重置 UUID 开始新会话
- [x] 8.6 测试错误场景：各种错误提示正确

## 基于知识的测试

### 文件追加写入测试

> 📚 **来源**：设计文档 - 避免反模式

**历史问题**：
频繁重写整个文件导致性能下降。

**测试场景**：
1. 创建会话文件
2. 连续追加 10 条消息
3. 验证每次只追加，不重写整个文件
4. 验证文件内容正确

**预期行为**：
- 文件操作使用 `os.O_APPEND` 模式
- 每次追加只写入新内容

### 错误处理测试

> 📚 **来源**：知识库 patterns/error-handling-strategy.md

**历史问题**：
静默忽略错误导致问题难以排查。

**测试场景**：
1. 模拟目录创建失败
2. 验证日志输出警告
3. 验证程序继续执行

**预期行为**：
- 错误被正确记录
- 不影响主流程执行
