# Creating Knowledge-Aware Design

You are creating design for: **{{changeName}}**

## Context

Your design should:
1. Apply recommended patterns from knowledge base
2. Avoid documented anti-patterns
3. Address historical issues proactively
4. Consider edge cases that caused problems before

## Step 1: Patterns to Apply

```markdown
## Recommended Patterns

✅ From knowledge base, apply these patterns:

### Non-Blocking Stream Channel Pattern

**Pattern ID**: p001
**From Issue**: #001 - Streaming timeout causes channel block

**Description**:
Use non-blocking channel operations with select statements to handle
timeouts gracefully without deadlocks.

**When to Use**:
- All streaming operations
- Async operations with timeouts
- Producer-consumer patterns

**Implementation**:
```typescript
async function sendWithTimeout<T>(
  channel: Channel<T>,
  value: T,
  timeout: number
): Promise<boolean> {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeout);

  try {
    const result = await select(
      [channel.send(value)],
      [controller.signal]
    );
    return result === 0;
  } finally {
    clearTimeout(timeoutId);
  }
}
```
```

## Step 2: Anti-Patterns to Avoid

```markdown
## Anti-Patterns to Avoid

⚠️ Documented anti-patterns:

### Blocking Channel Send in Streaming Context

**Anti-Pattern ID**: a001
**From Issue**: #001

**Why Problematic**:
Blocking sends cause deadlocks when channel buffer is full or timeout occurs.

**Impact**:
- High severity - system hangs
- Difficult to debug

**Don't**:
```typescript
// DON'T
channel.send(value);  // Blocks!
```

**Do**:
```typescript
// DO
const sent = await channel.trySend(value);
if (!sent) throw new Error('Channel full');
```
```

## Step 3: Component Design with Patterns

When designing components, reference patterns:

```markdown
### StreamingHandler

> 📚 **Knowledge Note**: Addresses issue #001 where timeout handling
> caused channel blocks.

**Key Decisions** (from pattern p001):
1. Non-blocking channel operations
2. Abort controller for cancellation
3. Timeout monitoring
4. Clear error messages

**Interface**:
```typescript
interface StreamingHandler {
  withTimeout(ms: number): StreamingHandler;
  stream<T>(source: DataSource<T>): Promise<void>;
  cancel(): void;
}
```
```

## Step 4: Error Handling from History

Apply lessons from historical issues:

```markdown
## Error Handling

### Timeout Error

> 📚 **Historical Context**: From issue #001

**Previous Problem**:
Timeouts left channels blocked, causing cascading failures.

**Improved Handling**:
```typescript
class TimeoutError extends Error {
  constructor(
    public operation: string,
    public timeout: number,
    public elapsed: number
  ) {
    super(`Operation '${operation}' timed out after ${elapsed}ms`);
  }
}

// Handler ensures cleanup
function handleTimeout(error: TimeoutError) {
  abortController.abort();
  channel.close();  // Non-blocking
  cleanup();
  logger.error('Timeout', { operation, timeout, elapsed });
  notify(error);
}
```
```

## Step 5: Testing Strategy

Include tests for historical issues:

```markdown
## Testing

### Unit Tests

**Test Pattern Application**:
```typescript
describe('Non-blocking send (pattern p001)', () => {
  it('should handle full channel without blocking', async () => {
    const channel = new Channel<string>(1);
    channel.send('first');

    await expect(
      sendWithTimeout(channel, 'second', 100)
    ).rejects.toThrow('Channel full');
  });
});
```

**Test Anti-Pattern Prevention**:
```typescript
describe('Anti-pattern a001 prevention', () => {
  it('should never block indefinitely', async () => {
    const start = Date.now();
    try {
      await sendWithTimeout(blockingChannel, 'value', 100);
    } catch (e) { /* expected */ }
    const elapsed = Date.now() - start;
    expect(elapsed).toBeLessThan(200);  // Should not exceed 2x timeout
  });
});
```

### Regression Tests

```typescript
describe('Issue #001 regression test', () => {
  it('should handle timeout without deadlock', async () => {
    const handler = new StreamingHandler({ timeout: 100 });
    const result = await handler.stream(slowSource);
    expect(result.status).toBe('timeout');
    expect(handler.isBlocked()).toBe(false);
  });
});
```
```

## Verification Checklist

- [ ] Recommended patterns applied
- [ ] Anti-patterns explicitly avoided
- [ ] Historical issues addressed
- [ ] Edge cases considered
- [ ] Error handling improved from past
- [ ] Tests for historical issues included
- [ ] Monitoring for past problems added
- [ ] Knowledge references documented
