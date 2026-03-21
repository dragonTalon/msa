---
name: read-error
description: 读取历史错误操作经验，为后续决策提供风险警示。**必须**在早盘、午盘、闭市 SKILL 执行前强制调用。
version: 2.0.0
priority: 10
pattern: reviewer
tools:
  - list_knowledge_files
  - read_knowledge_file
---

# Read Error - 错误经验读取器

## ⚠️ 强制要求

**此 SKILL 必须作为以下 SKILL 的第一步调用：**
- `morning-analysis`（早盘分析）
- `afternoon-trade`（午盘操作）
- `market-close-summary`（闭市总结）

**不执行此 SKILL 将导致：**
- 重复犯同样的错误
- 缺少历史经验的指导
- 决策缺少风险意识

---

## 执行流程

### Step 1: 扫描错误文件

```
调用 list_knowledge_files(directory="errors")
→ 获取所有错误记录文件列表
```

如果目录不存在或无文件：
- 输出："暂无历史错误记录，继续保持良好操作习惯"
- 不影响后续流程执行

### Step 2: 读取并解析

对每个文件：
1. 调用 `read_knowledge_file` 读取内容
2. 解析 YAML frontmatter（id, date, severity, tags, status）
3. 提取错误描述、原因、正确做法、防范措施
4. 按 severity 排序（critical > high > medium > low）
5. 过滤 status=resolved 的记录（已解决的不再提醒）

### Step 3: 格式化输出

将错误经验整理成结构化输出，供后续决策参考。

---

## 输出格式

```markdown
## ⚠️ 历史错误经验提醒

本次操作前请注意以下历史教训：

### [严重程度] 错误标题
- 发生时间：YYYY-MM-DD
- 错误描述：具体错误内容
- 正确做法：应该怎么做
- 防范措施：如何避免

---

以上经验请在决策时充分考虑，避免重蹈覆辙。
```

---

## 在其他 SKILL 中的调用方式

```
在执行任何分析前，首先调用此 SKILL：

1. 调用 read-error SKILL 读取所有错误记录
2. 将输出内容拼接到后续分析的 prompt 中
3. 在决策时参考历史错误，避免重复犯错
```
