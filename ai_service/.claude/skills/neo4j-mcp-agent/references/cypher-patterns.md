# Cypher Patterns for Neo4j MCP

Use these as starting points, not copy-paste destiny.

## 1. Schema-first exploration

Use `mcp__neo4j__get_neo4j_schema` first, then summarize:

- labels
- relationship types
- likely key properties
- suspicious overlaps or duplicate concepts

A good follow-up read query is small and bounded, for example:

```cypher
MATCH (n:SomeLabel)
RETURN n
LIMIT 10
```

Prefer returning selected properties instead of whole nodes once you know the shape.

## 2. Safe read pattern

When answering a user question, narrow the graph pattern first:

```cypher
MATCH (u:University)-[:HAS_PROGRAM]->(p:Program)
WHERE u.country = "Netherlands"
RETURN u.name AS university, collect(DISTINCT p.name) AS programs
LIMIT 25
```

Why this works well:

- explicit labels
- explicit relationship direction
- bounded output
- readable aliases for the final answer

## 3. Safe upsert pattern

When duplicates would be harmful, prefer `MERGE` on stable identifiers:

```cypher
MERGE (u:University {be_id: "uni-123"})
SET u.name = "Example University",
    u.country = "Netherlands"
MERGE (p:Program {name: "Computer Science"})
MERGE (u)-[:HAS_PROGRAM]->(p)
```

If the name is not stable enough to identify a node uniquely, do not use it as the only `MERGE` key unless the schema says it is safe.

## 4. Duplicate detection before cleanup

Start with a preview query:

```cypher
MATCH (s:Scholarship)
WITH toLower(trim(s.name)) AS normalized_name, collect(s) AS nodes
WHERE size(nodes) > 1
RETURN normalized_name, size(nodes) AS dup_count, [n IN nodes | properties(n)][0..5] AS examples
LIMIT 20
```

This tells you whether duplicates exist and gives the user something concrete to review before cleanup.

## 5. Cleanup with verification

After the user confirms a cleanup plan, keep the write query precise and then verify with a read query.

Example structure:

1. preview duplicates
2. perform the cleanup on a tightly scoped subset
3. re-run the preview query to confirm the duplicates are gone

A cautious cleanup template looks like this:

```cypher
// Example approach only: choose a canonical node first
MATCH (s:Scholarship)
WITH toLower(trim(s.name)) AS normalized_name, collect(s) AS nodes
WHERE size(nodes) > 1
WITH normalized_name, head(nodes) AS canonical, tail(nodes) AS duplicates
UNWIND duplicates AS dup
MATCH (dup)<-[r]-(source)
MERGE (source)-[newRel:OFFERS]->(canonical)
DELETE r
DETACH DELETE dup
```

Do not run a pattern like this blindly. First confirm:

- the canonical-node rule is actually correct
- the relationship type being reattached is the right one
- the cleanup scope is limited to the reviewed duplicate set

## 6. Debugging zero-result queries

If a query returns nothing:

- test one hop at a time
- check label spelling
- check relationship direction
- verify the property name exists in the schema
- relax the `WHERE` clause temporarily to find the failing assumption

Good debugging progression:

```cypher
MATCH (u:University)
RETURN u.name, u.country
LIMIT 10
```

then

```cypher
MATCH (u:University)-[:HAS_PROGRAM]->(p:Program)
RETURN u.name, p.name
LIMIT 10
```

then add filters only after the base pattern is confirmed.
