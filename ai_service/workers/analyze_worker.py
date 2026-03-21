import json
import logging
from claude_agent_sdk import query, ClaudeAgentOptions, ResultMessage
from graph import get_all_universities_flat
from job_db import update_job
from workers.callback import callback_be
from config import config
from models import AnalyzeResult
import os
import sys
from queuing.job_queue import make_worker_loop

logger = logging.getLogger(__name__)

NEO4J_MCP_URL = config.NEO4J_MCP_URL
SYSTEM = "You are an expert admissions counselor with access to Web Automation (TinyFish MCP) and Knowledge Graph (Neo4j MCP)."

def _hard_filter(unis: list[dict], input_data: dict) -> list[dict]:
    """Pure Python hard filter."""
    filtered = []
    for u in unis:
        if input_data.get("preferred_countries"):
            if u.get("country") not in input_data["preferred_countries"]:
                continue
        if not input_data.get("scholarship_required") and u.get("tuition_usd_per_year"):
            if u["tuition_usd_per_year"] > input_data["budget_usd_per_year"] * 1.4:
                continue
        if input_data.get("intended_major") and u.get("available_majors"):
            m = input_data["intended_major"].lower()
            if not any(m[:4] in major.lower() for major in u["available_majors"]):
                continue
        if input_data.get("ielts_overall") and u.get("ielts_min"):
            if input_data["ielts_overall"] < u["ielts_min"] - 1.5:
                continue
        filtered.append(u)

    filtered.sort(key=lambda x: x.get("qs_rank") or 999)
    return filtered

async def process_analyze_job(job: dict):
    case_id    = job["case_id"]
    input_data = job["input"]
    job_id     = job["job_id"]

    update_job(job_id, "processing")

    all_unis = await get_all_universities_flat()
    filtered = _hard_filter(all_unis, input_data)

    if not filtered:
        await callback_be(job["callback_url"], job_id, "analyze_profile", "done",
                          case_id=case_id, result={
                              "profile_summary": {"weaknesses": ["No matching universities in KB"]},
                              "recommendations": [],
                              "confidence_score": 0.3,
                              "escalation_needed": True,
                              "escalation_reason": "No universities in knowledge graph match the given criteria",
                          })
        update_job(job_id, "done")
        return

    with open("prompts/analyze.txt") as f:
        prompt_template = f.read()

    prompt = prompt_template.format(
        student_json=json.dumps(input_data, ensure_ascii=False, indent=2),
        universities_json=json.dumps(filtered, ensure_ascii=False, indent=2),
    )

    agentql_env = os.environ.copy()
    agentql_env["AGENTQL_API_KEY"] = config.TINYFISH_API_KEY
    npx_cmd = "npx.cmd" if sys.platform == "win32" else "npx"
    uvx_cmd = "uvx.exe" if sys.platform == "win32" else "uvx"

    neo4j_env = os.environ.copy()
    neo4j_env["NEO4J_URI"] = config.NEO4J_URI
    neo4j_env["NEO4J_USERNAME"] = config.NEO4J_USER
    neo4j_env["NEO4J_PASSWORD"] = config.NEO4J_PASSWORD

    options = ClaudeAgentOptions(
        system_prompt=SYSTEM,
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
            }
        },
        output_format=AnalyzeResult.model_json_schema(),
        max_turns=15,
    )

    agent_result = None
    try:
        async for message in query(prompt=prompt, options=options):
            if isinstance(message, ResultMessage):
                agent_result = message.structured_output
    except Exception as e:
        logger.error(f"Claude Agent failed: {e}")
        raise
        
    if not agent_result:
        raise ValueError("Agent returned no final result")

    if isinstance(agent_result, AnalyzeResult):
        result = agent_result.model_dump()
    elif isinstance(agent_result, dict):
        result = agent_result
    else:
        clean = str(agent_result).strip().lstrip("```json").lstrip("```").rstrip("```").strip()
        result = json.loads(clean)
    
    await callback_be(job["callback_url"], job_id, "analyze_profile", "done",
                      case_id=case_id, result=result)
    update_job(job_id, "done", result=result)

analyze_worker_loop = make_worker_loop("analyze_profile", process_analyze_job)
