# Skills System

MSA features a dynamic Skills system that allows modular, reusable prompt templates for domain-specific expertise. Skills are automatically selected based on conversation context or can be manually specified.

## Managing Skills

```bash
# List all available skills
msa skills list

# Show skill details
msa skills show base

# Disable a skill
msa skills disable stock-analysis

# Enable a skill
msa skills enable stock-analysis
```

## Using Skills

### Manual Selection (CLI)

```bash
# Use specific skills for the entire session
msa --skills=base,stock-analysis,output-formats
```

### Manual Selection (Chat)

```
/skills: base,stock-analysis
What's the technical analysis of AAPL?
```

### Automatic Selection

The system automatically selects relevant skills based on your question:

```
What's the current price of Tencent?
# Automatically uses: base, stock-analysis
```

## Creating Custom Skills

Create your own skills in `~/.msa/skills/your-skill/SKILL.md`:

```markdown
---
name: my-custom-skill
description: My specialized analysis framework
version: 1.0.0
priority: 7
---

# My Custom Skill

Your specialized knowledge and rules here...
```

### Skill Metadata

| Field | Description |
|-------|-------------|
| `name` | Unique skill identifier (used in CLI/config) |
| `description` | Brief description for skill listings |
| `version` | Semantic version number |
| `priority` | Higher = takes precedence in auto-selection |

## Built-in Skills

MSA ships with several built-in skills such as `base`, `stock-analysis`, and `output-formats`. Use `msa skills list` to see all available skills.
