---
name: trading-common
description: 交易系统公共定义，包含数据结构、文件格式、工具调用规范。其他交易相关 SKILL 应参考此文件的 references/ 目录。
version: 2.0.0
priority: 9
pattern: tool-wrapper
tools:
  - get_account_summary
  - get_positions
  - get_stock_quote
  - submit_buy_order
  - submit_sell_order
dependencies: []
---

# Trading Common - 交易系统公共定义

## 概述

此 SKILL 是交易系统的基础知识库，提供：
- 目录结构和文件格式规范
- 交易执行流程
- 仓位管理规则

**使用方式：** 其他交易 SKILL 应在需要时加载 `references/` 目录下的规则文件。

---

## 目录结构

```
.msa/knowledge/
├── user-profile/           # 用户画像
│   └── profile.md
├── strategies/             # 交易策略
│   └── *.md
├── insights/               # 市场洞察
│   └── *.md
├── notes/                  # 用户笔记
│   └── YYYY-MM-DD.md
├── summaries/              # 闭市总结
│   └── YYYY-MM-DD.md
└── errors/                 # 错误记录
    └── *.md
```

---

## Session 标签约定

用于区分不同时间段的会话：
- `morning-session` - 早盘会话（9:30-11:30 自动添加）
- `afternoon-session` - 午盘会话（13:00-14:30 自动添加）
- `close-session` - 闭市会话（16:00后 自动添加）

**注意**：Session 标签由系统自动添加，无需手动操作。

---

## 可用的 References

当执行交易相关操作时，按需加载以下文件：

| 文件 | 用途 | 加载时机 |
|------|------|----------|
| `position-rules.md` | 仓位管理规则 | 计算买入数量时 |
| `execution-rules.md` | 交易执行规范 | 执行买卖操作前 |
| `file-formats.md` | 文件格式定义 | 读写知识文件时 |

---

## 约束条件

1. **禁止透支消费**：可用余额不足时必须放弃买入操作
2. **禁止卖空**：卖出数量不能超过持仓数量
3. **整手交易**：买入数量必须是 100 的整数倍
4. **分段买入**：首次买入不超过计划仓位的 50%
5. **追高限制**：涨幅超过 50% 不买入
