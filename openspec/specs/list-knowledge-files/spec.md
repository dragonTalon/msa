# 规格：list-knowledge-files

## Purpose

系统 SHALL 提供工具来列出知识库目录下的文件，用于 SKILL 查找可用的知识文件。

## Requirements

### Requirement: 列出知识库文件

系统 SHALL 提供工具来列出知识库目录下的文件，用于 SKILL 查找可用的知识文件。

#### Scenario: 列出根目录下的所有子目录
- **WHEN** 调用工具不带参数
- **THEN** 返回 `.msa/knowledge/` 下所有子目录名称

#### Scenario: 列出指定目录下的文件
- **WHEN** 调用工具指定目录 `errors`
- **THEN** 返回 `.msa/knowledge/errors/` 下所有文件列表

#### Scenario: 列出不存在的目录
- **WHEN** 指定的目录不存在
- **THEN** 返回空列表

---

### Requirement: 路径安全校验

所有路径 MUST 经过安全校验，确保只能访问 `.msa/knowledge/` 目录下。

#### Scenario: 拒绝路径遍历
- **WHEN** 目录参数包含 `..` 或绝对路径前缀
- **THEN** 返回错误 "path traversal not allowed"

---

### Requirement: 文件信息返回

系统 SHALL 在返回的文件列表中包含基本信息。

#### Scenario: 返回文件元数据
- **WHEN** 列出文件成功
- **THEN** 返回每个文件的名称、大小、修改时间

#### Scenario: 按修改时间排序
- **WHEN** 返回文件列表
- **THEN** 按修改时间倒序排列（最新的在前）
