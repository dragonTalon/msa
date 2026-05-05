## Purpose

支持按日期范围筛选交易记录，用于当日交易次数统计和交易行为审计。

## ADDED Requirements

### Requirement: 交易记录日期范围筛选

`get_transactions` 工具必须新增 `date_from` 和 `date_to` 可选参数，支持按日期范围筛选交易记录。

参数定义：
- `date_from` (string, 可选) — 起始日期，格式 YYYY-MM-DD
- `date_to` (string, 可选) — 截止日期，格式 YYYY-MM-DD
- 两个参数均为可选，不传则保持现有行为（返回全部记录）

#### Scenario: 查询当日交易记录

- **GIVEN** 当前日期 2026-05-05，当日有 3 条交易记录
- **WHEN** AI 调用 `get_transactions(date_from="2026-05-05", date_to="2026-05-05")`
- **THEN** 应仅返回 2026-05-05 的交易记录，total=3

#### Scenario: 查询日期范围内的交易记录

- **GIVEN** 2026-05-01 至 2026-05-05 之间有 8 条交易记录
- **WHEN** AI 调用 `get_transactions(date_from="2026-05-01", date_to="2026-05-05")`
- **THEN** 应返回该范围内的 8 条记录

#### Scenario: 不传日期参数时保持向后兼容

- **GIVEN** 账户有 50 条交易记录
- **WHEN** AI 调用 `get_transactions()` 不传日期参数（或传空值）
- **THEN** 应返回最近 50 条记录，与现有行为完全一致

#### Scenario: 仅传 date_from 时查询从该日期至今

- **GIVEN** date_from="2026-05-01"，不传 date_to
- **WHEN** AI 调用 `get_transactions(date_from="2026-05-01")`
- **THEN** 应返回 2026-05-01 至今的所有交易记录

---

### Requirement: 当日交易次数统计

结合日期筛选能力，AI 必须能在交易前置检查中获取当日交易次数。

- 使用 `get_transactions(date_from="<today>", date_to="<today>")` 获取当日记录
- 根据返回的 `total` 判断是否超过每日上限（5次）
- 此能力配合 pre-trade-checklist 使用

#### Scenario: 统计当日买入次数

- **GIVEN** 当日有 2 条 BUY 类型交易和 1 条 SELL 类型交易
- **WHEN** AI 调用 `get_transactions(date_from="2026-05-05", date_to="2026-05-05")`
- **THEN** total=3，AI 可进一步按 type 字段统计 BUY 次数=2
