---
id: "{{AUTO_GENERATED_ID}}"
title: "Fix: {{SHORT_TITLE}}"
type: "{{runtime|compile|logic|config}}"
severity: "{{critical|high|medium|low}}"
component: "{{COMPONENT_NAME}}"
tags:
  - "{{tag1}}"
  - "{{tag2}}"
  - "{{tag3}}"
created_at: "{{TIMESTAMP_ISO8601}}"
resolved_at: "{{TIMESTAMP_ISO8601}}"
change: "changes/{{CHANGE_NAME}}"
status: "resolved"
---

# Fix: {{title}}

## Problem

### Description
{{Clear description of the problem encountered}}

### Reproduction Steps
1. {{Step 1 - What actions lead to the problem?}}
2. {{Step 2 - What are the specific conditions?}}
3. {{Step 3 - What triggers the error?}}

### Error Messages
```
{{Paste the exact error message or stack trace here}}
```

### Impact
- **User Impact:** {{How does this affect users?}}
- **System Impact:** {{How does this affect the system?}}
- **Frequency:** {{How often does this occur?}}

## Root Cause

### Analysis
{{Explanation of why the problem occurred}}

### Related Code Files
- `{{path/to/file.go}}:{{line_number}}`
- `{{path/to/another/file.go}}:{{line_number}}`

## Solution

### Approach
{{High-level description of the solution}}

### Changes Made
- {{Change 1 - what was modified?}}
- {{Change 2 - what was added?}}
- {{Change 3 - what was removed?}}

### Code Diff (Optional)
```diff
--- a/{{path/to/file.go}}
+++ b/{{path/to/file.go}}
@@ -{{line_number}},7 +{{line_number}},9 @@ {{function signature}}{
-    {{old code}}
+    {{new code}}
 }
```

## Verification
{{How to verify that the fix resolves the problem}}

Example:
1. {{Step 1 - How to test the fix}}
2. {{Step 2 - Expected behavior}}
3. {{Step 3 - How to confirm it's working}}

## Lessons Learned

### What Went Wrong
{{What mistakes were made? What assumptions were incorrect?}}

### What We Learned
{{What insights were gained? What would we do differently?}}

### Related Patterns
- **Pattern to follow:** `patterns/{{pattern-name}}.md`
- **Anti-pattern to avoid:** `anti-patterns/{{anti-pattern-name}}.md`

---

**Metadata**
- **Auto-collected:** Yes
- **Collected At:** {{TIMESTAMP_ISO8601}}
- **Collected By:** /opsx:collect
- **Status:** {{resolved|verified|archived}}

---

## Template Fields Guide

### Required Frontmatter Fields

| Field | Description | Example |
|-------|-------------|---------|
| `id` | Unique identifier (auto-generated) | `"001"` |
| `title` | Short descriptive title | `"Streaming timeout causes channel block"` |
| `type` | Type of issue | `runtime`, `compile`, `logic`, `config` |
| `severity` | Impact severity | `critical`, `high`, `medium`, `low` |
| `component` | Affected component | `streaming`, `agent`, `config`, etc. |
| `tags` | Searchable keywords | `timeout`, `channel`, `blocking` |
| `created_at` | When the issue was created | ISO 8601 timestamp |
| `resolved_at` | When the issue was resolved | ISO 8601 timestamp |
| `change` | Source change path | `changes/archive/example-fix` |
| `status` | Current status | `open`, `in-progress`, `resolved` |

### Content Sections

#### Problem
- **Description**: Clear, concise problem statement
- **Reproduction Steps**: Exact steps to reproduce
- **Error Messages**: Actual error output
- **Impact**: User and system impact assessment

#### Root Cause
- **Analysis**: Why the problem occurred
- **Related Code Files**: Affected files with line numbers

#### Solution
- **Approach**: High-level solution description
- **Changes Made**: List of modifications
- **Code Diff**: Optional diff of key changes

#### Verification
- Specific steps to verify the fix works

#### Lessons Learned
- What went wrong and what was learned
- Related patterns to follow or avoid

## Best Practices

1. **Be Specific**: Use concrete examples, not vague descriptions
2. **Include Context**: Add enough detail for someone else to understand
3. **Link Code**: Always reference specific files and line numbers
4. **Document Impact**: Clearly state user and system impact
5. **Extract Patterns**: If the solution is reusable, note it for extraction
6. **Tag Properly**: Use tags from `knowledge/metadata/tags.yaml`

## Example

See `openspec/changes/*/archive/*/fix.md` for real examples.
