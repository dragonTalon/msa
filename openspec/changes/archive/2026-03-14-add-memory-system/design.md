# 设计：add-memory-system

## 上下文

### 当前状态

MSA (My Stock Agent) 是一个基于 AI 的股票交易助手，使用 Bubble Tea 提供 TUI 终端界面。当前每次启动都是全新会话，AI 无法"记住"之前的交互，导致：

1. **上下文丢失**：用户需要重复说明偏好
2. **知识无法积累**：用户学过的概念无法被记住
3. **历史无法追溯**：无法查看之前的对话记录
4. **个性化缺失**：AI 无法基于历史提供个性化建议

### 约束条件

- **无新增外部依赖**：使用 Go 标准库和现有依赖
- **向后兼容**：记忆系统是可选功能，禁用时行为不变
- **性能影响可控**：启动时间增加 < 200ms，内存增加 < 20MB
- **隐私保护**：必须过滤敏感信息（API Key、密码等）

### 利益相关者

- **终端用户**：需要 AI 记住偏好和上下文
- **开发者**：需要可维护、可扩展的架构

### 相关历史问题

基于知识库分析，记忆系统是新功能领域，但应遵循以下模式：
- 问题 001：AI 助手不应自动执行影响性操作
- 问题 003：不安全的配置（需注意敏感信息脱敏）
- 模式 004：单例模式管理全局资源
- 模式 003：错误处理和重试策略
- 模式 002：AI 助手工作流程指南

## 目标 / 非目标

**目标：**
- 持久化所有聊天会话到本地文件系统
- 使用 AI 从对话中提取结构化知识（用户画像、概念、策略等）
- 提供跨会话搜索和浏览能力
- 将用户知识注入 AI Prompt 以提供个性化回复
- 实现 TUI 记忆浏览器（`/remember` 命令）

**非目标：**
- 多用户记忆（单用户设计）
- 云端同步（仅本地存储）
- 记忆导入/导出（可在后续版本添加）
- 实时协作或多设备同步

## 推荐模式

### 1. 记忆管理器单例模式

**模式ID**：browser-manager-singleton (pattern 004)
**来源问题**：避免频繁创建存储实例和资源浪费

**描述**：
使用 `sync.Once` 确保 MemoryManager 全局只初始化一次，避免：
- 频繁加载索引（性能开销）
- 多个存储实例导致数据不一致
- 资源没有正确清理（内存泄漏）

**实现**：
```go
type Manager struct {
    store        Store
    currentSession *SessionManager
    knowledgeCache *AggregatedKnowledge
    mu          sync.RWMutex
    once        sync.Once
}

var globalManager *Manager

func GetManager() *Manager {
    managerOnce.Do(func() {
        store, err := NewFileStore()
        if err != nil {
            log.Errorf("创建存储失败: %v", err)
            return
        }

        globalManager = &Manager{
            store:          store,
            knowledgeCache: &AggregatedKnowledge{},
        }

        // 启动时加载知识
        globalManager.LoadKnowledge()
    })

    return globalManager
}
```

**适用场景**：
- 全局唯一的存储访问点
- 启动时加载的知识缓存
- 避免并发写入冲突

### 2. 错误处理和重试策略

**模式ID**：error-handling-strategy (pattern 003)
**来源问题**：文件操作可能因临时故障失败

**描述**：
分类错误处理，根据错误类型采取不同策略：
- **参数错误**：立即返回，不重试
- **永久错误**（文件不存在、权限错误）：不重试
- **临时错误**（文件被占用、磁盘满）：重试 2-3 次

**实现**：
```go
func (s *FileStore) SaveSession(session *model.Session) error {
    const maxRetries = 2
    var lastErr error

    for attempt := 0; attempt <= maxRetries; attempt++ {
        err := s.writeSessionFile(session)
        if err == nil {
            return s.updateIndexes(session)
        }

        // 判断是否可重试
        if isTemporaryError(err) && attempt < maxRetries {
            lastErr = err
            time.Sleep(backoff(attempt))
            continue
        }

        // 永久错误或重试次数用尽
        return fmt.Errorf("保存会话失败: %w", err)
    }

    return nil
}

func isTemporaryError(err error) bool {
    // 检查是否为临时错误
    return errors.Is(err, syscall.EINTR) ||
           errors.Is(err, syscall.EAGAIN)
}
```

