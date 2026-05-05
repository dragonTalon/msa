# 规格：read-knowledge-file

## Purpose

系统 SHALL 提供工具来读取知识库目录下的文件，用于 SKILL 获取用户笔记、策略、错误经验等信息。

## Requirements

### Requirement: 安全读取知识文件

系统 SHALL 提供工具来读取知识库目录下的文件，用于 SKILL 获取用户笔记、策略、错误经验等信息。

#### Scenario: 读取存在的知识文件
- **WHEN** 调用工具读取 `.msa/knowledge/notes/2026-03-14.md`
- **THEN** 返回文件内容和元数据（YAML frontmatter 解析后的结构）

#### Scenario: 读取不存在的文件
- **WHEN** 调用工具读取不存在的文件路径
- **THEN** 返回错误信息，明确标识文件不存在

#### Scenario: 读取非知识库目录的文件
- **WHEN** 调用工具读取 `/etc/passwd` 或 `../../../etc/passwd`
- **THEN** 返回错误信息，拒绝访问（路径遍历攻击防护）

---

### Requirement: 路径安全校验

所有文件路径 SHALL 经过安全校验，确保只能访问 `.msa/knowledge/` 目录下的文件。

#### Scenario: 拒绝路径遍历
- **WHEN** 请求路径包含 `..` 或绝对路径前缀
- **THEN** 返回错误 "path traversal not allowed"

#### Scenario: 限制访问范围
- **WHEN** 请求路径为 `notes/2026-03-14.md`
- **THEN** 自动转换为 `.msa/knowledge/notes/2026-03-14.md`

---

### Requirement: YAML Frontmatter 解析

知识文件通常包含 YAML frontmatter，系统 SHALL 解析并返回结构化数据。

#### Scenario: 解析带 frontmatter 的文件
- **WHEN** 文件内容以 `---` 开头
- **THEN** 解析 YAML 元数据和 Markdown 正文，分别返回

#### Scenario: 处理无 frontmatter 的文件
- **WHEN** 文件不包含 YAML frontmatter
- **THEN** 尝试从 `**字段**: 值` 格式中提取字段值
- **AND** 如果仍未提取到，尝试从 Markdown 表格中提取
- **AND** 返回提取到的元数据和完整正文内容

#### Scenario: 从 Markdown 表格中提取 summary 字段
- **GIVEN** summary 文件使用表格格式存储账户数据：
  ```markdown
  | 项目 | 数值 |
  |------|------|
  | 当前总资产 | **493,233.00 元** |
  | 总盈亏 | **-6,767.00 元** |
  | 总收益率 | **-1.35%** |
  ```
- **WHEN** 调用 `ParseSummaryFile(path)`
- **THEN** `CurrAsset` 被提取为 `493233.00`
- **AND** `PNL` 被提取为 `-6767.00`
- **AND** `PNLRate` 被提取为 `-1.35`
- **AND** `Raw` 字段保留完整的原始文件内容
