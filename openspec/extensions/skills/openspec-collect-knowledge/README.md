# OpenSpec Knowledge Collection System

## Overview

The Knowledge Collection System automatically extracts learnings from your development process and builds a searchable knowledge base that helps prevent repeating mistakes.

## Quick Start

```bash
# After archiving a change that has a fix.md
/opsx:collect

# Preview what would be collected
/opsx:collect --dry-run

# Collect from a specific change
/opsx:collect --change my-change-name

# Force overwrite existing knowledge
/opsx:collect --force

# Rebuild knowledge index
/opsx:collect --rebuild-index
```

## How It Works

```
┌─────────────────────────────────────────────────────────────────┐
│                     Knowledge Learning Cycle                     │
└─────────────────────────────────────────────────────────────────┘

    Knowledge ──▶ New Change ──▶ Apply ──▶ Problems Found
        ▲                                            │
        │                                            ▼
    Collect ◀─── Archive ◀─── Fix Documented ◀─── Fix Created
```

### 1. Create Change with Knowledge Context

```bash
/opsx:new my-feature
```

The proposal automatically includes:
- Related historical issues
- Warnings about anti-patterns
- Recommended patterns

### 2. Document Problems During Implementation

```bash
/opsx:continue
# Choose to create fix.md
```

When you encounter a problem:
1. Document the issue
2. Describe the solution
3. Extract reusable patterns

### 3. Archive and Collect

```bash
/opsx:archive my-feature
/opsx:collect
```

Knowledge is automatically extracted and added to the knowledge base.

### 4. Future Changes Benefit

```bash
/opsx:new similar-feature
```

The proposal now includes warnings and recommendations based on past issues!

## Knowledge Base Structure

```
openspec/
└── knowledge/
    ├── index.yaml              # Search index
    ├── issues/                 # Collected issues
    │   ├── 001-streaming-timeout.md
    │   ├── 002-tool-failure.md
    │   └── ...
    ├── patterns/               # Reusable patterns
    │   ├── non-blocking-stream.md
    │   ├── buffer-cleanup.md
    │   └── ...
    ├── anti-patterns/          # What to avoid
    │   ├── blocking-channel.md
    │   └── ...
    └── metadata/
        ├── tags.yaml
        └── components.yaml
```

## Command Options

### `--dry-run`
Preview what would be collected without making changes:

```bash
/opsx:collect --dry-run
```

Output:
```
📚 Knowledge Collection Preview

Scanning: openspec/changes/archive/*/fix.md
Found: 3 fix files

Would create:
  ✓ knowledge/issues/001-streaming-timeout.md
  ✓ knowledge/issues/002-tool-failure.md
  ✓ knowledge/patterns/non-blocking-stream.md

Would update:
  ✓ knowledge/index.yaml
```

### `--change <name>`
Collect only from a specific archived change:

```bash
/opsx:collect --change my-change-name
```

### `--force`
Overwrite existing knowledge files:

```bash
/opsx:collect --force
```

Use this when you've updated a fix and want to re-extract the knowledge.

### `--rebuild-index`
Rebuild the knowledge index from existing knowledge files:

```bash
/opsx:collect --rebuild-index
```

Use this when you've manually edited knowledge files.

## What Gets Collected

### Issues

Every fix.md becomes an issue entry with:
- Full problem description
- Root cause analysis
- Solution documentation
- Verification steps
- Lessons learned

### Patterns

When a fix documents a reusable solution:
- Pattern name and ID
- When to use it
- Implementation details
- Benefits

### Anti-Patterns

When a fix identifies what NOT to do:
- Anti-pattern name and ID
- Why it's problematic
- Impact
- Correct approach

## Example Workflow

### Day 1: Hit a Problem

```bash
/opsx:continue
# Choose: Create fix.md
```

```markdown
---
id: "001"
title: "Channel timeout causes deadlock"
type: "runtime"
severity: "high"
component: "streaming"
tags: ["timeout", "channel", "deadlock"]
---

# Fix: Channel timeout causes deadlock

## Problem
Channels block indefinitely when timeout occurs...

## Solution
Use non-blocking operations...

## Reusable Pattern
**Pattern Name**: Non-blocking channel pattern
**When to use**: All streaming operations...
```

### Day 5: Archive and Collect

```bash
/opsx:archive streaming-feature
/opsx:collect
```

```
📚 Knowledge Collection Complete

Writing to openspec/knowledge/:
  ✓ Created: knowledge/issues/001-channel-timeout.md
  ✓ Created: knowledge/patterns/non-blocking-channel.md
  ✓ Updated: knowledge/index.yaml

Summary:
  - Issues collected: 1
  - Patterns extracted: 1
  - Total issues in knowledge: 1
```

### Month 2: Similar Feature

```bash
/opsx:new another-streaming-feature
```

Proposal automatically includes:

