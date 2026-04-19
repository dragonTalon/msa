# MSA (My Stock Agent) - Comprehensive Codebase Analysis

## Project Overview

**MSA** is a lightweight, open-source stock intelligence agent tool for investors and developers. It combines:
- Natural language AI chat interface (multiple LLM providers: OpenAI, Claude, SiliconFlow, Ollama)
- Stock data query and market analysis
- Terminal UI (Bubble Tea)
- SQLite local database for trading simulations
- Dynamic skills system (modular prompt templates)
- Memory system for conversation history and AI knowledge extraction

**Tech Stack:**
- Go 1.25.5
- Cobra (CLI), Bubble Tea (TUI)
- Eino (Cloudwego - AI framework with ReAct agent pattern)
- SQLite (pure Go), GORM (ORM)
- Logrus (logging)

---

## 1. Overall Directory Structure

```
msa/
├── main.go                          # Entry point, version info injection
├── cmd/                             # CLI command routing (Cobra)
│   ├── root.go                      # Root command, context creation, signal handling
│   ├── config/command.go            # Configuration TUI
│   ├── skill/                       # Skills management commands
│   ├── update/command.go            # Update functionality
│   └── version/command.go           # Version info
│
└── pkg/                             # Business logic layer
    ├── app/app.go                   # Application startup & TUI launcher
    ├── config/                      # Configuration management
    │   ├── local_config.go          # File-based config storage (~/.msa/msa_config.json)
    │   ├── env.go                   # Environment variable handling
    │   ├── validator.go             # Config validation
    │   └── logger.go                # Logrus setup
    │
    ├── db/                          # Database layer
    │   ├── db.go                    # SQLite connection setup, PRAGMA optimization
    │   ├── global.go                # Global DB singleton (sync.Once pattern)
    │   ├── migrate.go               # Schema migrations
    │   ├── account.go               # Account CRUD operations
    │   └── transaction.go           # Transaction recording
    │
    ├── model/                       # Data models & constants
    │   ├── account.go               # Account entity
    │   ├── transaction.go           # Transaction entity
    │   ├── stock.go                 # Stock data models
    │   ├── comment.go               # LLM provider registry & constants
    │   ├── tool_result.go           # Tool execution results (success/error JSON format)
    │   ├── stream.go                # Stream chunk & message types
    │   └── cmd.go                   # Command models
    │
    ├── session/                     # Session management (stateful conversation tracking)
    │   ├── manager.go               # Global session manager singleton
    │   ├── session.go               # Session entity & serialization
    │   ├── store.go                 # Session persistence (~/.msa/memory/sessions/)
    │   └── parser.go                # Session parsing & resumption
    │
    ├── logic/                       # Core business logic
    │   ├── agent/chat.go            # Eino ReAct agent setup, stream handling
    │   │                            # - toolCallChecker: processes non-standard tool formats (DSML, pipe format)
    │   │                            # - Ask: initiates async agent.Stream() call
    │   ├── command/cmd.go           # Command parsing & execution
    │   ├── message/stream_manage.go # StreamOutputManager - global broadcast system for streaming chunks
    │   ├── provider/                # LLM provider implementations
    │   │   ├── register.go          # Provider registry
    │   │   └── siliconflow/api.go   # SiliconFlow provider (OpenAI-compatible)
    │   ├── finsvc/                  # Finance service layer
    │   │   ├── trade_service.go     # Order execution, position tracking
    │   │   └── position_service.go  # P&L calculation
    │   ├── skills/                  # Dynamic skills system
    │   │   ├── manager.go           # Skills loading & initialization
    │   │   ├── loader.go            # Markdown YAML frontmatter parsing
    │   │   ├── selector.go          # Auto skill selection based on query
    │   │   └── registry.go          # Skill registry
    │   └── tools/                   # Tool implementations (called by Eino agent)
    │       ├── basic.go             # MsaTool interface, tool registry
    │       ├── register.go          # Tool registration
    │       ├── safetool/            # Panic recovery wrapper for all tools
    │       ├── finance/             # Trading tools (buy, sell, accounts, positions)
    │       ├── stock/               # Stock query tools (company code, K-line, info)
    │       ├── search/              # Web search & content fetching
    │       ├── skill/               # Skill content retrieval tools
    │       ├── knowledge/           # Memory system tools (read/write KB)
    │       └── todo/                # TODO list tools (daily tasks)
    │
    ├── tui/                         # Terminal UI (Bubble Tea)
    │   ├── chat.go                  # Main chat model (TUI stateful component)
    │   ├── model_selector.go        # Model selection UI
    │   ├── config/                  # Configuration UI components
    │   └── style/                   # Styling & markdown rendering
    │
    ├── extcli/chat.go               # CLI single-round chat mode (-q parameter)
    ├── updater/updater.go           # Binary update mechanism
    ├── version/version.go           # Version info storage
    └── utils/                       # Utility functions
        ├── file.go
        ├── http.go                  # HTTP client setup (Resty)
        └── format.go                # Data formatting
```

