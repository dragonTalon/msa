---
name: trading-common
description: 交易系统公共定义，包含数据结构、文件格式、工具调用规范。其他交易相关SKILL应参考此文件。
version: 1.0.0
priority: 9
---

# Trading Common Definitions

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

## 文件格式

### 用户笔记 (notes/YYYY-MM-DD.md)

```markdown
# YYYY-MM-DD 笔记

## 关注股票
- [股票代码] [股票名称]：关注原因

## 市场观点
- 观点内容

## 操作计划
- 计划操作内容
```

### 闭市总结 (summaries/YYYY-MM-DD.md)

```markdown
---
date: YYYY-MM-DD
session_ids: [session-id-1, session-id-2, session-id-3]
---

# YYYY-MM-DD 交易日总结

## 今日操作
| 时间 | 操作 | 股票 | 数量 | 价格 | 原因 |
|------|------|------|------|------|------|
| 早盘 | 买入 | XXX  | 100  | 10.5 | xxx  |

## 今日收益
- 总资产变化：+/- xxx 元
- 收益率：+/- x.xx%
- 盈亏股票：[股票: +/- xxx元]

## 操作复盘
### 正确操作
- 操作描述及原因分析

### 待改进操作
- 操作描述及改进方向

## 明日关注
- 关注要点
```

### 错误记录 (errors/*.md)

```markdown
---
id: err-XXX
date: YYYY-MM-DD
severity: high | medium | low
tags: [tag1, tag2]
status: active | resolved
---

## 错误描述
[具体发生了什么错误]

## 错误原因
[为什么会犯这个错误]

## 正确做法
[应该怎么做]

## 防范措施
[以后如何避免]

## 相关案例
[具体案例描述]
```

## 交易执行规范

### 执行前检查
1. 调用 `get_account_summary` 确认可用余额
2. 调用 `get_stock_quote` 获取当前价格
3. 计算所需金额：数量 × 价格 + 手续费
4. 确认金额充足，否则放弃操作

### 买入执行流程
```
1. get_account_summary → 获取 available_amt
2. get_stock_quote → 获取 current_price
3. 计算最大可买数量 = floor(available_amt / (price * 100)) * 100  # 整手
4. submit_buy_order(stock_code, stock_name, quantity, price, fee)
```

### 卖出执行流程
```
1. get_positions → 获取持仓数量
2. get_stock_quote → 获取 current_price
3. 确认卖出数量 <= 持仓数量
4. submit_sell_order(stock_code, stock_name, quantity, price, fee)
```

## Session 标签约定

用于区分不同时间段的会话：
- `morning-session` - 早盘会话
- `afternoon-session` - 午盘会话
- `close-session` - 闭市会话

## 约束条件

1. **禁止透支消费**：可用余额不足时必须放弃买入操作
2. **禁止卖空**：卖出数量不能超过持仓数量
3. **整手交易**：买入数量必须是100的整数倍
4. **执行不确认**：输出决策后直接执行，不需要用户确认