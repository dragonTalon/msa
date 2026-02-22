# 实现任务：{{changeName}}

{{#if riskAlerts}}
## ⚠️ 风险警报

基于历史问题识别的风险：

{{#each riskAlerts}}
### {{severity}}：{{message}}

- **相关问题**：{{issueId}}
- **缓解措施**：{{mitigation}}

{{/each}}
{{/if}}

## 1. <!-- 任务组名称 -->

- [ ] 1.1 <!-- 任务描述 -->
{{#if knowledgeContext}}
> 📚 **知识上下文**：{{knowledgeContext}}
{{/if}}
- [ ] 1.2 <!-- 任务描述 -->

## 2. <!-- 任务组名称 -->

- [ ] 2.1 <!-- 任务描述 -->
- [ ] 2.2 <!-- 任务描述 -->

{{#if hasHistoricalIssues}}
## 基于知识的测试

{{#each historicalTests}}
### {{issueTitle}} 回归测试

> 📚 **来源**：问题 {{issueId}}

**历史问题**：
{{historicalProblem}}

**测试场景**：
{{testScenario}}

**预期行为**：
{{expectedBehavior}}

{{/each}}
{{/if}}
