# 实现任务：add-knowledge-tools

## ⚠️ 风险警报

基于历史问题识别的风险：

### 高：写入失败导致数据丢失

- **相关问题**：模式 #003
- **缓解措施**：写入验证 + 自动重试 + 返回验证状态

### 中：Agent 忽略验证失败

- **相关问题**：问题 #001（AI 助手可能忽略规则）
- **缓解措施**：在 SKILL 中明确要求检查 verified 字段，使用强语气

---

## 1. 基础设施 - 路径管理

- [x] 1.1 创建 `pkg/logic/tools/knowledge/` 目录

- [x] 1.2 创建 `pkg/logic/tools/knowledge/paths.go`
  - 定义常量 `KnowledgeDir = ".msa"`
  - 定义常量 `ErrorsFile = "errors.md"`
  - 定义常量 `SummariesDir = "summaries"`
  - 实现 `GetKnowledgeDir()` 函数
  - 实现 `GetErrorsPath()` 函数
  - 实现 `GetSummariesDir()` 函数
  - 实现 `GetSummaryPath(date string)` 函数
  - 实现 `EnsureDir(path string)` 函数

> 📚 **知识上下文**：避免反模式 - 不要硬编码文件路径，使用函数动态获取用户目录

验收标准：
- [x] 所有路径函数返回正确的用户目录路径
- [x] EnsureDir 能自动创建不存在的目录

---

## 2. 文件解析器

- [x] 2.1 创建 `pkg/logic/tools/knowledge/parser.go`
  - 定义 `ErrorEntry` 结构体（date、title、stock、error、reason、lesson）
  - 定义 `SummaryData` 结构体（date、prev_asset、curr_asset、pnl、pnl_rate、raw）
  - 实现 `ParseErrorsFile(path)` 函数
  - 实现 `parseErrorsContent(content)` 函数
  - 实现 `FindPrevSummaryFile()` 函数
  - 实现 `ParseSummaryFile(path)` 函数
  - 实现 `isValidDate(date)` 验证函数

> 📚 **知识上下文**：参考模式 #003 - 文件不存在时返回空列表/nil，不报错

验收标准：
- [x] errors.md 不存在时返回空列表
- [x] summaries 目录不存在时返回 nil
- [x] 正确解析 Markdown 格式的错误条目
- [x] 正确解析 YAML frontmatter 的总结数据

---

## 3. read_knowledge 工具

- [x] 3.1 创建 `pkg/logic/tools/knowledge/read.go`
  - 定义 `ReadKnowledgeParam` 结构体（type 字段）
  - 定义 `ReadKnowledgeData` 结构体（errors、prev_summary、has_errors、has_summary）
  - 定义 `ReadKnowledgeTool` 结构体
  - 实现 `GetToolInfo()` 方法
  - 实现 `GetName()` 返回 "read_knowledge"
  - 实现 `GetDescription()` 返回中文描述
  - 实现 `GetToolGroup()` 返回 `model.KnowledgeToolGroup`
  - 实现 `ReadKnowledge()` 函数

> 📚 **知识上下文**：参考模式 #002 - 使用 Eino 框架的 tool 接口

验收标准：
- [x] type="errors" 只返回错误记录
- [x] type="summary" 只返回昨日总结
- [x] type="all" 返回两者
- [x] 文件不存在时不报错，返回空数据
- [x] 返回值使用 `model.NewSuccessResult` 格式

---

## 4. write_knowledge 工具

- [x] 4.1 创建 `pkg/logic/tools/knowledge/write.go`
  - 定义 `WriteKnowledgeParam` 结构体（type、content、date）
  - 定义 `WriteKnowledgeData` 结构体（type、path、verified、content_len、retry_count）
  - 定义常量 `MaxRetries = 3`
  - 定义常量 `RetryDelay = 100 * time.Millisecond`
  - 定义 `WriteKnowledgeTool` 结构体
  - 实现 `GetToolInfo()` 方法
  - 实现 `GetName()` 返回 "write_knowledge"
  - 实现 `GetDescription()` 返回中文描述
  - 实现 `GetToolGroup()` 返回 `model.KnowledgeToolGroup`

- [x] 4.2 实现 `WriteKnowledge()` 函数
  - 参数验证（type、content 必填）
  - 调用 `executeWrite()` 执行写入
  - 调用 `verifyWrite()` 验证写入
  - 失败时重试（最多 3 次）
  - 返回验证状态

- [x] 4.3 实现 `validateWriteParam()` 参数验证函数
  - 验证 type 必填且为 "error" 或 "summary"
  - 验证 content 必填
  - 验证 date 格式（仅 summary 类型）

- [x] 4.4 实现 `executeWrite()` 写入函数
  - type="error" 调用 `appendError()`
  - type="summary" 调用 `writeSummary()`

