# Creating Knowledge-Aware Tasks

You are creating tasks for: **{{changeName}}**

## Context

Tasks should:
1. Be aware of historical issues and complexity
2. Include tasks preventing past problems
3. Add verification based on past failures
4. Consider difficulty of similar past changes

## Step 1: Risk Assessment from Knowledge

```markdown
## ⚠️ Risk Alerts

Based on historical issues:

### HIGH: Timeout-Related Deadlocks
- **Issue**: #001 - Streaming timeout causes channel block
- **Risk**: Similar timeout handling could cause deadlocks
- **Mitigation**: Use pattern p001 for all channel operations

### MEDIUM: Memory Leaks in Streaming
- **Issue**: #008 - Memory leak in streaming operations
- **Risk**: Long-running operations may leak memory
- **Mitigation**: Add memory monitoring and cleanup tasks
```

## Step 2: Task Breakdown with Knowledge

### Implementation Tasks

Add knowledge context to tasks:

```markdown
### [T-1] Implement Timeout Configuration

**Type**: feature | **Priority**: high | **Effort**: 2h

> 📚 **Knowledge Context**:
> **From**: Issue #001 - Streaming timeout causes channel block
> **Lesson**: Hardcoded timeouts made deployment difficult.

**Acceptance Criteria**:
- [ ] Timeout is configurable
- [ ] Default: 30s, Min: 1s, Max: 5min
- [ ] Validation implemented
```

### Tasks Applying Patterns

```markdown
### [T-2] Implement Non-Blocking Channel Operations

**Type**: feature | **Priority**: high | **Effort**: 4h

> 📚 **Knowledge Context**:
> **From**: Issue #001 - blocking operations cause deadlocks
> **Lesson**: Use pattern p001

**Acceptance Criteria**:
- [ ] Apply pattern p001
- [ ] No blocking operations
- [ ] Timeout handling
- [ ] Clean cancellation

**Reference**:
```typescript
// Use pattern p001
async function sendWithTimeout<T>(
  channel: Channel<T>,
  value: T,
  timeout: number
): Promise<boolean> {
  // Implementation from pattern p001
}
```
```

### Tasks Preventing Issues

```markdown
### [T-3] Add Memory Monitoring

**Type**: feature | **Priority**: medium | **Effort**: 2h

> 📚 **Knowledge Context**:
> **From**: Issue #008 - Memory leak in streaming operations
> **Lesson**: Long-running operations leaked buffers

**Acceptance Criteria**:
- [ ] Track buffer size
- [ ] Log memory usage
- [ ] Alert on unusual growth
- [ ] Cleanup verified
```

## Step 3: Knowledge-Based Tests

```markdown
## Knowledge-Based Tests

### Test for Issue #001 - Channel Blocking

> 📚 **From**: Issue #001

**Historical Problem**:
Timeouts left channels blocked, causing deadlocks.

**Test Scenario**:
1. Create streaming with 100ms timeout
2. Trigger timeout
3. Verify channel not blocked
4. Verify cleanup
5. Verify no deadlock

**Expected**:
- Times out at ~100ms
- Channel unblocked
- Resources cleaned
- No deadlock
```

## Step 4: Enhanced Definition of Done

```markdown
## Definition of Done

**Standard**:
- [ ] Code implemented
- [ ] Tests passing
- [ ] Code reviewed
- [ ] Documentation updated

**Knowledge-Enhanced**:
- [ ] No anti-patterns used
- [ ] Recommended patterns followed
- [ ] Tests for historical issues pass
- [ ] No regressions
- [ ] Memory/performance acceptable
- [ ] Error handling covers past edge cases
```

## Step 5: During Implementation

```markdown
## If Issues Arise

1. **Document in fix.md**:
   ```bash
   /opsx:continue
   # Choose to create fix.md
   ```

2. **Check if known issue**:
   ```bash
   grep -r "error" openspec/knowledge/issues/
   ```

3. **Extract patterns**:
   - Is this reusable?
   - Should we document a pattern/anti-pattern?

4. **Update tasks**:
   - Add new tasks if needed
   - Update estimates
```

## Verification Checklist

- [ ] Risks from knowledge base addressed
- [ ] Tasks reference historical issues
- [ ] Test tasks include regression tests
- [ ] Estimates consider past complexity
- [ ] Dependencies are logical
- [ ] Definition of done includes knowledge checks
- [ ] Pre-checklist includes knowledge review
