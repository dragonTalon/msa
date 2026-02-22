# 提案：{{changeName}}

{{#if changeDescription}}
{{changeDescription}}
{{/if}}

## 知识上下文

### 相关历史问题

{{#if relatedIssues}}
过去遇到过的类似问题：

{{#each relatedIssues}}
- **{{id}}**：{{title}}
  - 类型：{{type}} | 严重程度：{{severity}}
  - 组件：{{component}}
  - 摘要：{{summary}}

{{#if resolution}}
  - 解决方案：{{resolution}}
{{/if}}
{{/each}}
{{else}}
知识库中未找到相关问题。
{{/if}}

### 需避免的反模式

{{#if antiPatterns}}
⚠️ **警告**：已为此类变更记录了以下反模式：

{{#each antiPatterns}}
- **{{id}}**：{{title}}
  - 问题所在：{{problem}}
  - 影响：{{impact}}

{{/each}}
{{else}}
未识别到特定反模式。
{{/if}}

### 推荐模式

{{#if patterns}}
✅ **推荐**：考虑使用以下模式：

{{#each patterns}}
- **{{id}}**：{{title}}
  - 描述：{{description}}
  - 使用时机：{{when}}

{{/each}}
{{else}}
未识别到特定模式。
{{/if}}

{{#if riskAlerts}}
## 风险评估

基于历史问题，此变更具有以下风险因素：

{{#each riskAlerts}}
- **{{severity}}**：{{message}}
  - 相关问题：{{issueId}}
  - 缓解措施：{{mitigation}}
{{/each}}
{{/if}}

{{#if mitigationStrategy}}
## 缓解策略

{{mitigationStrategy}}
{{/if}}

## 为什么

{{#if why}}
{{why}}
{{else}}
<!-- 解释此变更的动机。解决什么问题？为什么现在做？ -->
{{/if}}

## 变更内容

{{#if changes}}
{{#each changes}}
- {{this}}
{{/each}}
{{else}}
<!-- 描述变更内容。具体说明新增能力、修改或删除。 -->
{{/if}}

## 能力

### 新增能力
<!-- 引入的能力。使用kebab-case标识符（如 user-auth, data-export）。每个创建 specs/<name>/spec.md -->
{{#if capabilities}}
{{#each capabilities}}
- `{{name}}`：{{description}}
{{/each}}
{{else}}
- `<name>`：<此能力涵盖的简要描述>
{{/if}}

### 修改能力
<!-- 现有能力的需求在变更（不仅是实现）。
     仅当规格级别行为变更时在此列出。每个需要增量规格文件。
     使用 openspec/specs/ 中的现有规格名称。如无需求变更则留空。 -->
{{#if modifiedCapabilities}}
{{#each modifiedCapabilities}}
- `{{name}}`：{{description}}
{{/each}}
{{else}}
- `<existing-name>`：<什么需求在变更>
{{/if}}

## 影响

{{#if impact}}
{{impact}}
{{else}}
<!-- 受影响的代码、API、依赖、系统 -->
{{/if}}

---

**知识备注**：
- 此提案使用 `{{knowledgeBasePath}}` 中的知识生成
- 最后知识同步：{{lastKnowledgeSync}}
- 知识库总问题数：{{totalIssues}}
- 总模式数：{{totalPatterns}}
