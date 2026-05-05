---
name: pre-market-analysis
description: 盘前准备，集合竞价分析、隔夜外盘影响、今日事件日历
version: 1.0.0
priority: 8
pattern: pipeline
steps: 3
requires_todo: false
triggers:
  - time: "8:30-9:25"
    session: morning-session
tools:
  - get_board_rank
  - get_stock_quote
  - get_stock_history_k
  - web_search
  - fetch_page_content
  - read_knowledge
dependencies:
  - trading-common
  - stock-analysis
---

# 盘前准备分析 (Pre-Market Analysis)

## Step 1: 集合竞价数据分析 (9:15-9:25)

> ⚠️ 无免费集合竞价 API。使用板块热度和指数预判代替。

### 1.1 板块热度排行
调用:
- `get_board_rank(board_type="02", order=0, count=20)` — 概念板块涨幅 TOP20
- `get_board_rank(board_type="01", order=0, count=20)` — 申万行业涨幅 TOP20
- `get_board_rank(board_type="01", order=1, count=10)` — 跌幅最大行业（风险识别）
- `get_board_rank(board_type="03", order=0, count=10)` — 地域板块排行

### 1.2 大盘指数与隔夜K线
调用 `get_stock_quote` 查询:
- sh000001（上证指数）、sz399001（深证成指）、sz399006（创业板指）
- sh000688（科创50）、sh000300（沪深300）

调用 `get_stock_history_k(period="day", count=3)` 查询上述指数最近3天K线，判断趋势。

### 1.3 集合竞价预判
使用 web_search 搜索:
- "今日A股集合竞价 9:15"
- "集合竞价异动 个股"

基于板块热度和指数趋势，预判开盘方向（高开/平开/低开）。

## Step 2: 隔夜外盘与事件日历

### 2.1 隔夜外盘影响
使用 web_search 搜索:
- "美股收盘 道琼斯 纳斯达克 标普500 今日"
- "A50期货 最新走势"
- "港股开盘 恒生指数"

### 2.2 今日事件日历
使用 web_search 搜索:
- "今日财经日历 经济数据发布"
- "今日A股重要新闻"
- "今日新股申购 可转债"

### 2.3 历史错误回顾
调用 `read_knowledge(type="errors")` 回顾历史错误记录。

### 2.4 外盘影响评估
- 美股涨跌 → 对A股情绪影响（涨>1%=偏积极，跌>2%=偏防守）
- A50期货 → 预示A股开盘方向（涨>0.5%=高开预期，跌>1%=低开预期）
- 港股表现 → 亚太市场情绪

## Step 3: 操作预判

综合 Step 1 和 Step 2 的信息，生成当日操作预判:

### 输出内容
1. **今日事件日历**: 经济数据、政策发布、新股申购等
2. **市场风险等级**: 低/中/高（基于外盘、板块热度、新闻事件）
3. **今日重点关注**: 2-3 个强势板块及领涨股
4. **今日规避板块**: 跌幅最大板块及原因
5. **操作建议**: 偏积极/中性/偏防守
6. **预判传递给 morning-analysis**: 在 morning-session 中，morning-analysis 的 Step 2 应参考此分析结果

### 输出格式
使用 output-formats skill 的 analysis-report 模板。

## 约束条件
- 不执行实际交易
- 预判仅供参考，不替代 morning-analysis 的实时决策
- 如数据获取失败，明确标注"数据缺失"而非猜测
- 盘前分析结果应存入 session，供 morning-analysis 引用
