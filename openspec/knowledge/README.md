# OpenSpec Knowledge Base

This directory contains the accumulated knowledge from resolved issues, reusable patterns, and anti-patterns discovered during the development process.

## Structure

```
knowledge/
├── index.yaml              # Global index for quick lookup
├── issues/                 # Individual issue records
├── patterns/               # Reusable solution patterns
├── anti-patterns/          # Patterns to avoid
└── metadata/               # Tag definitions and component mappings
```

## Purpose

The knowledge base serves as a collective memory for the project, helping to:
- Avoid repeating the same mistakes
- Reuse proven solutions
- Onboard new developers faster
- Provide context for AI assistants

## How It Works

1. **Discovery**: During `/opsx:apply`, if issues are found, a `fix.md` artifact is created
2. **Collection**: After archiving a change, run `/opsx:collect` to extract fixes into knowledge
3. **Reuse**: When creating new changes with `/opsx:new`, relevant knowledge is automatically loaded

## File Formats

### Issue Records (`issues/*.md`)

Documents specific problems that were encountered and resolved:
- Problem description and reproduction steps
- Root cause analysis
- Solution applied
- Lessons learned

### Patterns (`patterns/*.md`)

Reusable solutions to common problems:
- When to use the pattern
- Implementation approach
- Benefits and trade-offs
- Real-world examples

### Anti-Patterns (`anti-patterns/*.md`)

Patterns that should be avoided:
- What the anti-pattern looks like
- Why it's dangerous
- Correct alternatives
- Real-world examples of failures

## Index

The `index.yaml` file provides a fast, machine-readable overview of all knowledge items:
- All issues with metadata
- Patterns and anti-patterns
- Component mappings
- Tag frequency statistics

## Contributing

Knowledge is collected automatically from `fix.md` artifacts in archived changes. To contribute:

1. When you encounter an issue during `/opsx:apply`, create a `fix.md`
2. Archive the change with `/opsx:archive`
3. Run `/opsx:collect` to extract the knowledge

## Maintenance

- **Keep it specific**: Issues should be concrete, not vague
- **Link related items**: Use `related_patterns` and `related_anti_patterns` fields
- **Update regularly**: Run `/opsx:collect` after archiving changes with fixes
- **Review periodically**: Remove outdated information, merge duplicates

---

**Last Updated**: 2025-02-23
**Knowledge Version**: 1.0