---

## 2. Main Packages & Their Responsibilities

### 2.1 **cmd/** - Command Routing Layer
- **root.go**: 
  - Creates context with signal handling (SIGINT, SIGTERM)
  - Calls `config.InitConfig()` to load configuration
  - Initializes global database: `db.InitGlobalDB()`
  - Routes to TUI (`app.Run(ctx)`) or CLI mode
  - Handles `--resume <session-id>` for session recovery
  - Context propagates down to `app.Run()` → `tui.NewChat(ctx)` → TUI update loop

### 2.2 **pkg/app/** - Application Startup
- **app.go**: 
  - Calls `tea.NewProgram(tui.NewChat(ctx))`
  - Runs Bubble Tea event loop
  - Gets session from manager after exit to display session ID

### 2.3 **pkg/config/** - Configuration Management
- **local_config.go**: 
  - Singleton pattern with `sync.Once` caching
  - Loads from `~/.msa/msa_config.json`
  - Merges CLI parameters via `SetCLIConfig()`
  - Priority: CLI args > env vars > config file > defaults
  - No context dependency

### 2.4 **pkg/db/** - Database Layer
- **db.go**: SQLite initialization with WAL mode, connection pooling
- **global.go**: 
  - Global singleton DB connection (also `sync.Once`)
  - `InitGlobalDB()` called once from `cmd/root.go`
  - `GetDB()` returns nil if not initialized
  - **Problem**: No context awareness - tools access via `db.GetDB()` without passing ctx

### 2.5 **pkg/session/** - Session Management
- **manager.go**: Global singleton session manager
- **session.go**: Session entity with ID, mode (TUI/CLI), timestamps
- **store.go**: Persists sessions to `~/.msa/memory/sessions/`
- **Problem**: Global singleton, no context passing

### 2.6 **pkg/logic/agent/** - Agent & Chat Orchestration
- **chat.go**:
  - `GetChatModel(ctx)`: Creates Eino ReAct agent (cached globally)
  - `toolCallChecker()`: Processes LLM streaming output, detects tool calls (DSML/pipe formats)
  - `Ask(ctx, messages, history)`: 
    - Initializes skills
    - Builds system prompt
    - **Launches goroutine**: `go func() { sr, err := agent.Stream(ctx, queryMsg) }()`
    - This goroutine calls `toolCallChecker` which broadcasts to `StreamOutputManager`
  - **Critical Pattern**: No explicit cancellation of agent goroutines - relies on context cancelation

### 2.7 **pkg/logic/message/** - Stream Broadcasting
- **stream_manage.go**:
  - `StreamOutputManager`: Global singleton (no context)
  - Maintains slice of subscriber channels (buffered, size 100)
  - `Broadcast(chunk)`: Sends to all subscribers
    - IsDone messages: **blocking send** (enforced delivery)
    - Normal messages: **non-blocking send** (skips if channel full)
  - Multiple goroutines can call `Broadcast()` concurrently (protected by RWMutex)

### 2.8 **pkg/logic/tools/** - Tool System
- **basic.go**: `MsaTool` interface - all tools must implement
- **register.go**: Registers all tools at init() time
- **safetool/safetool.go**: Panic recovery wrapper for tool execution
- **Finance Tools**: 
  - Access DB via `db.GetDB()` (no context)
  - Execute orders, update positions
  - Broadcast via `message.BroadcastToolStart()` / `BroadcastToolEnd()`
- **Stock Tools**: Query APIs, broadcast results
- **Search Tools**: Web search, content fetching
- **Problem**: All tools are stateless singletons, no request-scoped context passing

### 2.9 **pkg/tui/** - Terminal UI (Bubble Tea)
- **chat.go**:
  - `Chat` struct holds context (`ctx`)
  - `Update(msg)`: Event handler for keyboard, stream chunks
  - Subscribes to `StreamOutputManager` in `handleEnterKey()`
  - `chatStreamMsg()`: Processes incoming `*model.StreamChunk` messages
  - **Pattern**: TUI **receives** context from `cmd/root.go`, stores it, but context is largely unused
  - No explicit context cancellation mechanism

