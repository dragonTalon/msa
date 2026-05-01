---
name: market-close-summary
description: 闭市总结，分析今日收益，复盘操作，沉淀经验。触发时间 16:00 后。
version: 2.2.0
priority: 8
pattern: pipeline
steps: 5
requires_todo: true
triggers:
  - time: "16:00+"
    session: close-session
tools:
  - get_account_summary
  - get_positions
  - get_transactions
  - get_stock_quote
  - query_sessions_by_date
  - read_knowledge
  - write_knowledge
  - analyze_today_trades
dependencies:
  - trading-common
  - knowledge-read
  - knowledge-write
---

# Market Close Summary - 闭市总结

## 基本信息

| 属性 | 值 |
|------|-----|
| 触发时间 | 16:00 后（收盘后） |
| 主要任务 | 收益分析、操作复盘、经验沉淀、推理链审查 |

---

## ⚠️ 强制前置

**第一步 MUST 调用 `read_knowledge` 工具！**

```
不执行此步骤将无法正确评估今日操作。
```

---

## Pipeline 执行流程

**DO NOT 跳过任何步骤。按顺序执行。**

---

### Step 1: 信息收集【必须执行】

#### 1.1 读取历史错误【强制】
```
→ 调用 read_knowledge(type="all")
→ 获取历史错误和昨日总结
→ 了解历史错误，用于评估今日操作
```

#### 1.2 查询今日所有 Session
```
→ 调用 query_sessions_by_date(date="today")
→ 获取今日所有 Session：
  - morning-session → 早盘会话
  - afternoon-session → 午盘会话

提取：
- 早盘的决策和执行结果
- 午盘的决策和执行结果
- 早盘预测 vs 实际走势
- 午盘预测 vs 实际走势
- 【新增】每个决策的推理链（decision-log）
```

#### 1.3 获取今日交易记录
```
→ 调用 get_transactions(今日)
→ 统计：
  - 买入次数、总金额
  - 卖出次数、总金额
  - 手续费总计
  - 每笔交易详情
```

#### 1.4 获取今日收盘状态
```
→ 调用 get_account_summary → 总资产
→ 调用 get_positions → 持仓详情
→ 调用 get_stock_quote 获取每只持仓的收盘价
→ 计算今日盈亏
```

---

### Step 2: 收益分析

**加载 `references/review-criteria.md` 进行评估**

#### 计算今日收益
```
今日收益 = 当前总资产 - 昨日总资产
今日收益率 = 今日收益 / 昨日总资产 × 100%

个股盈亏 = (当前价 - 成本价) × 持仓数量
个股收益率 = (当前价 - 成本价) / 成本价 × 100%
```

#### 输出收益报表
使用 `assets/summary-template.md` 生成报告。

---

### Step 3: 操作复盘

---

#### 3.1 推理链审查【新增】

**加载 `references/reasoning-review-criteria.md`**

对今日每个决策的推理链进行质量审查：

```
1. 从 Session 中提取所有 decision-log
2. 逐条检查推理链完整性（6段是否齐全）
3. 按 reasoning-review-criteria.md 给出 A-E 评级

评级标准:
  A 级: 6段完整, 触发准确, 理由充分, 反向检查严格
  B 级: 6段完整, 主要理由充分, 细节可加强
  C 级: 基本完整, 缺 1-2 段, 需改进
  D 级: 严重不完整, 缺 ≥3 段, 视为情绪化决策
  E 级: 完全无推理链, 严重错误

4. 区分"正确决策运气不好"vs"错误决策运气好"
   → 见 reasoning-review-criteria.md 判定矩阵
```

**推理审查输出：**
```markdown
## 推理质量审查

| 决策ID | 类型 | 推理等级 | 结果 | 判定 |
|--------|------|----------|------|------|
| dec-001 | buy | B | +3% | ✅ 正确决策 |
| dec-002 | sell | A | -1% | ✅ 正确决策, 坏运气 |
| ... | ... | ... | ... | ... |

### 推理质量分布
A:[n] B:[n] C:[n] D:[n] E:[n]

### ⚠️ 需要警惕的"好运气"
[D/E 级但盈利的决策列表]

### ✅ 需要保留的"坏运气"
[A/B 级但亏损的决策列表]
```

