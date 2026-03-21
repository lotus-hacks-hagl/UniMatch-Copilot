import httpx
import asyncio
import logging
from typing import Any

logger = logging.getLogger(__name__)

async def callback_be(callback_url: str | None, job_id: str, job_type: str,
                      status: str, result: dict[str, Any] | None = None, error: str | None = None,
                      case_id: str | None = None, university_id: str | None = None):
    if not callback_url or not isinstance(callback_url, str) or not callback_url.strip():
        logger.warning(
            "Skip callback because callback_url is missing/invalid for job_id=%s job_type=%s status=%s",
            job_id,
            job_type,
            status,
        )
        return False

    payload = {"job_id": job_id, "job_type": job_type,
               "status": status, "error": error, "result": result}
    if case_id: payload["case_id"] = case_id
    if university_id: payload["university_id"] = university_id

    for attempt in range(2):
        try:
            async with httpx.AsyncClient() as c:
                resp = await c.post(callback_url, json=payload, timeout=10)
                resp.raise_for_status()
                return True
        except Exception as e:
            logger.error(f"Callback attempt {attempt+1} failed: {e}")
            if attempt == 0: await asyncio.sleep(5)

    return False