---

## 3. Context (ctx) Usage & Propagation

### 3.1 **Context Creation & Flow**
```
cmd/root.go::runRoot()
  └─ ctx := cmd.Context()  [from Cobra]
  └─ app.Run(ctx)
      └─ tea.NewProgram(tui.NewChat(ctx))
          └─ Chat.ctx = ctx
              └─ Chat.Update() [Bubble Tea event loop]
                  └─ handleEnterKey()
                      └─ agent.Ask(ctx, question, history)
                          └─ go func() { agent.Stream(ctx, messages) }()
                              └─ toolCallChecker(ctx, streamReader)
```

### 3.2 **Context Cancellation & Cleanup**
- **Root ctx creation**: `cmd/root.go::NotifySignal()`
  - Creates context with `context.WithCancel()`
  - Monitors OS signals (SIGINT, SIGTERM)
  - **First signal**: Cancels context
  - **Second signal**: Force exits via `os.Exit(128+n)`
- **Problem**: 
  - Context cancellation propagates to agent goroutines
  - BUT: No explicit cleanup of database connections or stream subscriptions
  - Stream subscribers never explicitly unregistered

### 3.3 **Where Context is NOT Used**
1. **Database access**: All tools call `db.GetDB()` directly
2. **Configuration**: Uses global singleton cache
3. **Session management**: Uses global singleton manager
4. **Stream broadcasting**: Global `StreamOutputManager` singleton
5. **Tool registry**: All tools are stateless singletons
6. **LLM provider**: Agent cached globally (no per-request context)

### 3.4 **Context Issues & Risks**

| Issue | Severity | Impact |
|-------|----------|--------|
| **No request-scoped DB transactions** | High | Multiple concurrent requests to same DB could cause locks or data races |
| **Global stream broadcaster** | High | Mixing streams from multiple concurrent chats in same process |
| **No stream subscriber cleanup** | Medium | Memory leaks if exception occurs during subscription |
| **Agent cached globally** | Medium | Model switch doesn't take effect until restart (relies on `ResetCache()` being called) |
| **No goroutine lifecycle management** | Medium | `Ask()` spawns goroutine that runs until ctx cancellation; no graceful shutdown |
| **Unused ctx in TUI** | Low | Context stored but mostly unused for cancellation or deadlines |

---

## 4. Communication Patterns Between Components

### 4.1 **Goroutine-Based Concurrency**

```
┌─ TUI Event Loop (Bubble Tea)
│  └─ textinput.Update()
│  └─ StreamChunk channel reader
│
├─ Agent Goroutine (spawned by Ask())
│  └─ agent.Stream(ctx, messages) [STREAMING]
│     └─ Eino internal goroutines for tool execution
│        └─ toolCallChecker(ctx, streamReader)
│           └─ StreamOutputManager.Broadcast(chunk)
│
└─ Tool Goroutines (spawned by Eino agent)
   └─ finance.SubmitBuyOrder()
   └─ stock.GetStockCompanyCode()
   └─ search.SearchTool.Call()
   └─ All broadcast via message.Broadcast*()
```

### 4.2 **Channel-Based Communication**
1. **StreamOutputManager** (global broadcast channel set):
   - **Publisher**: `toolCallChecker()` and tools via `message.Broadcast*()`
   - **Subscriber**: TUI's `Chat.Update()` via `streamOutputCh`
   - Pattern: **Multiple publishers → Multiple subscribers** (pub/sub)
   - Issue: If one subscriber is slow, others not affected (non-blocking sends)

2. **Thinking Progress Channel** (inside `toolCallChecker`):
   - `thinkingReset`: Signals LLM output received
   - `thinkingDone`: Closes the progress ticker
   - Local to `toolCallChecker()`, closed on completion

### 4.3 **Database Access Pattern**
- **Stateless**: All tools call `db.GetDB()` → `*gorm.DB`
- **No transactions**: Each tool operation uses GORM's auto-transaction per query
- **No context deadline**: Queries don't respect parent context timeouts
- **Concurrency**: SQLite WAL mode allows multiple readers + one writer

### 4.4 **Service Layer** (`pkg/logic/finsvc/`)
- **trade_service.go**: `SubmitBuyOrder(db, accountID, order) → transactionID`
- **position_service.go**: Calculates P&L from transactions
- These functions take `*gorm.DB` directly, not context
- Called by finance tools

