## ADDED Requirements

### Requirement: 历史 K 线数据工具

系统必须新增 `get_stock_history_k` Go 工具，封装腾讯财经 fqkline 接口。

工具规格:
- `name`: get_stock_history_k
- `ToolGroup`: stock
- 参数: `stock_code`(string), `period`(string: day/week/month), `count`(int: 1-640), `adjust`(string: qfq/hfq/空)
- 返回: 日期、开盘、收盘、最高、最低、成交量 的数组

#### Scenario: 获取日K线用于 ATR 计算

- **GIVEN** 需要计算 ATR(14)
- **WHEN** 调用 `get_stock_history_k(stock_code="sh600519", period="day", count=14, adjust="qfq")`
- **THEN** 应返回最近 14 个交易日的 [日期, 开盘, 收盘, 最高, 最低, 成交量] 数组

#### Scenario: 获取周K线用于趋势分析

- **GIVEN** 需要判断中长期趋势
- **WHEN** 调用 `get_stock_history_k(stock_code="sz000001", period="week", count=30, adjust="qfq")`
- **THEN** 应返回最近 30 周的周K线数据

#### Scenario: 代码为空时报错

- **GIVEN** 传入空的 stock_code
- **WHEN** 调用 `get_stock_history_k`
- **THEN** 应返回错误 "stock_code is empty"

---

### Requirement: API 响应解析

工具必须正确解析腾讯财经 fqkline API 的 JSON 响应。

API 端点: `https://web.ifzq.gtimg.cn/appstock/app/fqkline/get?param={market}{code},{period},,,{count},{adj}`

响应 `qfq{period}` 数组中每个元素为 `[日期, 开盘, 收盘, 最高, 最低, 成交量]`。

#### Scenario: 解析前复权日K线

- **GIVEN** API 返回 `{"code":0,"data":{"sh600519":{"qfqday":[...]}}}`
- **WHEN** 工具解析响应
- **THEN** 应正确提取 qfqday 数组，返回包含 date/open/close/high/low/volume 的结构化数据
