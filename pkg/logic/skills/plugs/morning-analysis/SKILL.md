---
name: morning-analysis
description: 早盘开市操作，执行以卖出为主的策略。触发时间 9:30-11:30，必须实际执行买卖操作。
version: 2.2.0
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
  - read_knowledge
dependencies:
  - trading-common
  - knowledge-read
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

**第一步 MUST 调用 `read_knowledge` 工具！**

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
→ 调用 read_knowledge(type="all")
→ 获取历史错误和昨日总结
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

决策流程新增 3 层过滤/增强:
```
web_search → 信息质量门 → 市场状态分类 → 推理链 → 决策
```

---

#### 2.1 [NEW] 信息质量门

**加载以下参考文件：**
- `trading-common/references/news-quality-gate.md`
- `trading-common/references/search-query-tactics.md`
- `references/news-caution-rules.md`

**执行步骤：**
```
1. 使用多角度搜索策略（3组并行查询）获取信息
   → 加载 search-query-tactics.md
   → 基本面 + 技术面 + 反向/风险 方向各一组查询

2. 对搜索获取的信息执行 4 步过滤
   → 加载 news-quality-gate.md
   → Step 1: 来源分类 (S/A/B/C/X)
   → Step 2: 时效检查 (>60min 可能已定价)
   → Step 3: 拥挤度检查 (板块涨 >3%=crowded)
   → Step 4: 完整性检查 (≥3条 + 反向观点)

3. 早盘新闻特殊处理
   → 加载 news-caution-rules.md
   → 隔夜新闻衰减检查
   → 开盘跳空检查 (>2%=已定价)
   → 早盘情绪检查

4. 输出信息质量摘要
```

---

#### 2.2 [NEW] 市场状态分类

**加载以下参考文件：**
- `trading-common/references/market-regime-classifier.md`
- `trading-common/references/market-regime-decision-logic.md`

**执行步骤：**
```
1. web_search 获取大盘数据
   → 大盘指数 vs MA20
   → 上涨/下跌家数
   → 成交量 vs 20日均量
   → 北向资金流向
   → VIX/波动率数据

2. 3 维度评分
   → 趋势 (Trend): 50% 权重
   → 波动率 (Volatility): 20% 权重
   → 情绪 (Sentiment): 30% 权重

3. 确定市场状态 (8种之一)
   → 输出综合得分和状态标签

4. 获取该状态下的动态参数
   → 仓位上限、追高阈值、止盈目标、止损线
   → 推荐的分段策略
```

---

#### 2.3 卖出决策（主要）

**加载 `references/sell-conditions.md` 进行卖出判断**

对每只持仓分析是否需要卖出：

| 检查项 | 判断标准 |
|--------|----------|
| 止盈位 | 是否触及？盈利是否超过目标？(使用动态止盈目标) |
| 技术面 | 是否走弱？跌破支撑？量能萎缩？ |
| 资金流向 | 是否有资金撤离迹象？ |
| 负面消息 | 是否有利空消息？ |
| 策略规则 | 是否符合用户策略的卖出规则？ |

---

#### 2.4 买入决策（次要）

**加载 `references/buy-conditions.md` 进行买入判断（现在引用动态参数）**

| 检查项 | 判断标准 |
|--------|----------|
| 资金情况 | 可用资金是否充足？ |
| 优质标的 | 是否有参考昨日笔记、策略的标的？ |
| 技术信号 | 是否出现买入信号？(信号数量要求视市场状态而定) |
| 策略规则 | 是否符合用户策略的买入规则？ |
| 信息质量 | 是否通过信息质量门？ |
| 市场状态 | 当前市场状态是否支持买入？ |

---

#### 2.5 [NEW] 推理链输出

**加载以下参考文件：**
- `trading-common/references/decision-reasoning-template.md`
- `trading-common/references/decision-log-format.md`

**每个决策必须用 6 段推理链输出：**

```
1. 触发条件评估 → 是什么触发了这个决策
2. 理由详述     → 支持和反对的全部理由
3. 反向因素     → 强制反向检查，对抗确认偏误
4. 信息质量     → 引用 2.1 质量门的输出
5. 市场状态     → 引用 2.2 分类器的输出
6. 决策输出     → 明确代码、数量、价格、止损止盈
```

**决策日志写入 Session：**
```
每个决策以 JSON 格式写入 Session:
  key: "decision-log"
  格式: 见 decision-log-format.md
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
| 错误必读 | 必须先调用 read_knowledge 工具 |
| 分段买入 | 首次买入不超过计划仓位的 50% |
| 信息质量门 | 所有搜索信息必须通过 4 步过滤 |
| 市场状态分类 | 每次决策前必须确定市场状态 |
| 推理链必输出 | 每个决策必须有完整 6 段推理链 |
