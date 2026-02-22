# 修复：{{changeName}}

---
id: "{{fixId}}"
title: "{{title}}"
type: "{{type}}"              <!-- runtime|compile|logic|config -->
severity: "{{severity}}"       <!-- critical|high|medium|low -->
component: "{{component}}"
tags:
  - "{{tag1}}"
  - "{{component}}"
created_at: "{{createdAt}}"
resolved_at: "{{resolvedAt}}"
change: "{{changePath}}"
status: "{{status}}"           <!-- open|investigating|resolved|cancelled -->
---

## 问题

### 问题描述

<!-- 发生了什么？应该发生什么？实际行为是什么？ -->

### 重现步骤

1. <!-- 步骤1 -->
2. <!-- 步骤2 -->
3. <!-- 观察到错误 -->

### 错误消息

```
<!-- 粘贴确切的错误消息和堆栈跟踪 -->
```

### 影响评估

- **用户影响**：<!-- 用户遇到什么 -->
- **系统影响**：<!-- 哪些系统部分受影响 -->
- **业务影响**：<!-- 业务后果 -->

## 根本原因

### 调查过程

<!-- 你如何调查的 -->

### 根本原因

<!-- 明确陈述根本原因 -->

### 相关文件

- `{{file_path}}`
  - 行：{{line_numbers}}
  - 说明：{{notes}}

## 解决方案

### 解决方案方法

<!-- 采取的方法、替代方案 -->

### 代码变更

#### 变更描述

**文件**：`{{file_path}}`

```diff
+ // 新增代码
- // 删除代码
  // 修改代码
```

**说明**：{{description}}
**原因**：{{rationale}}

## 验证

### 验证步骤

1. <!-- 步骤1 -->
2. <!-- 步骤2 -->

### 测试结果

- 单元测试：<!-- PASS/FAIL -->
- 集成测试：<!-- PASS/FAIL -->

{{#if hasPattern}}
## 可复用模式

**模式名称**：{{patternName}}
**模式ID**：{{patternId}}

**描述**：
{{patternDescription}}

**适用场景**：
- {{scenario1}}
- {{scenario2}}

**实现**：
```typescript
{{patternImplementation}}
```

**收益**：
- {{benefit1}}
- {{benefit2}}
{{/if}}

{{#if hasAntiPattern}}
## 反模式

**反模式名称**：{{antiPatternName}}
**反模式ID**：{{antiPatternId}}

**问题所在**：
{{antiPatternProblem}}

**避免**：
```typescript
// 不要这样做
{{badExample}}
```
{{/if}}

## 经验教训

### 出了什么问题

- <!-- 错误1 -->
- <!-- 错误2 -->

### 学到了什么

- <!-- 经验1 -->
- <!-- 经验2 -->

### 预防措施

1. <!-- 措施1 -->
2. <!-- 措施2 -->
