import json
import logging
from datetime import datetime
from claude_agent_sdk import query, ClaudeAgentOptions, ResultMessage, AssistantMessage
from models import UniversityMetadata
from job_db import update_job
from workers.callback import callback_be
from config import config
import os
import sys
from queuing.job_queue import make_worker_loop

logger = logging.getLogger(__name__)

NEO4J_MCP_URL      = config.NEO4J_MCP_URL

async def process_crawl_job(job: dict):
    university_id = job["university_id"]
    metadata      = job["metadata"]
    job_id        = job["job_id"]
    callback_url  = job["callback_url"]

    update_job(job_id, "processing")

    known_fields = {k: v for k, v in metadata.items() if v is not None}
    null_fields  = [k for k, v in metadata.items() if v is None]

    with open("prompts/crawl_system.txt") as f:
        system_prompt = (f.read()
                         .replace("{university_name}", metadata["name"])
                         .replace("{university_id}",   university_id)
                         .replace("{country}",         metadata["country"]))

    prompt = f"""Research this university and update the knowledge graph.

University: {metadata['name']}
Country: {metadata['country']}
BE ID: {university_id}

Already known: {json.dumps(known_fields, indent=2)}
Fields still null (need to find): {list(null_fields)}

Start by finding the official admission page, then extract all missing fields
and write them to Neo4j using write_neo4j_cypher.
"""

    agentql_env = os.environ.copy()
    agentql_env["AGENTQL_API_KEY"] = config.TINYFISH_API_KEY
    npx_cmd = "npx.cmd" if sys.platform == "win32" else "npx"
    uvx_cmd = "uvx.exe" if sys.platform == "win32" else "uvx"

    neo4j_env = os.environ.copy()
    neo4j_env["NEO4J_URI"] = config.NEO4J_URI
    neo4j_env["NEO4J_USERNAME"] = config.NEO4J_USER
    neo4j_env["NEO4J_PASSWORD"] = config.NEO4J_PASSWORD

    options = ClaudeAgentOptions(
        system_prompt=system_prompt,
        mcp_servers={
            "tinyfish": {
                "type": "stdio",
                "command": npx_cmd,
                "args": ["-y", "agentql-mcp"],
                "env": agentql_env
            },
            "neo4j": {
                "type": "stdio",
                "command": uvx_cmd,
                "args": ["mcp-neo4j-cypher", "--transport", "stdio"],
                "env": neo4j_env
            },
        },
        output_format=UniversityMetadata.model_json_schema(),
        max_turns=25,
    )

    agent_result = None
    source_urls = []

    try:
        async for message in query(prompt=prompt, options=options):
            if isinstance(message, ResultMessage):
                agent_result = message.structured_output
            elif isinstance(message, AssistantMessage):
                for block in message.content:
                    inp = getattr(block, "input", {}) or {}
                    if isinstance(inp, dict) and "url" in inp:
                        u = inp["url"]
                        if u and u not in source_urls:
                            source_urls.append(u)
    except Exception as e:
        logger.error(f"Claude Agent failed: {e}")
        raise

    if not agent_result:
        raise ValueError("Agent returned no final result")

    print("Agent returned result:", agent_result)
    if isinstance(agent_result, UniversityMetadata):
        agent_output = agent_result.model_dump()
    elif isinstance(agent_result, dict):
        agent_output = agent_result
    else:
        clean = str(agent_result).strip().lstrip("```json").lstrip("```").rstrip("```").strip()
        agent_output = json.loads(clean)

    fixed_fields = agent_output.get("fixed_fields", {})
    source_urls  = list(set(agent_output.get("source_urls", []) + source_urls))

    changes = [
        f"{f}: {metadata[f]} → {v}"
        for f, v in fixed_fields.items()
        if metadata.get(f) is not None and v is not None and metadata[f] != v
    ]

    result = {
        **metadata,
        **{k: v for k, v in fixed_fields.items() if v is not None},
        "crawl_status":     "changed" if changes else "ok",
        "changes_detected": changes,
        "source_urls":      source_urls,
        "crawled_at":       datetime.utcnow().isoformat() + "Z",
    }
    
    await callback_be(callback_url, job_id, "crawl_university", "done",
                      result=result, university_id=university_id)
    update_job(job_id, "done", result=result)

crawl_worker_loop = make_worker_loop("crawl_university", process_crawl_job)
