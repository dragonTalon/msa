# Knowledge Collection Test Script

This document provides test scenarios for the `/opsx:collect` command.

## Test Setup

### 1. Create Test Fix Files

```bash
# Create test archived changes with fixes
mkdir -p openspec/changes/archive/2025-02-23-test-timeout
mkdir -p openspec/changes/archive/2025-02-24-test-pattern
mkdir -p openspec/changes/archive/2025-02-25-test-antipattern

# Create fix with issue only
cat > openspec/changes/archive/2025-02-23-test-timeout/fix.md << 'EOF'
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
change: "add-streaming-timeout"
status: "resolved"
---

# Fix: Streaming timeout causes channel block

## Problem

### Description
Timeouts would leave channels blocked, causing deadlocks.

### Root Cause
Blocking channel operations without timeout handling.

## Solution

### Approach
Use non-blocking channel operations with timeout handling.

### Changes Made
#### Non-blocking send
**File**: `src/handlers/streaming.ts`
```typescript
async function sendWithTimeout<T>(channel: Channel<T>, value: T, timeout: number) {
  const controller = new AbortController();
  setTimeout(() => controller.abort(), timeout);
  // ... implementation
}
```
EOF

# Create fix with pattern
cat > openspec/changes/archive/2025-02-24-test-pattern/fix.md << 'EOF'
---
id: "002"
title: "Memory leak in buffer management"
type: "runtime"
severity: "medium"
component: "buffer-manager"
tags:
  - "memory"
  - "buffer"
  - "leak"
created_at: "2025-02-24T10:00:00Z"
resolved_at: "2025-02-24T11:00:00Z"
change: "fix-buffer-leak"
status: "resolved"
---

# Fix: Memory leak in buffer management

## Problem

Buffers were not being released after use.

## Solution

## Reusable Pattern

### Pattern Extracted

**Pattern Name**: Buffer Cleanup Pattern
**Pattern ID**: p002

**Description**:
Always release buffers in a finally block to ensure cleanup.

**When to Use**:
- All buffer operations
- Memory-intensive operations
- Long-running processes

**Implementation**:
```typescript
function processBuffer() {
  const buffer = allocateBuffer();
  try {
    // Process buffer
  } finally {
    buffer.release();  // Always cleanup
  }
}
```

**Benefits**:
- Prevents memory leaks
- Predictable memory usage
- Clean shutdown
EOF

# Create fix with anti-pattern
cat > openspec/changes/archive/2025-02-25-test-antipattern/fix.md << 'EOF'
---
id: "003"
title: "Synchronous file I/O blocks event loop"
type: "runtime"
severity: "high"
component: "file-handler"
tags:
  - "io"
  - "async"
  - "blocking"
created_at: "2025-02-25T10:00:00Z"
resolved_at: "2025-02-25T11:00:00Z"
change: "async-file-operations"
status: "resolved"
---

# Fix: Synchronous file I/O blocks event loop

## Problem

Using fs.readFileSync in request handlers blocked the event loop.

## Solution

## Reusable Pattern

### Pattern Extracted

**Pattern Name**: Async File Operations Pattern
**Pattern ID**: p003

**Description**:
Always use async file operations in request handlers.

**When to Use**:
- All file I/O in request handlers
- Server-side code
- High-throughput scenarios

**Implementation**:
```typescript
// DON'T: const data = fs.readFileSync(path);
// DO: const data = await fs.promises.readFile(path);
```

### Anti-Pattern Identified

**Anti-Pattern Name**: Synchronous File I/O in Event Loop
**Anti-Pattern ID**: a003

**Why It's Problematic**:
Blocking the event loop prevents handling other requests.

**Impact**:
- High severity - server appears frozen
- Poor performance under load
- Bad user experience

**What to Avoid**:
```typescript
// DON'T DO THIS
const data = fs.readFileSync(path);  // Blocks event loop!
```

**Correct Approach**:
```typescript
// DO THIS INSTEAD
const data = await fs.promises.readFile(path);
```
EOF
```

### 2. Run Collection Tests

```bash
# Test 1: Dry run
echo "=== Test 1: Dry Run ==="
/opsx:collect --dry-run

# Test 2: Full collection
echo "=== Test 2: Full Collection ==="
/opsx:collect

# Test 3: Verify knowledge files
echo "=== Test 3: Verify Knowledge Files ==="
ls -la openspec/knowledge/issues/
ls -la openspec/knowledge/patterns/
ls -la openspec/knowledge/anti-patterns/
cat openspec/knowledge/index.yaml

# Test 4: Collect specific change
echo "=== Test 4: Collect Specific Change ==="
rm -rf openspec/knowledge/
/opsx:collect --change test-timeout

# Test 5: Rebuild index
echo "=== Test 5: Rebuild Index ==="
/opsx:collect --rebuild-index
```

## Expected Results

### After Full Collection

**Directory Structure:**
```
openspec/knowledge/
├── index.yaml
├── issues/
│   ├── 001-streaming-timeout-causes-channel-block.md
│   ├── 002-memory-leak-in-buffer-management.md
│   └── 003-synchronous-file-io-blocks-event-loop.md
├── patterns/
│   ├── buffer-cleanup-pattern.md
│   └── async-file-operations-pattern.md
└── anti-patterns/
    └── synchronous-file-io-in-event-loop.md
```

**index.yaml should contain:**
```yaml
version: "1.0"
lastUpdated: "2025-02-25T..."
statistics:
  totalIssues: 3
  totalPatterns: 2
  totalAntiPatterns: 1
issues:
  - id: "001"
    title: "Streaming timeout causes channel block"
    type: "runtime"
    severity: "high"
    component: "streaming-handler"
    tags: ["streaming", "timeout", "channel"]
    file: "issues/001-streaming-timeout-causes-channel-block.md"
  # ... more issues
patterns:
  - id: "p002"
    title: "Buffer Cleanup Pattern"
    issues: ["002"]
    file: "patterns/buffer-cleanup-pattern.md"
  # ... more patterns
antiPatterns:
  - id: "a003"
    title: "Synchronous File I/O in Event Loop"
    issues: ["003"]
    file: "anti-patterns/synchronous-file-io-in-event-loop.md"
```

## Edge Cases to Test

### 1. No Fix Files
```bash
# Remove all fix files
find openspec/changes/archive -name "fix.md" -delete

# Expected: Message saying no fixes found
/opsx:collect
```

### 2. Invalid Fix (Missing Required Fields)
```bash
# Create invalid fix
mkdir -p openspec/changes/archive/invalid-fix
cat > openspec/changes/archive/invalid-fix/fix.md << 'EOF'
---
title: "Invalid Fix"
---
EOF

# Expected: Warning about missing id field, skip this file
/opsx:collect
```

### 3. Duplicate Knowledge
```bash
# Run collection twice
/opsx:collect
/opsx:collect  # Second run

# Expected: Warning about existing files, no overwrite
# Use --force to overwrite
/opsx:collect --force
```

### 4. Pattern Without Anti-Pattern
```bash
# Should handle gracefully
# Pattern section exists but no anti-pattern section
```

### 5. Anti-Pattern Without Pattern
```bash
# Should handle gracefully
# Anti-pattern section exists but no pattern section
```

## Cleanup

```bash
# Remove test data
rm -rf openspec/changes/archive/2025-02-23-test-*
rm -rf openspec/changes/archive/2025-02-24-test-*
rm -rf openspec/changes/archive/2025-02-25-test-*

# Remove generated knowledge
rm -rf openspec/knowledge/
```