**适用场景**：
- 所有文件写入操作
- 索引更新操作
- 知识保存操作

### 3. 异步任务工作流程

**模式ID**：ai-assistant-workflow (guide 002)
**来源问题**：001 - AI 助手不应自动执行影响性操作

**描述**：
知识提取是异步后台任务，遵循以下工作流程：
1. 用户退出时触发知识提取（不阻塞）
2. 异步调用 AI 分析对话内容
3. 提取完成后保存知识文件
4. 更新知识缓存（不影响当前会话）

**实现**：
```go
func (m *Manager) EndSession() error {
    // 1. 保存完整会话（同步）
    session := m.buildSession()
    if err := m.store.SaveSession(session); err != nil {
        log.Errorf("保存会话失败: %v", err)
        return err
    }

    // 2. 异步提取知识（不阻塞）
    go m.extractKnowledgeAsync(session)

    // 3. 清空当前会话
    m.currentSession = nil

    return nil
}

func (m *Manager) extractKnowledgeAsync(session *model.Session) {
    log.Infof("开始提取知识: %s", session.ID)

    knowledges, err := ExtractKnowledgeWithAI(session)
    if err != nil {
        log.Errorf("AI提取知识失败: %v", err)
        return
    }

    // 保存知识
    for _, knowledge := range knowledges {
        if err := m.store.SaveKnowledge(knowledge); err != nil {
            log.Errorf("保存知识失败: %v", err)
        }
    }

    // 刷新知识缓存
    m.LoadKnowledge()

    log.Infof("知识提取完成: %d条", len(knowledges))
}
```

**适用场景**：
- 会话结束时的知识提取
- 索引重建
- 大文件操作

### 4. 安全优先配置

**模式ID**：security-first-configuration (pattern 005)
**来源问题**：003 - 不安全的配置可能导致敏感信息泄露

**描述**：
记忆系统存储用户对话内容，必须实施隐私保护：
- 保存前过滤敏感关键词（API Key、密码）
- 检测并替换敏感内容为 `***REDACTED***`
- 不记录系统命令的输出（如 `/config`）

**实现**：
```go
var sensitivePatterns = []struct {
    pattern string
    context string
}{
    {`sk-[a-zA-Z0-9]{32,}`, "OpenAI/SiliconFlow API Key"},
    {`Bearer\s+[a-zA-Z0-9\-._~+/]+=*`, "Bearer Token"},
    {`(password|passwd|pwd):\s*\S+`, "密码"},
}

func redactSensitive(content string) string {
    result := content

    for _, p := range sensitivePatterns {
        re := regexp.MustCompile(p.pattern)
        result = re.ReplaceAllString(result, "***REDACTED***")
    }

    return result
}

func (m *Manager) AddMessage(msg model.Message) error {
    // 过滤敏感信息
    msg.Content = redactSensitive(msg.Content)

    // 记录消息
    m.currentSession.messages = append(m.currentSession.messages, msg)

    return nil
}
```

**适用场景**：
- 所有消息保存前
- 知识提取前
- 日志输出前

## 避免的反模式

### 1. 同步阻塞的知识提取

**反模式ID**：ai-auto-commit (issue 001)
**问题所在**：
在用户退出时同步调用 AI 提取知识，会阻塞退出流程，用户需要等待数秒。

**影响**：
- 用户体验差
- 如果 AI 调用失败，可能导致无法退出

**不要这样做**：
```go
// ❌ 错误：同步阻塞
func (m *Manager) EndSession() error {
    // 保存会话
    m.store.SaveSession(session)

    // 阻塞等待 AI 提取
    knowledges := ExtractKnowledgeWithAI(session)  // 可能需要 5-10 秒

    // 保存知识
    for _, k := range knowledges {
        m.store.SaveKnowledge(k)
    }

    return nil
}
```

**应该这样做**：
```go
// ✅ 正确：异步不阻塞
func (m *Manager) EndSession() error {
    // 保存会话
    m.store.SaveSession(session)

    // 异步提取知识
    go m.extractKnowledgeAsync(session)

    // 立即返回
    return nil
}
```

