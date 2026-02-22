# Creating Fix Artifacts

Create `fix.md` when encountering unexpected errors, bugs, performance issues, config problems, or design flaws.

## Quick Start

```bash
/opsx:continue
# Choose to create fix.md
```

## Frontmatter

```yaml
---
id: "003"
title: "Descriptive title"
type: "runtime"              # runtime|compile|logic|config
severity: "high"             # critical|high|medium|low
component: "component-name"
tags:
  - "keyword1"
  - "component-name"
created_at: "2025-02-23T10:00:00Z"
resolved_at: "2025-02-23T11:30:00Z"
change: "changes/your-change"
status: "resolved"           # open|investigating|resolved|cancelled
---
```

## Required Sections

### Problem

- What happened?
- What should have happened?
- Reproduction steps
- Error messages with stack traces
- Impact assessment (user/system/business)

### Root Cause

- Investigation process
- Root cause statement
- Contributing factors
- Related files with line numbers

### Solution

- Solution approach
- Alternatives considered
- Changes made (use diff format)

```markdown
#### Add timeout context

**File**: `src/handlers/streaming.ts`

```diff
+ export function createStreamingContext(timeout: number) {
+   const controller = new AbortController();
+   setTimeout(() => controller.abort(), timeout);
+   return { signal: controller.signal };
+ }
```

**Description**: Added abort controller for timeout handling
**Reason**: Allows proper cleanup when timeout occurs
```

### Verification

- How to verify (steps)
- Test results

## Critical: Pattern Extraction

This builds the knowledge base!

### Reusable Pattern (if applicable)

If the solution is generalizable:

```markdown
## Reusable Pattern

**Pattern Name**: Non-Blocking Channel Send with Timeout
**Pattern ID**: p005

**Description**:
Use non-blocking operations with timeout handling to prevent deadlocks.

**When to Use**:
- All streaming operations
- Any channel send in async context

**Implementation**:
```typescript
function sendWithTimeout<T>(
  channel: Channel<T>,
  value: T,
  timeout: number
): boolean {
  const result = channel.trySend(value);
  if (!result) {
    throw new Error(`Channel send timed out after ${timeout}ms`);
  }
  return result;
}
```
```

### Anti-Pattern (if applicable)

```markdown
## Anti-Pattern Identified

**Anti-Pattern Name**: Blocking Channel Send in Async Context
**Anti-Pattern ID**: a003

**Why Problematic**:
Blocking sends in async contexts cause deadlocks when buffer is full.

**Avoid**:
```typescript
// DON'T
channel.send(value);  // May block indefinitely
```
```

## Lessons Learned

```markdown
## Lessons Learned

### What Went Wrong
- Assumed channel would never be full
- Didn't consider timeout interaction with blocking sends
- Missing tests for timeout scenarios

### What We Learned
- Always consider channel capacity in async code
- Timeout handling requires non-blocking operations
- Need comprehensive tests for edge cases

### Preventive Measures
1. Add validation for channel capacity
2. Use non-blocking operations in async contexts
3. Add tests for timeout scenarios
4. Document channel behavior patterns
```

## Quality Checklist

- [ ] Clear, descriptive title
- [ ] Complete frontmatter
- [ ] Problem clearly described
- [ ] Reproduction steps
- [ ] Error messages/logs
- [ ] Impact assessed
- [ ] Root cause identified
- [ ] Related files linked
- [ ] Solution documented with code changes
- [ ] Verification steps
- [ ] Test results
- [ ] Pattern extracted (if applicable)
- [ ] Anti-pattern documented (if applicable)
- [ ] Lessons learned documented

## After Creating Fix

1. Complete the fix
2. Verify thoroughly
3. Archive: `/opsx:archive`
4. Collect knowledge: `/opsx:collect`

The fix will be processed into the knowledge base automatically!

## Common Mistakes

❌ **Too vague**: "Fixed a bug"
✅ **Specific**: "Fixed channel blocking by adding non-blocking send"

❌ **No root cause**: "It didn't work"
✅ **Root cause**: "Channel.send() blocked because buffer was full"

❌ **No verification**: "Should work now"
✅ **Verified**: "Tested with 1000 concurrent requests, no deadlocks"

❌ **No pattern**: "Just fixed it"
✅ **Pattern extracted**: "Created pattern p005 for non-blocking sends"
