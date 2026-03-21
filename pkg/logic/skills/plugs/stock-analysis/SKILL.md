---
name: stock-analysis
description: 股票分析专家，提供技术面、基本面、市场面三维分析框架。当用户要求分析股票、获取股票建议、询问技术分析（MACD、RSI、支撑阻力）、基本面分析（PE、PB、财务数据）或市场分析（资金流向、市场情绪）时使用。
version: 2.0.0
priority: 8
pattern: tool-wrapper
tools:
  - web_search
  - fetch_page_content
dependencies:
  - output-formats
---

# Stock Analysis - 股票分析专家

## 概述

此 SKILL 让 Agent 成为股票分析专家，提供结构化的分析框架。

**使用方式：** 在分析股票时，按需加载 `references/` 目录下的分析框架。

---

## 数据处理流程

### 信息收集阶段

1. 先调用 `web_search` 工具检索相关信息，获取链接列表
2. 必须从搜索结果中选取 1~10 个最相关的链接（优先选择发布时间在 3 个月内的权威来源）
3. 调用 `fetch_page_content` 工具获取完整页面内容
4. **禁止**仅使用搜索摘要进行分析，必须基于 `fetch_page_content` 返回的完整内容
5. 对多源信息进行交叉验证，确保数据准确性

---

## 可用的 References

| 文件 | 用途 | 加载时机 |
|------|------|----------|
| `technical-analysis.md` | 技术面分析框架 | 分析趋势、指标时 |
| `fundamental-analysis.md` | 基本面分析框架 | 分析财务数据时 |
| `market-analysis.md` | 市场面分析框架 | 分析资金流向、情绪时 |

---

## 输出规范

应结合 `output-formats` SKILL 中的投资建议模板进行结构化输出。

---

## 约束条件

1. 必须基于完整页面内容分析，禁止仅使用搜索摘要
2. 多源信息必须交叉验证
3. 分析结果必须包含风险提示
