---
name: output-formats
description: 标准化输出格式模板库。当生成投资建议、分析报告、账户总结等结构化输出时使用。
version: 2.0.0
priority: 8
pattern: generator
output-format: markdown
---

# Output Formats - 输出格式模板库

## 概述

此 SKILL 提供 JSON 格式的输出模板，用于生成结构化的分析报告。

**使用方式：** 在需要生成报告时，加载 `assets/` 目录下的模板文件，填充数据后输出。

---

## 可用的模板

| 文件 | 用途 | 使用场景 |
|------|------|----------|
| `investment-advice.md` | 投资建议模板 | 给出买入/卖出建议时 |
| `analysis-report.md` | 分析报告模板 | 完整股票分析时 |
| `account-report.md` | 账户报告模板 | 展示账户状态时 |

---

## 使用流程

```
Step 1: 根据场景选择合适的模板
Step 2: 加载模板文件
Step 3: 根据分析结果填充模板中的占位符
Step 4: 输出完整的 Markdown 文档
```

---

## 模板使用注意事项

1. 模板中的 `[xxx]` 为必须填充的占位符
2. 每个部分都必须填充，不能跳过
3. 风险提示部分必须包含至少 2 条风险