### 4.5 **Session State Management**
- **Global manager**: `session.GetManager()` (singleton)
- **Per-chat session**: Created in `NewChat()` or resumed from `--resume <id>`
- **Session persistence**: Asynchronously written to `~/.msa/memory/sessions/`
- **No locking**: Session state updates not protected by mutex (potential race condition)

---

## 5. Obvious Architectural Issues with Context & Inter-Component Communication

### **Critical Issues**

#### 1. **No Request-Scoped Context in Database Operations**
**Problem**: Tools access database via global `db.GetDB()` without context
```go
// Current pattern (PROBLEMATIC)
func SubmitBuyOrder(ctx context.Context, param *SubmitBuyOrderParam) (string, error) {
    database := msadb.GetDB()  // ← Ignores ctx!
    // ...
    transID, err := finsvc.SubmitBuyOrder(database, account.ID, order)
}
```
**Impact**: 
- No timeout enforcement on database queries
- No cancellation propagation to long-running queries
- In concurrent scenarios, impossible to associate operations with requests
- Hard to debug which operation belongs to which chat session

**Fix Required**: Pass `ctx` to all database operations
```go
// Correct pattern
func SubmitBuyOrder(ctx context.Context, param *SubmitBuyOrderParam) (string, error) {
    database := msadb.GetDB()
    transID, err := finsvc.SubmitBuyOrder(ctx, database, account.ID, order)
}
```

#### 2. **Global Singleton StreamOutputManager Creates Mixing Issues**
**Problem**: Multiple concurrent chats (if supported) share the same subscriber list
```go
// Current pattern (PROBLEMATIC)
var globalStreamManager = &StreamOutputManager{}  // ← Single instance for ALL chats!

// If two chats run concurrently:
chat1.streamOutputCh = RegisterStreamOutput(100)  // Adds to globalStreamManager
chat2.streamOutputCh = RegisterStreamOutput(100)  // Adds to same list
message.BroadcastToolStart("tool_a", "params")   // ← Broadcasts to BOTH!
```
**Impact**: 
- If TUI supports multiple concurrent chats (future), streams would mix
- Debugging becomes impossible (which tool belongs to which chat?)
- Performance issue: Slow consumer blocks others

**Fix Required**: Request-scoped stream managers
```go
// Should store in context or per-chat instance
type ChatRequest struct {
    ctx    context.Context
    stream *StreamOutputManager  // ← Per-request
}
```

#### 3. **No Goroutine Lifecycle Management**
**Problem**: `Ask()` spawns a goroutine that runs until context cancellation
```go
func Ask(ctx context.Context, messages string, history []model.Message) error {
    go func() {
        sr, err := agent.Stream(ctx, queryMsg)  // ← If ctx cancels mid-execution
        // ...                                     what happens to active tool calls?
    }()
    return nil  // ← Returns immediately without waiting!
}
```
**Issues**:
- Caller doesn't know if operation completed successfully
- No error feedback to caller
- If tools modify database and context cancels, operations might roll back inconsistently
- TUI might exit before agent goroutine completes

**Fix Required**: Use sync.WaitGroup or channels to track goroutine completion
```go
func Ask(ctx context.Context, messages string, history []model.Message) (chan error, error) {
    doneCh := make(chan error, 1)
    go func() {
        // ... agent execution
        doneCh <- err
    }()
    return doneCh, nil
}
```

#### 4. **Agent Cached Globally Without Thread Safety**
**Problem**: `agentCache` is global and mutable
```go
// Current pattern (PROBLEMATIC)
var agentCache *react.Agent  // ← Global, no mutex!

func GetChatModel(ctx context.Context) (*react.Agent, error) {
    if agentCache != nil {
        return agentCache, nil  // ← If called during ResetCache()...
    }
    // ... create new agent
    agentCache = agent
    return agent, nil
}

func ResetCache() {
    agentCache = nil  // ← Race condition!
}
```
**Impact**: 
- If two goroutines call `GetChatModel()` and `ResetCache()` concurrently, UB
- Model switch requires manual call to `ResetCache()` before new context

**Fix Required**: Protect with sync.Mutex or move to request-scoped storage
```go
type ChatContext struct {
    agent *react.Agent
    mu    sync.RWMutex
}
```

