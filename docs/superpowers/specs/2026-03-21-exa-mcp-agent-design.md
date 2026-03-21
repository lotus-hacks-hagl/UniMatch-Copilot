# EXA MCP Agent Design

Date: 2026-03-21

## Goal

Add an EXA MCP skill and wire EXA MCP into the UniMatch crawl flow so the agent can discover likely official admissions-related URLs before using TinyFish for structured extraction and Neo4j for persistence.

## Scope

- Create a hybrid EXA MCP skill under `ai_service/.claude/skills/exa-mcp-agent`
- Add EXA runtime configuration via environment variables
- Expose EXA MCP to the crawl agent in `workers/crawl_worker.py`
- Update `prompts/crawl_system.txt` so EXA is used for source discovery and TinyFish is used for extraction
- Add eval prompts and a review plan for the new skill

## Architecture

1. EXA MCP performs web discovery and candidate-source selection.
2. TinyFish performs targeted extraction from chosen official URLs.
3. Neo4j MCP writes confirmed data into the knowledge graph.

This keeps the tools separated by strength: EXA for search, TinyFish for page extraction, Neo4j for graph mutation.

## Runtime design

- Add `EXA_API_KEY` to configuration and environment.
- Start EXA MCP via stdio using `npx -y exa-mcp-server --tools=web_search_exa`.
- Register the EXA MCP server alongside Neo4j in `ClaudeAgentOptions.mcp_servers`.
- Avoid broad deep-research tooling for now; the crawl flow only needs reliable source discovery.

## Prompt design

The crawl system prompt should instruct the agent to:

1. use EXA first to identify likely official university pages,
2. prefer official domains over third-party summaries,
3. pass chosen URLs to TinyFish one page at a time,
4. write only reasonably verified data to Neo4j.

## Verification

- Validate edited files for syntax/lint issues.
- Confirm the new skill files are well-formed.
- Ensure the prompt and worker agree on tool names and responsibilities.

## Risks

- EXA MCP packaging may differ across environments; use the published `npx -y exa-mcp-server` pattern.
- Missing `EXA_API_KEY` should fail clearly for crawl jobs that rely on EXA discovery.
