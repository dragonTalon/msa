# 设计：enhance-trading-system

## 上下文

当前交易系统依赖 AI 模型手动调用工具来完成状态传递和错误记录，但由于模型行为不可预测，经常遗漏关键步骤：

1. **Session 标签问题**：`add_session_tag` 工具调用经常被遗漏，导致午盘无法通过标签查询早盘 Session
2. **错误记录问题**：收盘时模型经常跳过错误记录步骤，导致错误无法沉淀
3. **仓位管理问题**：SKILL 中没有仓位限制，模型直接计算"最大可买数量"导致 ALL IN

**当前状态**：
- Session 标签由 `manager.go` 中的 `AddSessionTag()` 方法管理
- 买入逻辑在 SKILL 中定义，使用 `floor(可用余额 / (价格 × 100)) × 100` 计算
- 错误记录需要手动调用 `write_knowledge_file` 工具

**约束**：
- 保持向后兼容，不破坏现有 SKILL 流程
- 遵循 Go 标准命名规范
- 使用 logrus 结构化日志

---

## 目标 / 非目标

**目标：**
- 在 Session 初始化时自动添加时间相关标签
- 新增交易分析工具辅助错误识别
- 在 SKILL 中定义清晰的仓位管理策略
- 强化错误记录的 SKILL 指导

**非目标：**
- 不在代码层面强制执行仓位限制（由 SKILL 指导模型决策）
- 不修改现有的交易执行工具（`submit_buy_order`, `submit_sell_order`）
- 不引入新的数据库表或持久化结构

---

## 推荐模式

### 安全优先配置原则

**模式ID**：p005
**来源**：知识库 Pattern #005

**描述**：
在涉及资金操作的系统中，采用保守的默认配置，确保即使出错也不会造成重大损失。

**适用场景**：
- 仓位比例计算
- 止损阈值设定
- 买入限制参数

**实现**：
```go
// 仓位管理参数 - 保守默认值
const (
    MaxPositionPercent     = 0.30  // 单只股票最大仓位 30%
    MaxSingleBuyPercent    = 0.50  // 单次买入上限 50% 计划仓位
    MaxDailyBuyPercent     = 0.70  // 同日累计买入上限 70%
    StopLossPercent        = 0.20  // 止损阈值 20%
    MaxChaseGainPercent    = 0.50  // 追高限制 50%
)
```

### 错误处理和重试策略

**模式ID**：p003
**来源**：知识库 Pattern #003

**描述**：
在分析交易时，需要优雅地处理数据获取失败的情况，确保工具不会因部分失败而完全中断。

**适用场景**：
- analyze_today_trades 工具获取交易记录
- 获取 Session 数据
- 计算仓位比例

**实现**：
```go
func AnalyzeTodayTrades(ctx context.Context) (*AnalysisResult, error) {
    // 获取交易记录，失败时返回空列表继续
    transactions, err := getTransactions(today)
    if err != nil {
        log.Warnf("获取交易记录失败: %v，使用空列表继续", err)
        transactions = []Transaction{}
    }

    // 获取 Session，失败时返回空列表继续
    sessions, err := querySessionsByDate(today)
    if err != nil {
        log.Warnf("获取 Session 失败: %v，使用空列表继续", err)
        sessions = []Session{}
    }

    // 继续分析...
}
```

---

## 避免的反模式

### 硬编码时间判断

**问题所在**：
在 Session 初始化时硬编码时间范围，可能导致时区问题或夏令时问题。

**影响**：
用户在不同时区使用系统时，Session 标签可能不正确。

**不要这样做**：
```go
// 不要这样做 - 硬编码小时判断
if hour >= 9 && hour <= 11 {
    tag = "morning-session"
}
```

**应该这样做**：
```go
// 应该这样做 - 使用配置或常量，并考虑分钟
const (
    MorningStart   = "09:30"
    MorningEnd     = "11:30"
    AfternoonStart = "13:00"
    AfternoonEnd   = "14:30"
    CloseStart     = "16:00"
)

func getAutoTag(now time.Time) string {
    currentTime := now.Format("15:04")

    switch {
    case currentTime >= MorningStart && currentTime <= MorningEnd:
        return "morning-session"
    case currentTime >= AfternoonStart && currentTime <= AfternoonEnd:
        return "afternoon-session"
    case currentTime >= CloseStart:
        return "close-session"
    default:
        return ""
    }
}
```