### 2. 全文件扫描搜索

**反模式ID**：performance-issue (issue 004)
**问题所在**：
每次搜索都扫描所有会话文件，随着会话数量增加，搜索性能会线性下降。

**影响**：
- 大量会话时搜索响应慢（> 1 秒）
- 磁盘 I/O 开销大
- 用户体验差

**不要这样做**：
```go
// ❌ 错误：全文件扫描
func (s *FileStore) Search(query string) []SearchResult {
    var results []SearchResult

    // 扫描所有文件
    files := filepath.Glob("sessions/**/*.json")
    for _, file := range files {
        data := readFile(file)      // 磁盘 I/O
        session := parseJSON(data)  // CPU 密集
        if contains(session, query) {
            results = append(results, session)
        }
    }

    return results
}
```

**应该这样做**：
```go
// ✅ 正确：使用内存索引
func (s *FileStore) Search(query string) []SearchResult {
    var results []SearchResult

    // 使用内存索引（已在启动时加载）
    for _, session := range s.indexes.Sessions {
        if s.matchQuery(query, session.Summary, session.Keywords) {
            results = append(results, toSearchResult(session))
        }
    }

    return results
}
```

### 3. 索引丢失时崩溃

**反模式ID**：critical-failure
**问题所在**：
索引文件损坏或不存在时，程序直接崩溃或无法启动。

**影响**：
- 用户无法使用 MSA
- 需要手动删除索引才能恢复

**不要这样做**：
```go
// ❌ 错误：索引加载失败时崩溃
func (s *FileStore) loadIndexes() error {
    data := readFile("indexes/all.json")
    if err != nil {
        log.Fatal("加载索引失败:", err)  // 程序崩溃
    }
    return json.Unmarshal(data, s.indexes)
}
```

**应该这样做**：
```go
// ✅ 正确：回退到扫描重建
func (s *FileStore) loadIndexes() error {
    data, err := readFile("indexes/all.json")
    if err != nil {
        log.Warnf("加载索引失败: %v，将重建索引", err)
        return s.rebuildIndexes()  // 自动重建
    }
    return json.Unmarshal(data, s.indexes)
}
```

## 设计决策

### 决策 1：为什么选择文件系统存储而非数据库

**选项对比**：

| 选项 | 优点 | 缺点 | 选择 |
|------|------|------|------|
| SQLite | 结构化查询、事务支持 | 增加依赖、查询复杂性 | ❌ |
| 文件系统 | 简单、无需依赖、易于调试 | 无事务、需手动索引 | ✅ |

**理由**：
1. **无新增依赖**：使用 Go 标准库的 `encoding/json`
2. **易于调试**：JSON 文件可直接查看和编辑
3. **简化架构**：无需额外的 ORM 层
4. **足够的性能**：使用索引后，查询性能可接受

**权衡**：
- 失去事务支持 → 通过单例写入避免并发冲突
- 需要手动索引 → 启动时构建内存索引

### 决策 2：为什么选择 JSON 格式而非 Markdown

**选项对比**：

| 选项 | 优点 | 缺点 | 选择 |
|------|------|------|------|
| Markdown | 人类可读、易于编辑 | 解析复杂、结构不一致 | ❌ |
| JSON | 结构化、易解析、类型安全 | 不可读、需要工具 | ✅ |

**理由**：
1. **类型安全**：Go 的 `encoding/json` 提供类型安全
2. **易于解析**：无需复杂的文本解析
3. **一致性**：所有数据使用统一格式
4. **性能**：JSON 解析比文本解析快

**权衡**：
- 人类不可读 → 通过 TUI 视图提供友好界面
- 编辑困难 → 提供 `/remember` 命令管理

### 决策 3：为什么知识不合并而是创建新的

**选项对比**：

| 选项 | 优点 | 缺点 | 选择 |
|------|------|------|------|
| 合并知识 | 单一对象、简单 | 丢失历史、可能冲突 | ❌ |
| 创建新知识 | 保留历史、可追溯 | 需要聚合逻辑 | ✅ |

**理由**：
1. **保留完整历史**：可以看到用户偏好的演变
2. **无冲突风险**：不会因为合并而丢失信息
3. **简单可靠**：无需复杂的合并逻辑
4. **可追溯**：每个知识都有来源会话

