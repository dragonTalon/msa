---
name: "OPSX: Collect"
description: Collect knowledge from archived changes that contain fix.md files
category: Workflow
tags: [workflow, collect, knowledge, knowledge-driven]
---

Collect knowledge from archived changes and extract it into the knowledge base.

**Input**: Optional flags for collection behavior after `/opsx:collect` (e.g., `/opsx:collect --dry-run`).

**Steps**

1. **Parse command options**

   Check for these options in the user's request:
   - `--dry-run` - Preview what would be collected without making changes
   - `--change <name>` - Collect only from a specific archived change
   - `--force` - Overwrite existing knowledge files
   - `--rebuild-index` - Rebuild the knowledge index from existing knowledge

   If no options provided, perform a full collection.

2. **Execute openspec collect command**

   Run the appropriate openspec command based on options:
   ```bash
   openspec collect
   openspec collect --dry-run
   openspec collect --change <name>
   openspec collect --force
   openspec collect --rebuild-index
   ```

3. **Display results**

   Show the output from the openspec collect command to the user.

**Output On Success**

The command will display a summary like:
```
📚 Knowledge Collection Report

Scanning: openspec/changes/archive/*/fix.md
Found: N fix files

Processing:
  ✓ 001-streaming-timeout.md → issues/
  ✓ 002-tool-failure.md → issues/ → patterns/

Writing to openspec/knowledge/:
  ✓ Created: knowledge/issues/001-streaming-timeout.md
  ✓ Updated: knowledge/index.yaml

Summary:
  - Knowledge items collected: N
  - Total items in knowledge: X
```

**Guardrails**
- Always pass through the exact options provided by the user
- Display the full output from the openspec command
- If the command fails, show the error to the user
- Don't modify the behavior of the openspec collect command
