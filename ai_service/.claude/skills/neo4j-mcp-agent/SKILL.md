---
name: neo4j-mcp-agent
description: Use this skill whenever the user needs to inspect a Neo4j schema, read or write property-graph data with Cypher, clean up duplicates, debug Neo4j queries, or explain graph results through Neo4j MCP. Trigger even when the user mentions Neo4j, Cypher, labels, relationships, deduplication, knowledge-graph cleanup, or property-graph modeling without explicitly naming Neo4j MCP.
---

# Neo4j MCP Agent

Use this skill when working through Neo4j via MCP rather than guessing at Cypher from memory. The goal is to make graph work safer, clearer, and easier to verify.

## Compatibility

This skill assumes these Neo4j MCP tools are available:
- `mcp__neo4j__get_neo4j_schema`
- `mcp__neo4j__read_neo4j_cypher`
- `mcp__neo4j__write_neo4j_cypher`

If one or more tools are unavailable, say so clearly and avoid pretending the graph was queried.

## Quick-start workflow

When the task is not trivial, default to this order:
1. inspect schema
2. run one bounded read query
3. write only if the task truly needs a change
4. verify with a follow-up read query
5. explain the result in plain language

## Default workflow

1. Decide whether the task is exploration, reading, writing, cleanup, or debugging.
2. For any non-trivial task, inspect the schema first with `mcp__neo4j__get_neo4j_schema`.
3. Summarize the relevant labels, relationship types, and likely key properties before writing Cypher.
4. Prefer a small, bounded read query before any write that changes existing data.
5. Use `MERGE` for idempotent creation or connection work when duplicates would be harmful.
6. Use `MATCH` plus explicit `SET` clauses for precise updates when identity is already clear.
7. For destructive or high-risk changes, preview impact first and ask for confirmation unless the user explicitly requested the destructive change.
8. After any write, run a verification read query and summarize what changed.

## Query-writing rules

- Keep queries narrow. Prefer explicit labels and bounded patterns over vague graph scans.
- Return only the fields needed to answer the question.
- Add `LIMIT` when exploring or debugging.
- When checking duplicates or anomalies, start with aggregation queries that show counts and representative examples.
- Favor stable business keys for `MERGE`; avoid merging on mutable display names when a stable identifier exists.
- When the schema is unclear, state the assumption you are making.
- If the tool does not support parameters, keep literal values precise and minimize blast radius.

## Operating patterns

### Explore the graph

Start by discovering:
- important labels
- relationship types
- likely primary keys or unique identifiers
- high-value subgraphs relevant to the task

Then explain which part of the graph matters for the user's goal.

### Read data

When the user asks analytical questions:
- map the question to the smallest useful graph pattern
- write readable Cypher
- explain the result in plain language, not just raw rows

### Write or upsert data

When the user wants to create or update data:
- identify the target node key first
- use `MERGE` for nodes or relationships that should be unique
- keep `SET` clauses explicit
- avoid accidental fan-out by matching too broadly

### Cleanup and deduplication

When the user wants to clean data:
- detect candidate duplicates with a read query first
- show how many records would be affected
- present a safe cleanup query
- verify the result afterward

### Debugging

When a query returns no rows or too many rows:
- check labels and relationship directions against the schema
- test the pattern in smaller pieces
- confirm the expected properties actually exist
- explain the likely failure mode before rewriting the query

## Response structure

For non-trivial tasks, prefer this structure:
1. **Goal** — what you are trying to answer or change
2. **Schema notes** — relevant labels, relationships, keys
3. **Cypher** — the query or sequence of queries
4. **What it does** — short explanation in plain language
5. **Verification** — how the result was checked
6. **Risks / next step** — only if relevant

For simple tasks, be shorter.

## High-risk operations

Treat these as risky and worth previewing first whenever possible:
- `DETACH DELETE`
- broad `DELETE`
- removing relationships without a tight `MATCH`
- large-scale rewrites
- `MERGE` on ambiguous properties

If the user clearly wants one of these, still show impact and verification steps instead of performing mystery surgery.

## Examples and reusable patterns

See `references/cypher-patterns.md` for reusable templates covering:
- schema-first exploration
- safe upserts
- duplicate detection
- cleanup verification
- debugging zero-result queries
