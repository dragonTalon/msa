# MSA Codebase - Quick Reference Card

## At a Glance

| Aspect | Details |
|--------|---------|
| **Project** | My Stock Agent (MSA) - AI-powered stock analysis tool |
| **Language** | Go 1.25.5 |
| **Main Frameworks** | Cobra (CLI), Bubble Tea (TUI), Eino (AI), GORM (ORM), SQLite |
| **Architecture** | Layered: cmd → app → tui → logic → db |
| **Current State** | ✅ Single-user workflows | ❌ Multi-chat concurrent scenarios |
| **Analysis Doc** | `ARCHITECTURE_ANALYSIS.md` (661 lines, comprehensive) |

---

## Key Entry Points

```
main.go
  └─ cmd/root.go::Execute()
      └─ Creates context with signal handling
      └─ Calls config.InitConfig() + db.InitGlobalDB()
      └─ Routes to:
          ├─ app.Run(ctx) → TUI mode (default)
          ├─ extcli.Run(ctx) → CLI single-round mode (-q flag)
          └─ cmd.runResume() → Resume session (--resume flag)
```

---

## Package Responsibilities (Quick Map)

| Package | What It Does | Issues |
|---------|-------------|--------|
| `cmd/` | CLI routing, context creation, signal handling | ✅ Good |
| `config/` | Config file + env vars (singleton cache) | ✅ Good |
| `db/` | SQLite/GORM setup (global singleton) | ❌ No ctx in queries |
| `model/` | Data models (Message, StreamChunk, Account, etc.) | ✅ Good |
| `session/` | Session tracking (global singleton) | ❌ Race conditions |
| `logic/agent/` | Eino ReAct agent + toolCallChecker + Ask() | ❌ No goroutine lifecycle |
| `logic/message/` | StreamOutputManager (global pub/sub) | ❌ Mixes concurrent chats |
| `logic/tools/` | 25+ tools (stateless singletons) | ❌ No ctx in DB ops |
| `logic/finsvc/` | Finance services (orders, positions) | ❌ No ctx in DB ops |
| `tui/` | Bubble Tea chat interface | ⚠️ Subscribes to global broadcaster |
| `extcli/` | CLI single-round mode | ⚠️ Uses global components |

---

## Context Flow (10-second Overview)

```
cmd/root.go creates ctx with signal handling
    ↓
app.Run(ctx)
    ↓
tui.NewChat(ctx) - stores ctx in Chat.ctx
    ↓
Chat.Update() - Bubble Tea event loop
    ↓
handleEnterKey() - spawns agent.Ask(ctx, ...)
    ↓
go func() { agent.Stream(ctx, messages) } - goroutine spawned
    ↓
toolCallChecker() - broadcasts to StreamOutputManager
    ↓
Chat.Update() - receives StreamChunk from channel

❌ PROBLEM: ctx not passed to database operations!
```

---

## 8 Critical Architectural Issues

### CRITICAL (Break concurrency)
1. **No request-scoped context in database** - tools call `db.GetDB()` without ctx
2. **Global StreamOutputManager mixes streams** - concurrent chats share broadcaster
3. **No goroutine lifecycle management** - `Ask()` spawns goroutine with no tracking

### HIGH (Breaks reliability)
4. **Agent cached globally without mutex** - race condition on ResetCache()
5. **No stream subscriber cleanup on error** - memory leaks
6. **Session state race condition** - not thread-safe
7. **No context deadline/timeout** - can hang indefinitely

### MEDIUM (Breaks correctness)
8. **Tool broadcasting outside context scope** - wrong chat receives messages

---

## Global Singletons (Anti-Pattern Alert!)

```go
// ALL of these are global, non-context-aware singletons:

db.globalDB                          // sync.Once protected ✅
config.configCache                  // sync.Once protected ✅
session.globalManager               // sync.Once protected ✅
message.globalStreamManager         // ❌ NO SYNC! Race condition
agent.agentCache                    // ❌ NO SYNC! Race condition
tools.toolMap                       // ❌ NO SYNC! Race condition
```

---

## Quick Fixes (Priority Order)

| Fix | Component | Effort | Impact |
|-----|-----------|--------|--------|
| Add ctx to DB ops | db/, all tools | 1-2h | CRITICAL |
| Protect agentCache with mutex | agent/ | 15m | HIGH |
| Add defer for stream cleanup | tui/ | 30m | HIGH |
| Add context deadlines | cmd/ | 30m | MEDIUM |
| Thread-safe session manager | session/ | 1h | HIGH |
| Request-scoped StreamOutputManager | message/ | 2-3h | CRITICAL |
| Goroutine lifecycle tracking | agent/ | 2-3h | HIGH |

---

## Code Patterns to Watch

### ❌ Anti-Pattern: Global DB Access (Used Everywhere)
```go
func SubmitBuyOrder(ctx context.Context, param *Param) (string, error) {
    database := msadb.GetDB()  // ← IGNORES CTX!
    // ...
}
```

