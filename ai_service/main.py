from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from contextlib import asynccontextmanager
import asyncio
import logging

from graph import init_graph_constraints, neo4j_driver, get_university_flat
from job_db import create_job, get_job
from queuing.job_queue import enqueue, start_all_workers
from models import CrawlJobRequest, AnalyzeJobRequest
from config import config

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

@asynccontextmanager
async def lifespan(app: FastAPI):
    logger.info("Starting AI Service...")
    await init_graph_constraints()
    asyncio.create_task(start_all_workers())
    yield
    logger.info("Shutting down AI Service...")

app = FastAPI(
    title="UniMatch AI Service API",
    description="Background jobs handler for UniMatch Copilot. Interfaces with Neo4j and Claude agents.",
    version="1.0.0",
    lifespan=lifespan
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.post("/jobs/crawl", tags=["Jobs"], summary="Submit a University Crawl Job")
async def submit_crawl(body: CrawlJobRequest):
    """
    Enqueue a background job to crawl and merge university data via TinyFish MCP and Neo4j MCP.
    """
    try:
        async with neo4j_driver.session() as s:
            await s.run("""
                MERGE (u:University {be_id: $id})
                SET u.name = $name, u.country = $country, u.updated_at = datetime()
            """, id=body.university_id,
                 name=body.metadata.name,
                 country=body.metadata.country)
    except Exception as e:
        logger.warning(f"Failed to pre-create university node (Neo4j might be down): {e}")

    create_job(body.job_id, "crawl_university", body.callback_url, body.model_dump(mode="json"))
    await enqueue("crawl_university", body.model_dump(mode="json"))
    return {"accepted": True, "job_id": body.job_id}

@app.post("/jobs/analyze", tags=["Jobs"], summary="Submit an Analyze Profile Job")
async def submit_analyze(body: AnalyzeJobRequest):
    """
    Enqueue a background job to match a student profile against the internal Knowledge Graph.
    """
    create_job(body.job_id, "analyze_profile", body.callback_url, body.model_dump(mode="json"))
    await enqueue("analyze_profile", body.model_dump(mode="json"))
    return {"accepted": True, "job_id": body.job_id}

@app.delete("/jobs/university/{university_id}", tags=["Graph Sync"], summary="Delete University from Graph")
async def delete_university_graph(university_id: str):
    """
    Sync deletion from BE: deletes everything referencing the university ID in Neo4j graph.
    """
    from graph import delete_university_and_orphans
    try:
        await delete_university_and_orphans(university_id)
        return {"deleted": True, "university_id": university_id}
    except Exception as e:
        logger.error(f"Failed to delete university {university_id}: {e}")
        raise HTTPException(500, detail=str(e))

@app.get("/jobs/{job_id}", tags=["Jobs"], summary="Get Job Status")
async def get_job_status(job_id: str):
    job = get_job(job_id)
    if not job:
        raise HTTPException(404, "Job not found")
    return {"job_id": job_id, "status": job.status, "error": job.error, "result": job.result}

@app.get("/graph/university/{university_id}", tags=["Debug"], summary="Get Flat Graph Node")
async def get_graph_node(university_id: str):
    """Debug: view graph data for 1 university."""
    data = await get_university_flat(university_id)
    if not data:
        raise HTTPException(404, "University not found in graph")
    return data

@app.get("/health", tags=["Health"], summary="Healthcheck API")
async def health():
    try:
        async with neo4j_driver.session() as s:
            result = await s.run("MATCH (u:University) RETURN count(u) as n")
            record = await result.single()
            uni_count = record["n"] if record else 0
    except Exception:
        uni_count = -1
    return {"status": "ok", "graph_university_count": uni_count}
