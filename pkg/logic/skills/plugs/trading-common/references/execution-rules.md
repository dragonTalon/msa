# 交易执行规范

---

## Phase 2 风险管理规则 (v2.3.0)

在执行任何交易前，加载并执行以下规则:

### 买入前附加检查
- 加载 `trading-common/references/correlation-checklist.md` 执行行业集中度检查
- 加载 `trading-common/references/risk-management.md` 检查回撤熔断状态
- 如熔断触发（日内回撤 >5%、周内 >10%、月内 >20%）→ 拒绝开仓
- 加载 `trading-common/references/operation-types.md` 确定操作类型

### 卖出前附加检查
- 加载 `trading-common/references/risk-management.md` 计算 ATR 止损/止盈位
- 对盈利持仓更新移动止损位
- 加载 `trading-common/references/operation-types.md` 确定卖出操作类型

### 仓位计算
- 参考 `trading-common/references/operation-types.md` 的置信度-仓位联动表
- 凯利公式仓位 vs 市场状态仓位限制 → 取较小值

---

## 执行前检查清单

在执行任何交易前，必须完成以下检查：

### 买入前检查

```
□ 调用 get_account_summary 确认可用余额
□ 调用 get_stock_quote 获取当前价格
□ 计算所需金额：数量 × 价格 + 手续费
□ 确认金额充足，否则放弃操作
□ 检查股票涨幅是否超过追高限制（50%）
□ 确认买入数量为 100 的整数倍
```

### 卖出前检查

```
□ 调用 get_positions 确认持仓数量
□ 调用 get_stock_quote 获取当前价格
□ 确认卖出数量 <= 持仓数量
□ 确认卖出理由是否成立
```

---

## 买入执行流程

```
1. get_account_summary → 获取 available_amt
2. get_stock_quote → 获取 current_price
3. 计算最大可买数量 = floor(available_amt / (price * 100)) * 100  # 整手
4. submit_buy_order(stock_code, stock_name, quantity, price, fee)
```

### 示例调用

```
submit_buy_order(
  stock_code: "sz000858",
  stock_name: "五粮液",
  quantity: 100,
  price: 142.00,
  fee: 5.0
)
```

---

## 卖出执行流程

```
1. get_positions → 获取持仓数量
2. get_stock_quote → 获取 current_price
3. 确认卖出数量 <= 持仓数量
4. submit_sell_order(stock_code, stock_name, quantity, price, fee)
```

### 示例调用

```
submit_sell_order(
  stock_code: "sh600519",
  stock_name: "贵州茅台",
  quantity: 50,
  price: 1850.00,
  fee: 5.0
)
```

---

## 错误处理

### 常见错误及处理

| 错误类型 | 处理方式 |
|----------|----------|
| 余额不足 | 放弃买入，输出提示 |
| 持仓不足 | 放弃卖出，输出提示 |
| 价格获取失败 | 暂停操作，提示用户 |
| 下单失败 | 记录错误，不重试 |

### 手续费

- 默认手续费：5.0 元/笔
- 在计算所需金额时必须加上手续费
