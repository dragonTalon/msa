# Knowledge Collection Implementation Guide

This guide helps implement the `/opsx:collect` command functionality.

## Overview

The knowledge collection system:
1. Scans archived changes for `fix.md` files
2. Parses fix.md frontmatter and content
3. Extracts issues, patterns, and anti-patterns
4. Writes to `openspec/knowledge/` directory
5. Updates `openspec/knowledge/index.yaml`

## Directory Structure

```
openspec/
├── knowledge/
│   ├── index.yaml              # Search index and metadata
│   ├── issues/                 # Collected issues
│   │   ├── 001-streaming-timeout.md
│   │   └── 002-tool-failure.md
│   ├── patterns/               # Reusable patterns
│   │   └── non-blocking-stream.md
│   ├── anti-patterns/          # Patterns to avoid
│   │   └── blocking-channel-send.md
│   └── metadata/               # Additional metadata
│       ├── tags.yaml           # Tag definitions
│       └── components.yaml     # Component mappings
└── changes/
    └── archive/
        ├── 2025-02-23-change-a/
        │   └── fix.md          # Source of knowledge
        └── 2025-02-24-change-b/
            └── fix.md
```

## Implementation Commands

### 1. Scan for Fix Files

```bash
# Find all fix.md in archive
find openspec/changes/archive -name "fix.md" -type f

# Find specific change
find openspec/changes/archive -name "*<change-name>*/fix.md"
```

### 2. Parse Fix Frontmatter

Fix.md files have YAML frontmatter:

```yaml
---
id: "001"
title: "Streaming timeout causes channel block"
type: "runtime"
severity: "high"
component: "streaming-handler"
tags:
  - "streaming"
  - "timeout"
  - "channel"
created_at: "2025-02-23T10:00:00Z"
resolved_at: "2025-02-23T11:30:00Z"
change: "changes/add-streaming"
status: "resolved"
---
```

### 3. Extract Knowledge Sections

From fix.md content, look for:

**Issue Section** (always present):
- Problem description
- Root cause analysis
- Solution implemented

**Pattern Section** (optional):
```markdown
## Reusable Pattern

### Pattern Extracted
**Pattern Name**: Non-Blocking Channel Send
**Pattern ID**: p001
**Description**: ...
**When to Use**: ...
**Implementation**: ...
```

**Anti-Pattern Section** (optional):
```markdown
### Anti-Pattern Identified
**Anti-Pattern Name**: Blocking Channel Send
**Anti-Pattern ID**: a001
**Why It's Problematic**: ...
**What to Avoid**: ...
```

### 4. Create Knowledge Files

**Issue File** (`knowledge/issues/<id>-<slug>.md`):
```markdown
---
id: "001"
title: "Streaming timeout causes channel block"
type: "runtime"
severity: "high"
component: "streaming-handler"
tags: ["streaming", "timeout", "channel"]
createdAt: "2025-02-23T10:00:00Z"
resolvedAt: "2025-02-23T11:30:00Z"
sourceArchive: "changes/archive/2025-02-23-add-streaming"
---

# Fix: Streaming timeout causes channel block

[Full fix.md content]
```

**Pattern File** (`knowledge/patterns/<slug>.md`):
```markdown
---
id: "p001"
title: "Non-Blocking Channel Send Pattern"
sourceIssue: "001"
component: "streaming"
tags: ["async", "channel", "non-blocking"]
---

# Non-Blocking Channel Send Pattern

**Pattern ID**: p001
**From Issue**: #001 - Streaming timeout causes channel block

## Description
[Pattern description from fix.md]

## When to Use
[Use cases from fix.md]

## Implementation
```typescript
[Code from fix.md]
```

## Benefits
- Prevents deadlocks
- Handles timeouts gracefully
- Clear error messages

## Related Issues
- #001: Streaming timeout causes channel block
```

**Anti-Pattern File** (`knowledge/anti-patterns/<slug>.md`):
```markdown
---
id: "a001"
title: "Blocking Channel Send in Async Context"
sourceIssue: "001"
component: "streaming"
tags: ["async", "channel", "anti-pattern"]
---

# Blocking Channel Send (Anti-Pattern)

**Anti-Pattern ID**: a001
**From Issue**: #001 - Streaming timeout causes channel block

## Why It's Problematic
[Problem description from fix.md]

## Impact
- High severity - system hangs
- Difficult to debug
- Affects all connected operations

## What NOT to Do
```typescript
// DON'T DO THIS
channel.send(value);  // Blocks indefinitely if full!
```

## Correct Approach
```typescript
// DO THIS INSTEAD
async function sendValue<T>(channel: Channel<T>, value: T) {
  const sent = await channel.trySend(value);
  if (!sent) {
    throw new Error('Channel full - use backpressure');
  }
}
```

## Related Issues
- #001: Streaming timeout causes channel block
```

### 5. Update Index

`knowledge/index.yaml`:

