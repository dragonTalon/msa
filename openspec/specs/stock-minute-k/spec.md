## Purpose

提供分钟级K线数据获取能力，支持VWAP计算、多时间周期框架和短线价格查询。

## ADDED Requirements

### Requirement: 分钟级 K 线数据获取

系统必须新增 `get_stock_minute_k` 工具，复用现有 `minute/query` API（已由 `FetchStockData()` 调用），解析其 `Data` 字段返回结构化分钟 K 线数据。

**数据来源**: `https://web.ifzq.gtimg.cn/appstock/app/minute/query?code={stockCode}`
**原始格式**: `HHmm price cumulative_volume cumulative_turnover`（空格分隔）
**数据量**: 全天约 242 个分钟级数据点（9:30-15:00）

工具元数据：
- `name`: get_stock_minute_k
- `tool_group`: stock
- 参数: `stock_code` (string, required) — 股票代码
- 返回: `StockMinuteKResp` — 包含 `bars: [{time, price, volume, turnover}]`

#### Scenario: 获取当日分钟 K 线数据

- **GIVEN** 股票代码 sh600519（贵州茅台），当前为交易日盘中
- **WHEN** AI 调用 `get_stock_minute_k("sh600519")`
- **THEN** 应返回 JSON 包含分钟级数据数组，每项含 time(HHmm)、price(元)、volume(手，已差分)、turnover(元)

#### Scenario: 非交易日返回空数据

- **GIVEN** 股票代码 sh600519，当前为周末
- **WHEN** AI 调用 `get_stock_minute_k("sh600519")`
- **THEN** 应返回空 bars 数组，message 提示 "非交易日或无分钟数据"

---

### Requirement: 分钟级成交量差分计算

工具必须将 API 返回的累计成交量转换为每分钟的实际成交量。

- 累计成交量序列: `V_cum[0], V_cum[1], ..., V_cum[n]`
- 每分钟成交量: `V[i] = V_cum[i] - V_cum[i-1]`（首分钟 V[0] = V_cum[0]）
- 成交额同理

#### Scenario: 成交量差分计算正确

- **GIVEN** API 返回累计成交量 [408, 1259, 2259]
- **WHEN** 系统计算每分钟成交量
- **THEN** 每分钟成交量应为 [408, 851, 1000]

---

### Requirement: 分钟数据支持价格查询

工具必须支持查询指定时间点的价格，用于短线追高检测。

- 返回的 bars 数组按时间升序排列
- AI 可通过查找 N 分钟前的 bar 获取历史价格
- 如果查询时间点无数据（如尚未到该时间），返回最近可用数据

#### Scenario: 查询 30 分钟前价格

- **GIVEN** 当前时间 14:00，bars 数组包含 13:30 的数据点
- **WHEN** AI 查找 30 分钟前（13:30）的价格
- **THEN** 应返回 13:30 对应的 price 值，用于计算 30 分钟涨跌幅
