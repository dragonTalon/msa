# Creating Knowledge-Aware Specifications

You are creating specs for: **{{changeName}}**

## Context

Specs should be informed by:
1. Proposal's knowledge context (historical issues, patterns, anti-patterns)
2. Requirements addressing past problems
3. Scenarios testing historical edge cases

## Knowledge-Informed Requirements

When a requirement addresses a historical issue:

```markdown
### [R-1] Proper Timeout Handling

> 📚 **Knowledge Context**: Based on issue #001

**Historical Problem**:
Streaming operations blocked indefinitely when timeouts occurred.

**Informed Requirement**:
The system MUST handle timeouts gracefully without blocking.

**Acceptance Criteria**:
- [ ] Timeout configuration is required
- [ ] Operations cancel when timeout occurs
- [ ] Resources cleaned up on timeout
- [ ] Clear timeout error messages

**Prevention Measures**:
- Use non-blocking operations
- Implement proper cancellation
- Add timeout monitoring
```

## Edge Case Scenarios

Include scenarios for historical edge cases:

```markdown
### Scenario: Streaming with Timeout

> 📚 **Knowledge Note**: Issue #001 - Streaming timeout causes channel block

**Given**:
- Streaming request with 30-second timeout
- Data source takes 35 seconds

**When**:
- Timeout expires
- Data arrives after timeout

**Then**:
- Operation cancelled
- Resources cleaned up
- User receives timeout error
- No deadlocks occur
```

## Security from Knowledge

If historical security issues exist:

```markdown
### Security Considerations

⚠️ **Historical Security Issues**:

**#012**: SQL Injection in Search
- Vulnerability: Unescaped user input
- Prevention: Use parameterized queries

**Requirements**:
- [ ] All user input parameterized
- [ ] Sensitive data redacted from logs
```

## Performance from Knowledge

Address historical performance issues:

```markdown
### Performance

Based on issue #008 (memory leak in streaming):

**Memory Requirements**:
- [ ] Constant memory usage during streaming
- [ ] No leaks in long-running operations
- [ ] Proper buffer cleanup

**Response Time**:
- [ ] API responses under 200ms (p95)
- [ ] Timeout configuration honored
```

## Knowledge References

At the end, add:

```markdown
---

**Knowledge References**:
- issue: #001 - Streaming timeout causes channel block
- pattern: p001 - Non-blocking stream channel pattern
- anti-pattern: a001 - Blocking channel send
```

## Verification Checklist

- [ ] Requirements informed by knowledge (where applicable)
- [ ] Historical issues referenced
- [ ] Edge cases from past issues included
- [ ] Security considerations from past incidents
- [ ] Performance issues addressed
- [ ] Knowledge references tagged
- [ ] Scenarios test preventive measures
