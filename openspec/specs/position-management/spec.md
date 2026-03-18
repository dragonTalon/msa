# position-management Specification

## Purpose
TBD - created by archiving change enhance-trading-system. Update Purpose after archive.
## Requirements
### Requirement: 计划仓位计算

系统 SHALL 根据总资产和可用资金计算计划仓位，取两者较小值。

#### Scenario: 计算单只股票计划仓位
- **WHEN** 用户计划买入某只股票
- **THEN** 计划仓位 = min(总资产 × 30%, 可用资金)

#### Scenario: 计划仓位的整手处理
- **WHEN** 计算出计划仓位金额
- **THEN** 转换为整手数量（100股整数倍）

---

### Requirement: 分段买入策略

系统 SHALL 支持三种分段买入方式，由模型根据市场情况选择。

#### Scenario: 按时间分段买入
- **WHEN** 模型选择时间分段策略
- **THEN** 首次买入计划仓位 1/3，午盘确认后加 1/3，次日确认后加满

#### Scenario: 按价格信号分段买入
- **WHEN** 模型选择价格信号分段策略
- **THEN** 买入信号出现时买入 1/3，上涨 2-3% 时加 1/3，趋势确认时加满

#### Scenario: 固定比例分段买入
- **WHEN** 模型选择固定比例分段策略
- **THEN** 第一次买入 40%，第二次买入 40%，第三次买入 20%

---

### Requirement: 买入限制

系统 MUST 强制执行买入限制规则。

#### Scenario: 单次买入上限
- **WHEN** 执行买入操作
- **THEN** 单次买入数量不超过计划仓位的 50%

#### Scenario: 同日累计买入上限
- **WHEN** 同一交易日对同一股票多次买入
- **THEN** 累计买入数量不超过计划仓位的 70%

#### Scenario: 追高限制
- **WHEN** 股票涨幅超过 50%
- **THEN** 不执行买入操作

---

### Requirement: 分段卖出策略

系统 SHALL 支持分段卖出，但止损情况例外。

#### Scenario: 正常分批卖出
- **WHEN** 正常卖出操作（非止损）
- **THEN** 单次最多卖出持仓的 50%

#### Scenario: 止损一次性卖出
- **WHEN** 持仓亏损超过 20%
- **THEN** 允许一次性全部卖出

---

### Requirement: SKILL 文档更新

trading-common SKILL MUST 包含完整的仓位管理规则文档。

#### Scenario: 文档内容完整性
- **WHEN** 查阅 trading-common/SKILL.md
- **THEN** 包含计划仓位计算、分段买入策略、买入限制、分段卖出策略四个章节

