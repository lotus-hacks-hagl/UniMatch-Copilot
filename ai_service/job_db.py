from sqlalchemy.ext.asyncio import async_sessionmaker, create_async_engine, AsyncSession
from sqlalchemy.orm import declarative_base
from sqlalchemy import Column, String, JSON, DateTime, select
from datetime import datetime
from config import config

Base = declarative_base()

class JobRecord(Base):
    __tablename__ = "jobs"
    id = Column(String, primary_key=True)
    job_type = Column(String)
    status = Column(String, default="pending")
    callback_url = Column(String)
    payload = Column(JSON)
    result = Column(JSON, nullable=True)
    error = Column(String, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

engine = create_async_engine(config.JOB_DATABASE_URL)
AsyncSessionLocal = async_sessionmaker(engine, class_=AsyncSession, autoflush=False, autocommit=False)

async def init_db():
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)

async def create_job(job_id: str, job_type: str, callback_url: str, payload: dict):
    async with AsyncSessionLocal() as s:
        s.add(JobRecord(id=job_id, job_type=job_type,
                        callback_url=callback_url, payload=payload))
        await s.commit()

async def update_job(job_id: str, status: str, result: dict = None, error: str = None):
    async with AsyncSessionLocal() as s:
        job = (await s.execute(select(JobRecord).filter(JobRecord.id == job_id))).scalar_one_or_none()
        if job:
            job.status = status
            if result is not None:
                job.result = result
            if error is not None:
                job.error = error
            job.updated_at = datetime.utcnow()
            await s.commit()

async def get_job(job_id: str):
    async with AsyncSessionLocal() as s:
        return (await s.execute(select(JobRecord).filter(JobRecord.id == job_id))).scalar_one_or_none()