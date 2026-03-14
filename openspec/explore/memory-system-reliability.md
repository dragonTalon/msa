# 记忆系统探索：可靠性、容错和资源控制

> **探索日期**：2026-03-13
> **探索者**：Claude
> **上下文**：add-memory-system change 设计阶段

## 概述

本次探索深入讨论了记忆系统的三个核心可靠性问题：
1. **文件损坏恢复** - JSON 损坏如何恢复数据？
2. **错误重试机制** - AI 提取失败如何处理？
3. **内存控制** - 知识缓存如何限制大小？

---

## 1. 知识冲突处理

### 问题场景

用户的偏好、知识、理解会随时间演变，系统需要处理：
- 投资风格演变（保守 → 激进 → 保守）
- 概念理解加深（30% → 70% → 95%）
- 策略优化（v1 → v2 → v3）

### 设计策略：分类处理

不同类型知识使用不同处理策略：

| 知识类型 | 处理方式 | 理由 |
|---------|---------|------|
| **用户画像** | 覆盖模式 | 当前偏好最重要，历史无意义 |
| **学过概念** | 保留模式 | 可以看到学习进度（confidence 递增） |
| **交易策略** | 主记录+历史 | 策略可演变，历史版本可回滚 |
| **常见问答** | 覆盖模式 | 答案应该是最新的理解 |

### 数据模型扩展

```go
type UserProfileSnapshot struct {
    ID              string    `json:"id"`
    ValidFrom       time.Time `json:"validFrom"`       // 生效时间
    ValidTo         time.Time `json:"validTo"`         // 失效时间（零=当前）

    // 元数据
    Source          string   `json:"source"`          // "ai-extracted" | "user-edited"
    Confidence      float64  `json:"confidence"`      // AI 提取的置信度
    IsMajorChange   bool     `json:"isMajorChange"`   // 是否重大变化
    PreviousProfile string   `json:"previousProfile"` // 上一个版本 ID
}

type ConceptLearned struct {
    ID          string             `json:"id"`
    Confidence  float64            `json:"confidence"`      // 当前理解程度
    History     []ConfidenceHistory `json:"history"`        // 理解程度变化
}

type ConfidenceHistory struct {
    Confidence float64  `json:"confidence"`
    ChangedAt  time.Time `json:"changedAt"`
    Reason     string    `json:"reason"`     // "首次询问" "成功应用" "复习"
    Source     string    `json:"source"`     // "ai-assessed" | "user-confirmed"
}
```

### 关键特性

1. **最小有效期（7天）**：避免记录短期波动
2. **相似度检测**：重大变化时标记提示
3. **学习进度可视化**：显示 30% → 95% 的提升
4. **手动编辑机制**：允许用户纠正 AI 提取错误

### 用户界面：知识演变可视化

```
📈 学习进度
100% █████████████████████████████████████  3/20
 70% ████████████████████████████         3/10
 30% ████████████                               3/1

📅 学习历史
• 3/1   首次询问 "什么是MACD？"              [理解: 30%]
• 3/10  学习金叉死叉 "MACD金叉是什么意思？"   [理解: 70%]
• 3/20  成功应用 "MACD显示金叉，我买入了"     [理解: 95%]
```

---

## 2. 文件损坏恢复

### 损坏类型分析

| 损坏类型 | 检测方式 | 可恢复性 | 恢复难度 |
|---------|---------|---------|---------|
| JSON 语法错误 | json.Unmarshal | 高 | 低 - 修复语法 |
| 文件截断 | 检查文件大小 | 中 | 中 - 可能恢复部分 |
| 数据类型错误 | schema 验证 | 高 | 中 - 记录错误字段 |
| 磁盘坏块 | 读取 I/O 错误 | 低 | 高 - 需要专业工具 |
| 索引损坏 | 索引重建 | 高 | 低 - 重新扫描 |

### 恢复策略

