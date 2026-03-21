# Iteration 1 Review Setup

This skill is ready for an initial skill-creator review loop.

## Suggested focus areas

When you compare the skill-enabled run against a baseline, look for these differences:

- Does the skill inspect schema first instead of guessing?
- Does it keep read queries bounded and readable?
- Does it use `MERGE` thoughtfully for upserts?
- Does it preview risky cleanup work before destructive writes?
- Does it verify writes with a follow-up read query?
- Does it explain graph reasoning in plain language instead of dumping raw Cypher and vanishing into the mist?

## Suggested human review rubric

For each eval, score the result qualitatively on:

1. correctness of graph reasoning
2. safety of read/write strategy
3. clarity of Cypher
4. verification discipline
5. usefulness of the explanation

## Suggested directory layout for the first run

When you run the formal review loop, use a sibling workspace like this:

- `neo4j-mcp-agent-workspace/iteration-1/schema-exploration/with_skill/`
- `neo4j-mcp-agent-workspace/iteration-1/schema-exploration/without_skill/`
- `neo4j-mcp-agent-workspace/iteration-1/safe-upsert-verification/with_skill/`
- `neo4j-mcp-agent-workspace/iteration-1/safe-upsert-verification/without_skill/`
- `neo4j-mcp-agent-workspace/iteration-1/duplicate-detection-cleanup/with_skill/`
- `neo4j-mcp-agent-workspace/iteration-1/duplicate-detection-cleanup/without_skill/`
- `neo4j-mcp-agent-workspace/iteration-1/debugging-zero-results/with_skill/`
- `neo4j-mcp-agent-workspace/iteration-1/debugging-zero-results/without_skill/`

Inside each eval directory, create an `eval_metadata.json` file with:

- `eval_id`
- `eval_name`
- `prompt`
- `assertions` (start empty for iteration 1, then draft them after runs are launched)

This keeps the review setup aligned with the normal skill-creator loop instead of relying on mystery folders named only by number.

## Suggested next step

Run the four prompts in `evals.json` with and without the skill, then compare:

- whether the skill causes schema-first behavior
- whether it reduces risky Cypher
- whether it improves verification and explanation quality

Once you have outputs:

1. add assertions to each `eval_metadata.json`
2. grade the runs
3. aggregate a benchmark
4. generate the human review viewer

That gives you both the qualitative output review and the quantitative comparison for iteration 1.
