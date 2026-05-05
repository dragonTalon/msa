## Purpose

扩展技术指标体系，新增VWAP、ATR、OBV和多时间周期分析框架。

## ADDED Requirements

### Requirement: VWAP 指标

技术分析框架必须新增 VWAP（成交量加权平均价）指标。

计算公式：VWAP = Σ(Price_i × Volume_i) / Σ(Volume_i)

数据来源：
- 日内精确 VWAP：基于 `get_stock_minute_k` 的分钟级数据计算
- 日线近似 VWAP：基于 `get_stock_history_k` 的日线 OHLCV 估算（(High+Low+Close)/3 × Volume）

#### Scenario: 基于分钟数据计算日内 VWAP

- **GIVEN** `get_stock_minute_k` 返回全天 242 个分钟数据点
- **WHEN** AI 执行技术分析，需要计算 VWAP
- **THEN** 应对每分钟的 price × volume 求和，除以总成交量，得到 VWAP 值，与当前价对比判断机构成本位

#### Scenario: 无分钟数据时用日线近似 VWAP

- **GIVEN** `get_stock_minute_k` 返回空数据（非交易日或历史日期）
- **WHEN** AI 需要 VWAP 作为参考
- **THEN** 应使用日线 (H+L+C)/3 作为典型价格，标注 "VWAP 近似值（基于日线）"

---

### Requirement: ATR 指标集成

技术分析框架必须集成 ATR(14) 指标（已在 Phase 2 risk-management.md 中定义计算方法）。

- ATR 计算公式引用 risk-management.md 的完整定义
- technical-analysis.md 中仅需包含简要说明和引用
- ATR 值用于波动率判断：ATR/价格 > 3% = 高波动，1-3% = 正常，<1% = 低波动

#### Scenario: 技术分析报告中包含 ATR 波动率评估

- **GIVEN** 某股票价格 100 元，ATR(14)=3.5 元
- **WHEN** AI 输出技术分析报告
- **THEN** 报告中应包含 "ATR(14)=3.5，波动率 3.5%（高波动），止损位建议参考 risk-management.md"

---

### Requirement: OBV 指标

技术分析框架必须新增 OBV（能量潮）指标。

计算公式：
- 若当日收盘 > 昨日收盘：OBV = 前日 OBV + 当日成交量
- 若当日收盘 < 昨日收盘：OBV = 前日 OBV - 当日成交量
- 若当日收盘 = 昨日收盘：OBV = 前日 OBV

数据来源：`get_stock_history_k` 返回的日线 close 和 volume

#### Scenario: OBV 与价格背离检测

- **GIVEN** 近 10 日价格创新高但 OBV 未创新高（量价背离）
- **WHEN** AI 分析技术面
- **THEN** 应标注 "OBV 与价格顶背离，上涨动能减弱，谨慎追多"

---

### Requirement: 多时间周期框架

技术分析框架必须建立多时间周期分析框架。

三个周期：
- 周线定方向（长期趋势）：使用 `get_stock_history_k(period="week")`
- 日线找入场（中期信号）：使用 `get_stock_history_k(period="day")`
- 60分钟精确定时（短期时机）：使用 `get_stock_minute_k` 聚合为 60 分钟 K 线

60 分钟 K 线聚合规则：
- 将分钟数据按 60 分钟分组
- OHLC：组内首价=Open，末价=Close，最高=High，最低=Low
- 成交量：组内求和

#### Scenario: 多周期共振买入信号

- **GIVEN** 周线 MACD 金叉 + 日线站上 5 日均线 + 60 分钟线放量突破
- **WHEN** AI 执行多周期分析
- **THEN** 应判定为 "多周期共振，信号强度：强"，置信度 ≥80%

#### Scenario: 60 分钟 K 线聚合

- **GIVEN** `get_stock_minute_k` 返回 09:30-10:30 的 60 个分钟数据点
- **WHEN** AI 需要 60 分钟 K 线
- **THEN** 第一根 60 分钟 K 线应为：Open=09:30 价，Close=10:29 价，High=60 分钟内最高，Low=60 分钟内最低，Volume=60 分钟成交量之和