```go
// 自动修复常见错误
1. 缺少闭合括号 → 自动添加
2. 尾部有逗号 → 自动移除
3. 格式化美化 → 重新格式化

// 部分解析（JSON streaming）
1. 使用 json.Decoder 逐 token 解析
2. 恢复已解析的消息
3. 丢弃损坏部分

// 无法自动修复
1. 移动到 `.corrupt/` 目录
2. 保留原始内容供手动恢复
3. 提供 `/remember repair` 命令
```

### 损坏报告

```go
type CorruptionReport struct {
    File        string
    Severity    CorruptionSeverity  // Low/Medium/High/Critical
    Error       error
    Recoverable bool
    Recovery    RecoveryAction
}

type RecoveryAction struct {
    Type        string  // "skip" | "repair" | "partial" | "backup"
    Description string
    Execute     func() error
}
```

### 用户界面

```
⚠️ 记忆系统完整性检查

扫描结果:
✓ 127 个会话文件正常
⚠️  2 个文件损坏

📄 sessions/2024-03/def456.json
  严重程度: 中等
  错误: JSON 格式错误
  恢复: 自动添加闭合括号
  [r] 自动修复  [s] 跳过  [b] 备份后手动编辑
```

---

## 3. 错误重试机制

### 失败场景

| 场景 | 策略 |
|------|------|
| 网络超时 | 指数退避重试 2 次 |
| API 限流 | 等待 10s 后重试 |
| 无效响应 | 使用降级策略 |
| 内容过少 | 跳过提取 |

### 重试配置

```go
const (
    DefaultMaxAttempts   = 2
    DefaultInitialDelay  = 5 * time.Second
    DefaultMaxDelay      = 30 * time.Second
    DefaultBackoffFactor = 2.0
)
```

### 降级策略

```go
// AI 提取完全失败时，使用规则提取
func fallbackExtraction(session *model.Session) *ExtractionResult {
    // 规则 1: 提取股票代码
    if symbols := extractSymbols(session); len(symbols) > 0 {
        // 创建关注列表知识
    }

    // 规则 2: 检测技术指标关键词
    if keywords := extractTechnicalIndicators(session); len(keywords) > 0 {
        // 创建概念知识（低置信度）
    }

    return &ExtractionResult{
        Strategy: "fallback",
        Knowledges: knowledges,
    }
}
```

### 待重试队列

```go
type PendingExtraction struct {
    SessionID   string
    CreatedAt   time.Time
    Attempts   int
    NextRetry  time.Time
}

// 1 小时后重试
// 最多重试 3 次
// 放弃后记录日志
```

### 管理命令

```bash
/remember pending        # 查看待重试队列
/remember retry <id>      # 手动重试
/remember pending clear  # 清空队列
```

---

## 4. 内存控制

### 问题：知识缓存无限增长

```
时间    概念数    缓存大小    查询时间
───────────────────────────────────────
第1月    10 个     50 KB      10 ms
第3月    200 个    1 MB       200 ms
第6月    1000 个   5 MB       1 s
第1年    10000 个  50 MB      5 s   ← 启动变慢
```

### 控制策略

```go
const (
    MaxConcepts    = 100   // 最多缓存 100 个概念
    MaxWatchlist   = 20    // 最多缓存 20 只股票
    MaxStrategies  = 10    // 最多缓存 10 个策略
    MaxQAndA       = 50    // 最多缓存 50 个问答

    DefaultCacheTTL = 30 * 24 * time.Hour  // 默认 30 天过期
)
```

### 淘汰策略

1. **固定知识**（Pinned）：用户标记的重要知识
2. **高优先级**：最近访问
3. **中优先级**：访问频率高
4. **低优先级**：长期未访问

```go
func findEvictionCandidates(items []*CachedKnowledge, count int) []*CachedKnowledge {
    sort.Slice(items, func(i, j int) bool {
        // 优先级低的先淘汰
        if items[i].Priority != items[j].Priority {
            return priorityToScore(items[i].Priority) < priorityToScore(items[j].Priority)
        }

        // 固定的不淘汰
        if items[i].Pinned {
            return false
        }

        // 最后访问时间早的先淘汰（LRU）
        return items[i].LastAccessed.Before(items[j].LastAccessed)
    })
}
```

