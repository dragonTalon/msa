## ADDED Requirements

### Requirement: 九种操作类型定义

系统必须支持 9 种交易操作类型，替代当前仅有的买/卖二元操作。

操作类型分为三类:

**买入类**: BUY_TRIAL(试探,10%仓位), BUY_STANDARD(标准,30%仓位), BUY_ADD(加仓,50%仓位)
**卖出类**: SELL_REDUCE(减持,30%), SELL_HALF(减半,50%), SELL_ALL(清仓,100%)
**特殊类**: SWAP(换股), HOLD(主动观望), REVERSE_REPO(逆回购)

#### Scenario: 弱买入信号触发试探性持仓

- **GIVEN** 市场状态允许交易，且仅有 1 个弱买入信号
- **WHEN** AI 分析完信号强度后做出买入决策
- **THEN** 应选择 BUY_TRIAL 操作类型，使用计划仓位的 10%

#### Scenario: 强卖出信号触发清仓

- **GIVEN** 持仓中存在某股票，且出现 2 强+3 中卖出信号
- **WHEN** AI 判断需要卖出
- **THEN** 应选择 SELL_ALL 操作类型，卖出全部持仓

#### Scenario: 市场不适合操作时主动观望

- **GIVEN** 市场状态为"极端"，或预检查发现 3 项以上风险因素
- **WHEN** AI 完成分析决策
- **THEN** 应输出 HOLD 操作，不执行任何买卖

---

### Requirement: 信号强度到操作类型映射

必须定义信号强度等级与对应操作类型的明确映射关系。

| 信号强度 | 买入操作 | 卖出操作 |
|----------|----------|----------|
| 1个弱信号 | HOLD | SELL_REDUCE |
| 2个中信号 | BUY_TRIAL | SELL_HALF |
| 1强+2中 | BUY_STANDARD | SELL_ALL |
| 2强+3中 | BUY_ADD | SELL_ALL |

#### Scenario: 中等信号触发试探买入

- **GIVEN** 某股票出现 2 个中等强度买入信号
- **WHEN** AI 查询操作类型映射表
- **THEN** 应选择 BUY_TRIAL，置信度要求 ≥60%
