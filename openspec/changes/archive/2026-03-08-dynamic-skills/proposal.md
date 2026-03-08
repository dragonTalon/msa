# 提案：dynamic-skills

实现动态 Skill 系统，支持 LLM 自动选择和按需加载 Prompt 模板，解决当前硬编码 Prompt 不灵活的问题。

## 知识上下文

### 相关历史问题

知识库中未找到与 Skill 系统直接相关的历史问题。这是新领域的变更。

### 需避免的反模式

未识别到特定反模式。但需注意以下潜在问题：
- 避免过度设计：从简单开始，逐步扩展
- 避免每次对话都产生额外 LLM 调用的性能问题

### 推荐模式

基于现有项目模式：

- **006**：Provider 切换时自动填充配置
  - 描述：配置变更时的自动化处理模式
  - 使用时机：当配置变更需要自动触发其他操作时

- **003**：错误处理和重试策略
  - 描述：LLM 调用失败时的降级策略
  - 使用时机：Skill 选择失败时应有兜底方案

### 风险评估

此变更属于架构级变更，具有以下风险因素：

- **中等**：LLM 调用开销增加（每次对话多一次选择调用）
  - 缓解措施：使用轻量级 prompt，缓存选择结果

- **低**：现有 Prompt 兼容性
  - 缓解措施：保留 prompt.go 作为兜底，渐进式迁移

### 缓解策略

1. 懒加载 Skill 内容，只在需要时读取文件
2. LLM 选择失败时降级到 base skill
3. 保留现有 prompt.go，新系统作为增强而非替换
4. 实现前先完成 PoC 验证思路

## 为什么

当前 MSA 的 Prompt 模板硬编码在 `pkg/logic/agent/prompt.go` 中，包含超过 100 行的单一大模板。这带来以下问题：

1. **不灵活**：修改 Prompt 需要重新编译部署
2. **不可扩展**：无法动态添加新的分析能力或策略
3. **难以维护**：所有规则混在一起，职责不清晰
4. **用户体验受限**：无法根据不同场景选择不同的能力

通过引入 Skill 系统，可以实现：
- 动态加载和管理 Prompt 模板
- LLM 自动选择合适的 Skill
- 支持用户自定义 Skill
- 更清晰的职责分离

## 变更内容

1. **新增 Skill 管理系统**：`pkg/logic/skills/` 包
   - `skill.go`：Skill 结构定义
   - `loader.go`：扫描和加载 Skill 文件
   - `registry.go`：Skill 注册表
   - `selector.go`：LLM 自动选择器
   - `builder.go`：Prompt 构建器
   - `manager.go`：对外管理 API
   - `config.go`：配置管理

2. **内置 Skills**：`pkg/logic/skills/plugs/`
   - `base/`：核心系统规则（工具调用、时间敏感性等）
   - `stock-analysis/`：股票分析框架
   - `account-management/`：账户管理流程
   - `output-formats/`：输出模板（投资建议、报告格式）

3. **用户自定义目录**：`~/.msa/skills/`
   - 支持用户创建自己的 Skill

4. **配置扩展**：`~/.msa/config.json`
   - 新增 `disable_skills` 字段

5. **修改 chat.go**：集成 Skill 选择流程

## 能力

### 新增能力

- `skill-management`：Skill 管理能力
  - 扫描和加载 Skill 文件（内置 + 用户目录）
  - 懒加载 Skill 内容
  - Skill 注册和索引

- `skill-selector`：LLM 自动选择能力
  - 根据用户输入自动选择合适的 Skills
  - 使用模板变量构建选择 Prompt
  - 失败降级到 base skill

- `dynamic-prompts`：动态 Prompt 构建
  - 合并多个 Skills 内容
  - 支持 Priority 优先级排序
  - 支持用户手动覆盖

- `skill-cli`：Skill 管理 CLI 命令
  - `msa skills list`：列出所有 Skills
  - `msa skills show <name>`：查看 Skill 详情
  - `msa skills enable/disable <name>`：启用/禁用 Skill

### 修改能力

- `agent-chat`：现有对话能力扩展
  - 集成 Skill 选择流程
  - 保持向后兼容

## 影响

### 代码变更

- **新增**：`pkg/logic/skills/` 整个目录
- **修改**：`pkg/logic/agent/chat.go`（集成 Skill 选择）
- **保留**：`pkg/logic/agent/prompt.go`（作为兜底）

### 数据变更

- 配置文件新增字段：`disable_skills: []`

### 依赖变更

无新增外部依赖

### API 变更

- `GetDefaultTemplate` 函数签名增加 `selectedSkills` 参数
- 新增 `skills.GetManager()` 对外 API

### 向后兼容性

✅ 完全兼容：现有功能保持不变，新系统作为增强

---

**知识备注**：
- 此提案使用 `openspec/knowledge/index.yaml` 中的知识生成
- 最后知识同步：2026-03-08
- 知识库总问题数：5
- 总模式数：6
