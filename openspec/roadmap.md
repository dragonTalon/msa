# MSA 量化交易系统优化路线图

> 创建日期: 2026-05-01
> 基于对全部技能、提示、参考文件的深度分析

---

## 问题等级定义

| 等级 | 含义 | 标准 |
|------|------|------|
| **CRITICAL** | 导致系统性亏损的核心缺陷 | 必须修复才能稳定运行 |
| **HIGH** | 显著影响策略质量或风险控制 | 修复后有明显收益提升 |
| **MEDIUM** | 影响系统完善度和可靠性 | 不直接导致亏损 |
| **LOW** | 锦上添花的长期优化 | 可排到后续迭代 |

---

## 问题全景

### CRITICAL (3项)

| # | 问题 | 现象 | 影响面 |
|---|------|------|--------|
| C1 | 搜索结果耦合/羊群效应 | web_search 返回"热门"结果 → AI追热点 → 高位接盘 | 所有买入决策 |
| C2 | 缺少市场状态分类器 | 牛熊震荡用同一套规则 | 全部策略基准失效 |
| C3 | 决策黑箱无推理链 | 只有结论无推导过程，复盘无法追溯决策逻辑 | 错误无法定位 |

### HIGH (5项)

| # | 问题 | 现象 | 影响面 |
|---|------|------|--------|
| H1 | 操作类型过于单一 | 只有买/卖，无观望/试探/调仓/分批 | 5种日内走势覆盖不全 |
| H2 | 止损机制粗糙 | "亏损20%清仓"，无ATR动态止损、移动止盈 | 止损过宽或过窄 |
| H3 | 缺少置信度标注 | 低把握当高把握执行 | 小信号大仓位 |
| H4 | 缺少持仓相关性分析 | 3只白酒股合规但合计暴露同行业 | 隐性集中度风险 |
| H5 | 缺少盘前准备 | 9:30直接操作，无集合竞价/隔夜外盘分析 | 开盘操作盲目 |

### MEDIUM (4项)

| # | 问题 | 现象 | 影响面 |
|---|------|------|--------|
| M1 | 技术指标不足 | 缺 VWAP、ATR、OBV、多时间周期框架 | 分析精度 |
| M2 | 盘中监控缺失 | 只在早午盘检查，盘中异常无触发 | 极端行情损失 |
| M3 | 执行缺少前置检查 | 买卖前未问"现在适合交易吗" | 不当交易 |
| M4 | 错误记录滞后 | 追高收盘才发现，当天无法纠正 | 错误无法实时规避 |

### LOW (3项)

| # | 问题 | 现象 | 影响面 |
|---|------|------|--------|
| L1 | 缺少策略回测 | 无法验证策略参数的历史表现 | 策略优化靠猜测 |
| L2 | 缺少绩效归因 | 不知道盈亏来源(Alpha/Beta/行业) | 无法定向优化 |
| L3 | 缺少多策略对比 | 只有一套策略，无备选 | 策略单一化风险 |

---

## 排期总览

```
Phase 1 (本周): 基础能力补全
    CRITICAL ×3 — 不改策略逻辑，增加推理和过滤能力

Phase 2 (下周): 策略体系升级
    HIGH ×5 — 扩展操作类型，重构决策框架

Phase 3 (后续): 风险管理增强
    MEDIUM ×4 — 相关性、动态止损、盘中监控

Phase 4 (远期): 量化体系建立
    LOW ×3 — 回测、绩效归因、多策略
```

---

## Phase 1: 基础能力补全

**目标**: 不改动现有买/卖规则，给系统加上推理和过滤能力

**改动量**: ~7个文件 (新增3个, 修改4个)

### C3 — 推理链模板

**新增**: `pkg/logic/skills/plugs/trading-common/references/reasoning-chain.md`

每个交易决策必须填写的推理模板：