#### 5. **No Stream Subscriber Cleanup on Error**
**Problem**: If exception occurs after `RegisterStreamOutput()`, subscription leaks
```go
func handleEnterKey() {
    streamOutputCh, streamUnregister := message.RegisterStreamOutput(streamBufferSize)
    c.streamOutputCh = streamOutputCh
    c.streamUnregister = streamUnregister
    
    if _, err := agent.Ask(c.ctx, question, c.history); err != nil {
        // ← If panic occurs in Ask(), streamUnregister is never called!
        return
    }
}
```
**Impact**: 
- Memory leak if exception occurs
- After 100 exceptions, 100 channels leak

**Fix Required**: Use defer
```go
streamUnregister := defer streamUnregister  // Always clean up
```

---

### **Major Issues**

#### 6. **Session State Race Condition**
**Problem**: `session.Manager` is not thread-safe for state updates
```go
// In chat.go:
c.sessionMgr.SetCurrent(sess)  // ← Not protected during TUI event loop

// In any goroutine:
sess := sessionMgr.Current()  // ← Could be inconsistent state
```

#### 7. **No Context Deadline/Timeout**
**Problem**: Context is never given a deadline
```go
ctx, cancel := NotifySignal(context.Background(), ...)  // ← No deadline!
// Could hang indefinitely on:
// - LLM API calls
// - Database queries
// - Tool execution
```

#### 8. **Tool Result Broadcasting Outside Context Scope**
**Problem**: Tools broadcast messages that outlive their execution context
```go
func GetStockCompanyCode(ctx context.Context, param *CompanyParam) (string, error) {
    message.BroadcastToolStart(...)  // ← Uses global broadcaster
    if ctx.Err() != nil {
        message.BroadcastToolEnd(..., ctx.Err())  // ← Broadcasts to wrong chat?
    }
}
```
**If two chats run tools concurrently**: Both broadcast to same StreamOutputManager

---

## 6. Key Data Structures & Interfaces

### 6.1 **Core Models** (`pkg/model/`)
```go
// Message types
type MessageRole string
const (
    RoleUser, RoleAssistant, RoleSystem, RoleLogo
)

type StreamMsgType string
const (
    StreamMsgTypeText, StreamMsgTypeReason, StreamMsgTypeTool
)

type Message struct {
    Role    MessageRole
    Content string
    MsgType StreamMsgType
}

// Stream chunk for async updates
type StreamChunk struct {
    Content  string
    MsgType  StreamMsgType
    Err      error
    IsDone   bool  // ← Marks end of stream
}

// Tool result format (JSON)
type ToolResult struct {
    Success bool          // true = success, false = error
    Data    interface{}   // Result data or error message
    Message string        // Human-readable description
}

// Account & Transaction entities
type Account struct {
    ID      int64
    Balance int64  // Stored in 毫 (1/10000 Yuan)
    Status  string
}

type Transaction struct {
    ID          int64
    AccountID   int64
    Type        string  // "BUY" or "SELL"
    StockCode   string
    Quantity    int64
    Price       int64   // Stored in 毫
    Fee         int64   // Stored in 毫
    Status      string  // "PENDING", "COMPLETED", "FAILED"
    CreatedAt   time.Time
}
```

### 6.2 **Tool Interface** (`pkg/logic/tools/`)
```go
type MsaTool interface {
    GetToolInfo() (tool.BaseTool, error)    // ← Returns Eino tool
    GetName() string
    GetDescription() string
    GetToolGroup() model.ToolGroup
}

// Example implementation
type CompanyCode struct {}
func (c *CompanyCode) GetToolInfo() (tool.BaseTool, error) {
    return utils.InferTool(c.GetName(), c.GetDescription(), GetStockCompanyCode)
}
```

### 6.3 **Session** (`pkg/session/`)
```go
type Session struct {
    id       string  // Format: YYYY-MM-DD_uuid
    mode     string  // "TUI" or "CLI"
    messages []Message  // ← Conversation history
    createdAt time.Time
}

type Manager struct {
    current *Session  // ← Current active session (not thread-safe!)
}
```

### 6.4 **Chat TUI** (`pkg/tui/`)
```go
type Chat struct {
    ctx               context.Context
    textInput         textinput.Model
    history           []model.Message
    streamOutputCh    <-chan *model.StreamChunk  // ← Subscription channel
    streamUnregister  func()
    sessionMgr        *session.Manager
    isStreaming       bool
    fullStreamContent strings.Builder
}
```

---

## 7. Recent Git Commits Analysis

