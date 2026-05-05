# summary-table-parser 规格说明

## Purpose

扩展 `parseSummaryFields` 函数，使其能够从 Markdown 表格格式中提取结构化字段值，解决知识库 summary 文件数值字段全为 0 的问题。

## Requirements

### Requirement: table-field-extraction

`parseSummaryFields` SHALL 在 YAML frontmatter 和 `extractField()` 解析失败后，尝试从 Markdown 表格中提取字段值。

#### Scenario: 从表格提取资产字段

- **GIVEN** summary 文件内容包含表格：
  ```markdown
  | 项目 | 数值 |
  |------|------|
  | 初始资金 | **500,000.00 元** |
  | 当前总资产 | **493,233.00 元** |
  ```
- **WHEN** 调用 `ParseSummaryFile(path)`
- **THEN** `SummaryData.CurrAsset` 被设为 `493233.00`
- **AND** 从 "初始资金" 行提取的值可用于推导 `PrevAsset`

#### Scenario: 从表格提取盈亏字段

- **GIVEN** summary 表格包含：
  ```markdown
  | 总盈亏 | **-6,767.00 元** |
  | 总收益率 | **-1.35%** |
  ```
- **WHEN** 调用 `ParseSummaryFile(path)`
- **THEN** `SummaryData.PNL` 被设为 `-6767.00`
- **AND** `SummaryData.PNLRate` 被设为 `-1.35`

#### Scenario: 表格中无匹配字段

- **GIVEN** summary 表格不包含任何已知字段名（如 "当前总资产"、"总盈亏" 等）
- **WHEN** 调用 `ParseSummaryFile(path)`
- **THEN** 数值字段保持零值
- **AND** `SummaryData.Raw` 仍包含完整原始内容
- **AND** 不发生 panic 或错误

---

### Requirement: field-name-mapping

表格解析 SHALL 使用字段名映射表，将表格中的中文字段名映射到 `SummaryData` 结构体字段。

#### Scenario: 字段名映射表

- **GIVEN** 以下字段名映射：
  - "初始资金" / "昨日资产" → `PrevAsset`
  - "当前总资产" / "今日资产" → `CurrAsset`
  - "总盈亏" / "今日盈亏" → `PNL`
  - "总收益率" / "收益率" → `PNLRate`
- **WHEN** 在表格中匹配任意一个映射名称
- **THEN** 提取对应的数值

#### Scenario: 解析优先级

- **GIVEN** summary 文件同时包含 YAML frontmatter、`**字段**: 值` 格式、和表格
- **WHEN** 调用 `ParseSummaryFile(path)`
- **THEN** 优先使用 YAML frontmatter 解析
- **AND** 其次尝试 `**字段**: 值` 格式
- **AND** 最后尝试表格解析（作为 fallback）
- **AND** 先成功的解析结果被使用