```markdown
## 决策推理链

### 观察
[基于什么数据/信号得出结论]

### 假设
[基于什么逻辑推导，市场会怎么走]

### 备选方案
[考虑过但放弃的方案及原因]

### 验证条件
[什么信号会验证此假设正确? 什么信号会否定?]

### 最坏情况
[如果判断错误，最大损失 = ? 是否有止损保护?]

### 退出计划
[什么条件下退出这笔交易? 时间/价格/事件]

### 置信度
[高(>80%) / 中(60-80%) / 低(<60%)]
```

**修改**:
- `morning-analysis/SKILL.md` — Step 2 决策时加载并填写 reasoning-chain
- `afternoon-trade/SKILL.md` — Step 2 同理
- `stock-analysis/SKILL.md` — 分析输出中包含推理链

### C2 — 市场状态分类器

**新增**: `pkg/logic/skills/plugs/trading-common/references/market-regime.md`

```markdown
## 市场状态判断规则

### 数据来源
- 上证指数/沪深300 日K线、均线系统
- 涨跌停家数 (市场广度)
- 成交量变化趋势

### 四种状态
| 状态 | 技术特征 | 仓位上限 | 策略偏好 | 操作频率 |
|------|----------|----------|----------|----------|
| 牛市 | 大盘>年线, 均线多头, 新高>新低 | 80% | 持有为主 | 低 |
| 熊市 | 大盘<年线, 均线空头, 新低>新高 | 30% | 现金为主 | 极低 |
| 震荡 | 均线缠绕, 布林带收窄 | 50% | 波段操作 | 中 |
| 极端 | 单日涨跌>5% | 10% | 暂停交易 | 停止 |

### 判断流程
1. 获取上证指数/沪深300日K线 (近60日)
2. 计算: 均线排列、新高新低比例、布林带宽度
3. 结合成交量确认
```

**修改**:
- `morning-analysis/SKILL.md` Step 2 开头增加: "先判断市场状态"
- `afternoon-trade/SKILL.md` Step 2 开头同理

### C1 — 信息去耦合

**新增**: `pkg/logic/skills/plugs/stock-analysis/references/information-quality.md`

核心机制:

```markdown
## 信息时效性分层
| 时效 | 定义 | 交易价值 |
|------|------|----------|
| 实时 | <1小时 | 有交易价值，可能未完全定价 |
| 当日 | 1-6小时 | 谨慎参考，部分已定价 |
| 日内 | 6-24小时 | 仅作背景，大概率已定价 |
| 历史 | >24小时 | 无短线交易价值 |

## 来源可信度分层
| 等级 | 来源 | 用途 |
|------|------|------|
| S级 | 交易所公告、证监会文件 | 直接作为决策依据 |
| A级 | 权威财经(财新、证券时报) | 交叉验证后使用 |
| B级 | 一般财经、券商研报 | 仅作参考背景 |
| C级 | 自媒体、论坛、股吧 | 仅作情绪参考, 不作为买卖依据 |

## 反身性检查 (必做)
1. 这个新闻的发布时间? (>24小时则已定价)
2. 价格在新闻后已经涨了多少? (>3%则已被定价)
3. 这个新闻90%的散户都能看到吗? (如果是, 大概率已定价)
4. 这是新信息, 还是旧信息被重新传播?

## 交叉验证要求
- 任何交易信号必须至少有 3 个独立来源支持
- 单一来源的"独家消息"视为噪音, 忽略
- 如果多个来源矛盾, 默认观望
```

**修改**:
- `morning-analysis/SKILL.md` Step 2 — web_search 后加载此文件做质量评估
- `afternoon-trade/SKILL.md` Step 2 同理
- `stock-analysis/SKILL.md` — 数据处理阶段增加质量评估
- `pkg/core/agent/prompt.go` `BasePromptTemplate` — 增加信息质量段落

---

## Phase 2: 策略体系升级

**目标**: 扩展操作空间，精细化管理

**改动量**: ~10个文件 (新增4个, 修改6个)

### H1 — 扩展操作类型

**新增**: `pkg/logic/skills/plugs/trading-common/references/operation-types.md`