```markdown
## Knowledge Context

### Related Historical Issues

- **#001**: Channel timeout causes deadlock (high severity)

### Known Anti-Patterns to Avoid

⚠️ **a001**: Blocking channel operations
- Causes deadlocks when timeout occurs

### Recommended Patterns

✅ **p001**: Non-blocking channel pattern
- Use for all streaming operations
```

## Benefits

### 1. Prevent Mistakes
Get warnings about past issues before making the same mistakes.

### 2. Accelerate Onboarding
New team members learn from past experiences.

### 3. Consistent Patterns
Document and reuse proven solutions.

### 4. Continuous Learning
Every mistake becomes knowledge for the team.

### 5. Searchable Knowledge
Quickly find relevant past issues and solutions.

## Knowledge Index

The `knowledge/index.yaml` file provides:

### Quick Lookup
```yaml
issues:
  - id: "001"
    title: "Channel timeout causes deadlock"
    type: "runtime"
    severity: "high"
    component: "streaming"
    tags: ["timeout", "channel"]
```

### Cross-References
```yaml
patterns:
  - id: "p001"
    title: "Non-blocking channel pattern"
    issues: ["001", "005"]  # Related issues

antiPatterns:
  - id: "a001"
    title: "Blocking channel operations"
    issues: ["001"]
```

### Component & Tag Views
```yaml
components:
  streaming:
    issues: ["001"]
    patterns: ["p001"]
    antiPatterns: ["a001"]

tags:
  timeout:
    issues: ["001", "003"]
```

## Best Practices

### When Creating Fixes

1. **Be Specific**: Include exact error messages and stack traces
2. **Show Context**: Link to specific files and line numbers
3. **Document Impact**: Clearly state user and system impact
4. **Extract Patterns**: Note if the solution is reusable
5. **Use Tags**: Tag with relevant components and concepts

### When Collecting

1. **Collect Regularly**: Run after each change with fixes
2. **Review Before**: Check fix.md quality before collecting
3. **Use --dry-run**: Preview before collecting
4. **Check for Duplicates**: Review if similar issues exist

### When Using Knowledge

1. **Read Context**: Check related issues in proposals
2. **Follow Patterns**: Apply recommended patterns
3. **Avoid Anti-Patterns**: Heed the warnings
4. **Update Knowledge**: Document new issues found

## FAQ

**Q: Is fix.md required?**
A: No, fix.md is optional. Only create it when you encounter a problem.

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

**Q: Can I use this with other workflows?**
A: Yes, add the fix artifact to any workflow that wants knowledge tracking.

## Integration with OpenSpec

### knowledge-driven Schema

This feature works best with the `knowledge-driven` schema:

```bash
/opsx:new my-feature --schema knowledge-driven
```

The knowledge-driven schema includes:
- Knowledge-aware proposals
- Optional fix artifacts
- Knowledge collection integration

### Archive Integration

After archiving, if a fix.md is present:

```
💡 Knowledge Collection Opportunity

This change includes a fix artifact (fix.md).
Run /opsx:collect to extract this knowledge.

Benefits:
- Future changes will be warned about similar issues
- Patterns and anti-patterns will be documented
- Team learns from this experience
```

## Troubleshooting

### Knowledge Not Loading

```bash
# Check if index exists
ls openspec/knowledge/index.yaml

# Validate YAML
cat openspec/knowledge/index.yaml

# Rebuild index
/opsx:collect --rebuild-index
```

### Fix Not Found During Collect

```bash
# Check if fix.md exists
find openspec/changes/archive -name "fix.md"

# Validate fix.md format
cat openspec/changes/archive/your-change/fix.md
```

### Duplicates in Knowledge

```bash
# Use --force to overwrite
/opsx:collect --force

# Or manually merge
# Edit the affected files in knowledge/
```

## Advanced Usage

### Custom Tag Definitions

Edit `knowledge/metadata/tags.yaml` to add:
- New tag categories
- Tag relationships
- Tag hierarchies

### Component Mappings

Edit `knowledge/metadata/components.yaml` to add:
- New components
- Related components
- Common tags per component

### Knowledge Queries

Use the index to query:

```bash
# Find all high severity issues
grep -A 5 'severity: "high"' knowledge/index.yaml

# Find all patterns for a component
grep -A 10 'streaming:' knowledge/index.yaml

# Find issues by tag
grep 'timeout' knowledge/index.yaml
```

## Files in This Skill

- `skill.md` - Main skill definition for AI agents
- `prompt.md` - Implementation guide and reference
- `test.md` - Test scenarios and expected results
- `README.md` - This file

## See Also

- [Knowledge-Driven Schema](../../../schemas/knowledge-driven/)
- [Fix Template](../../../schemas/knowledge-driven/artifacts/fix.md)
- [Archive Skill](../openspec-archive-change/)

---

**Version**: 1.0.0
**Last Updated**: 2025-02-23