```yaml
version: "1.0"
lastUpdated: "2025-02-23T12:00:00Z"
statistics:
  totalIssues: 15
  totalPatterns: 8
  totalAntiPatterns: 6

issues:
  - id: "001"
    title: "Streaming timeout causes channel block"
    type: "runtime"
    severity: "high"
    component: "streaming-handler"
    tags: ["streaming", "timeout", "channel"]
    file: "issues/001-streaming-timeout.md"
    createdAt: "2025-02-23T10:00:00Z"
    sourceArchive: "changes/archive/2025-02-23-add-streaming"

patterns:
  - id: "p001"
    title: "Non-blocking stream channel pattern"
    issues: ["001"]
    file: "patterns/non-blocking-stream.md"

antiPatterns:
  - id: "a001"
    title: "Blocking channel send"
    issues: ["001"]
    file: "anti-patterns/blocking-channel-send.md"

components:
  streaming-handler:
    issues: ["001"]
    patterns: ["p001"]
    antiPatterns: ["a001"]
    description: "Handles streaming operations"

tags:
  streaming:
    issues: ["001"]
    patterns: ["p001"]
    description: "Streaming data operations"
  timeout:
    issues: ["001"]
    description: "Timeout handling"
  channel:
    issues: ["001"]
    patterns: ["p001"]
    antiPatterns: ["a001"]
    description: "Channel communication"
```

## Command Options

### `--dry-run`
Show what would be collected without making changes:
```bash
/opsx:collect --dry-run
```

Output:
```
📚 Knowledge Collection Preview

Scanning: openspec/changes/archive/*/fix.md
Found: 3 fix files

Would create:
  - knowledge/issues/001-streaming-timeout.md
  - knowledge/issues/002-tool-failure.md
  - knowledge/issues/003-config-error.md
  - knowledge/patterns/non-blocking-stream.md
  - knowledge/anti-patterns/blocking-channel-send.md

Would update:
  - knowledge/index.yaml

Use --force to overwrite existing files.
```

### `--change <name>`
Collect only from a specific archived change:
```bash
/opsx:collect --change add-streaming-timeout
```

### `--force`
Overwrite existing knowledge files:
```bash
/opsx:collect --force
```

### `--rebuild-index`
Rebuild index from existing knowledge (don't scan fixes):
```bash
/opsx:collect --rebuild-index
```

## Error Handling

### Invalid Fix Files
If fix.md has missing required fields:
```
⚠️  Warning: Skipping changes/archive/2025-02-23-bad-fix/fix.md
  Missing required field: id
  Fix file will not be collected
```

### Duplicate Knowledge
If knowledge file already exists (without --force):
```
⚠️  Warning: Skipping knowledge/issues/001-streaming-timeout.md
  File already exists
  Use --force to overwrite
```

### No Fixes Found
If no fix.md files found:
```
📚 Knowledge Collection Report

Scanning: openspec/changes/archive/*/fix.md
Found: 0 fix files

No fixes to collect. Knowledge base unchanged.

To build knowledge:
1. Create fix.md artifacts when you encounter problems
2. Archive changes with fixes
3. Run /opsx:collect again
```

## Helper Functions

### Slugify Title
```typescript
function slugify(title: string): string {
  return title
    .toLowerCase()
    .replace(/[^\w\s-]/g, '')  // Remove special chars
    .replace(/\s+/g, '-')        // Spaces to hyphens
    .replace(/-+/g, '-')         // Collapse multiple hyphens
    .trim();
}
```

### Extract Pattern ID
```typescript
function extractPatternId(fixId: string): string {
  return `p${fixId.padStart(3, '0')}`;
}

// Example: "001" → "p001"
```

### Extract Anti-Pattern ID
```typescript
function extractAntiPatternId(fixId: string): string {
  return `a${fixId.padStart(3, '0')}`;
}

// Example: "001" → "a001"
```

### Parse Frontmatter
```typescript
function parseFrontmatter(content: string): Record<string, any> {
  const match = content.match(/^---\n([\s\S]+?)\n---/);
  if (!match) return {};

  const yaml = match[1];
  // Use YAML parser (e.g., js-yaml)
  return parseYaml(yaml);
}
```

## Testing

### Test Collection
```bash
# Create test fix
mkdir -p openspec/changes/archive/test-change
cat > openspec/changes/archive/test-change/fix.md << 'EOF'
---
id: "999"
title: "Test Issue"
type: "runtime"
severity: "low"
component: "test"
tags: ["test"]
created_at: "2025-02-23T10:00:00Z"
resolved_at: "2025-02-23T11:00:00Z"
change: "test"
status: "resolved"
---

# Test Fix

This is a test.
EOF

# Run collection
/opsx:collect

# Verify
ls openspec/knowledge/issues/
cat openspec/knowledge/index.yaml
```

### Test Rebuild
```bash
# Add manual knowledge
cat > openspec/knowledge/issues/manual-001.md << 'EOF'
---
id: "manual-001"
title: "Manual Issue"
---
EOF

# Rebuild index
/opsx:collect --rebuild-index

# Verify index includes manual issue
grep "manual-001" openspec/knowledge/index.yaml
```

## Integration with Existing Skills

The collect skill integrates with:
- **openspec-archive-change**: After archiving, suggest collecting
- **openspec-new-change**: Use knowledge to inform proposals
- **openspec-continue-change**: Create fix artifacts

## Future Enhancements

Potential improvements:
1. **Knowledge Search**: Search issues by tags/components
2. **Knowledge Graph**: Visualize relationships
3. **Knowledge Merge**: Merge duplicate issues
4. **Knowledge Export**: Export to different formats
5. **Knowledge Stats**: Show statistics and trends
6. **Knowledge Alerts**: Alert on related issues
7. **Knowledge Validation**: Validate knowledge consistency
