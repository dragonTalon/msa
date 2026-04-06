---
name: morning-analysis
description: 早盘开市操作，执行以卖出为主的策略。触发时间 9:30-11:30，必须实际执行买卖操作。
version: 2.0.0
priority: 8
pattern: pipeline
steps: 4
requires_todo: true
triggers:
  - time: "9:30-11:30"
    session: morning-session
tools:
  - get_account_summary
  - get_positions
  - get_transactions
  - get_stock_quote
  - submit_buy_order
  - submit_sell_order
  - web_search
  - fetch_page_content
dependencies:
  - trading-common
  - read-error
  - output-formats
---

# Morning Analysis - 早盘开市操作

## 基本信息

| 属性 | 值 |
|------|-----|
| 触发时间 | 9:30 - 11:30 |
| 操作倾向 | **以卖出为主，买入为辅** |
| 执行方式 | **必须实际执行交易** |

---

## ⚠️ 强制前置

**第一步 MUST 调用 `read-error` SKILL！**

```
在执行任何操作前，必须先读取历史错误经验。
不执行此步骤将导致重复犯错。
```

---

## Pipeline 执行流程

**DO NOT 跳过任何步骤。按顺序执行。**

---

### Step 1: 信息收集【必须执行】

#### 1.1 读取历史错误【强制】
```
→ 调用 read-error SKILL
→ 将错误经验拼接到后续分析的 prompt 中
```

#### 1.2 读取用户知识库
```
→ 读取 .msa/knowledge/notes/YYYY-MM-DD.md（昨日笔记）
  - 如果不存在，记录"无昨日笔记"
  - 提取：关注股票、市场观点、操作计划

→ 读取 .msa/knowledge/user-profile/
  - 了解用户风险偏好、投资风格、关注领域

→ 读取 .msa/knowledge/strategies/
  - 获取用户的交易策略规则

→ 读取 .msa/knowledge/summaries/ 最近一个工作日的文件
  - 获取上一交易日的复盘信息
  - 获取"明日关注"要点
```

#### 1.3 获取账户状态
```
→ 调用 get_account_summary → 总资产、可用余额、持仓市值
→ 调用 get_positions → 当前持仓详情
→ 调用 get_transactions(limit=10) → 最近交易记录
```

---

### Step 2: 分析决策

**加载 `references/sell-conditions.md` 进行卖出判断**
**加载 `references/buy-conditions.md` 进行买入判断**

#### 卖出决策（主要）

对每只持仓分析是否需要卖出：

| 检查项 | 判断标准 |
|--------|----------|
| 止盈位 | 是否触及？盈利是否超过目标？ |
| 技术面 | 是否走弱？跌破支撑？量能萎缩？ |
| 资金流向 | 是否有资金撤离迹象？ |
| 负面消息 | 是否有利空消息？ |
| 策略规则 | 是否符合用户策略的卖出规则？ |

**卖出决策格式：**
```markdown
卖出决策：
- 持仓代码: [code]
- 持仓名称: [name]
- 当前持仓: [qty]股
- 建议卖出: [qty]股（减半仓/清仓）
- 卖出原因: [原因]
```

#### 买入决策（次要）

| 检查项 | 判断标准 |
|--------|----------|
| 资金情况 | 可用资金是否充足？ |
| 优质标的 | 是否有参考昨日笔记、策略的标的？ |
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

---

### Step 3: 执行交易【强制执行】

**重要：先输出决策，再自动执行，不需要用户确认**

**⚠️ GATE: 执行前必须确认：**
- 卖出数量 <= 持仓数量
- 买入金额 <= 可用余额

#### 执行卖出

```
FOR each 卖出决策:
  1. 调用 get_stock_quote(stock_code) 获取当前价格
  2. 输出决策：
     "【卖出执行】准备卖出 [名称]([代码]) [数量]股，当前价 [价格]元"
  3. 调用 submit_sell_order:
     - stock_code, stock_name, quantity, price, fee: 5.0
  4. 输出执行结果
  5. 记录到当前 Session
```

#### 执行买入

```
FOR each 买入决策:
  1. 调用 get_account_summary 确认可用余额
  2. 调用 get_stock_quote(stock_code) 获取当前价格
  3. 计算最大可买数量（整手）:
     max_qty = floor(可用余额 / (价格 × 100)) × 100
  4. 检查金额是否充足
  5. 输出决策：
     "【买入执行】准备买入 [名称]([代码]) [数量]股，当前价 [价格]元"
  6. 调用 submit_buy_order:
     - stock_code, stock_name, quantity, price, fee: 5.0
  7. 输出执行结果
  8. 记录到当前 Session
```

---

### Step 4: 输出与记录

**加载 `assets/report-template.md` 生成报告**

#### 输出格式

使用 `output-formats` SKILL 的模板生成结构化报告。

#### 记录到 Session

**注意**：Session 标签 `morning-session` 由系统自动添加，无需手动操作。

记录关键决策和执行结果，供午盘和闭市会话读取。

---

## 约束条件

| 约束 | 说明 |
|------|------|
| 禁止透支 | 可用余额不足时必须放弃买入 |
| 禁止卖空 | 卖出数量不能超过持仓数量 |
| 整手交易 | 买入数量必须是 100 的整数倍 |
| 执行不确认 | 输出决策后直接执行，不需要用户确认 |
| 错误必读 | 必须先执行 read-error SKILL |
| 分段买入 | 首次买入不超过计划仓位的 50% |