```markdown
## 操作类型定义

### 买入类
| 类型 | 比例 | 条件 | 置信度要求 |
|------|------|------|------------|
| BUY_TRIAL | 计划仓位10% | 信号弱但方向对 | 低(60%+) |
| BUY_STANDARD | 计划仓位30% | 多信号共振 | 中(70%+) |
| BUY_ADD | 剩余仓位50% | 趋势确认 | 高(80%+) |

### 卖出类
| 类型 | 比例 | 条件 | 置信度要求 |
|------|------|------|------------|
| SELL_REDUCE | 持仓30% | 单一信号走弱 | 低 |
| SELL_HALF | 持仓50% | 盈利达标/信号走弱 | 中 |
| SELL_ALL | 持仓100% | 多信号共振/止损 | 高 |

### 特殊操作
| 类型 | 说明 | 条件 |
|------|------|------|
| SWAP | 卖A买B | 行业轮动、比价优势 |
| HOLD | 主动观望 | 市场不适合交易 |
| REVERSE_REPO | 闲置资金逆回购 | 不操作时的收益 |

### 信号强度 → 操作映射
| 信号强度 | 买入操作 | 卖出操作 |
|----------|----------|----------|
| 1个弱信号 | HOLD | SELL_REDUCE |
| 2个中信号 | BUY_TRIAL | SELL_HALF |
| 1强+2中 | BUY_STANDARD | SELL_ALL |
| 2强+3中 | BUY_ADD | SELL_ALL |
```

**修改**:
- `morning-analysis/references/buy-conditions.md` — 操作类型映射
- `morning-analysis/references/sell-conditions.md` — 同理
- `afternoon-trade/references/buy-conditions.md` — 同理

### H2 — ATR 动态止损

**新增**: `pkg/logic/skills/plugs/trading-common/references/risk-management.md`

```markdown
## 风险管理规则

### ATR 动态止损
- 初始止损 = 入场价 - 2 × ATR(14)
- 移动止盈: 盈利后止损位跟随 (最高价 - 1 × ATR(14))
- 止损位只能上移，不能下移

### 最大回撤熔断
- 日内回撤 > 5%: 暂停所有开仓
- 周内回撤 > 10%: 减仓至 30% 以下
- 月内回撤 > 20%: 暂停交易，全面复盘

### 时间止损
- 持仓 N 日未达预期收益: 减半仓
- 持仓 2N 日仍亏损: 清仓

### 凯利公式仓位优化
- 仓位 = (胜率 × 赔率 - 失败率) / 赔率
- 作为参考上限，不超过固定规则的限制
```

**修改**:
- `trading-common/references/execution-rules.md` — 执行前引用风险管理规则

### H3 — 置信度标注

在 Phase 1 推理链模板中已包含，本阶段增加仓位联动:

```markdown
## 置信度-仓位联动
| 置信度 | 仓位系数 | 适用操作 |
|--------|----------|----------|
| 高(>80%) | 100% | BUY_STANDARD, BUY_ADD, SELL_ALL |
| 中(60-80%) | 50% | BUY_TRIAL, SELL_HALF |
| 低(<60%) | 20% | HOLD, 最多 BUY_TRIAL |
```

### H4 — 持仓相关性检查

**新增**: `pkg/logic/skills/plugs/trading-common/references/correlation-checklist.md`

```markdown
## 持仓相关性检查

### 买入前必查
1. 与现有持仓是否同行业? 同概念?
2. 行业总暴露是否超过总资产 50%?
3. 新标的与现有标的历史相关性?

### 相关性矩阵 (概念级别)
- 如果新标的与已有持仓高度相关 (>0.7): 降低仓位或放弃
- 如果新标的提供对冲效果 (负相关): 可适当增加仓位

### 行业集中度限制
| 行业 | 最大暴露 |
|------|----------|
| 单一行业 | 总资产 50% |
| 单一子行业 | 总资产 30% |
| 单一概念 | 总资产 20% |
```

**修改**:
- `morning-analysis/references/buy-conditions.md` — 增加相关性检查
- `afternoon-trade/references/buy-conditions.md` — 同理

### H5 — 盘前准备模块

