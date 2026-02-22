---
name: openspec-collect-knowledge
description: Collect and extract knowledge from archived changes that contain fix.md files. Use after archiving changes with fixes to build the knowledge base.
license: MIT
compatibility: Requires openspec CLI and knowledge-driven schema.
metadata:
  author: openspec
  version: "1.0"
  generatedBy: "1.1.1"
---

Collect knowledge from archived changes and extract it into the knowledge base.

**Input**: Optional flags for collection behavior.

**Steps**

1. **Parse command options**

   Check for these options in the user's request:
   - `--dry-run` - Preview what would be collected without making changes
   - `--change <name>` - Collect only from a specific archived change
   - `--force` - Overwrite existing knowledge files
   - `--rebuild-index` - Rebuild the knowledge index from existing knowledge

   If no options provided, perform a full collection.

2. **Scan for fix.md files**

   Search for fix.md files in archived changes:
   ```bash
   find openspec/changes/archive -name "fix.md" -type f
   ```

   **If --change <name> is specified:**
   Only check `openspec/changes/archive/*<name>*/fix.md`

   **If no fix.md files found:**
   - Inform user no fixes found to collect
   - Suggest creating fix artifacts during implementation
   - Exit successfully

3. **Preview collection (unless --rebuild-index)**

   For each fix.md file found, parse the frontmatter to extract:
   - `id` - Fix identifier
   - `title` - Fix title
   - `type` - runtime|compile|logic|config
   - `severity` - critical|high|medium|low
   - `component` - Affected component
   - `tags` - Searchable tags
   - `status` - open|investigating|resolved|cancelled

   Display preview:
   ```
   📚 Knowledge Collection Report

   Scanning: openspec/changes/archive/*/fix.md
   Found: N fix files

   Processing:
     ✓ 001-streaming-timeout.md → issues/
     ✓ 002-tool-failure.md → issues/ → patterns/
     ✓ 003-config-error.md → issues/
   ```

   **If --dry-run:**
   - Show the preview
   - Show what would be created
   - Exit without making changes

4. **Process each fix.md**

   For each fix file:

   a. **Read and parse fix.md**
      - Extract YAML frontmatter
      - Extract markdown content sections
      - Validate required fields (id, title, type, severity)

   b. **Create issue entry**
      - Target: `openspec/knowledge/issues/<id>-<slug>.md`
      - Copy full fix.md content
      - Add metadata header

      **If file exists and --force not specified:**
      - Skip with warning
      - Suggest using --force to overwrite

   c. **Extract pattern if present**
      - Check for "Reusable Pattern" section in fix.md
      - If present, create: `openspec/knowledge/patterns/<pattern-slug>.md`
      - Extract pattern name, description, implementation

   d. **Extract anti-pattern if present**
      - Check for "Anti-Pattern Identified" section
      - If present, create: `openspec/knowledge/anti-patterns/<anti-pattern-slug>.md`
      - Extract anti-pattern name, problem, what to avoid

5. **Update or create knowledge index**

   Read existing `openspec/knowledge/index.yaml` or create new.

   **Index structure:**
   ```yaml
   version: "1.0"
   lastUpdated: "2025-02-23T10:00:00Z"
   statistics:
     totalIssues: N
     totalPatterns: M
     totalAntiPatterns: O
   issues:
     - id: "001"
       title: "Streaming timeout causes channel block"
       type: "runtime"
       severity: "high"
       component: "streaming-handler"
       tags: ["streaming", "timeout", "channel"]
       file: "issues/001-streaming-timeout.md"
       createdAt: "2025-02-23T10:00:00Z"
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
   tags:
     streaming:
       issues: ["001"]
       patterns: ["p001"]
     timeout:
       issues: ["001"]
   ```

6. **Handle --rebuild-index**

   If `--rebuild-index` flag is present:
   - Skip fix.md scanning
   - Scan existing knowledge/ directory
   - Rebuild index.yaml from all knowledge files
   - Update statistics

7. **Create knowledge directories if needed**

   Ensure these directories exist:
   ```bash
   mkdir -p openspec/knowledge/issues
   mkdir -p openspec/knowledge/patterns
   mkdir -p openspec/knowledge/anti-patterns
   mkdir -p openspec/knowledge/metadata
   ```

8. **Display summary**

   Show collection summary:
   ```
   📚 Knowledge Collection Complete

   Writing to openspec/knowledge/:
     ✓ Created: knowledge/issues/003-config-validation-error.md
     ✓ Created: knowledge/patterns/error-propagation.md
     ✓ Updated: knowledge/index.yaml

   Summary:
     - Issues collected: N
     - Patterns extracted: M
     - Anti-patterns extracted: O
     - Total issues in knowledge: X
     - Total patterns: Y
     - Total anti-patterns: Z

   Knowledge base is now ready for future changes!
   ```

**Output On Success**

```
## 📚 Knowledge Collection Report

Scanning: openspec/changes/archive/*/fix.md
Found: 3 fix files

Processing:
  ✓ 001-streaming-timeout.md → issues/
  ✓ 002-tool-failure.md → issues/ → patterns/
  ✓ 003-config-error.md → issues/

Writing to openspec/knowledge/:
  ✓ Created: knowledge/issues/001-streaming-timeout.md
  ✓ Created: knowledge/issues/002-tool-failure.md
  ✓ Created: knowledge/issues/003-config-error.md
  ✓ Created: knowledge/patterns/non-blocking-stream.md
  ✓ Updated: knowledge/index.yaml

Summary:
  - Issues collected: 3
  - Patterns extracted: 1
  - Total issues in knowledge: 15
  - Total patterns: 8
```

**Guardrails**
- Always show preview of what will be collected
- Don't overwrite existing knowledge without --force
- Validate fix.md frontmatter before processing
- Skip invalid fix files with warning
- Create knowledge directories if they don't exist
- Handle --rebuild-index separately (don't scan fixes)
- Maintain index.yaml as single source of truth
- Use ISO 8601 timestamps for all dates
- Slug-ify filenames (kebab-case, no special chars)
