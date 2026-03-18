# trade-analysis Specification

## Purpose
TBD - created by archiving change enhance-trading-system. Update Purpose after archive.
## Requirements
### Requirement: 今日交易获取

工具 SHALL 能够自动获取今日所有交易记录和 Session。

#### Scenario: 获取今日交易记录
- **WHEN** 调用 analyze_today_trades 工具
- **THEN** 自动获取今日所有交易记录（通过 get_transactions）

#### Scenario: 获取今日 Session
- **WHEN** 调用 analyze_today_trades 工具
- **THEN** 自动获取今日所有 Session（通过 query_sessions_by_date）

---

### Requirement: 错误类型识别

工具 MUST 能够识别以下错误类型。

#### Scenario: 识别追高买入
- **WHEN** 买入时股票涨幅超过 3%
- **THEN** 标记为"追高买入"潜在错误

#### Scenario: 识别仓位过重
- **WHEN** 单次买入超过计划仓位的 50%
- **THEN** 标记为"仓位过重"潜在错误

#### Scenario: 识别未及时止损
- **WHEN** 持仓亏损超过 20% 且未卖出
- **THEN** 标记为"未及时止损"潜在错误

#### Scenario: 识别同日过度买入
- **WHEN** 同一交易日对同一股票累计买入超过计划仓位的 70%
- **THEN** 标记为"同日过度买入"潜在错误

---

### Requirement: 分析结果输出

工具 SHALL 输出结构化的分析结果。

#### Scenario: 输出潜在错误列表
- **WHEN** 分析完成
- **THEN** 输出包含错误类型、股票名称、详情描述的列表

#### Scenario: 输出建议记录
- **WHEN** 发现潜在错误
- **THEN** 输出建议写入 errors/ 目录的文件内容模板

---

### Requirement: 工具注册

工具 MUST 注册到知识库工具组。

#### Scenario: 工具可被发现
- **WHEN** 系统列出可用工具
- **THEN** analyze_today_trades 出现在 knowledge 工具组中

#### Scenario: 工具描述
- **WHEN** 查看工具描述
- **THEN** 显示"分析今日交易，识别潜在错误操作"