**新增**: `pkg/logic/skills/plugs/pre-market-analysis/SKILL.md`

```yaml
name: pre-market-analysis
description: 盘前准备，集合竞价分析、隔夜外盘影响、今日事件日历
version: 1.0.0
priority: 8
pattern: pipeline
steps: 3
triggers:
  - time: "8:30-9:25"
    session: pre-market-session
tools:
  - web_search
  - fetch_page_content
  - read_knowledge
dependencies:
  - trading-common
  - stock-analysis
```

步骤:
1. 集合竞价数据分析 (9:15-9:25)
2. 隔夜外盘影响 (美股收盘、A50期货、港股)
3. 今日事件日历 + 操作预判 → 传递给 morning-analysis

---

## Phase 3: 风险管理增强

**改动量**: ~6个文件 (新增2个, 修改4个, 可能有 Go 代码)

### M1 — 技术指标体系升级

**修改**: `stock-analysis/references/technical-analysis.md`

新增指标:
- **VWAP** (成交量加权平均价): 机构建仓成本参考
- **ATR(14)**: 波动率测量，止损位计算基础
- **OBV** (能量潮): 量价关系验证
- **多时间周期框架**: 周线定方向，日线找入场，60分钟精确定时

### M2 — 盘中监控

**新增**: `pkg/logic/skills/plugs/realtime-monitor/SKILL.md` (或 Go 代码层实现)

- 价格预警: 突破/跌破关键位 → 触发通知
- 止损触发: 实时检测 ATR 止损位
- 异常波动: 突然放量/闪崩 → 触发通知
- 实现方式: 轻量级定时轮询 (每5分钟) vs 事件驱动

> 注意: 此功能可能需要 Go 代码层支持

### M3 — 交易前置检查

**新增**: `pkg/logic/skills/plugs/trading-common/references/pre-trade-checklist.md`

每次交易前:
1. 当前市场适合交易吗? (检查市场状态)
2. 今天已经交易几次了? (过度交易 >5次/日 警告)
3. 这个决策是否与上次决策矛盾?
4. 如果什么都不做会怎样? (强制考虑 HOLD)

### M4 — 实时错误规避

**修改**: `trading-common/references/execution-rules.md`

买入前增加:
- 当前涨幅 > 3%? → 警告，要求标记为"追高风险"，需要更严格理由
- 最近30分钟涨幅? → 短线追高检查
- 今日已买入同板块? → 集中度警告

---

## Phase 4: 量化体系建立

**改动量**: 独立模块开发

### L1 — 策略回测

独立模块: `backtest/`
- 输入: 策略规则 + 历史K线数据
- 输出: 收益曲线、最大回撤、夏普比率、胜率、盈亏比
- 技术选型: Go 实现或 Python 脚本

### L2 — 绩效归因

**增强**: `market-close-summary/references/review-criteria.md`
- Alpha/Beta 分解 (相对大盘的超额收益)
- 选股能力 vs 择时能力 归因
- 行业暴露归因分析

### L3 — 多策略框架

**新增**: `strategies/` 多策略定义
- 趋势跟踪策略
- 均值回归策略
- 动量策略
- 策略 A/B 对比和切换机制

---

## 文件改动汇总

