---
name: platform-docs-writer
description: >
  Write professional, structured documentation for any software platform, SaaS product, AI tool,
  API, mobile app, or developer-facing product. Use this skill whenever the user asks to write
  docs, create a documentation page, write a guide, create a tutorial, document a feature, write
  onboarding instructions, or produce any content that looks like product documentation. Trigger
  even if the user says "write instructions for X", "explain how to use X for users", "create a
  getting-started page", "document this feature", or "write a help article". Covers all
  documentation types: user guides, API references, onboarding flows, feature explanations,
  troubleshooting pages, FAQs, changelogs, and platform overviews.
---

# Platform Documentation Writer

A skill for writing high-quality, structured documentation for any software platform or product.
This skill is platform-agnostic and applies to SaaS, mobile apps, developer tools, APIs,
Web3 products, AI platforms, internal tools, and consumer-facing products.

---

## Step 0: Identify Context Before Writing

Before writing anything, answer these questions (infer from context or ask the user):

1. **Who is the reader?**
   - End user (non-technical) → plain language, outcome-focused
   - Developer / integrator → precise, code-heavy, spec-driven
   - Admin / operator → configuration, permissions, system-level
   - Mixed → progressive disclosure (simple intro, advanced later)

2. **What type of page is needed?**
   → See Page Types section below

3. **What platform/product is being documented?**
   → Extract from description, screenshot, or conversation

4. **What is the user trying to accomplish?**
   → This becomes the page's goal and shapes the structure

---

## Page Types

### Type A — Getting Started / Onboarding

**When to use:** First-time user introduction to the platform.

**Structure:**
```markdown
# Getting Started with [Product Name]

[One sentence: what this product does and who it's for.]

## Prerequisites
- [Account, tools, or knowledge needed before starting]

## Step 1: [First Action]
1. [Instruction]
2. [Instruction]
3. [Instruction]

> 💡 [Tip: common mistake or helpful note]

## Step 2: [Next Action]
...

## What's Next?
- [Link to next logical page]
- [Link to related feature]
```

**Rules:**
- Steps must be **numbered**, not bulleted
- UI elements (buttons, tabs, fields) must be **bolded**
- Each step should have one clear action
- End with "What's Next?" to guide the user forward

---

### Type B — Feature Explanation

**When to use:** Documenting a specific feature, module, or concept.

**Structure:**
```markdown
# [Feature Name]

## Overview
[What this feature does and why it's useful — 2-3 sentences max.]

## How It Works
[Brief explanation of the mechanism. No jargon for end users.]

## Using [Feature Name]

1. Go to **[Location in UI]**
2. Click **[Button/Action]**
3. Configure the options:

| Field / Option | Description | Required |
|---|---|---|
| [Field 1] | [What it does] | ✅ Yes |
| [Field 2] | [What it does] | ❌ No |

4. Click **Save** / **Submit** / **Confirm**

> ⚠️ [Warning if any destructive or irreversible action]

## Example
[A realistic scenario showing the feature in use]

## Related
- [Link to related feature or concept]
```

---

### Type C — API Reference

**When to use:** Documenting endpoints, SDKs, or programmatic interfaces.

**Structure:**
```markdown
# [METHOD] /path/to/endpoint

[One-line description of what this endpoint does.]

## Authentication
[How to authenticate — token, API key, OAuth, etc.]

## Request Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| [param] | string | ✅ | [What it does] |
| [param] | integer | ❌ | [What it does, default value] |

## Request Example

```bash
curl -X POST https://api.example.com/v1/resource \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "param1": "value",
    "param2": 42
  }'
```

## Response

```json
{
  "status": "success",
  "data": { ... },
  "created_at": "2026-01-01T00:00:00Z"
}
```

## Error Codes

| Code | Message | Cause |
|---|---|---|
| 400 | bad_request | Malformed input |
| 401 | unauthorized | Missing or invalid API key |
| 404 | not_found | Resource does not exist |
| 500 | server_error | Internal error |
```

---

### Type D — Troubleshooting / FAQ

**When to use:** Helping users resolve issues or answer common questions.

**Structure:**
```markdown
# Troubleshooting [Feature or Topic]

## [Problem Statement as a Question or Symptom]

**Cause:** [Why this happens]

**Solution:**
1. [Step to fix]
2. [Step to fix]

> 💡 If the issue persists, [next action: contact support, check logs, etc.]

---

## [Next Problem]
...
```

