# skill-frontmatter-access 规格说明

## Purpose

为 Skill 提供单独访问 YAML frontmatter 元数据的能力，使 LLM 可以获取 skill 的 name、version、priority、pattern、output-format 等结构化元数据。

## Requirements

### Requirement: get-frontmatter-method

Skill 结构体 SHALL 提供 `GetFrontmatter()` 方法，返回 SKILL.md 的 YAML frontmatter 原始内容（字符串形式）。

#### Scenario: 获取有 frontmatter 的 skill 的元数据

- **GIVEN** skill 的 SKILL.md 包含 YAML frontmatter：
  ```yaml
  ---
  name: output-formats
  version: 2.0.0
  priority: 8
  pattern: generator
  output-format: markdown
  ---
  ```
- **WHEN** 调用 `skill.GetFrontmatter()`
- **THEN** 返回 YAML frontmatter 的原始字符串内容
- **AND** 包含 `name: output-formats`、`version: 2.0.0` 等字段

#### Scenario: 获取无 frontmatter 的 skill

- **GIVEN** skill 的 SKILL.md 不包含 YAML frontmatter（不以 `---` 开头）
- **WHEN** 调用 `skill.GetFrontmatter()`
- **THEN** 返回空字符串 `""`
- **AND** 不返回 error

---

### Requirement: existing-api-unchanged

`GetContent()` 的现有行为 SHALL 保持不变（仍剥离 frontmatter 返回 body）。

#### Scenario: GetContent 行为不变

- **GIVEN** 任意 skill 的 SKILL.md 文件
- **WHEN** 调用 `skill.GetContent()`
- **THEN** 返回不含 YAML frontmatter 的 body 内容
- **AND** 行为与变更前完全一致

---

### Requirement: lazy-loading-consistency

`GetFrontmatter()` SHALL 使用与 `GetContent()` 相同的懒加载和双重检查锁定模式。

#### Scenario: 线程安全

- **GIVEN** 多个 goroutine 同时调用 `GetContent()` 和 `GetFrontmatter()`
- **WHEN** skill 尚未加载
- **THEN** SKILL.md 文件只被读取一次
- **AND** 不发生 data race