### ❌ Anti-Pattern: Fire-and-Forget Goroutine
```go
func Ask(ctx context.Context, messages string, history []Message) error {
    go func() {
        sr, err := agent.Stream(ctx, queryMsg)
        // ← No way to track completion or get errors!
    }()
    return nil  // Returns immediately
}
```

### ✅ Pattern: Stream Subscriber (At Least Has Cleanup)
```go
streamOutputCh, streamUnregister := message.RegisterStreamOutput(100)
defer streamUnregister()  // Good - but done in handler, not safe on panic
```

---

## How Tools Work (Data Flow)

```
User input in TUI
    ↓
Chat.handleEnterKey()
    ↓
agent.Ask(ctx, question, history)
    ↓
go agent.Stream(ctx, messages) [GOROUTINE]
    ↓
Eino detects tool calls
    ↓
toolCallChecker() extracts tool name + parameters
    ↓
Tool function called (e.g., SubmitBuyOrder(ctx, param))
    ↓
Tool calls message.BroadcastToolStart(toolName, params)
    ↓
Tool operates on database: db := db.GetDB(); ...
    ↓
Tool calls message.BroadcastToolEnd(toolName, result, err)
    ↓
StreamOutputManager broadcasts to TUI Chat.streamOutputCh
    ↓
TUI Chat.Update() receives StreamChunk and renders
```

---

## Testing Entry Points

### Single-Round Chat (CLI Mode)
```bash
msa -q "What's AAPL stock code?" -m "model-name"
# Calls: extcli.Run(ctx, question, modelOverride)
```

### Interactive Chat (TUI Mode)
```bash
msa
# Calls: app.Run(ctx) → tea.NewProgram(tui.NewChat(ctx)).Run()
```

### Resume Session
```bash
msa --resume YYYY-MM-DD_uuid
# Calls: session.GetManager().LoadSession(id) → tui.NewChat(ctx, WithResumeSession(parsed))
```

---

## Important Files to Know

| File | Purpose | Lines |
|------|---------|-------|
| `cmd/root.go` | Context creation, signal handling, routing | ~240 |
| `pkg/app/app.go` | TUI launcher | ~33 |
| `pkg/logic/agent/chat.go` | Eino agent setup, Ask(), toolCallChecker() | ~539 |
| `pkg/logic/message/stream_manage.go` | StreamOutputManager (global broadcaster) | ~124 |
| `pkg/tui/chat.go` | Bubble Tea chat component | ~400+ |
| `pkg/db/global.go` | Global DB singleton | ~85 |
| `pkg/logic/tools/basic.go` | MsaTool interface & registry | ~49 |
| `pkg/session/manager.go` | Global session manager | ~65 |

---

## Gotchas & Traps

| Gotcha | Impact | How to Avoid |
|--------|--------|--------------|
| `db.GetDB()` called without ctx | No timeout enforcement | Pass ctx to all DB ops |
| `agent.agentCache` accessed without mutex | Race condition | Protect with sync.RWMutex |
| Stream subscribers never unregistered | Memory leaks | Use defer immediately after RegisterStreamOutput() |
| Multiple tools broadcast to same global | Stream mixing | Move StreamOutputManager to per-request |
| `agent.Ask()` spawns goroutine with no tracking | Caller doesn't know completion | Return error channel or use sync.WaitGroup |
| Session manager state not locked | Race condition | Add sync.Mutex to Manager struct |
| Context never given deadline | Can hang indefinitely | Add context.WithTimeout() in cmd/root.go |

---

## Performance Notes

- **SQLite**: WAL mode allows multi-reader/single-writer (good for this use case)
- **StreamOutputManager**: Non-blocking sends on normal messages (skips if full), blocking on IsDone
- **Tool registry**: All tools loaded at init() time as singletons (fast lookup, but no request scoping)
- **Config caching**: Single load with sync.Once (very efficient)

---

## Future Improvements

**If you want to support concurrent chats in same process:**
1. Move from global singletons to dependency injection
2. Create per-request context with timeout
3. Make StreamOutputManager per-request (store in context)
4. Implement proper goroutine lifecycle management
5. Add request correlation IDs for logging

**Otherwise, current design is fine for single-user TUI.**

---

## Resources

- **Full Analysis**: `ARCHITECTURE_ANALYSIS.md` (661 lines)
- **Project Docs**: `README.md`, `docs/SKILLS.md`, `docs/memory-guide.md`
- **Recent Commits**: Focus on tool error handling & memory system
- **Tech Stack**: Eino (Cloudwego AI), Bubble Tea, Cobra, GORM, SQLite

---

**Last Updated**: 2026-04-19
**Analysis Depth**: Comprehensive (8 issues identified, 8+ improvement recommendations)
