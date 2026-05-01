# 决策日志格式

## 概述

每个决策的推理链需以 JSON 格式写入 Session，供复盘时追溯决策逻辑。此格式与 decision-reasoning-template.md 的 6 段推理链一一对应。

---

## JSON Schema

```json
{
  "decision_id": "dec-YYYYMMDD-HHMMSS-XXX",
  "timestamp": "2026-05-01T09:45:00+08:00",
  "session": "morning-session",
  "decision_type": "buy | sell | hold | watch",
  "confidence": "high | medium | low",
  "status": "executed | pending | cancelled",

  "trigger": {
    "type": "stop_profit | stop_loss | technical_signal | news_driven | strategy_rule | periodic_check",
    "source": "",
    "strength": "strong | medium | weak"
  },

  "reasoning": {
    "supporting": [
      {"reason": "", "weight": "high | medium | low"}
    ],
    "opposing": [
      {"reason": "", "weight": "high | medium | low"}
    ]
  },

  "counter_check": {
    "fomo_check": true,
    "emotion_check": true,
    "independent_judgment": true,
    "worst_case": "",
    "counter_search_done": true
  },

  "info_quality": {
    "source_credibility": "high | medium | low",
    "timeliness": "fresh | acceptable | priced_in",
    "crowding": "normal | watch | crowded | overcrowded",
    "completeness": "sufficient | acceptable | insufficient",
    "verdict": "proceed | supplement | suspend"
  },

  "market_regime": {
    "regime": "STRONG_BULL | BULL | WEAK_BULL | RANGING_BULLISH | RANGING_BEARISH | WEAK_BEAR | BEAR | STRONG_BEAR",
    "trend_score": 0,
    "volatility_score": 0,
    "sentiment_score": 0,
    "position_cap_pct": 0,
    "chase_limit_pct": 0
  },

  "order": {
    "stock_code": "",
    "stock_name": "",
    "quantity": 0,
    "price": 0.00,
    "fee": 5.0
  },

  "expected": {
    "target_return_pct": 0,
    "stop_loss_price": 0.00,
    "stop_loss_pct": 0,
    "target_price": 0.00,
    "holding_period_days": 0
  },

  "actual": {
    "executed_at": "",
    "executed_price": 0.00,
    "executed_quantity": 0
  }
}
```

---

## 字段说明

### 基础字段
| 字段 | 类型 | 说明 |
|------|------|------|
| decision_id | string | 唯一标识，格式: dec-YYYYMMDD-HHMMSS-序号 |
| timestamp | ISO 8601 | 决策时间 |
| session | enum | 所属会话 |
| decision_type | enum | 决策类型 |
| confidence | enum | 置信度 |
| status | enum | 执行状态 |

### trigger - 对应推理链第1段
| 字段 | 说明 |
|------|------|
| type | 触发类型枚举 |
| source | 具体触发来源描述 |
| strength | 触发强度 |

### reasoning - 对应推理链第2段
| 字段 | 说明 |
|------|------|
| supporting | 支持理由列表，每条有原因和权重 |
| opposing | 反对理由列表 |

### counter_check - 对应推理链第3段
| 字段 | 说明 |
|------|------|
| fomo_check | 是否通过 FOMO 检查 |
| emotion_check | 是否通过情绪驱动检查 |
| independent_judgment | 是否独立判断 |
| worst_case | 最坏情况描述 |
| counter_search_done | 是否完成反向搜索 |

### info_quality - 对应推理链第4段
直接引用 news-quality-gate.md 的输出。

### market_regime - 对应推理链第5段
直接引用 market-regime-classifier.md 的输出。

### order - 操作详情
执行时填写的订单信息。

### expected - 预期结果
| 字段 | 说明 |
|------|------|
| target_return_pct | 预期收益率 |
| stop_loss_price | 止损价位 |
| stop_loss_pct | 止损比例 |
| target_price | 目标价位 |
| holding_period_days | 预期持仓天数 |

### actual - 实际执行结果
执行后回填。

---

## 写入方式

决策日志写入当前 Session：

```
在 Session 中记录:
  key: "decision-log"
  value: JSON 字符串
```

每个决策追加到 Session 的决策日志列表中。

---

## 复盘使用

market-close-summary 在复盘时：
1. 从 Session 中提取所有 decision-log
2. 逐条对比决策逻辑与实际结果
3. 评估推理质量（见 reasoning-review-criteria.md）
4. 区分"正确决策运气不好"vs"错误决策运气好"
