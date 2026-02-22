# Knowledge-Driven Development Schema

A workflow for building a knowledge base from your development experiences and using that knowledge to avoid repeating mistakes.

## Overview

The **knowledge-driven** schema extends traditional spec-driven development with a continuous learning loop:

```
┌──────────────────────────────────────────────────────────────────┐
│                   Knowledge Learning Cycle                       │
└──────────────────────────────────────────────────────────────────┘

    Knowledge ──▶ New Change ──▶ Apply ──▶ Problems Found
        ▲                                            │
        │                                            ▼
    Collect ◀─── Archive ◀─── Fix Documented ◀─── Fix Created
```

## Key Features

### 📚 Knowledge-Aware Planning

- **Proposals** automatically include relevant historical issues
- **Warnings** about anti-patterns before you make mistakes
- **Recommendations** for patterns that have worked before

### 🔧 Fix Artifacts

- Document problems encountered during implementation
- Extract reusable patterns from solutions
- Build an anti-patterns library of what NOT to do

### 🔄 Knowledge Collection

- Automatically extract knowledge from archived changes
- Build a searchable knowledge base over time
- Get smarter with every change

### 🎯 Pattern Recognition

- Identify reusable solutions from fixes
- Document anti-patterns to avoid
- Build a team's collective wisdom

## When to Use This Schema

Use **knowledge-driven** when:

✅ You want to learn from past mistakes
✅ Building a long-term project with a team
✅ Complex domain with many edge cases
✅ High-value system where mistakes are costly
✅ Growing team that needs knowledge sharing

## Artifacts

### Required Artifacts

| Artifact | Description | Dependencies |
|----------|-------------|--------------|
| `proposal` | Change proposal with historical context | - |
| `specs` | Requirements informed by past issues | proposal |
| `design` | Technical design with pattern applications | specs |
| `tasks` | Implementation tasks with risk alerts | design |

### Optional Artifacts

| Artifact | Description | When to Create |
|----------|-------------|----------------|
| `fix` | Document a problem encountered | When you find a bug/issue |

## Workflow

### 1. Create New Change

```bash
/opsx:new add-feature --schema knowledge-driven
```

**What happens**:
- Knowledge base is scanned for related issues
- Relevant patterns and anti-patterns are identified
- Proposal template is pre-populated with warnings and recommendations

**Example output**:
```markdown
# Proposal: add-streaming-feature

## Knowledge Context

### Related Historical Issues

- **#001**: Streaming timeout causes channel block (high severity)
  - This issue occurred when implementing similar features

### Known Anti-Patterns to Avoid

⚠️ **a001**: Blocking channel send in streaming context
- Why it's problematic: Causes deadlocks
- Impact: High - system hangs

### Recommended Patterns

✅ **p001**: Non-blocking stream channel pattern
- Use for: All streaming operations
- Prevents: Deadlocks and hangs
```

### 2. Implement

```bash
/opsx:continue
```

Work through specs, design, and tasks as usual, but with:
- Pattern recommendations in design phase
- Risk alerts in task breakdown
- Test cases based on past failures

### 3. Document Problems (If Found)

```bash
/opsx:continue
# Choose to create fix.md
```

When you encounter a problem:
1. Document the issue in `fix.md`
2. Describe root cause and solution
3. Extract reusable patterns if applicable
4. Note anti-patterns to avoid

### 4. Archive and Collect

```bash
/opsx:archive
/opsx:collect
```

After archiving:
- `fix.md` is processed
- Knowledge is extracted to `knowledge/` directory
- Index is updated for future searches

## Knowledge Base Structure

```
openspec/
├── knowledge/
│   ├── index.yaml              # Search index
│   ├── issues/                 # Collected issues
│   │   ├── 001-streaming-timeout.md
│   │   ├── 002-tool-failure.md
│   │   └── ...
│   ├── patterns/               # Reusable patterns
│   │   ├── non-blocking-stream.md
│   │   └── ...
│   ├── anti-patterns/          # What to avoid
│   │   ├── blocking-channel-send.md
│   │   └── ...
│   └── metadata/
│       ├── tags.yaml           # Tag definitions
│       └── components.yaml     # Component mappings
```

## Knowledge Lifecycle

### 1. Problem Encountered

During implementation, you find an issue:
```markdown
<!-- changes/your-change/fix.md -->
---
id: "001"
title: "Streaming timeout causes channel block"
type: "runtime"
severity: "high"
---
```

### 2. Change Archived

```bash
/opsx:archive your-change
```

Fix is moved to:
```
changes/archive/2025-02-23-your-change/fix.md
```

### 3. Knowledge Collected

```bash
/opsx:collect
```

Knowledge is extracted:
```
knowledge/issues/001-streaming-timeout.md
knowledge/patterns/non-blocking-stream.md
knowledge/anti-patterns/blocking-channel-send.md
```

### 4. Future Changes Informed

Next time someone creates a similar change:
```bash
/opsx:new another-streaming-feature
```