**Rules:**
- Frame each issue as the user would experience it ("Why can't I...?", "I'm getting error X")
- Always explain the cause before the fix
- End with an escalation path if self-service fails

---

### Type E — Concept / Reference

**When to use:** Explaining a technical concept, data model, or system behavior (not a how-to).

**Structure:**
```markdown
# [Concept Name]

## What Is [Concept]?
[Plain-language definition — 1-3 sentences]

## Why It Matters
[How understanding this helps the user accomplish their goal]

## Key Terms

| Term | Definition |
|---|---|
| [Term] | [Definition] |

## How [Concept] Works in [Product]
[Product-specific explanation with diagram description or example if useful]

## Related
- [Link]
```

---

### Type F — Changelog / Release Notes

**When to use:** Communicating what changed in a new version or release.

**Structure:**
```markdown
# Changelog — v[X.Y.Z] ([Date])

## New Features
- **[Feature Name]:** [What it does and why it's useful]

## Improvements
- [What improved and how it affects users]

## Bug Fixes
- Fixed: [Description of the bug that was fixed]

## Breaking Changes
> ⚠️ **Action required:** [What users must update or change]
- [Specific breaking change]

## Deprecations
- [What is deprecated and what to use instead]
```

---

## Sidebar / Navigation Structure

When building a full documentation site, organize pages into a logical sidebar.
Use this general template and adapt to the product's domain:

```
Getting Started
  ├── Introduction
  ├── Quick Start
  ├── Installation / Sign Up
  └── Core Concepts

[Core Domain 1]           ← e.g., "Models", "Projects", "Workspaces"
  ├── Overview
  ├── Create / Add
  ├── Manage
  └── Settings

[Core Domain 2]
  ├── ...

Integrations / API
  ├── Authentication
  ├── Endpoints
  └── Webhooks / SDKs

Account & Billing
  ├── Plans & Pricing
  ├── Usage & Limits
  └── Invoices

Troubleshooting
  ├── Common Issues
  ├── Error Reference
  └── Contact Support
```

**Rules for sidebar:**
- Group by user task / domain, not by internal system structure
- Most important sections first
- Limit to 2 levels of nesting maximum
- Each leaf page should be completable in one reading

---

## Universal Writing Rules

These rules apply to every documentation page regardless of type:

### Language
- Use **active voice**: "Click Save" not "The Save button should be clicked"
- Use **present tense**: "The system returns a token" not "The system will return a token"
- Use **second person**: "You can configure..." not "Users can configure..."
- Avoid jargon unless audience is confirmed technical; define terms on first use

### Formatting
- **Bold** all UI elements: buttons, field names, tab labels, menu items
- Use `inline code` for values, variables, filenames, and commands
- Use code blocks (with language tag) for all multi-line code or CLI commands
- Use tables for parameters, options, comparisons — not prose lists
- Use numbered lists for sequential steps, bulleted lists for non-ordered items
- Use blockquotes (`>`) for tips (💡), warnings (⚠️), and important notes (📌)

### Page anatomy
Every page must have:
- [ ] A clear **H1 title** (one per page)
- [ ] A **1-2 sentence intro** explaining what the page covers
- [ ] **Logical section headers** (H2 / H3)
- [ ] A **"Related" or "What's Next"** section at the end

### Quality checklist before output
- [ ] No unexplained acronyms or jargon for the target audience
- [ ] All steps are actionable and testable
- [ ] Every UI element name matches what is actually on screen
- [ ] Code examples are complete and runnable
- [ ] Tables have a header row and consistent column alignment
- [ ] Page can be understood without reading other pages (self-contained)

---

## Extracting Docs from Screenshots or UI Descriptions

When the user provides a screenshot or describes a UI screen:

1. **Identify the screen type** — Is it a form, dashboard, settings panel, list view?
2. **List all visible UI elements** — Labels, buttons, fields, navigation items, states
3. **Infer the user's goal** — What task does this screen support?
4. **Choose the matching page type** (A–F above)
5. **Write the doc** — Reference UI labels exactly as shown; do not paraphrase button names
6. **Add contextual tips** — Flag anything that may confuse a first-time user

---

## Output Format

- Default output: **Markdown (.md)**
- For a single page: one `.md` file
- For a full docs site: a folder structure with one `.md` per page
- If the user needs HTML, convert Markdown to clean semantic HTML
- If the user needs a Word doc, follow the `docx` skill

Always present output as a file, not inline text, so the user can download or deploy it directly.