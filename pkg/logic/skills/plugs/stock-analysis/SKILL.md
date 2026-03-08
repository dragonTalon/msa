---
name: stock-analysis
description: This skill should be used when the user wants to analyze stocks, get stock recommendations, asks about stock technical analysis (MACD, RSI, support/resistance levels, volume analysis), fundamental analysis (PE ratio, PB ratio, financial data, revenue, profit), market analysis (fund flow, market sentiment, policy impact), or requests investment advice.
version: 1.0.0
priority: 8
---

# Stock Analysis

## 数据处理流程

### 信息收集阶段
1. 先调用 web_search 工具检索相关信息，获取链接列表
2. 必须从搜索结果中选取1~10个最相关的链接（优先选择发布时间在3个月内的权威来源），调用 fetch_page_content 工具获取完整页面内容
3. 禁止仅使用搜索摘要进行分析，必须基于 fetch_page_content 返回的完整内容
4. 对多源信息进行交叉验证，确保数据准确性

### 分析框架

#### 技术面分析
需覆盖：
- 趋势：日线级别处于上升趋势/下降趋势
- 支撑阻力位：支撑位、阻力位的具体价格
- 成交量：近3日成交量较上周的变化
- 技术指标：至少2个技术指标（MACD、RSI、KDJ等）

#### 基本面分析
需覆盖：
- 财务数据：2024年净利润同比增长、毛利率变化
- 行业地位：白酒行业市占率
- 成长性：未来3年营收复合增长率预计
- 估值水平：PE为25倍，低于行业平均30倍

#### 市场面分析
需覆盖：
- 资金流向：北向资金近5日净买入金额
- 市场情绪：行业板块涨跌幅、市场热度
- 政策影响：白酒消费税政策调整预期

#### 风险分析
需覆盖：
- 系统性风险：大盘面临3000点压力位
- 个股特有风险：公司一季度业绩不及预期
- 流动性风险：股票近30日平均换手率低于1%

## 使用此技能时
应结合 output-formats skill 中的投资建议模板进行结构化输出。