They automatically get:
- Warning about issue #001
- Recommendation for pattern p001
- Warning about anti-pattern a001

## Template Features

### Proposal Template

- **Knowledge Context Section**: Related issues, patterns, anti-patterns
- **Risk Assessment**: Based on historical issues
- **Mitigation Strategy**: How to avoid past problems

### Fix Template

- **Root Cause Analysis**: What went wrong and why
- **Solution Documentation**: How it was fixed
- **Pattern Extraction**: Reusable solution pattern
- **Anti-Pattern Documentation**: What to avoid
- **Lessons Learned**: What the team learned

### Task Template

- **Risk Alerts**: Warnings based on past issues
- **Knowledge Context**: References to related issues
- **Knowledge-Based Tests**: Regression tests for past problems
- **Progress Tracking**: With risk levels

## Comparison with Spec-Driven

| Feature | Spec-Driven | Knowledge-Driven |
|---------|-------------|-----------------|
| Proposal | Standard | With historical context |
| Design | Best practices | Pattern-based |
| Tasks | Implementation | Risk-aware |
| Fix Artifact | No | Yes (optional) |
| Knowledge Collection | No | Yes |
| Pattern Extraction | No | Yes |
| Anti-Pattern Documentation | No | Yes |

## Best Practices

### When Creating Fixes

1. **Be Specific**: Include exact error messages and stack traces
2. **Show Context**: Link to specific files and line numbers
3. **Document Impact**: Clearly state user and system impact
4. **Extract Patterns**: Note if the solution is reusable
5. **Use Tags**: Tag with relevant components and concepts

### When Collecting Knowledge

1. **Collect Regularly**: Run after each change with fixes
2. **Review Before**: Check fix.md quality
3. **Use --dry-run**: Preview before collecting
4. **Check for Duplicates**: Review if similar issues exist

### When Creating New Changes

1. **Read Context**: Check related issues in proposal
2. **Follow Patterns**: Apply recommended patterns
3. **Avoid Anti-Patterns**: Heed the warnings
4. **Update Knowledge**: Document new issues found

## Commands

```bash
# Create new change with knowledge context
/opsx:new feature-name --schema knowledge-driven

# Continue work (creates fix.md if needed)
/opsx:continue

# Archive change
/opsx:archive

# Collect knowledge from archived changes
/opsx:collect

# Preview what would be collected
/opsx:collect --dry-run

# Collect specific change
/opsx:collect --change change-name

# Rebuild knowledge index
/opsx:collect --rebuild-index
```

## Migration from Spec-Driven

You can migrate existing changes:

```bash
# Your current change is using spec-driven
/opsx:status
# Current schema: spec-driven

# Switch to knowledge-driven
/opsx:new feature-name --schema knowledge-driven

# Or update existing change
# Edit openspec/config.yaml: schema: knowledge-driven
/opsx:continue  # Will use knowledge-driven templates
```

## Example Workflow

### Day 1: Create Change

```bash
/opsx:new add-user-auth
```

**Generated proposal includes**:
- No related issues (first time)
- No patterns yet (knowledge base is new)

### Day 5: Hit a Problem

```bash
/opsx:continue
# Choose: Create fix.md
```

**Document the problem**:
- Issue: JWT token validation timing issue
- Root cause: Clock skew between services
- Solution: Add leeway to token validation

### Day 10: Archive and Collect

```bash
/opsx:archive add-user-auth
/opsx:collect
```

**Knowledge extracted**:
- `knowledge/issues/001-jwt-clock-skew.md`
- `knowledge/patterns/token-validation-leeway.md`

### Month 2: Next Change

```bash
/opsx:new add-oauth-integration
```

**Generated proposal includes**:
- Related issue: #001 (both involve tokens)
- Recommended pattern: p001 (token validation leeway)
- Warning: Watch for clock skew issues

**Result**: You avoid the same mistake!

## FAQ

**Q: Is fix.md required?**
A: No, only create it when you encounter a problem.

**Q: Can I edit knowledge files directly?**
A: Yes, but run `/opsx:collect --rebuild-index` after editing.

**Q: How do I search for specific issues?**
A: Use `grep` on `knowledge/index.yaml` or search in `knowledge/issues/`.

**Q: Can knowledge be version controlled?**
A: Yes, commit `knowledge/` to Git like any other files.

**Q: What if I disagree with a pattern?**
A: Knowledge files are guidance. You can edit them or create your own.

**Q: How does this differ from post-mortems?**
A: It's continuous and automated. Every fix becomes knowledge automatically.

**Q: Can I use this with other schemas?**
A: Yes, you can add the fix artifact to any workflow.

## See Also

- [Knowledge System Workflow](../KNOWLEDGE_WORKFLOW.md) - Complete workflow guide
- [Fix Template](./artifacts/fix.md) - Fix artifact template
- [Schema Definition](./schema.yaml) - Schema configuration

---

**Version**: 1.0.0
**Last Updated**: 2025-02-23