---

#### 3.2 对比分析

**早盘预测 vs 实际走势：**
```
预测：[早盘预测内容]
实际：[实际走势]
判断：正确 ✓ / 错误 ✗
推理质量: [A-E]
```

**午盘预测 vs 实际走势：**
```
预测：[午盘预测内容]
实际：[实际走势]
判断：正确 ✓ / 错误 ✗
推理质量: [A-E]
```

---

#### 3.3 操作评估（新增推理质量维度）

**正确操作：**
- 操作描述
- 推理链等级: [A/B]
- 为什么正确
- 可复用经验

**待改进操作：**
- 操作描述
- 推理链等级: [B/C]
- 改进方向

**错误操作（如有）：**
- 错误描述
- 推理链等级: [D/E]
- 原因分析
- 正确做法
- 需记录到错误库

---

### Step 4: 错误检查【强制执行】

**加载 `references/error-checklist.md` 逐项检查**

#### 错误检查清单

```
□ 今日是否有追高买入？（涨幅 > 追高阈值时买入）
□ 今日是否一次性买入超过计划仓位的 50%？
□ 今日是否有持仓亏损超过 [动态止损线] 未止损？
□ 今日是否有情绪化决策？
□ 今日操作是否符合用户策略？
□ 【新增】今日是否有决策缺少推理链？（D/E 级 = 情绪化决策）
□ 【新增】今日是否有 D/E 级推理但结果盈利的决策？（最危险）
```

#### 如果有任何一项为"是"【必须执行】

1. **必须** 调用 `analyze_today_trades` 工具分析
2. **必须** 使用 `write_knowledge(type="error", content=错误记录)` 写入错误
3. **必须** 检查返回的 `verified` 字段是否为 true

---

### Step 5: 生成总结文件【强制执行】

**⚠️ GATE: 必须生成总结文件才算完成**

#### 必须生成的文件

**必须调用**: `write_knowledge(type="summary", date="YYYY-MM-DD", content=总结内容)`

**重要**：调用后必须检查返回的 `verified` 字段：
- 如果 `verified: true`，写入成功
- 如果 `verified: false`，必须重试或报告错误

#### 总结文件格式

使用 `assets/summary-template.md` 的格式。总结中必须包含：
- 推理质量审查结果
- 正确决策/错误决策的判定
- 经验沉淀

---

## 经验沉淀

根据复盘结果，使用 `write_knowledge` 工具写入：

| 类型 | 调用方式 | 条件 |
|------|----------|------|
| 错误操作 | write_knowledge(type="error", content=错误记录) | 有错误操作时 |
| 每日总结 | write_knowledge(type="summary", date="YYYY-MM-DD", content=总结) | 每日收盘后 |
| 推理质量问题 | write_knowledge(type="error", content=推理错误) | D/E 级推理存在时 |

---

## 约束条件

| 约束 | 说明 |
|------|------|
| 必须查询 Session | 查询今日所有 Session |
| 必须执行错误检查 | 如有错误必须调用 write_knowledge 记录 |
| 必须写入总结 | 使用 write_knowledge(type="summary") |
| 必须先调用 read_knowledge | 历史错误指导评估 |
| 必须检查 verified 字段 | 写入后验证 verified=true |
| 必须审查推理链 | 每个决策的推理链给出 A-E 评级 |
| 必须区分决策与运气 | 使用判定矩阵区分正确决策与好运气 |

---

## 执行检查表

完成以下所有步骤才算执行成功：

```
□ 已调用 read_knowledge(type="all")
□ 已查询今日所有 Session
□ 已获取今日交易记录
□ 已计算今日收益
□ 已执行推理链审查（每个决策 A-E 评级）
□ 已区分"正确决策运气不好"vs"错误决策运气好"
□ 已执行错误检查清单（含推理链完整性检查）
□ 如有错误，已调用 write_knowledge(type="error") 并验证成功
□ 已调用 write_knowledge(type="summary") 并验证成功
□ 已输出完整的交易日总结（含推理质量审查结果）
```
