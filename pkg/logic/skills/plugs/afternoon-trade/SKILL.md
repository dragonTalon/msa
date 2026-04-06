---
name: afternoon-trade
description: 午盘交易操作，执行以买入为主的策略。触发时间 13:00-14:30，必须实际执行买卖操作。
version: 2.1.0
priority: 8
pattern: pipeline
steps: 4
requires_todo: true
triggers:
  - time: "13:00-14:30"
    session: afternoon-session
tools:
  - get_account_summary
  - get_positions
  - get_transactions
  - get_stock_quote
  - submit_buy_order
  - submit_sell_order
  - web_search
  - fetch_page_content
  - query_sessions_by_date
  - read_knowledge
dependencies:
  - trading-common
  - knowledge-read
  - output-formats
---

# Afternoon Trade - 午盘操作

## 基本信息

| 属性 | 值 |
|------|-----|
| 触发时间 | 13:00 - 14:30 |
| 操作倾向 | **以买入为主，卖出为辅** |
| 执行方式 | **必须实际执行交易** |

---

## ⚠️ 强制前置

**第一步 MUST 调用 `read_knowledge` 工具！**

```
在执行任何操作前，必须先读取历史错误经验。
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
→ 将错误经验拼接到后续分析的 prompt 中
```

#### 1.2 搜索午间新闻
```
→ 调用 web_search("今日午间财经新闻 A股")
→ 选取 1-5 个最相关链接
→ 调用 fetch_page_content 获取完整内容
→ 分析新闻对市场的影响
→ 识别可能的利好/利空板块
```

#### 1.3 读取早盘 Session
```
→ 调用 query_sessions_by_date(date="today", tag="morning-session")
→ 提取早盘的：
  - 操作决策和执行结果
  - 预测和建议
  - 关注要点
→ 对比早盘预测与上午实际走势
```

#### 1.4 读取用户知识库
```
→ 读取 .msa/knowledge/user-profile/
→ 读取 .msa/knowledge/strategies/
→ 读取 .msa/knowledge/insights/
```

#### 1.5 获取当前账户状态
```
→ 调用 get_account_summary → 可用余额
→ 调用 get_positions → 当前持仓（早盘操作后的最新状态）
→ 调用 get_transactions(今日) → 今日已执行交易
```

---

### Step 2: 分析决策

**加载 `references/buy-conditions.md` 进行买入判断**

#### 早盘回顾与验证

```
对比早盘预测与实际走势：
- 早盘判断是否正确？
- 建议的操作是否已执行？
- 市场走势是否符合预期？
- 是否需要调整策略？
```

#### 买入决策（主要）

| 检查项 | 判断标准 |
|--------|----------|
| 资金情况 | 可用资金是否充足？ |
| 早盘标的 | 早盘关注的股票是否出现买入机会？ |
| 午间新闻 | 是否有利好消息？ |
| 技术信号 | 是否出现买入信号？ |
| 策略规则 | 是否符合用户策略的买入规则？ |

**买入决策格式：**
```markdown
买入决策：
- 目标代码: [code]
- 目标名称: [name]
- 建议买入: [qty]股
- 买入原因: [原因]
```

#### 卖出决策（次要）

如果早盘未执行的卖出，评估是否仍需执行：
- 卖出理由是否仍然成立？
- 是否有新的卖出理由？
- 当前价格是否更优？

---

### Step 3: 执行交易【强制执行】

**重要：先输出决策，再自动执行，不需要用户确认**

**⚠️ GATE: 执行前必须确认：**
- 买入金额 <= 可用余额
- 时间 < 14:30（14:30 后不建议新开仓）

#### 执行买入（主要）

```
FOR each 买入决策:
  1. 调用 get_account_summary 确认可用余额
  2. 调用 get_stock_quote(stock_code) 获取当前价格
  3. 计算最大可买数量（整手）
  4. 检查金额是否充足
  5. 输出决策：
     "【买入执行】准备买入 [名称]([代码]) [数量]股，当前价 [价格]元"
  6. 调用 submit_buy_order 执行
  7. 输出执行结果
  8. 记录到当前 Session
```

#### 执行卖出（如有）

```
FOR each 卖出决策:
  1. 调用 get_positions 确认持仓
  2. 调用 get_stock_quote 获取价格
  3. 确认卖出数量 <= 持仓数量
  4. 输出决策
  5. 调用 submit_sell_order 执行
  6. 输出执行结果
  7. 记录到当前 Session
```

---

### Step 4: 输出与记录

**加载 `assets/report-template.md` 生成报告**

#### 记录到 Session

**注意**：Session 标签 `afternoon-session` 由系统自动添加。

记录关键决策和执行结果，供闭市会话读取。

---

## 约束条件

| 约束 | 说明 |
|------|------|
| 禁止透支 | 金额不足必须放弃买入 |
| 时间敏感 | 14:30 后不建议新开仓 |
| 参考早盘 | 必须读取早盘 Session 的决策 |
| 执行不确认 | 直接执行，不需要确认 |
| 错误必读 | 必须先调用 read_knowledge 工具 |
| 分段买入 | 首次买入不超过计划仓位的 50% |
