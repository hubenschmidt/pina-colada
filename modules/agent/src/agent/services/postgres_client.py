"""Postgres service using SQLAlchemy for local development."""

import logging
import os
from typing import Dict, List, Optional, Set
from sqlalchemy import create_engine, select
from sqlalchemy.orm import Session, sessionmaker
from agent.models import Base
from agent.models.Job import Job
from agent.models.LeadStatus import LeadStatus  # Import to register with Base.metadata

logger = logging.getLogger(__name__)

# Module-level state
_engine = None
_session_factory = None
_applied_jobs_cache: Set[str] = set()
_jobs_details_cache: List[Dict[str, str]] = []


def _get_connection_string() -> Optional[str]:
    """Get Postgres connection string from environment."""
    db_host = os.getenv("POSTGRES_HOST", "postgres")
    db_port = os.getenv("POSTGRES_PORT", "5432")
    db_user = os.getenv("POSTGRES_USER", "postgres")
    db_password = os.getenv("POSTGRES_PASSWORD", "postgres")
    db_name = os.getenv("POSTGRES_DB", "development")
    
    return f"postgresql://{db_user}:{db_password}@{db_host}:{db_port}/{db_name}"


def _get_engine():
    """Get or create SQLAlchemy engine."""
    global _engine
    if _engine is not None:
        return _engine
    
    conn_string = _get_connection_string()
    if not conn_string:
        logger.warning("Postgres connection string not available")
        return None
    
    _engine = create_engine(conn_string, pool_pre_ping=True)
    Base.metadata.create_all(_engine)
    return _engine


def _get_session() -> Optional[Session]:
    """Get database session."""
    engine = _get_engine()
    if not engine:
        return None
    
    global _session_factory
    if _session_factory is None:
        _session_factory = sessionmaker(bind=engine)
    
    return _session_factory()


def _normalize_job_identifier(company: str, title: str) -> str:
    """Normalize job identifier for comparison."""
    return f"{company.lower().strip()}|{title.lower().strip()}"


def _map_db_record_to_job(record: Job) -> Dict[str, str]:
    """Map database record to job dictionary."""
    return {
        "company": record.company or "Unknown Company",
        "title": record.job_title or "",
        "date_applied": str(record.date) if record.date else "Not specified",
        "link": record.job_url or "",
        "status": record.status or "applied",
        "salary_range": record.salary_range or "",
        "notes": record.notes or "",
        "source": record.source or "manual",
        "id": str(record.id) if record.id else "",
    }


def _clear_cache() -> None:
    """Clear module-level cache."""
    global _applied_jobs_cache, _jobs_details_cache
    _applied_jobs_cache = set()
    _jobs_details_cache = []


def fetch_applied_jobs(refresh: bool = False) -> Set[str]:
    """Fetch list of companies/job titles from applied jobs table."""
    global _applied_jobs_cache, _jobs_details_cache
    
    if _applied_jobs_cache and not refresh:
        return _applied_jobs_cache
    
    session = _get_session()
    if not session:
        logger.warning("Postgres session unavailable - returning empty set")
        return set()
    
    try:
        # Filter by status='applied' only
        stmt = select(Job).where(Job.status == "applied")
        records = session.execute(stmt).scalars().all()
        
        logger.info(f"✓ Found {len(records)} rows in Job table with status='applied'")
        
        applied_jobs = set()
        jobs_details = []
        
        def _process_record(record) -> Optional[Dict[str, str]]:
            job_data = _map_db_record_to_job(record)
            if not job_data["title"]:
                return None
            # Double-check status (should already be filtered by query, but be safe)
            if job_data.get("status", "").lower() != "applied":
                return None
            return job_data
        
        def _add_job(job_data: Dict[str, str]) -> None:
            identifier = _normalize_job_identifier(job_data["company"], job_data["title"])
            applied_jobs.add(identifier)
            jobs_details.append(job_data)
        
        for record in records:
            job_data = _process_record(record)
            if job_data:
                _add_job(job_data)
        
        logger.info(f"✓ Loaded {len(applied_jobs)} applied jobs (status='applied') from Postgres")
        _applied_jobs_cache = applied_jobs
        _jobs_details_cache = jobs_details
        return applied_jobs
    
    except Exception as e:
        logger.error(f"Failed to fetch applied jobs from Postgres: {e}")
        return _applied_jobs_cache
    finally:
        session.close()


def get_jobs_details(refresh: bool = False) -> List[Dict[str, str]]:
    """Get detailed list of applied jobs."""
    global _jobs_details_cache
    
    if not _jobs_details_cache or refresh:
        fetch_applied_jobs(refresh=refresh)
    return _jobs_details_cache.copy()


def is_job_applied(company: str, title: str) -> bool:
    """Check if a job has already been applied to."""
    if not _applied_jobs_cache:
        fetch_applied_jobs()
    
    identifier = _normalize_job_identifier(company, title)
    return identifier in _applied_jobs_cache


def filter_jobs(jobs: List[Dict[str, str]]) -> List[Dict[str, str]]:
    """Filter out jobs that have already been applied to."""
    if not _applied_jobs_cache:
        fetch_applied_jobs()
    
    if not _applied_jobs_cache:
        logger.warning("No applied jobs loaded - returning all jobs")
        return jobs
    
    filtered = [
        job for job in jobs
        if not is_job_applied(job.get("company", ""), job.get("title", ""))
    ]
    
    filtered_count = len(jobs) - len(filtered)
    if filtered_count > 0:
        logger.info(f"Filtered {filtered_count} jobs, {len(filtered)} remaining")
    
    return filtered


def add_applied_job(
    company: str,
    job_title: str,
    job_url: str = "",
    location: str = "",  # Deprecated, kept for API compatibility but ignored
    salary_range: str = "",
    notes: str = "",
    status: str = "applied",
    source: str = "agent"
) -> Optional[Dict]:
    """Add a new job application to the database."""
    session = _get_session()
    if not session:
        logger.error("Cannot add job - Postgres session unavailable")
        return None
    
    try:
        job = Job(
            company=company,
            job_title=job_title,
            job_url=job_url,
            salary_range=salary_range,
            notes=notes,
            status=status,
            source=source
        )
        
        session.add(job)
        session.commit()
        
        logger.info(f"✓ Added job application: {company} - {job_title}")
        _clear_cache()
        
        return _map_db_record_to_job(job)
    
    except Exception as e:
        logger.error(f"Failed to add job application: {e}")
        session.rollback()
        return None
    finally:
        session.close()


def get_postgres_jobs_tracker():
    """Get Postgres jobs tracker (functional dict interface)."""
    return {
        "get_jobs_details": lambda refresh=False: get_jobs_details(refresh=refresh),
        "fetch_applied_jobs": lambda refresh=False: fetch_applied_jobs(refresh=refresh),
        "is_job_applied": lambda company, title: is_job_applied(company, title),
        "filter_jobs": lambda jobs: filter_jobs(jobs),
        "add_applied_job": lambda **kwargs: add_applied_job(**kwargs),
    }
