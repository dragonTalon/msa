# 规格：write-knowledge-file

## Purpose

系统 SHALL 提供工具来写入知识库目录下的文件，用于 SKILL 记录错误经验、交易总结等信息。

## Requirements

### Requirement: 安全写入知识文件

系统 SHALL 提供工具来写入知识库目录下的文件，用于 SKILL 记录错误经验、交易总结等信息。

#### Scenario: 创建新文件
- **WHEN** 调用工具写入新文件 `summaries/2026-03-15.md`
- **THEN** 在 `.msa/knowledge/summaries/` 下创建文件

#### Scenario: 覆盖已存在的文件
- **WHEN** 调用工具写入已存在的文件
- **THEN** 使用新内容覆盖旧文件

#### Scenario: 自动创建目录
- **WHEN** 目标目录不存在
- **THEN** 自动创建必要的目录结构

---

### Requirement: 路径安全校验

所有写入路径 SHALL 经过安全校验，确保只能写入 `.msa/knowledge/` 目录下。

#### Scenario: 拒绝路径遍历
- **WHEN** 写入路径包含 `..` 或绝对路径前缀
- **THEN** 返回错误 "path traversal not allowed"

---

### Requirement: 原子写入

写入操作 SHALL 使用原子方式，避免部分写入导致的数据损坏。

#### Scenario: 原子写入文件
- **WHEN** 执行写入操作
- **THEN** 先写入临时文件，成功后重命名为目标文件

---

### Requirement: YAML Frontmatter 支持

写入内容可能包含 YAML frontmatter，系统 SHALL 正确处理。

#### Scenario: 写入带 frontmatter 的内容
- **WHEN** 内容包含 YAML frontmatter
- **THEN** 保持原样写入，不做修改

#### Scenario: 自动生成 frontmatter
- **WHEN** 提供 metadata 参数
- **THEN** 自动生成 YAML frontmatter 并追加内容
