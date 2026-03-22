import json
import logging
import sys
from datetime import UTC, datetime
from pathlib import Path
from claude_agent_sdk import (
    query,
    ClaudeAgentOptions,
    ResultMessage,
    AssistantMessage,
    UserMessage,
)
from models import UniversityMetadata, CrawlResult
from job_db import update_job
from workers.callback import callback_be
from config import config, build_claude_cli_env
from queuing.job_queue import make_worker_loop

logger = logging.getLogger(__name__)


async def process_crawl_job(job: dict):
    university_id = job["university_id"]
    metadata = job["metadata"]
    job_id = job["job_id"]
    callback_url = job.get("callback_url")

    if not config.TINYFISH_API_KEY:
        raise RuntimeError(
            "Missing TINYFISH_API_KEY in environment. Configure it in ai_service/.env and restart containers."
        )
    if not config.EXA_API_KEY:
        raise RuntimeError(
            "Missing EXA_API_KEY in environment. Configure it in ai_service/.env and restart containers."
        )

    update_job(job_id, "processing")

    is_metadata_only = job.get("is_metadata_only", False)
    known_fields = {k: v for k, v in metadata.items() if v is not None}
    null_fields = [k for k, v in metadata.items() if v is None]

    # Optimization: Skip crawling if metadata is complete or manually requested
    if is_metadata_only or not null_fields:
        logger.info("Skipping crawl for %s (is_metadata_only=%s, null_fields=%s)", 
                    metadata["name"], is_metadata_only, len(null_fields))
        result = {
            **metadata,
            "crawl_status": "ok",
            "changes_detected": [],
            "source_urls": [],
            "crawled_at": datetime.now(UTC).isoformat().replace("+00:00", "Z"),
        }
        if callback_url:
            await callback_be(callback_url, job_id, "crawl_university", "done", result=result, university_id=university_id)
        update_job(job_id, "done", result=result)
        return


    with open("prompts/crawl_system.txt") as f:
        system_prompt = (
            f.read()
            .replace("{university_name}", metadata["name"])
            .replace("{university_id}", university_id)
            .replace("{country}", metadata["country"])
        )

    prompt = f"""Research this university and update the knowledge graph.

University: {metadata["name"]}
Country: {metadata["country"]}
BE ID: {university_id}

Already known: {json.dumps(known_fields, indent=2)}
Fields still null (need to find): {list(null_fields)}

Start by finding the official admission page, then extract all missing fields
and write them to Neo4j using write_neo4j_cypher.
"""

    uvx_cmd = "uvx.exe" if sys.platform == "win32" else "uvx"
    npx_cmd = "npx.cmd" if sys.platform == "win32" else "npx"

    neo4j_env = {}
    neo4j_env["NEO4J_URI"] = config.NEO4J_URI
    neo4j_env["NEO4J_USERNAME"] = config.NEO4J_USER
    neo4j_env["NEO4J_PASSWORD"] = config.NEO4J_PASSWORD

    options = ClaudeAgentOptions(
        system_prompt={
            "type": "preset",
            "preset": "claude_code",
        },
        permission_mode="bypassPermissions",
        env=build_claude_cli_env(),
        mcp_servers={
            "neo4j": {
                "type": "stdio",
                "command": uvx_cmd,
                "args": ["mcp-neo4j-cypher", "--transport", "stdio"],
                "env": neo4j_env,
            },
            "exa": {
                "type": "http",
                "url": f"https://mcp.exa.ai/mcp?exaApiKey={config.EXA_API_KEY}",
            },
            "agentql": {
                "type": "stdio",
                "command": npx_cmd,
                "args": ["-y", "agentql-mcp"],
                "env": {
                    "AGENTQL_API_KEY": config.AGENTQL_API_KEY,
                },
            },
        },
        setting_sources=["local", "project", "user"],
        disallowed_tools=[
            "WebSearch",
        ],
        output_format={
            "type": "json_schema",
            "schema": CrawlResult.model_json_schema(),
        },
    )

    agent_result = None
    source_urls = []

    try:
        async for message in query(prompt=prompt, options=options):
            if isinstance(message, AssistantMessage):
                logger.info("AssistantMessage: %s", message)
            elif isinstance(message, UserMessage):
                logger.info("UserMessage: %s", message)
            if isinstance(message, ResultMessage):
                logger.info("Received ResultMessage from agent: %s", message)
                agent_result = message.structured_output
    except Exception as e:
        logger.error(f"Claude Agent failed: {e}")
        raise

    if not agent_result:
        raise ValueError("Agent returned no final result")

    if isinstance(agent_result, CrawlResult):
        agent_output = agent_result.model_dump()
    elif isinstance(agent_result, UniversityMetadata):
        agent_output = agent_result.model_dump()
    elif isinstance(agent_result, dict):
        agent_output = agent_result
    else:
        clean = (
            str(agent_result)
            .strip()
            .lstrip("```json")
            .lstrip("```")
            .rstrip("```")
            .strip()
        )
        agent_output = json.loads(clean)

    fixed_fields = agent_output.get("fixed_fields", {}) or {}
    source_urls = list(set((agent_output.get("source_urls", []) or []) + source_urls))

    fixed_fields = {
        key: value
        for key, value in fixed_fields.items()
        if key not in {"name", "country"}
    }

    has_researched_value = any(v is not None for v in fixed_fields.values())

    if null_fields and has_researched_value and not source_urls:
        raise ValueError("Agent returned researched fields without any source_urls")

    if null_fields and not source_urls and not has_researched_value:
        raise ValueError("Agent finished without any source_urls or researched fields")

    changes = [
        f"{f}: {metadata[f]} → {v}"
        for f, v in fixed_fields.items()
        if metadata.get(f) is not None and v is not None and metadata[f] != v
    ]

    result = {
        **metadata,
        **{k: v for k, v in fixed_fields.items() if v is not None},
        "crawl_status": "changed" if changes else "ok",
        "changes_detected": changes,
        "source_urls": source_urls,
        "crawled_at": datetime.now(UTC).isoformat().replace("+00:00", "Z"),
    }

    if callback_url:
        await callback_be(
            callback_url,
            job_id,
            "crawl_university",
            "done",
            result=result,
            university_id=university_id,
        )
    else:
        logger.info(
            "No callback_url for optional crawl job_id=%s; skipping done callback",
            job_id,
        )
    update_job(job_id, "done", result=result)


crawl_worker_loop = make_worker_loop("crawl_university", process_crawl_job)
