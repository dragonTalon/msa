# MSA 记忆系统设计文档

> 设计日期：2026-03-13
> 状态：探索阶段

## 概述

为 MSA (My Stock Agent) CLI 工具设计一个持久化的对话记忆系统，支持：
- 跨会话上下文：下次启动时能继续之前的对话
- 分析回顾：查看历史对话记录
- AI 上下文增强：让 AI 能"记住"之前的交互
- 调试/审计：用于问题排查

## 核心设计理念

### 与 OpenSpec 的类比

```
OpenSpec                  记忆系统
┌─────────┐               ┌─────────┐
│  spec   │  ←→           │ 中期记忆 │  对话总结/知识
│ (规范)  │               │ (summary)│  "我学到了什么"
└─────────┘               └─────────┘
     │                         │
┌─────────┐               ┌─────────┐
│ change  │  ←→           │ 长期记忆 │  完整会话
│ (变更)  │               │ (session)│  "发生了什么"
└─────────┘               └─────────┘
```

- **中期记忆** = spec（知识沉淀）：从会话中提取的结构化知识
- **长期记忆** = change（完整记录）：完整的对话会话

### 记忆分类

1. **短期记忆**：运行时，前10条对话（内存）
2. **中期记忆**：对话信息总结（JSON 文件，按类型存储）
3. **长期记忆**：完整会话记录（JSON 文件，按日期存储）

### 知识创建策略

**总是创建新的**（不合并）：
- 保留完整历史
- 无冲突风险
- 可追溯每个知识的来源会话
- 查询时动态聚合

## 文件组织结构

```
~/.msa/remember/
├── sessions/                          # 长期记忆
│   ├── 2024-03/                      # 按年月组织
│   │   ├── abc123-def456.json        # 完整会话
│   │   └── index.json                # 月度索引
│   └── index.json                    # 全局索引
│
├── knowledge/                         # 中期记忆
│   ├── profile/                      # 用户画像
│   ├── watchlist/                    # 关注列表
│   ├── concepts/                     # 学过的概念
│   ├── strategies/                   # 交易策略
│   ├── qanda/                        # 问答
│   └── insights/                     # 洞察
│
├── short-term/                       # 短期记忆
│   └── current.json                  # 当前会话的最近消息
│
├── indexes/                          # 搜索索引
│   └── all.json                      # 全局索引
│
└── stats.json                        # 统计信息
```

## 1. 数据模型（全 JSON）

### 核心模型

```go
// Session 完整的对话会话（长期记忆）
type Session struct {
    ID           string
    StartTime    time.Time
    EndTime      time.Time
    MessageCount int
    Summary      string
    Tags         []string
    Context      *SessionContext
    Messages     []Message
    KnowledgeID  string
}

// SessionContext 会话上下文
type SessionContext struct {
    Provider    string
    Model       string
    Symbols     []string
    Commands    []string
    Tools       []string
    HasTrading  bool
    HasSearch   bool
}

// Knowledge 中期记忆
type Knowledge struct {
    ID          string
    SessionID   string
    GeneratedAt time.Time
    Type        KnowledgeType

    // 根据类型存储不同的数据
    UserProfile  *UserProfile
    Watchlist    *Watchlist
    Concept      *ConceptLearned
    Strategy     *TradingStrategy
    QAndA        *QAndAItem
    Insight      *Insight
}

// KnowledgeType 知识类型
type KnowledgeType string

const (
    KnowledgeTypeProfile   KnowledgeType = "profile"
    KnowledgeTypeWatchlist KnowledgeType = "watchlist"
    KnowledgeTypeConcept   KnowledgeType = "concept"
    KnowledgeTypeStrategy  KnowledgeType = "strategy"
    KnowledgeTypeQAndA     KnowledgeType = "qanda"
    KnowledgeTypeInsight   KnowledgeType = "insight"
)
```

### 知识数据结构

```go
// UserProfile 用户画像
type UserProfile struct {
    InvestmentStyle []string
    RiskPreference  string
    TimeHorizon     string
    ExperienceLevel string
    FocusAreas      []string
    UpdatedAt       time.Time
}

// ConceptLearned 学过的概念
type ConceptLearned struct {
    Name        string
    Category    string
    Description string
    Confidence  float64  // 0-1 理解程度
    Related     []string
    LearnedAt   time.Time
}

// QAndAItem 常见问答
type QAndAItem struct {
    Question     string
    Answer       string
    FirstAskedAt time.Time
    AskCount     int
    LastAskedAt  time.Time
}
```

## 2. TUI 视图设计

