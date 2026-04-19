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
- **THEN** 返回空元数据和完整正文内容