- [x] 4.5 实现 `appendError()` 追加错误函数
  - 文件不存在时创建并写入标题
  - 文件存在时追加内容
  - 自动添加分隔符

- [x] 4.6 实现 `writeSummary()` 写入总结函数
  - 自动创建 summaries 目录
  - 覆盖同名文件

- [x] 4.7 实现 `verifyWrite()` 验证函数
  - 重新读取文件
  - 验证内容是否包含写入的内容
  - 调用 `verifyErrorContent()` 或 `verifySummaryContent()`

> 📚 **知识上下文**：参考模式 #003 - 错误处理和重试策略，验证失败时自动重试

验收标准：
- [x] 写入后验证内容完整性
- [x] 验证失败时自动重试最多 3 次
- [x] 返回 `verified: true/false`
- [x] 目录不存在时自动创建
- [x] 返回值使用 `model.NewSuccessResult` 或 `model.NewErrorResult`

---

## 5. 工具注册

- [x] 5.1 创建 `pkg/logic/tools/knowledge/register.go`
  - 定义 `KnowledgeTools` 变量列表

- [x] 5.2 修改 `pkg/logic/tools/register.go`
  - 导入 knowledge 包
  - 添加 `registerKnowledge()` 函数
  - 在 `init()` 中调用 `registerKnowledge()`

验收标准：
- [x] 运行 `go vet ./...` 无错误
- [x] 工具可通过 Agent 调用

---

## 6. 更新 SKILL 文件

- [x] 6.1 修改 `pkg/logic/skills/plugs/morning-analysis/SKILL.md`
  - 移除 `dependencies` 中的 `read-error`
  - 添加 `dependencies` 中的 `knowledge-read`
  - 更新 Step 1.1 为调用 `read_knowledge(type="all")`
  - 更新 tools 列表

- [x] 6.2 修改 `pkg/logic/skills/plugs/afternoon-trade/SKILL.md`
  - 移除 `dependencies` 中的 `read-error`
  - 添加 `dependencies` 中的 `knowledge-read`
  - 更新 Step 1.1 为调用 `read_knowledge(type="all")`
  - 更新 tools 列表

- [x] 6.3 修改 `pkg/logic/skills/plugs/market-close-summary/SKILL.md`
  - 移除 `dependencies` 中的 `read-error`
  - 添加 `dependencies` 中的 `knowledge-read` 和 `knowledge-write`
  - 更新 Step 1.1 为调用 `read_knowledge(type="all")`
  - 更新 Step 5 为调用 `write_knowledge`
  - 更新 tools 列表

验收标准：
- [x] SKILL.md 格式正确
- [x] dependencies 正确引用新能力

---

## 7. 删除旧组件

- [x] 7.1 删除 `pkg/logic/skills/plugs/read-error/` 目录

验收标准：
- [x] 目录已删除
- [x] 无其他代码引用该 SKILL

---

## 8. 集成测试

- [x] 8.1 创建 `pkg/logic/tools/knowledge/parser_test.go`
  - 测试 `ParseErrorsFile()` 空文件情况
  - 测试 `ParseErrorsFile()` 正常解析
  - 测试 `FindPrevSummaryFile()` 空目录情况
  - 测试 `FindPrevSummaryFile()` 正常查找
  - 测试 `ParseSummaryFile()` 正常解析

- [x] 8.2 测试 read_knowledge 工具
  - 测试 type="errors"
  - 测试 type="summary"
  - 测试 type="all"
  - 测试文件不存在情况

- [x] 8.3 测试 write_knowledge 工具
  - 测试写入错误记录
  - 测试写入总结
  - 测试验证失败重试
  - 测试参数验证

验收标准：
- [x] 运行 `go test ./pkg/logic/tools/knowledge/...` 全部通过
- [x] 运行 `go vet ./...` 无错误

---

## 基于知识的测试

### 模式 #003 回归测试 - 错误处理和重试

> 📚 **来源**：模式 #003 - 错误处理和重试策略

**历史经验**：
文件写入失败可能导致数据丢失。

**测试场景**：
1. 模拟写入失败（权限问题）
2. 验证自动重试（最多 3 次）
3. 验证返回 `verified: false`
4. 验证错误信息清晰

**预期行为**：
- 重试次数不超过 3 次
- 返回明确的错误信息
- 不 panic

### 问题 #001 回归测试 - Agent 忽略规则

> 📚 **来源**：问题 #001 - AI 助手不应自动执行 git commit

**历史问题**：
AI 助手可能忽略规则，跳过必要步骤。

**测试场景**：
1. SKILL 执行时检查是否调用了 read_knowledge
2. 写入失败时检查是否检查了 verified 字段
3. 验证 SKILL 中的强语气提示有效

**预期行为**：
- Agent 必须先调用 read_knowledge
- Agent 必须检查 write_knowledge 的 verified 字段
- verified=false 时必须重试或报告错误