**权衡**：
- 查询时需要聚合 → 启动时聚合到缓存
- 文件数量增加 → 按类型分目录组织

### 决策 4：为什么短期记忆限制为 10 条

**选项对比**：

| 选项 | 优点 | 缺点 | 选择 |
|------|------|------|------|
| 5 条 | 内存占用小 | 上下文太少 | ❌ |
| 10 条 | 平衡内存和上下文 | 固定大小 | ✅ |
| 20 条 | 上下文更多 | 内存占用大 | ❌ |
| 动态大小 | 灵活 | 复杂度高 | ❌ |

**理由**：
1. **足够上下文**：10 条消息（5 轮对话）提供足够的短期上下文
2. **内存可控**：约 5-10KB 内存占用
3. **简单实现**：固定大小的循环缓冲区
4. **符合 UX**：大多数对话在 5 轮内完成

**权衡**：
- 可能截断长对话 → 完整会话仍保存在文件中

## 风险 / 权衡

### 风险 1：启动性能退化

**风险**：
随着会话数量增加（> 1000），加载索引可能变慢。

**缓解措施**：
- 使用增量索引（只加载最新的）
- 实施索引分页（按月分割）
- 后台加载索引（非阻塞）

### 风险 2：磁盘空间占用

**风险**：
每个会话 10-50KB，1000 个会话约 10-50MB。

**缓解措施**：
- 提供 `/remember prune` 命令清理旧会话
- 实施自动归档（> 6 个月的会话）
- 压缩旧会话文件

### 风险 3：敏感信息泄露

**风险**：
用户可能在对话中透露 API Key、密码等敏感信息。

**缓解措施**：
- 实施关键词过滤（已实现）
- 提供 `/remember clear` 命令清空记忆
- 在文档中明确告知隐私策略

### 风险 4：AI 提取知识失败

**风险**：
网络故障或 API 问题导致知识提取失败。

**缓解措施**：
- 异步执行，不影响用户体验
- 失败时记录日志，不阻塞
- 可以手动重新提取（`/remember reextract <session-id>`）

## 迁移计划

### 阶段 1：核心存储层（第 1 周）

1. 实现 `pkg/model/memory.go` - 数据模型
2. 实现 `pkg/logic/memory/store.go` - 存储接口和 FileStore
3. 实现索引构建和加载
4. 单元测试：文件读写、索引操作

### 阶段 2：记忆管理器（第 2 周）

1. 实现 `pkg/logic/memory/manager.go` - 单例管理器
2. 实现 `pkg/logic/memory/extraction.go` - AI 知识提取
3. 集成到 `pkg/tui/chat.go` - 消息记录
4. 集成到 `pkg/logic/agent/chat.go` - 知识注入

### 阶段 3：TUI 视图（第 3 周）

1. 实现 `pkg/tui/memory_home.go` - 主页
2. 实现 `pkg/tui/memory_session_list.go` - 会话列表
3. 实现 `pkg/tui/memory_session_detail.go` - 会话详情
4. 实现 `pkg/tui/memory_knowledge_list.go` - 知识列表
5. 实现 `pkg/tui/memory_search.go` - 搜索视图

### 阶段 4：命令和优化（第 4 周）

1. 实现 `pkg/logic/command/memory_cmd.go` - `/remember` 命令
2. 性能优化：索引缓存、增量更新
3. 隐私保护：敏感信息过滤
4. 文档和测试

### 回滚策略

如果出现严重问题：
1. 设置环境变量 `MSA_MEMORY_ENABLED=false` 禁用记忆系统
2. 删除 `~/.msa/remember/` 目录清空所有记忆
3. 行为恢复到现有版本

### 验证步骤

1. **功能测试**：
   - 创建会话并保存
   - 查看、搜索会话
   - 查看知识库
   - AI 上下文增强

2. **性能测试**：
   - 启动时间 < 200ms
   - 搜索响应 < 500ms
   - 内存占用 < 20MB

3. **隐私测试**：
   - API Key 被过滤
   - 密码被脱敏
   - 敏感命令不记录

4. **兼容性测试**：
   - 禁用时行为不变
   - 索引损坏时自动恢复
