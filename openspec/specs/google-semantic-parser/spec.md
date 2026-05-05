# google-semantic-parser 规格说明

## Purpose

Google 搜索结果解析采用多策略 fallback，替代单一 `class="g"` 依赖，确保 Google 页面结构变化时仍能提取结果。

## Requirements

### Requirement: multi-strategy-fallback

`parseSearchResults` SHALL 按以下优先级尝试解析 Google 搜索结果：

1. 策略1: 按已知 class 名匹配（`class="g"` 等，当前逻辑）
2. 策略2: 按语义结构匹配（`<h3>` + `<a>` 组合）
3. 策略3: 通用链接提取（页面中所有外部链接）

#### Scenario: 策略1成功时

- **GIVEN** Google 页面包含 `class="g"` 的搜索结果容器
- **WHEN** 调用 `parseSearchResults(html)`
- **THEN** 使用策略1解析并返回结果
- **AND** 不触发策略2和策略3

#### Scenario: 策略1失败策略2成功

- **GIVEN** Google 页面结构变化，`class="g"` 不再存在，但 `<h3>` + `<a>` 结构仍在
- **WHEN** 调用 `parseSearchResults(html)`
- **THEN** 策略1返回空 → 自动切换到策略2
- **AND** 策略2从 `<h3>` 下的 `<a>` 标签提取标题/URL
- **AND** 返回结构化结果

#### Scenario: 策略1和2都失败策略3兜底

- **GIVEN** Google 页面结构完全未知
- **WHEN** 调用 `parseSearchResults(html)`
- **THEN** 策略1/2失败 → 策略3提取所有外部链接
- **AND** 至少返回部分可用结果
- **AND** 不发生错误或返回空

---

### Requirement: result-quality

语义解析提取的结果 SHALL 与 class 名解析结果格式一致。

#### Scenario: 结果格式统一

- **GIVEN** 任意策略解析成功
- **WHEN** 返回搜索结果
- **THEN** 每条结果包含 `Title`, `URL`, `Snippet`, `Source`
- **AND** 无必需字段为空
