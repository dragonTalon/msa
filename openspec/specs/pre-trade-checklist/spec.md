## Purpose

定义交易前的强制执行检查清单，包括市场环境、交易频率、决策一致性和HOLD替代方案。

## ADDED Requirements

### Requirement: 交易前置检查清单

系统必须新增 `trading-common/references/pre-trade-checklist.md` reference 文件，定义每次交易前 AI 必须执行的 4 项检查。

检查项：
1. 当前市场适合交易吗？（引用 market-regime-classifier.md）
2. 今天已经交易几次了？（过度交易 >5次/日 警告）
3. 这个决策是否与上次决策矛盾？（对比 session 中的上次决策）
4. 如果什么都不做会怎样？（强制考虑 HOLD）

#### Scenario: 检测到过度交易

- **GIVEN** 当日已执行 6 次交易（5 次买入 + 1 次卖出），账户可用余额充足
- **WHEN** AI 准备执行新的买入决策，执行前置检查第 2 项
- **THEN** 应输出 "当日交易 6 次，超过 5 次上限，建议 HOLD"，不执行新交易

#### Scenario: 检测到决策矛盾

- **GIVEN** 上次决策为 "卖出贵州茅台，原因：技术面走弱"，当前准备买入贵州茅台
- **WHEN** AI 执行前置检查第 3 项，对比上次决策
- **THEN** 应输出 "与上次决策矛盾（上次卖出，现拟买入），请重新评估或提供反转理由"

#### Scenario: 强制考虑 HOLD

- **GIVEN** 市场状态为震荡偏熊，信号模糊
- **WHEN** AI 执行前置检查第 4 项
- **THEN** 应在推理链中明确记录 "如果什么都不做：无亏损风险，等待更明确信号"，并将 HOLD 列为首选方案之一

---

### Requirement: 与现有交易流程集成

前置检查清单必须嵌入现有的买入/卖出决策流程。

加载时机：
- morning-analysis/SKILL.md Step 2 — 决策前加载
- afternoon-trade/SKILL.md Step 2 — 决策前加载
- execution-rules.md — 执行前最终检查

#### Scenario: 买入决策流程中触发前置检查

- **GIVEN** morning-analysis 执行到 Step 2 决策阶段
- **WHEN** AI 加载 buy-conditions.md 判断买入条件
- **THEN** 必须先加载 pre-trade-checklist.md 执行 4 项检查，通过后方可继续买入决策
