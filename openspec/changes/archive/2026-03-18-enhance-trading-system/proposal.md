# 提案：enhance-trading-system

增强交易系统的核心策略和经验沉淀机制，解决分段操作、状态传递和错误记录三大问题。

## 知识上下文

### 相关历史问题

知识库中未找到与交易系统直接相关的问题。这是新领域，如遇问题请创建 fix.md 来构建知识库。

### 需避免的反模式

未识别到特定反模式。

### 推荐模式

- **Pattern #003**: 错误处理和重试策略 - 适用于 analyze_today_trades 工具的错误分析逻辑
- **Pattern #005**: 安全优先配置原则 - 适用于仓位管理策略的保守设计

---

## 为什么

当前交易系统存在三个核心问题：

1. **午盘无法读取早盘操作**：Session 标签依赖模型手动调用 `add_session_tag`，经常被遗忘，导致午盘决策缺少早盘上下文。

2. **收盘无错误记录和反思**：SKILL.md 定义了错误记录流程，但缺少强制机制和辅助工具，模型经常忽略这一步，导致错误无法沉淀和复用。

3. **首次买入直接 ALL IN**：缺少仓位管理策略，买入逻辑是"最大可买数量"，风险过高。需要分段买入机制。

---

## 变更内容

### 1. Session 标签自动添加

在 `pkg/logic/memory/manager.go` 的 `Initialize()` 方法中，根据当前时间自动添加标签：
- 9:30-11:30 → `morning-session`
- 13:00-14:30 → `afternoon-session`
- 16:00+ → `close-session`

### 2. 分段买入/卖出策略

在 `trading-common/SKILL.md` 中新增：
- 计划仓位计算规则（总资产 30% 上限）
- 三种分段买入方式（时间分段、价格信号分段、固定比例分段）
- 分段卖出规则（正常分批、亏损超 20% 可一次性止损）
- 买入限制（单次不超过 50%、涨幅超 50% 不追高）

### 3. Error 记录强化

在 `market-close-summary/SKILL.md` 中新增：
- 明确的错误判断标准（追高买入、未分段、未止损、情绪化交易等）
- 错误检查清单

### 4. 新增 analyze_today_trades 工具

新增知识库工具，辅助模型识别今日交易中的潜在错误，输出建议记录的 error 列表。

---

## 能力

### 新增能力

- `position-management`: 仓位管理策略，包含分段买入/卖出规则和限制参数
- `trade-analysis`: 交易分析工具，自动识别潜在错误操作

### 修改能力

- `session-tagging`: Session 标签从手动调用改为自动添加

---

## 影响

### 代码变更

| 文件 | 变更类型 |
|------|----------|
| `pkg/logic/memory/manager.go` | 修改 - 自动添加 Session 标签 |
| `pkg/logic/tools/knowledge/analyze.go` | 新增 - analyze_today_trades 工具 |

### SKILL 变更

| 文件 | 变更类型 |
|------|----------|
| `trading-common/SKILL.md` | 修改 - 新增分段策略 |
| `morning-analysis/SKILL.md` | 修改 - 移除手动标签说明 |
| `afternoon-trade/SKILL.md` | 修改 - 移除手动标签说明，强化早盘读取 |
| `market-close-summary/SKILL.md` | 修改 - 强化 error 记录要求 |

### 向后兼容性

- Session 标签自动添加不影响现有逻辑，只是补充
- 新增工具不影响现有工具
- SKILL 变更是增强，不破坏现有流程

---

**知识备注**：
- 此提案使用 `openspec/knowledge/index.yaml` 中的知识生成
- 最后知识同步：2026-02-23
- 知识库总问题数：5
- 总模式数：6
