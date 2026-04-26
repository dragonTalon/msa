# {{changeName}} Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** [一句话描述此变更要实现什么]

**Architecture:** [2-3 句关于实现方法]

**Tech Stack:** [关键技术和库]

---

{{#if riskAlerts}}
## ⚠️ 风险警报

基于历史问题识别的风险：

{{#each riskAlerts}}
### {{severity}}：{{message}}

- **相关问题**：{{issueId}}
- **缓解措施**：{{mitigation}}

{{/each}}
{{/if}}

## 1. [任务组名称]

**Files:**
- Modify: `exact/path/to/file`

- [ ] **Step 1: [动作描述]**

```bash
# 完整代码或命令
```

- [ ] **Step 2: [动作描述]**

Run: `command`
Expected: [预期输出]

- [ ] **Step 3: Commit**

```bash
git add <files>
git commit -m "feat: [描述]"
```

## 2. [任务组名称]

**Files:**
- Modify: `exact/path/to/file`

- [ ] **Step 1: [动作描述]**

```bash
# 完整代码
```

---

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

---

**知识引用**：
- 模式：[相关模式]
- 问题：[相关问题]
