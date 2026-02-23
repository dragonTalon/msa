---
name: "OPSX: Collect"
description: Collect knowledge from archived changes that contain fix.md files
category: Workflow
tags: [workflow, collect, knowledge, knowledge-driven]
---

Collect knowledge from archived changes and extract it into the knowledge base.

**Note**: The `openspec` CLI does not currently have a `collect` subcommand. This skill implements knowledge collection manually by scanning archived changes and extracting knowledge from fix.md files.

**Input**: Optional flags for collection behavior after `/opsx:collect` (e.g., `/opsx:collect --dry-run`).

**Steps**

1. **Parse command options**

   Check for these options in the user's request:
   - `--dry-run` - Preview what would be collected without making changes
   - `--change <name>` - Collect only from a specific archived change
   - `--force` - Overwrite existing knowledge files

   If no options provided, perform a full collection.

2. **Scan for fix.md files**

   Find all fix.md files in the archive:
   ```bash
   find openspec/changes/archive/ -name "fix.md" -type f
   ```

3. **Process each fix.md**

   For each fix.md file found:
   - Parse the content to identify:
      - **Issues** (problems encountered, with emoji indicators)
      - **Patterns** (solutions, best practices, design patterns)
      - **Guides** (step-by-step processes, workflows)
   - Extract relevant knowledge items
   - Categorize into appropriate knowledge directories

4. **Write knowledge files**

   Create or update files in `openspec/knowledge/`:
   - `openspec/knowledge/issues/*.md` - Problems encountered
   - `openspec/knowledge/patterns/*.md` - Solutions and best practices
   - `openspec/knowledge/guides/*.md` - Processes and workflows
   - Update `openspec/knowledge/index.yaml` with new entries

5. **Display results**

   Show a summary of collected knowledge items.

**Knowledge Categorization**

When extracting from fix.md, use these rules:

| Content Type | Indicator | Destination |
|--------------|-----------|-------------|
| Critical problem | ⚠️ emoji, "关键" | issues/ |
| Known limitation | 🚫 emoji, "已知限制" | issues/ |
| Solution/workaround | "解决方案", "预防措施" | patterns/ |
| Design pattern | "设计模式" | patterns/ |
| Best practice | "最佳实践", "经验总结" | patterns/ |
| Workflow | "工作流程", "步骤" | guides/ |

**Output On Success**

```
📚 Knowledge Collection Report

Scanning: openspec/changes/archive/*/fix.md
Found: 1 fix file(s)

Processing:
  ✓ 2026-02-23-add-web-search-tool/fix.md

Extracted knowledge items:
  ✓ Issue: AI auto-commit problem (流程问题)
  ✓ Issue: Google CAPTCHA detection (已知限制)
  ✓ Issue: chromedp timeout configuration
  ✓ Pattern: Browser manager singleton
  ✓ Pattern: Error handling strategy
  ✓ Pattern: Security-first configuration

Writing to openspec/knowledge/:
  ✓ Created: knowledge/issues/ai-auto-commit-problem.md
  ✓ Created: knowledge/issues/google-captcha-detection.md
  ✓ Created: knowledge/patterns/browser-manager-singleton.md
  ✓ Updated: knowledge/index.yaml

Summary:
  - Knowledge items collected: 6
  - Total items in knowledge: X
```

**Guardrails**
- Always check for `--dry-run` before making any file changes
- Preserve existing knowledge files unless `--force` is specified
- Use descriptive filenames based on the knowledge content
- Update the index.yaml file after writing new knowledge
- If no fix.md files are found, inform the user
