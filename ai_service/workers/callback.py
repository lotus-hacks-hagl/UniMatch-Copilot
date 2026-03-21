import httpx
import asyncio
import logging

logger = logging.getLogger(__name__)

async def callback_be(callback_url: str, job_id: str, job_type: str,
                      status: str, result: dict = None, error: str = None,
                      case_id: str = None, university_id: str = None):
    payload = {"job_id": job_id, "job_type": job_type,
               "status": status, "error": error, "result": result}
    if case_id: payload["case_id"] = case_id
    if university_id: payload["university_id"] = university_id

    for attempt in range(2):
        try:
            async with httpx.AsyncClient() as c:
                resp = await c.post(callback_url, json=payload, timeout=10)
                resp.raise_for_status()
                return
        except Exception as e:
            logger.error(f"Callback attempt {attempt+1} failed: {e}")
            if attempt == 0: await asyncio.sleep(5)
