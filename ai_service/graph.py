from neo4j import AsyncGraphDatabase
from config import config
import logging
from typing import Any

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

def _flatten(record: Any) -> dict:
    if not isinstance(record, dict):
        record = dict(record)
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

async def ingest_crawl_result_to_graph(
    university_id: str,
    payload: dict,
    source_urls: list[str] | None = None,
    crawled_at: str | None = None,
):
    """Persist crawl output into the knowledge graph using the structured crawl result."""
    source_urls = source_urls or []

    university_params = {
        "id": university_id,
        "name": payload.get("name"),
        "country": payload.get("country"),
        "qs_rank": payload.get("qs_rank"),
        "ielts_min": payload.get("ielts_min"),
        "sat_required": payload.get("sat_required"),
        "gpa_expectation_normalized": payload.get("gpa_expectation_normalized"),
        "tuition_usd_per_year": payload.get("tuition_usd_per_year"),
        "scholarship_available": payload.get("scholarship_available"),
        "scholarship_notes": payload.get("scholarship_notes"),
        "application_deadline": payload.get("application_deadline"),
        "acceptance_rate": payload.get("acceptance_rate"),
        "source_urls": source_urls,
        "crawled_at": crawled_at,
    }

    async with neo4j_driver.session() as s:
        await s.run(
            """
            MERGE (u:University {be_id: $id})
            SET u.name = $name,
                u.country = $country,
                u.qs_rank = $qs_rank,
                u.ielts_min = $ielts_min,
                u.sat_required = $sat_required,
                u.gpa_expectation_normalized = $gpa_expectation_normalized,
                u.tuition_usd_per_year = $tuition_usd_per_year,
                u.scholarship_available = $scholarship_available,
                u.scholarship_notes = $scholarship_notes,
                u.application_deadline = $application_deadline,
                u.acceptance_rate = $acceptance_rate,
                u.source_urls = $source_urls,
                u.last_crawled_at = $crawled_at
            """,
            **university_params,
        )

        if payload.get("available_majors") is not None:
            await s.run(
                """
                MATCH (u:University {be_id: $id})-[r:HAS_PROGRAM]->(:Program)
                DELETE r
                """,
                id=university_id,
            )
            majors = [m for m in payload.get("available_majors", []) if m]
            if majors:
                await s.run(
                    """
                    MATCH (u:University {be_id: $id})
                    UNWIND $majors AS major_name
                    MERGE (p:Program {name: major_name})
                    MERGE (u)-[:HAS_PROGRAM]->(p)
                    """,
                    id=university_id,
                    majors=majors,
                )

        await s.run(
            """
            MATCH (u:University {be_id: $id})-[r:REQUIRES]->(req:AdmissionReq {university_be_id: $id})
            DETACH DELETE req
            """,
            id=university_id,
        )

        if payload.get("ielts_min") is not None:
            await s.run(
                """
                MATCH (u:University {be_id: $id})
                MERGE (req:AdmissionReq {type: 'IELTS', university_be_id: $id})
                SET req.min_score = $ielts_min,
                    req.notes = 'Extracted from official admissions-related sources'
                MERGE (u)-[:REQUIRES]->(req)
                """,
                id=university_id,
                ielts_min=payload.get("ielts_min"),
            )

        if payload.get("gpa_expectation_normalized") is not None:
            await s.run(
                """
                MATCH (u:University {be_id: $id})
                MERGE (req:AdmissionReq {type: 'GPA', university_be_id: $id})
                SET req.min_score = $gpa_expectation_normalized,
                    req.notes = 'Normalized to 4.0 scale from crawl output'
                MERGE (u)-[:REQUIRES]->(req)
                """,
                id=university_id,
                gpa_expectation_normalized=payload.get("gpa_expectation_normalized"),
            )

        await s.run(
            """
            MATCH (u:University {be_id: $id})-[r:OFFERS]->(sc:Scholarship {university_be_id: $id})
            DETACH DELETE sc
            """,
            id=university_id,
        )

        if payload.get("scholarship_available"):
            scholarship_name = payload.get("scholarship_notes") or f"Scholarship for {payload.get('name', university_id)}"
            await s.run(
                """
                MATCH (u:University {be_id: $id})
                MERGE (sc:Scholarship {name: $name, university_be_id: $id})
                SET sc.notes = $notes
                MERGE (u)-[:OFFERS]->(sc)
                """,
                id=university_id,
                name=scholarship_name,
                notes=payload.get("scholarship_notes"),
            )

        await s.run(
            """
            MATCH (u:University {be_id: $id})-[r:HAS_DEADLINE]->(d:Deadline {university_be_id: $id})
            DETACH DELETE d
            """,
            id=university_id,
        )

        if payload.get("application_deadline"):
            await s.run(
                """
                MATCH (u:University {be_id: $id})
                MERGE (d:Deadline {university_be_id: $id, intake: 'general'})
                SET d.date = $date
                MERGE (u)-[:HAS_DEADLINE]->(d)
                """,
                id=university_id,
                date=payload.get("application_deadline"),
            )

    logger.info("Ingested crawl result into Neo4j for university_id=%s", university_id)

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
