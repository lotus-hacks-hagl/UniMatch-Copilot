import os
from dotenv import load_dotenv

load_dotenv()

class Config:
    PORT = int(os.getenv("PORT", 9000))
    JOB_DATABASE_URL = os.getenv("JOB_DATABASE_URL", "postgresql://postgres:password@localhost:5433/unimatch_jobs")
    ANTHROPIC_API_KEY = os.getenv("ANTHROPIC_API_KEY", "")
    TINYFISH_API_KEY = os.getenv("TINYFISH_API_KEY", "")
    NEO4J_URI = os.getenv("NEO4J_URI", "bolt://localhost:7687")
    NEO4J_USER = os.getenv("NEO4J_USER", "neo4j")
    NEO4J_PASSWORD = os.getenv("NEO4J_PASSWORD", "password")
    NEO4J_MCP_URL = os.getenv("NEO4J_MCP_URL", "http://127.0.0.1:8081/api/mcp/")

config = Config()