### 视图层次结构

```
Chat (聊天)
  │ /remember
  ▼
MemoryHome (主页)
  ├─ SessionList (会话列表) → SessionDetail (会话详情)
  ├─ KnowledgeList (知识列表) → KnowledgeDetail (知识详情)
  ├─ SearchView (搜索) → SearchResult
  └─ StatsView (统计)
```

### 主页视图

```
🧠 MSA 记忆系统

▶ [1] 📖 历史会话 (23)
  • 查看所有历史对话记录

  [2] 💡 知识库 (12)
  • 查看提取的知识和洞察

  [3] 🔍 搜索记忆
  • 按关键词、股票、日期搜索

  [4] 📊 统计信息
  • 查看使用统计和趋势

⌨️  ↑↓:选择  Enter:确认  1-4:快捷键  ESC:返回
```

### 会话列表视图

```
📖 历史会话

🔍 搜索: tencent_
────────────────────────────────────────────────────────────────

📍 位置: 3/23  |  已过滤: 5/23  |  显示: 1-10

▶ 1. 2024-03-10 10:30  [23条] 腾讯股价+MACD学习
     标签: 股票查询,技术指标,MACD  股票: 00700.HK

  2. 2024-03-09 15:20  [15条] 创建模拟交易账户
     标签: 账户管理

────────────────────────────────────────────────────────────────
⌨️  ↑↓:移动  PgUp/Dn:翻页  Enter:查看详情  d:删除  /:搜索
```

### 会话详情视图

```
📖 会话详情: 腾讯股价+MACD学习

📋 元数据
  • 时间: 2024-03-10 10:30 - 11:45 (1小时15分钟)
  • 消息: 23条
  • 模型: Qwen2.5-7B-Instruct
  • 股票: 00700.HK 腾讯控股
  • 标签: 股票查询,技术指标,MACD

💡 提取的知识
  • [概念] MACD - 异同移动平均线

────────────────────────────────────────────────────────────────

👤 你: 帮我查一下腾讯的股价
10:30:15

🤖 MSA: 腾讯控股(00700.HK)当前股价为365.20港元...

────────────────────────────────────────────────────────────────
⌨️  ↑↓:滚动  PgUp/Dn:翻页  /:搜索  g:开头  G:结尾  e:导出
```

### 知识列表视图

```
💡 知识库

[1] 👤 用户画像 (3)  [2] 📊 关注列表 (4)
[3] 📚 学过概念 (8)  [4] 📋 交易策略 (2)
[5] ❓ 常见问答 (5)  [6] 🔍 其他洞察 (6)

────────────────────────────────────────────────────────────────

📚 学过的概念 (8)
────────────────────────────────────────────────────────────────

▶ 1. MACD - 异同移动平均线                        [技术分析]
     学习于: 2024-03-10  理解程度: 80%

  2. KDJ - 随机指标                                   [技术分析]
     学习于: 2024-03-09  理解程度: 60%

────────────────────────────────────────────────────────────────
⌨️  1-6:切换类别  ↑↓:移动  Enter:查看详情  /:搜索
```

## 3. 存储层实现

### Store 接口

```go
type Store interface {
    // Session 操作
    SaveSession(session *model.Session) error
    GetSession(id string) (*model.Session, error)
    ListSessions(offset, limit int) ([]model.SessionIndex, error)
    DeleteSession(id string) error

    // Knowledge 操作
    SaveKnowledge(knowledge *model.Knowledge) error
    GetKnowledge(id string) (*model.Knowledge, error)
    ListKnowledgeByType(kType model.KnowledgeType) ([]model.Knowledge, error)
    GetAllKnowledge() (map[model.KnowledgeType][]model.Knowledge, error)

    // ShortTermMemory 操作
    SaveShortTermMemory(memory *model.ShortTermMemory) error
    GetShortTermMemory() (*model.ShortTermMemory, error)

    // 索引和搜索
    BuildIndexes() error
    Search(query string, filters SearchFilters) ([]SearchResult, error)

    // 统计
    GetStats() (*model.MemoryStats, error)
}
```

### FileStore 关键方法

1. **SaveSession**：按年月组织目录，保存完整会话，更新索引
2. **SaveKnowledge**：按类型分目录保存，更新知识索引
3. **Search**：基于索引的快速搜索，支持关键词、日期、股票代码过滤
4. **GetStats**：聚合统计信息

## 4. 与 chat.go 的集成点

### 初始化阶段

```go
func NewChat(ctx context.Context) *Chat {
    chat := &Chat{
        memoryManager: memory.GetManager(),
    }

    // 初始化新会话
    chat.memoryManager.Initialize()

    return chat
}
```

