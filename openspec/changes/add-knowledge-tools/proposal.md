# 提案：add-knowledge-tools

## 知识上下文

### 相关历史问题

知识库中未找到直接相关问题。

### 需避免的反模式

未识别到特定反模式。

### 推荐模式

✅ **推荐**：考虑使用以下模式：

- **003**：错误处理和重试策略
  - 描述：分类错误处理，根据错误类型采取不同的处理策略
  - 使用时机：文件写入需要验证成功，失败时重试

## 风险评估

基于知识库分析：

- **中**：文件写入失败可能导致知识丢失
  - 相关模式：003
  - 缓解措施：写入验证 + 自动重试（最多 3 次）

## 缓解策略

1. 写入后立即验证：重新读取文件确认内容完整性
2. 失败自动重试：最多 3 次，指数退避
3. 返回验证状态：`verified: true/false` 明确告知调用者

## 为什么

当前知识库系统存在以下问题：

1. **结构过于复杂**：原设计包含 `errors/`、`summaries/`、`strategies/`、`insights/`、`notes/` 等多个目录，维护成本高
2. **知识库工具已删除**：`pkg/logic/tools/knowledge/` 目录不存在
3. **SKILL 依赖断裂**：`morning-analysis`、`afternoon-trade`、`market-close-summary` 依赖 `read-error` SKILL，但该 SKILL 设计不够灵活

需要一个简单的知识库系统：
- 存储历史错误（避免重复犯错）
- 存储每日总结（复盘参考）
- 提供可靠的读写工具

## 变更内容

1. **新增工具**：
   - `read_knowledge`：读取历史错误 + 昨日总结
   - `write_knowledge`：写入错误记录 / 每日总结（带验证和重试）

2. **简化文件结构**：
   - `~/.msa/errors.md`：历史错误记录（追加写入）
   - `~/.msa/summaries/YYYY-MM-DD.md`：每日总结

3. **删除旧组件**：
   - 删除 `pkg/logic/skills/plugs/read-error/` SKILL

4. **更新依赖 SKILL**：
   - `morning-analysis`：改用 `read_knowledge`
   - `afternoon-trade`：改用 `read_knowledge`
   - `market-close-summary`：改用 `read_knowledge` + `write_knowledge`

## 能力

### 新增能力

- `knowledge-read`：读取交易知识库（历史错误 + 昨日总结）
- `knowledge-write`：写入交易知识库（错误记录 / 每日总结），带验证和重试

### 修改能力

无。这是新增能力，不修改现有能力。

## 影响

- **新增代码**：`pkg/logic/tools/knowledge/` 目录
- **删除代码**：`pkg/logic/skills/plugs/read-error/` 目录
- **修改 SKILL**：`morning-analysis`、`afternoon-trade`、`market-close-summary` 的 SKILL.md
- **用户数据**：`~/.msa/` 下新增 `errors.md` 和 `summaries/` 目录

## 非目标

- 不实现复杂的知识分类（策略、洞察等）
- 不使用数据库存储
- 不支持知识搜索或查询
- 不实现知识的自动清理或归档

---

**知识备注**：
- 此提案使用 `openspec/knowledge/index.yaml` 中的知识生成
- 知识库总问题数：5
- 总模式数：6