```
Phase 1 (CRITICAL):
├── 新增
│   ├── pkg/logic/skills/plugs/trading-common/references/reasoning-chain.md
│   ├── pkg/logic/skills/plugs/trading-common/references/market-regime.md
│   └── pkg/logic/skills/plugs/stock-analysis/references/information-quality.md
├── 修改
│   ├── pkg/logic/skills/plugs/morning-analysis/SKILL.md
│   ├── pkg/logic/skills/plugs/afternoon-trade/SKILL.md
│   ├── pkg/logic/skills/plugs/stock-analysis/SKILL.md
│   └── pkg/core/agent/prompt.go

Phase 2 (HIGH):
├── 新增
│   ├── pkg/logic/skills/plugs/trading-common/references/operation-types.md
│   ├── pkg/logic/skills/plugs/trading-common/references/risk-management.md
│   ├── pkg/logic/skills/plugs/trading-common/references/correlation-checklist.md
│   └── pkg/logic/skills/plugs/pre-market-analysis/SKILL.md
├── 修改
│   ├── pkg/logic/skills/plugs/morning-analysis/references/buy-conditions.md
│   ├── pkg/logic/skills/plugs/morning-analysis/references/sell-conditions.md
│   ├── pkg/logic/skills/plugs/afternoon-trade/references/buy-conditions.md
│   └── pkg/logic/skills/plugs/trading-common/references/execution-rules.md

Phase 3 (MEDIUM):
├── 新增
│   ├── pkg/logic/skills/plugs/trading-common/references/pre-trade-checklist.md
│   └── pkg/logic/skills/plugs/realtime-monitor/SKILL.md (待定)
├── 修改
│   ├── pkg/logic/skills/plugs/stock-analysis/references/technical-analysis.md
│   └── pkg/logic/skills/plugs/trading-common/references/execution-rules.md

Phase 4 (LOW):
├── 独立模块 backtest/
├── 修改 review-criteria.md
└── 新增 strategies/ 多策略
```

---

## 系统架构演进

### 当前架构
```
用户输入 → Skill选择 → 信息收集 → 分析决策 → 执行
                              ↑
                         web_search (无过滤)
```

### Phase 1 后
```
用户输入 → Skill选择 → 信息收集 → 质量评估 → 市场判断 → 推理链 → 决策 → 执行
                              │         │          │
                              ▼         ▼          ▼
                         时效性分层  状态分类   观察/假设/验证
                         来源可信度             备选/风险/置信度
                         反身性检查
```

### Phase 2 后
```
                      ┌─────────────┐
                      │ 盘前准备     │ 8:30
                      │ 集合竞价     │
                      │ 隔夜外盘     │
                      │ 事件日历     │
                      └──────┬──────┘
                             ▼
用户输入 → info → quality → regime → reasoning → decision → execute
                                            │          │
                                            ▼          ▼
                                       置信度联动   操作类型选择
                                       仓位计算     {试探/标准/加仓
                                                      减仓/半仓/清仓
                                                      调仓/观望/逆回购}
```

### 最终目标架构
```
         ┌────────────────────────────────────────────┐
         │              策略层                         │
         │  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────────┐  │
         │  │趋势  │ │均值  │ │动量  │ │用户自定义 │  │
         │  │跟踪  │ │回归  │ │策略  │ │策略      │  │
         │  └──┬───┘ └──┬───┘ └──┬───┘ └────┬─────┘  │
         │     └────────┼────────┼──────────┘         │
         │              ▼                            │
         │     ┌────────────────┐                    │
         │     │  策略选择/切换  │                    │
         │     └───────┬────────┘                    │
         └─────────────┼─────────────────────────────┘
                       ▼
         ┌────────────────────────────────────────────┐
         │              决策层                         │
         │  市场状态 → 信息质量 → 推理链 → 置信度      │
         └─────────────────────┬──────────────────────┘
                               ▼
         ┌────────────────────────────────────────────┐
         │              执行层                         │
         │  前置检查 → 操作类型 → 仓位计算 → 执行      │
         │  盘中监控 → 止损触发 → 风险控制             │
         └─────────────────────┬──────────────────────┘
                               ▼
         ┌────────────────────────────────────────────┐
         │              反馈层                         │
         │  复盘总结 → 错误记录 → 策略评估 → 回测      │
         └────────────────────────────────────────────┘
```

---

## 关键设计决策 (待确认)

| # | 决策 | 选项A | 选项B |
|---|------|-------|-------|
| Q1 | 推理链实现方式 | skill reference (AI填写) | Go代码结构化输出 |
| Q2 | 市场状态判断 | 技术指标自动计算 | AI搜索市场概况 |
| Q3 | 盘前准备模块 | 独立 skill | 合并到 morning-analysis |
| Q4 | 盘中监控实现 | 定时轮询 (轻量) | WebSocket推送 (重量) |
