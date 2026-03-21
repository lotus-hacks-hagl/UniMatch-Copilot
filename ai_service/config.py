import os
from dotenv import load_dotenv

load_dotenv()


def _env_stripped(name: str, default: str = "") -> str:
    value = os.getenv(name, default)
    if value is None:
        return default
    return value.strip()

class Config:
    PORT = int(os.getenv("PORT", 9000))
    JOB_DATABASE_URL = _env_stripped("JOB_DATABASE_URL", "postgresql://postgres:password@localhost:5433/unimatch_jobs")
    ANTHROPIC_API_KEY = _env_stripped("ANTHROPIC_API_KEY", "")
    ANTHROPIC_AUTH_TOKEN = _env_stripped("ANTHROPIC_AUTH_TOKEN", "")
    ANTHROPIC_BASE_URL = _env_stripped("ANTHROPIC_BASE_URL", "")
    TINYFISH_API_KEY = _env_stripped("TINYFISH_API_KEY", "")
    NEO4J_URI = _env_stripped("NEO4J_URI", "bolt://localhost:7687")
    NEO4J_USER = _env_stripped("NEO4J_USER", "neo4j")
    NEO4J_PASSWORD = _env_stripped("NEO4J_PASSWORD", "password")
    NEO4J_MCP_URL = _env_stripped("NEO4J_MCP_URL", "http://127.0.0.1:8081/api/mcp/")
    EXA_API_KEY = _env_stripped("EXA_API_KEY", "")
    AGENTQL_API_KEY = _env_stripped("AGENTQL_API_KEY", "")
    
config = Config()


def build_claude_cli_env() -> dict[str, str]:
    env: dict[str, str] = {}

    if config.ANTHROPIC_BASE_URL:
        env["ANTHROPIC_BASE_URL"] = config.ANTHROPIC_BASE_URL

    if config.TINYFISH_API_KEY:
        env["TINYFISH_API_KEY"] = config.TINYFISH_API_KEY

    if config.ANTHROPIC_AUTH_TOKEN:
        env["ANTHROPIC_AUTH_TOKEN"] = config.ANTHROPIC_AUTH_TOKEN
    elif config.ANTHROPIC_API_KEY:
        env["ANTHROPIC_API_KEY"] = config.ANTHROPIC_API_KEY
    elif config.ANTHROPIC_BASE_URL:
        env["ANTHROPIC_API_KEY"] = "proxy-placeholder"
    else:
        raise RuntimeError(
            "Missing Claude authentication. Set ANTHROPIC_API_KEY or ANTHROPIC_AUTH_TOKEN, or configure ANTHROPIC_BASE_URL for proxy mode."
        )

    return env
