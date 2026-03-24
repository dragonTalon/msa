# 设计：session-resume

## 上下文

MSA 当前缺少会话持久化与恢复能力。用户对话仅在内存中保存，退出后丢失。本设计引入会话管理模块，实现对话历史的自动持久化和恢复。

**当前状态**：
- TUI 模式：对话历史存储在 `pkg/tui/chat.go` 的 `history []model.Message` 中
- CLI 模式：单轮问答，无历史记录
- 无持久化机制

**约束**：
- 使用 Markdown 文件存储（人类可读）
- 按日期分目录组织
- 追加写入，避免频繁重写

## 目标 / 非目标

**目标：**
- 实现会话自动持久化（用户消息 + 助手回复）
- 实现 `--resume` 参数恢复历史会话
- CLI 模式输出会话 ID
- TUI 模式 `clear` 命令重置会话

**非目标：**
- 会话列表选择功能
- 会话数量限制或自动清理
- 敏感信息加密
- 会话导出/导入

## 推荐模式

### 单例模式管理 SessionManager

**模式ID**：browser-manager-singleton
**来源**：知识库 patterns/browser-manager-singleton.md

**描述**：
使用单例模式 + `sync.Once` 管理全局 SessionManager，确保会话状态全局唯一。

**适用场景**：
- 全局会话状态管理
- 需要在多个模块间共享会话信息

**实现**：
```go
type Manager struct {
    mu        sync.RWMutex
    current   *Session
    homeDir   string
    memoryDir string
}

var (
    globalManager *Manager
    once          sync.Once
)

func GetManager() *Manager {
    once.Do(func() {
        homeDir, _ := os.UserHomeDir()
        globalManager = &Manager{
            homeDir:   homeDir,
            memoryDir: filepath.Join(homeDir, ".msa", "memory"),
        }
    })
    return globalManager
}
```

### 错误分类处理

**模式ID**：error-handling-strategy
**来源**：知识库 patterns/error-handling-strategy.md

**描述**：
文件操作错误分类处理，区分可恢复和不可恢复错误。

**适用场景**：
- 文件写入失败
- 目录创建失败
- 文件解析失败

**实现**：
```go
// 参数错误：立即返回
if sessionID == "" {
    return fmt.Errorf("请指定要恢复的会话 ID")
}

// 可降级错误：记录日志继续执行
if err := m.AppendMessage(msg); err != nil {
    log.Warnf("追加消息失败: %v", err)
    // 继续执行，不影响主流程
}
```

## 避免的反模式

### 频繁重写整个文件

**问题所在**：
每次消息都重写整个文件，文件变大后性能下降。

**影响**：
- 大文件写入延迟
- 磁盘 I/O 压力

**不要这样做**：
```go
// 每次都读取、修改、重写整个文件
content, _ := os.ReadFile(path)
content = append(content, newMsg...)
os.WriteFile(path, content, 0644)
```

**应该这样做**：
```go
// 使用追加模式
f, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
defer f.Close()
f.WriteString(msgContent)
```

### 忽略文件操作错误

**问题所在**：
文件操作失败时静默忽略，用户无法感知问题。

**影响**：
- 会话丢失用户无感知
- 难以排查问题

**不要这样做**：
```go
_ = os.MkdirAll(dir, 0755)
_ = f.WriteString(msg)
```

**应该这样做**：
```go
if err := os.MkdirAll(dir, 0755); err != nil {
    log.Warnf("创建会话目录失败: %v", err)
}
if err := f.WriteString(msg); err != nil {
    log.Errorf("追加消息失败: %v", err)
}
```

## 设计决策

### 1. 文件存储格式

**决策**：Markdown + YAML Frontmatter

**理由**：
- 人类可读，便于调试
- 支持元数据（frontmatter）
- 与 AI 回复格式一致（Markdown）

**文件结构**：
```
~/.msa/memory/2024/03/24/a1b2c3d4-5678-....md
```

### 2. 模块划分

**决策**：新建 `pkg/session/` 模块

**理由**：
- 职责单一，易于维护
- 与现有模块解耦
- 便于单元测试

**模块结构**：
```
pkg/session/
├── manager.go    # SessionManager（单例）
├── session.go    # Session 数据结构
├── store.go      # 文件存储（追加写入）
└── parser.go     # Markdown 解析
```

### 3. 写入时机

**决策**：
- 用户发问时：立即写入用户消息
- 回复完成时：追加写入助手消息

**理由**：
- 用户消息可立即持久化
- 避免回复中途写入不完整内容
- 即使崩溃也保留用户问题

### 4. frontmatter 更新

**决策**：使用固定大小的 frontmatter 占位符

**理由**：
- 追加写入时无法修改文件开头
- 使用固定宽度时间戳预留空间

**实现**：
```go
// frontmatter 模板，时间戳固定 25 字符
const frontmatterTemplate = `---
uuid: %s
created_at: %s
updated_at: %s
mode: %s
---
`
```

## 架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                         调用流程                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  msa (TUI)              msa -q "问题"        msa --resume ID    │
│      │                       │                      │           │
│      ▼                       ▼                      ▼           │
│  ┌───────┐              ┌───────┐              ┌───────┐       │
│  │app.Run│              │extcli │              │ root  │       │
│  └───┬───┘              │.Run   │              └───┬───┘       │
│      │                  └───┬───┘                  │            │
│      ▼                      ▼                      ▼            │
│  ┌───────────────────────────────────────────────────────┐     │
│  │              session.GetManager()                      │     │
│  │                                                       │     │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐   │     │
│  │  │ NewSession  │  │ LoadSession │  │ AppendMsg   │   │     │
│  │  └─────────────┘  └─────────────┘  └─────────────┘   │     │
│  └───────────────────────────────────────────────────────┘     │
│                              │                                  │
│                              ▼                                  │
│                     ~/.msa/memory/                              │
│                     YYYY/MM/DD/{uuid}.md                        │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## 风险 / 权衡

| 风险 | 缓解措施 |
|------|----------|
| 文件并发写入冲突 | 使用 sync.Mutex 保护文件操作 |
| 磁盘空间占用 | 暂不处理，用户可手动清理 |
| 解析兼容性 | 严格解析格式，错误时给出清晰提示 |
| 异常退出丢失最后消息 | 用户消息立即写入，助手消息完成时写入 |

## 迁移计划

**部署步骤**：
1. 新增 `pkg/session/` 模块
2. 修改 `cmd/root.go` 添加 `--resume` 参数
3. 修改 `pkg/tui/chat.go` 注入 SessionManager
4. 修改 `pkg/extcli/chat.go` 注入 SessionManager
5. 测试验证

**回滚策略**：
- 删除 `pkg/session/` 模块
- 回滚 `cmd/root.go`、`pkg/tui/chat.go`、`pkg/extcli/chat.go` 的修改
- 无数据库迁移，无需数据回滚
