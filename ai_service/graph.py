from neo4j import AsyncGraphDatabase
from config import config
import logging

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

neo4j_driver = AsyncGraphDatabase.driver(
    config.NEO4J_URI,
    auth=(config.NEO4J_USER, config.NEO4J_PASSWORD),
)

async def init_graph_constraints():
    """Run once on startup to create constraints."""
    try:
        async with neo4j_driver.session() as s:
            await s.run("CREATE CONSTRAINT IF NOT EXISTS FOR (u:University) REQUIRE u.be_id IS UNIQUE")
            await s.run("CREATE CONSTRAINT IF NOT EXISTS FOR (p:Program) REQUIRE p.name IS UNIQUE")
            await s.run("CREATE CONSTRAINT IF NOT EXISTS FOR (c:City) REQUIRE c.name IS UNIQUE")
        logger.info("Neo4j constraints ready")
    except Exception as e:
        logger.warning(f"Could not create Neo4j constraints (is DB running?): {e}")

def _flatten(record: dict) -> dict:
    u = dict(record["u"])
    reqs = [dict(r) for r in record.get("reqs", [])] if record.get("reqs") else []
    scholarships = [dict(sc) for sc in record.get("scholarships", [])] if record.get("scholarships") else []
    
    ielts = next((r for r in reqs if r.get("type") == "IELTS"), None)
    gpa   = next((r for r in reqs if r.get("type") == "GPA"), None)
    return {
        "university_id":              u.get("be_id"),
        "name":                       u.get("name"),
        "country":                    u.get("country"),
        "qs_rank":                    u.get("qs_rank"),
        "ielts_min":                  ielts.get("min_score") if ielts else u.get("ielts_min"),
        "sat_required":               u.get("sat_required"),
        "gpa_expectation_normalized": gpa.get("min_score") if gpa else u.get("gpa_expectation_normalized"),
        "tuition_usd_per_year":       u.get("tuition_usd_per_year"),
        "scholarship_available":      len(scholarships) > 0,
        "scholarship_notes":          scholarships[0].get("notes") if scholarships else None,
        "application_deadline":       u.get("application_deadline"),
        "available_majors":           record.get("majors", []),
        "acceptance_rate":            u.get("acceptance_rate"),
    }

async def get_all_universities_flat() -> list[dict]:
    """Analyze worker: get all universities with requirements."""
    async with neo4j_driver.session() as s:
        result = await s.run("""
            MATCH (u:University) WHERE u.be_id IS NOT NULL
            OPTIONAL MATCH (u)-[:REQUIRES]->(r:AdmissionReq)
            OPTIONAL MATCH (u)-[:HAS_PROGRAM]->(p:Program)
            OPTIONAL MATCH (u)-[:OFFERS]->(sc:Scholarship)
            RETURN u,
                   collect(DISTINCT r) AS reqs,
                   collect(DISTINCT p.name) AS majors,
                   collect(DISTINCT sc) AS scholarships
        """)
        records = await result.data()
        return [_flatten(r) for r in records]

async def get_university_flat(university_id: str) -> dict | None:
    """Report worker: get a single university by be_id."""
    async with neo4j_driver.session() as s:
        result = await s.run("""
            MATCH (u:University {be_id: $id})
            OPTIONAL MATCH (u)-[:REQUIRES]->(r:AdmissionReq)
            OPTIONAL MATCH (u)-[:HAS_PROGRAM]->(p:Program)
            OPTIONAL MATCH (u)-[:OFFERS]->(sc:Scholarship)
            OPTIONAL MATCH (u)-[:HAS_DEADLINE]->(d:Deadline)
            RETURN u,
                   collect(DISTINCT r) AS reqs,
                   collect(DISTINCT p.name) AS majors,
                   collect(DISTINCT sc) AS scholarships,
                   collect(DISTINCT d) AS deadlines
        """, id=university_id)
        record = await result.single()
        return _flatten(record) if record else None

async def delete_university_and_orphans(university_id: str):
    """DELETE endpoint: delete University node + orphaned connected nodes."""
    async with neo4j_driver.session() as s:
        await s.run("""
            MATCH (u:University {be_id: $id})-[r]->(n)
            WHERE size([(n)<--(:University) | n]) <= 1
            DETACH DELETE n
        """, id=university_id)
        await s.run(
            "MATCH (u:University {be_id: $id}) DETACH DELETE u",
            id=university_id
        )