### 管理命令

```bash
/remember cache stats      # 查看缓存状态
/remember cache prune      # 清理过期知识
/remember cache pin <id>   # 固定重要知识
/remember cache unpin <id> # 取消固定
```

### 用户界面

```
💾 知识缓存管理

📊 缓存统计
  概念: 45/100 (45%)
  ████████████████████████████████░░░░░░░░░░░░░░░░░░░░░░░░

  关注列表: 12/20 (60%)
  ████████████████████████████████████████░░░░░░░░░░░░░

总内存: 2.3 MB  |  固定: 3 个  | 待清理: 5 个

⚠️ 待清理的知识（30天内未访问）
  • KDJ 指标 - 最后访问: 2024-02-10
  • RSI 指标 - 最后访问: 2024-02-15
  • Bollinger 带 - 最后访问: 2024-02-20

🔒 固定知识（不会被清理）
  • MACD 指标 - 2024-03-10 固定
  • 用户画像 - 2024-03-12 固定
  • 腾讯控股 - 2024-03-12 固定

[p] 清理过期  [r] 手动固定  [c] 配置限制
```

---

## 关键设计决策

### 文件损坏恢复

| 方案 | 优点 | 缺点 | 选择 |
|------|------|------|------|
| 跳过损坏文件 | 简单、安全 | 丢失数据 | 作为默认 |
| 自动修复 | 恢复数据 | 可能产生错误数据 | 尝试，但备份 |
| 部分解析 | 最大化恢复 | 复杂度高 | 作为 fallback |
| 手动恢复 | 用户控制 | 需要用户介入 | 提供命令 |

### 错误重试

| 方案 | 适用场景 | 重试次数 | 延迟策略 |
|------|---------|---------|---------|
| 指数退避 | 网络超时 | 2 次 | 5s → 10s |
| 固定延迟 | API 限流 | 3 次 | 10s 固定 |
| 不重试 | 永久错误 | 0 | - |
| 降级策略 | 所有失败 | - | 规则提取 |

### 内存控制

| 方案 | 实现难度 | 用户体验 | 选择 |
|------|---------|---------|------|
| 硬限制（100个） | 简单 | 无法扩展 | ✓ 作为默认 |
| TTL 自动过期 | 中等 | 平衡 | ✓ 启用 |
| 用户固定 | 简单 | 灵活 | ✓ 提供 |
| LRU 淘汰 | 复杂 | 自动化 | ✓ 作为 fallback |

---

## 待深入的问题

以下问题可能在实现时遇到，需要进一步探索：

1. **并发安全**：索引的原子性更新
2. **数据一致性**：会话和知识的关联
3. **启动性能**：大量会话时的优化
4. **搜索算法**：相关性评分的细节

---

## 相关文档

- Proposal: `/openspec/changes/add-memory-system/proposal.md`
- Specs: `/openspec/changes/add-memory-system/specs/**/*.md`
- Design: `/openspec/changes/add-memory-system/design.md`
- Tasks: `/openspec/changes/add-memory-system/tasks.md`
- 记忆设计文档: `/Users/dragon/.claude/projects/-Users-dragon-Documents-project-msa/memory/memory-system-design.md`

---

**探索模式总结**：

这次探索深入讨论了记忆系统的三个核心可靠性问题，设计了：
1. **分类处理的知识冲突机制** - 根据知识类型选择不同的处理策略
2. **多层次的文件恢复策略** - 从自动修复到手动恢复的完整体系
3. **智能的错误重试和降级机制** - 确保知识提取的可靠性
4. **可控的内存管理策略** - 通过限制、TTL、固定和 LRU 平衡性能

这些设计决策确保了记忆系统在实际使用中的健壮性和可靠性。
