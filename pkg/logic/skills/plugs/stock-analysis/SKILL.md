---
name: stock-analysis
description: 股票分析专家，提供技术面、基本面、市场面三维分析框架。当用户要求分析股票、获取股票建议、询问技术分析（MACD、RSI、支撑阻力）、基本面分析（PE、PB、财务数据）或市场分析（资金流向、市场情绪）时使用。
version: 2.1.0
priority: 8
pattern: tool-wrapper
tools:
  - web_search
  - fetch_page_content
dependencies:
  - trading-common
  - output-formats
---

# Stock Analysis - 股票分析专家

## 概述

此 SKILL 让 Agent 成为股票分析专家，提供结构化的分析框架。

**使用方式：** 在分析股票时，按需加载 `references/` 目录下的分析框架。

---

## 数据处理流程

### 信息收集阶段

1. **多角度搜索（新增）**
   - 加载 `trading-common/references/search-query-tactics.md`
   - 使用 3 组并行查询策略：
     - 查询组 1: 基本面方向（财报、估值、行业对比）
     - 查询组 2: 技术面方向（走势、资金流向、板块轮动）
     - 查询组 3: 反向/风险方向（利空、风险、减持）
   - 每组查询包含时间锚定词

2. 从搜索结果中选取 1~10 个最相关的链接（优先选择发布时间在 3 个月内的权威来源）

3. 调用 `fetch_page_content` 工具获取完整页面内容

4. **禁止**仅使用搜索摘要进行分析，必须基于 `fetch_page_content` 返回的完整内容

5. 对多源信息进行交叉验证，确保数据准确性

---

### 信息质量门（新增）

**加载 `trading-common/references/news-quality-gate.md`**

对收集的信息执行 4 步过滤：

```
Step 1: 来源分类 → S/A/B/C/X 五级分类
Step 2: 时效检查 → > 60min 可能已定价
Step 3: 拥挤度检查 → 板块涨 > 3% = crowded（牛市 > 5%）
Step 4: 完整性检查 → ≥ 3 条有效信息 + 至少 1 条反向观点
```

输出信息质量摘要，标注每条分析所用信息的质量等级。

---

### 市场状态前置分析（新增）

在股票分析前，先加载市场状态：

```
→ 加载 trading-common/references/market-regime-classifier.md
→ 确定当前市场状态（8种之一）
→ 了解当前市场的仓位上限、追高阈值等参数
→ 在分析结论中标注市场状态对个股的影响
```

市场状态影响个股分析的维度：
- 牛市中：放宽估值容忍度，关注趋势强度
- 熊市中：收紧估值要求，关注防御属性
- 震荡中：关注区间边界，高抛低吸策略

---

## 可用的 References

| 文件 | 用途 | 加载时机 |
|------|------|----------|
| `technical-analysis.md` | 技术面分析框架 | 分析趋势、指标时 |
| `fundamental-analysis.md` | 基本面分析框架 | 分析财务数据时 |
| `market-analysis.md` | 市场面分析框架 | 分析资金流向、情绪时 |

### 新增可用 References（来自 trading-common）

| 文件 | 用途 | 加载时机 |
|------|------|----------|
| `trading-common/references/search-query-tactics.md` | 多角度搜索策略 | 搜索信息时 |
| `trading-common/references/news-quality-gate.md` | 信息质量门 | 分析前过滤信息 |
| `trading-common/references/market-regime-classifier.md` | 市场状态分类 | 分析前了解市场环境 |

---

## 输出规范

应结合 `output-formats` SKILL 中的投资建议模板进行结构化输出。

### 分析输出应包含
1. 信息质量摘要（来源评级、时效、拥挤度）
2. 市场状态上下文（当前状态及对个股的影响）
3. 三维分析（技术面 + 基本面 + 市场面）
4. 风险提示

---

## 约束条件

1. 必须基于完整页面内容分析，禁止仅使用搜索摘要
2. 多源信息必须交叉验证
3. 分析结果必须包含风险提示
4. 必须使用多角度搜索策略，避免信息茧房
5. 必须通过信息质量门过滤信息来源
6. 分析时必须标注当前市场状态及其影响
