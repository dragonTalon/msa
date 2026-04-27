# Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** <!-- 一句话描述此变更要实现什么 -->

**Architecture:** <!-- 2-3 句关于实现方法 -->

**Tech Stack:** Go 1.21+, Cobra, Bubble Tea, Eino, GORM/SQLite

---

## ⚠️ 风险警报

<!-- 基于历史问题识别的风险，格式：
### 高/中/低：风险描述
- **相关问题**：#ID - 标题
- **缓解措施**：具体措施
如无风险，删除此节。 -->

---

## 1. <!-- 任务组名称 -->

**Files:**
- Modify: `<!-- exact/path/to/file.go -->`

- [ ] **Step 1: Write the failing test**

```go
// 完整测试代码
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./... -run TestXxx`
Expected: FAIL with "..."

- [ ] **Step 3: Implement**

```go
// 完整实现代码
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./... -run TestXxx`
Expected: PASS

- [ ] **Step 5: Verify compilation**

Run: `go vet ./...`
Expected: no output

- [ ] **Step 6: Commit**

```bash
git add <files>
git commit -m "feat: <!-- 描述 -->"
```

## 2. <!-- 任务组名称 -->

**Files:**
- Modify: `<!-- exact/path/to/file.go -->`

- [ ] **Step 1: <!-- 动作描述 -->**

```go
// 完整代码
```

---

## 基于知识的测试

<!-- 针对历史问题的回归测试，格式：
### 问题 #ID 回归测试 - 标题
> 📚 **来源**：问题 #ID
**历史问题**：问题描述
**测试场景**：测试步骤
**预期行为**：预期结果
如无历史问题，删除此节。 -->

---

**知识引用**：
- 模式：<!-- 相关模式 ID，如 p001 -->
- 问题：<!-- 相关问题 ID，如 #001 -->