### 运行阶段

```go
// 用户输入时
func (c *Chat) handleEnterKey() {
    // 添加到历史
    c.history = append(c.history, userMessage)

    // 同时添加到记忆
    c.memoryManager.AddMessage(userMessage)
}

// AI 回复时注入知识
func BuildQueryMessages() {
    // 获取用户知识
    userKnowledge := memoryManager.GetKnowledgeForPrompt()

    // 注入到系统提示
    templateData["userKnowledge"] = userKnowledge
}
```

### 退出阶段

```go
func (c *Chat) Update(msg tea.Msg) {
    case tea.KeyMsg:
    case tea.KeyCtrlC, tea.KeyEsc:
        // 保存会话
        c.memoryManager.EndSession()
        // → 保存完整会话
        // → 异步调用 AI 提取知识
        // → 更新索引
}
```

## 5. AI Prompt 实现

### 知识提取 Prompt

会话结束时调用，从对话中提取结构化知识：

```prompt
你是一个专业的知识提取助手。请分析以下对话，提取关键信息。

# 对话内容
{conversation}

# 提取任务

## 1. 用户画像
- 投资风格、风险偏好、经验水平、关注领域

## 2. 关注列表
- 股票代码、名称、关注原因

## 3. 学过的概念
- 概念名称、分类、描述、理解程度(0-1)

## 4. 交易策略
- 策略名称、描述、使用条件、优缺点

## 5. 常见问答
- 问题、简短回答

# 输出格式

```json
[
  {
    "type": "concept",
    "concept": {
      "name": "MACD",
      "category": "技术分析指标",
      "description": "...",
      "confidence": 0.8,
      "learnedAt": "2024-03-10T12:00:00Z"
    }
  }
]
```
```

### 知识注入 Prompt

AI 回复时注入到系统提示：

```prompt
# 【记忆与上下文】

以下是关于用户的历史知识：

# 用户画像
- 投资风格: 价值投资，中长期持有
- 风险偏好: 稳健型
- 经验水平: 初级

# 关注列表
- 00700.HK 腾讯控股: 优质科技股，业绩稳定
- 600519 贵州茅台: 白酒龙头，品牌价值高

# 学过的概念
- MACD: 异同移动平均线，用于判断买卖时机 (理解程度: 80%)
- PE/PB: 估值指标 (理解程度: 90%)

请基于这些知识更好地理解用户的问题和偏好。
```

## 6. 完整流程图

```
启动阶段
1. NewChat()
   └─▶ memory.GetManager()
       └─▶ LoadKnowledge() - 加载知识缓存

运行阶段
1. 用户输入
   └─▶ memoryManager.AddMessage()
       ├─▶ 更新当前会话消息
       ├─▶ 更新短期记忆（最多10条）
       └─▶ 持久化 short-term/current.json

2. AI 回复
   └─▶ 注入知识到系统提示
       └─▶ 更个性化的回复

退出阶段
1. Ctrl+C / ESC
   └─▶ memoryManager.EndSession()
       ├─▶ 保存完整会话
       ├─▶ AI 提取知识（异步）
       │   └─▶ 保存到 knowledge/
       └─▶ 更新索引
```

## 新增文件列表

```
pkg/
├── logic/
│   └── memory/                    # 记忆系统逻辑
│       ├── store.go               # 存储抽象层和FileStore实现
│       ├── manager.go             # 记忆管理器
│       └── extraction.go          # AI知识提取
│
├── tui/
│   ├── memory_home.go             # 记忆主页
│   ├── memory_session_list.go     # 会话列表
│   ├── memory_session_detail.go   # 会话详情
│   ├── memory_knowledge_list.go   # 知识列表
│   ├── memory_search.go           # 搜索视图
│   └── memory_helper.go           # 辅助函数
│
├── model/
│   └── memory.go                  # 记忆相关数据模型
│
└── logic/
    └── command/
        └── memory_cmd.go          # /remember 命令
```

## 待实现细节

1. **索引优化**：使用倒排索引提升搜索性能
2. **增量更新**：避免每次会话结束都重建索引
3. **知识去重**：检测相似概念避免重复
4. **隐私过滤**：敏感信息（API Key、密码）自动脱敏
5. **性能监控**：记录存储操作的耗时

## 参考资料
- 现有架构：pkg/tui/chat.go, pkg/logic/agent/
- Bubble Tea: https://github.com/charmbracelet/bubbletea
- OpenSpec 设计模式：记忆系统类比 spec/change
