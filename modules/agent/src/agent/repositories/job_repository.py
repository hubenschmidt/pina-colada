"""Repository layer for job data access."""

import logging
import os
from typing import List, Optional
from sqlalchemy import create_engine, select, func as sql_func
from sqlalchemy.orm import Session, sessionmaker
from agent.models.job import AppliedJob, Base, JobCreateData, JobUpdateData, orm_to_dict, dict_to_orm, update_orm_from_dict

logger = logging.getLogger(__name__)

_engine = None
_session_factory = None


def _get_connection_string() -> str:
    """Get Postgres connection string from environment."""
    db_host = os.getenv("POSTGRES_HOST", "postgres")
    db_port = os.getenv("POSTGRES_PORT", "5432")
    db_user = os.getenv("POSTGRES_USER", "postgres")
    db_password = os.getenv("POSTGRES_PASSWORD", "postgres")
    db_name = os.getenv("POSTGRES_DB", "pina_colada")
    
    return f"postgresql://{db_user}:{db_password}@{db_host}:{db_port}/{db_name}"


def _get_engine():
    """Get or create SQLAlchemy engine."""
    global _engine
    if _engine is not None:
        return _engine
    
    conn_string = _get_connection_string()
    _engine = create_engine(conn_string, pool_pre_ping=True)
    Base.metadata.create_all(_engine)
    return _engine


def _get_session() -> Session:
    """Get database session."""
    global _session_factory
    if _session_factory is None:
        engine = _get_engine()
        _session_factory = sessionmaker(bind=engine)
    
    return _session_factory()


def find_all_jobs() -> List[AppliedJob]:
    """Find all jobs."""
    session = _get_session()
    try:
        stmt = select(AppliedJob)
        return list(session.execute(stmt).scalars().all())
    finally:
        session.close()


def count_jobs() -> int:
    """Count total jobs."""
    session = _get_session()
    try:
        stmt = select(sql_func.count(AppliedJob.id))
        return session.execute(stmt).scalar() or 0
    finally:
        session.close()


def find_job_by_id(job_id: str) -> Optional[AppliedJob]:
    """Find job by ID."""
    session = _get_session()
    try:
        return session.get(AppliedJob, job_id)
    finally:
        session.close()


def create_job(data: JobCreateData) -> AppliedJob:
    """Create a new job."""
    session = _get_session()
    try:
        job = dict_to_orm(data)
        session.add(job)
        session.commit()
        session.refresh(job)
        return job
    except Exception as e:
        session.rollback()
        logger.error(f"Failed to create job: {e}")
        raise
    finally:
        session.close()


def update_job(job_id: str, data: JobUpdateData) -> Optional[AppliedJob]:
    """Update an existing job."""
    session = _get_session()
    try:
        job = session.get(AppliedJob, job_id)
        if not job:
            return None
        
        update_orm_from_dict(job, data)
        session.commit()
        session.refresh(job)
        return job
    except Exception as e:
        session.rollback()
        logger.error(f"Failed to update job: {e}")
        raise
    finally:
        session.close()


def delete_job(job_id: str) -> bool:
    """Delete a job by ID."""
    session = _get_session()
    try:
        job = session.get(AppliedJob, job_id)
        if not job:
            return False
        session.delete(job)
        session.commit()
        return True
    except Exception as e:
        session.rollback()
        logger.error(f"Failed to delete job: {e}")
        raise
    finally:
        session.close()


def find_job_by_company_and_title(company: str, title: str) -> Optional[AppliedJob]:
    """Find job by company and title."""
    session = _get_session()
    try:
        stmt = select(AppliedJob).where(
            AppliedJob.company.ilike(company.strip()),
            AppliedJob.job_title.ilike(title.strip())
        )
        return session.execute(stmt).scalar_one_or_none()
    finally:
        session.close()


# Helper function to create job from ORM model (for compatibility)
def create_job_from_orm(job: AppliedJob) -> AppliedJob:
    """Create a job from an ORM model instance."""
    session = _get_session()
    try:
        session.add(job)
        session.commit()
        session.refresh(job)
        return job
    except Exception as e:
        session.rollback()
        logger.error(f"Failed to create job: {e}")
        raise
    finally:
        session.close()


def update_job_orm(job: AppliedJob) -> AppliedJob:
    """Update a job ORM model instance."""
    session = _get_session()
    try:
        session.merge(job)
        session.commit()
        session.refresh(job)
        return job
    except Exception as e:
        session.rollback()
        logger.error(f"Failed to update job: {e}")
        raise
    finally:
        session.close()
