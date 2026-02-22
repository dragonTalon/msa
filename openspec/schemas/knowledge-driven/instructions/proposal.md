# Creating Knowledge-Aware Proposals

**⚠️ MANDATORY: Must read openspec/knowledge directory before creating proposal**

## Quick Check

First, verify the knowledge base exists:
```bash
ls openspec/knowledge/index.yaml
```

If missing, initialize:
```bash
mkdir -p openspec/knowledge/{issues,patterns,anti-patterns,lessons}
cat > openspec/knowledge/index.yaml << 'EOF'
version: 1
lastUpdated: $(date -u +"%Y-%m-%dT%H:%M:%SZ")
categories:
  issues: 0
  patterns: 0
  antiPatterns: 0
  lessons: 0
items: []
EOF
```

## Step 1: Load Knowledge Base

Read `openspec/knowledge/index.yaml` and extract:
- `items` array - all knowledge entries
- `categories` - statistics
- `lastUpdated` - freshness check

## Step 2: Analyze Change Context

From the change name, identify:
- **Components**: Technical components involved
- **Keywords**: Technical and business terms
- **Categories**: performance, security, integration, config, etc.

## Step 3: Search Related Knowledge

Match against knowledge base by:
- Component name (highest priority)
- Tags/keywords (medium priority)
- Title text (lower priority)

Calculate relevance score:
- Component match: +10
- Tag match: +5
- Category match: +5
- Title keyword: +2

Include items with score >= 5 in proposal.

## Step 4: Create Proposal with Knowledge Context

Structure (in this order):

### 1. Knowledge Context (REQUIRED)

```markdown
## Knowledge Context

### Related Historical Issues
{{#if relatedIssues}}
The following similar issues have been encountered:
{{#each relatedIssues}}
- **{{id}}**: {{title}}
  - Type: {{type}} | Severity: {{severity}}
  - Component: {{component}}
  - Summary: {{summary}}
  - Resolution: {{resolution}}
{{/each}}
{{else}}
No related issues found.
{{/if}}

### Known Anti-Patterns to Avoid
{{#if antiPatterns}}
⚠️ **Warning**: Documented anti-patterns:
{{#each antiPatterns}}
- **{{id}}**: {{title}}
  - Problem: {{problem}}
  - Impact: {{impact}}
{{/each}}
{{else}}
None identified.
{{/if}}

### Recommended Patterns
{{#if patterns}}
✅ **Recommended**:
{{#each patterns}}
- **{{id}}**: {{title}}
  - Description: {{description}}
  - When: {{when}}
{{/each}}
{{else}}
None identified.
{{/if}}
```

### 2. Risk Assessment (REQUIRED)

```markdown
## Risk Assessment

{{#if riskAlerts}}
{{#each riskAlerts}}
- **{{severity}}**: {{message}}
  - Related: {{issueId}}
  - Mitigation: {{mitigation}}
{{/each}}
{{else}}
No previous issues - this is a new area.
{{/if}}
```

### 3. Mitigation Strategy (REQUIRED)

```markdown
## Mitigation Strategy

{{#if riskAlerts}}
To mitigate risks:
{{#each riskAlerts}}
1. Follow pattern {{patternId}}
2. Address {{issueId}}
3. Test {{area}} thoroughly
{{/each}}
{{else}}
Consider creating fix.md if issues emerge. This builds the knowledge base.
{{/if}}
```

### 4. Standard Proposal Sections

Then add: Why, What Changes, Capabilities, Impact (per standard proposal format).

## Verification

Before completing:
- [ ] Knowledge base loaded
- [ ] Change context analyzed
- [ ] Related knowledge searched
- [ ] Knowledge context sections included
- [ ] Risk assessment completed
- [ ] Mitigation strategy provided
- [ ] Standard sections completed

## If Knowledge Base is Empty

Add this note:

```markdown
## Knowledge Context

### Related Historical Issues
No related issues found in knowledge base.

### Known Anti-Patterns to Avoid
No specific anti-patterns identified.

### Recommended Patterns
No specific patterns identified.

## Risk Assessment
Based on historical issues, this change has the following risk factors:
- No previous issues found - this is a new area

## Mitigation Strategy
As this is a new area, be extra careful:
1. Document any issues encountered
2. Create fix.md when problems arise
3. Run /opsx:collect after archiving
4. Build the knowledge base for future changes
```