### 工具返回错误信息过于简单

**问题所在**：
analyze_today_trades 工具如果只返回"有错误"或"无错误"，模型无法理解具体是什么问题。

**影响**：
模型无法准确记录错误或给出建议。

**不要这样做**：
```go
// 不要这样做 - 只返回简单结果
return "发现 2 个潜在错误"
```

**应该这样做**：
```go
// 应该这样做 - 返回结构化详细信息
type PotentialError struct {
    Type        string `json:"type"`         // 错误类型
    StockCode   string `json:"stock_code"`   // 股票代码
    StockName   string `json:"stock_name"`   // 股票名称
    Detail      string `json:"detail"`       // 详情描述
    Suggestion  string `json:"suggestion"`   // 建议操作
}

type AnalysisResult struct {
    Date         string           `json:"date"`
    TotalTrades  int              `json:"total_trades"`
    Errors       []PotentialError `json:"errors"`
    ErrorTemplate string          `json:"error_template"` // 可直接写入 errors/ 的模板
}
```

---

## 设计决策

### 决策 1：Session 标签自动添加位置

**选择**：在 `Manager.Initialize()` 方法中添加

**理由**：
- Session 创建时就需要标签，Initialize 是最早的时机
- 避免遗忘，不需要依赖外部调用
- 与现有 `AddSessionTag()` 方法兼容，自动添加的标签也可以被后续修改

**替代方案**：在 SKILL 中强制调用
- ❌ 依赖模型行为，不可靠
- ❌ 已被证明存在问题

### 决策 2：仓位策略实现位置

**选择**：在 SKILL 文档中定义，不在代码中强制执行

**理由**：
- 仓位策略需要灵活性，不同用户有不同偏好
- SKILL 可以被用户自定义修改
- 模型可以根据市场情况选择不同策略

**替代方案**：在代码中强制执行
- ❌ 过于僵化，无法适应不同投资风格
- ❌ 需要大量配置参数

### 决策 3：analyze_today_trades 工具职责

**选择**：只分析，不自动写入

**理由**：
- 错误记录需要人工判断是否真的是错误
- 保持模型决策权
- 工具只提供信息，写入由模型决定

**替代方案**：自动写入 errors/
- ❌ 可能误判
- ❌ 剥夺模型决策权

---

## 风险 / 权衡

| 风险 | 缓解措施 |
|------|----------|
| 时间判断可能受时区影响 | 使用本地时间，格式化为 "15:04" 比较 |
| 模型可能忽略 SKILL 中的仓位策略 | 强化 SKILL 文档，增加检查清单 |
| analyze_today_trades 可能误判 | 只标记为"潜在错误"，由模型判断 |
| 自动标签可能与手动标签冲突 | 使用幂等检查，避免重复 |

---

## 迁移计划

### 阶段 1：代码修改（无破坏性变更）

1. 修改 `pkg/logic/memory/manager.go`：
   - 添加 `getAutoTag()` 辅助函数
   - 在 `Initialize()` 中调用自动标签逻辑

2. 新增 `pkg/logic/tools/knowledge/analyze.go`：
   - 实现 `AnalyzeTodayTradesTool`
   - 注册到 knowledge 工具组

### 阶段 2：SKILL 文档更新

1. 更新 `trading-common/SKILL.md`：
   - 新增仓位管理章节
   - 新增分段买入/卖出策略

2. 更新 `morning-analysis/SKILL.md`：
   - 移除手动标签说明
   - 引用 trading-common 的仓位策略

3. 更新 `afternoon-trade/SKILL.md`：
   - 移除手动标签说明
   - 强化早盘 Session 读取说明

4. 更新 `market-close-summary/SKILL.md`：
   - 移除手动标签说明
   - 新增错误检查清单
   - 引用 analyze_today_trades 工具

### 回滚策略

- 代码变更：直接回滚 git commit
- SKILL 变更：SKILL 是纯文档，不影响代码运行
