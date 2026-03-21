from sqlalchemy import create_engine, Column, String, JSON, DateTime
from sqlalchemy.orm import declarative_base, sessionmaker
from datetime import datetime
from config import config

Base = declarative_base()

class JobRecord(Base):
    __tablename__ = "jobs"
    id = Column(String, primary_key=True)      # job_id from BE
    job_type = Column(String)
    status = Column(String, default="pending") # pending|processing|done|failed
    callback_url = Column(String)
    payload = Column(JSON)
    result = Column(JSON, nullable=True)
    error = Column(String, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

engine = create_engine(config.JOB_DATABASE_URL)
Base.metadata.create_all(engine)
SessionLocal = sessionmaker(bind=engine, autoflush=False, autocommit=False)

def create_job(job_id: str, job_type: str, callback_url: str, payload: dict):
    with SessionLocal() as s:
        s.add(JobRecord(id=job_id, job_type=job_type,
                        callback_url=callback_url, payload=payload))
        s.commit()

def update_job(job_id: str, status: str, result: dict = None, error: str = None):
    with SessionLocal() as s:
        job = s.query(JobRecord).filter(JobRecord.id == job_id).first()
        if job:
            job.status = status
            if result is not None:
                job.result = result
            if error is not None:
                job.error = error
            job.updated_at = datetime.utcnow()
            s.commit()

def get_job(job_id: str):
    with SessionLocal() as s:
        return s.query(JobRecord).filter(JobRecord.id == job_id).first()
