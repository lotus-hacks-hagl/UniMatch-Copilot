import asyncio
import logging
from collections import defaultdict

logger = logging.getLogger(__name__)

_queues: dict[str, asyncio.Queue] = defaultdict(asyncio.Queue)

async def enqueue(job_type: str, data: dict):
    await _queues[job_type].put(data)

async def start_all_workers():
    from workers.crawl_worker import crawl_worker_loop
    from workers.analyze_worker import analyze_worker_loop
    
    logger.info("Starting all worker loops...")
    await asyncio.gather(
        crawl_worker_loop(),
        analyze_worker_loop(),
    )

def make_worker_loop(job_type: str, process_fn):
    """Generic worker loop factory."""
    async def loop():
        logger.info(f"Started worker loop for {job_type}")
        while True:
            job = await _queues[job_type].get()
            try:
                await process_fn(job)
            except Exception as e:
                logger.error(f"[{job_type}] job {job.get('job_id')} failed: {e}", exc_info=True)
                from workers.callback import callback_be
                await callback_be(job["callback_url"], job["job_id"],
                                  job_type, "failed", error=str(e),
                                  university_id=job.get("university_id"),
                                  case_id=job.get("case_id"))
            finally:
                _queues[job_type].task_done()
    return loop