```
f826b46 fix: tool格式化问题+日志问题                  [Most recent]
3875484 fix: tool格式化问题+日志问题
a1b6afe feat: 优化一下tools的处理，防止panic情况     ← SafeTool wrapper added
4ac6be7 feat: 优化一下tools的处理，防止panic情况
ee3d010 feat: 重构错题总结逻辑                        ← Refactored error summarization
5ed4617 feat: 增加每日开盘执行的TODO列表（harness方式）  ← TODO system
b1cca3d feat: 修改会话记录功能，清除记忆实现          ← Session & memory work
d205073 fix: 日志确实问题
ba0c058 fix: tool格式化问题+日志问题
f3d1643 feat: 提供版本更新和安装能力
69a2943 feat: 账户信息统计
4158651 feat: 优化skill机制，参考ADK的SKILL blog     ← Skills system optimization
d3d4562 fix: 修复早午盘与总结没有落文件问题
cd1a073 fix: tool报错导致agent中断问题               ← Critical: Tool errors break agent
9ac45af fix: 修复搜索问题 和记忆存储错误问题
c32c845 fix: 修复搜索问题 和记忆存储错误问题
b86c1aa feat: 增加早（9:30~11:30） 午(1:00~14:30) 晚(归纳总结) 三种股票操作逻辑
863a3d7 style:格式化问题修复
5a90d82 feat: 添加命令行单轮对话功能，支持 `-q` 和 `-m` 参数快速查询。
```

**Key Observations**:
- Recent work focused on **tool error handling** (SafeTool wrapper)
- **Session & memory system** still in active development
- **Tool formatting issues** keep appearing (suggests fragility)
- **No commits addressing context or concurrency issues**

---

## 8. Documentation & Existing Docs

### Present:
- ✅ `README.md` (comprehensive, in English & Chinese)
- ✅ `README-zh.md` (Chinese version)
- ✅ `docs/SKILLS.md` (Skills creation guide)
- ✅ `docs/memory-guide.md` (Memory system docs)
- ❌ No architecture documentation
- ❌ No context propagation guide
- ❌ No concurrency model documentation

---

## 9. Summary of Critical Issues

| Issue | Component | Severity | Impact | Effort to Fix |
|-------|-----------|----------|--------|---------------|
| No request-scoped DB context | `db/`, all tools | **CRITICAL** | Breaks multi-concurrency, deadlocks | High |
| Global StreamOutputManager | `message/`, `tui/` | **CRITICAL** | Mixing between concurrent chats | Medium |
| No goroutine lifecycle management | `agent/chat.go` | **HIGH** | Memory leaks, unclear completion | Medium |
| Agent cached globally without mutex | `agent/chat.go` | **HIGH** | Race conditions on model switch | Low |
| No stream subscriber cleanup | `tui/chat.go` | **HIGH** | Memory leaks on errors | Low |
| Session state race condition | `session/manager.go` | **HIGH** | Inconsistent state | Low |
| No context deadline/timeout | `cmd/root.go` | **MEDIUM** | Hangs on stuck operations | Low |
| Tool broadcasting outside context | all tools | **MEDIUM** | Wrong chat receives messages | High |

---

## 10. Recommendations for Improvement

### **Short-term (High Priority)**
1. Add context to all database operations: `func (...) DoSomething(ctx context.Context, db *gorm.DB, ...)`
2. Add mutex protection to `agentCache`
3. Add defer for stream subscriber cleanup
4. Add request-scoped StreamOutputManager (pass via context)

### **Medium-term**
1. Implement proper goroutine lifecycle management with WaitGroup
2. Add context deadlines in `cmd/root.go`
3. Move session state updates behind thread-safe mutex
4. Add request correlation ID for logging/debugging

### **Long-term**
1. Consider moving from global singletons to dependency injection
2. Add integration tests for concurrent chat scenarios
3. Add performance benchmarks for stream broadcasting
4. Document concurrency model and context flow

---

## Conclusion

MSA is a well-structured CLI/TUI application with clear separation of concerns (cmd → app → tui → logic → db). However, **context propagation and inter-component communication are fundamentally flawed for concurrent scenarios**:

1. **Context is created but underutilized**: Created at top level but not passed to database operations
2. **Global singletons everywhere**: StreamOutputManager, SessionManager, database, config - all non-context-aware
3. **No goroutine lifecycle tracking**: Agent spawns goroutine with no completion feedback
4. **Ready to scale but needs architectural refactoring**: Current design works for single-user TUI but breaks with concurrent operations

The codebase is production-ready for **single-user synchronous workflows** but needs significant refactoring for **robust concurrent/async operations**.
