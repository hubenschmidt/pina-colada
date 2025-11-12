"""Repository layer for job data access."""

import logging
import os
from typing import List, Optional, Dict, Any
from sqlalchemy import create_engine, select, func as sql_func
from sqlalchemy.orm import Session, sessionmaker
from models.Job import Job, Base, JobCreateData, JobUpdateData, orm_to_dict, dict_to_orm, update_orm_from_dict
from models.LeadStatus import LeadStatus, LeadStatusData, orm_to_dict as lead_status_to_dict

logger = logging.getLogger(__name__)

_engine = None
_session_factory = None
_supabase_client = None

USE_SUPABASE = os.getenv("USE_SUPABASE", "false").lower() == "true"
USE_LOCAL_POSTGRES = os.getenv("USE_LOCAL_POSTGRES", "false").lower() == "true"


def _should_use_supabase() -> bool:
    """Determine if Supabase should be used based on environment."""
    if USE_LOCAL_POSTGRES:
        return False
    if USE_SUPABASE:
        return True
    # Default: use Supabase if SUPABASE_URL is set, otherwise use local Postgres
    return bool(os.getenv("SUPABASE_URL"))


def _get_supabase_client():
    """Get or create Supabase client."""
    global _supabase_client
    
    if _supabase_client is not None:
        return _supabase_client
    
    if not _should_use_supabase():
        return None
    
    try:
        from supabase import create_client, Client
    except ImportError:
        logger.warning("supabase-py not installed. Run: pip install supabase")
        return None
    
    supabase_url = os.getenv("SUPABASE_URL")
    supabase_key = os.getenv("SUPABASE_SERVICE_KEY")
    
    if not supabase_url or not supabase_key:
        logger.warning("SUPABASE_URL or SUPABASE_SERVICE_KEY not set")
        return None
    
    _supabase_client = create_client(supabase_url, supabase_key)
    logger.info("✓ Supabase client initialized")
    return _supabase_client


def _get_connection_string() -> str:
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


def _supabase_to_job(record: Dict[str, Any]) -> Job:
    """Convert Supabase record to Job ORM model."""
    from datetime import datetime
    import uuid
    
    job = Job()
    job.id = uuid.UUID(record["id"]) if record.get("id") else uuid.uuid4()
    job.company = record.get("company", "")
    job.job_title = record.get("job_title", "")
    job.date = datetime.fromisoformat(record["date"].replace("Z", "+00:00")) if record.get("date") else None
    job.status = record.get("status", "applied")
    job.job_url = record.get("job_url")
    job.notes = record.get("notes")
    job.resume = datetime.fromisoformat(record["resume"].replace("Z", "+00:00")) if record.get("resume") else None
    job.salary_range = record.get("salary_range")
    job.source = record.get("source", "manual")
    job.lead_status_id = uuid.UUID(record["lead_status_id"]) if record.get("lead_status_id") else None
    job.created_at = datetime.fromisoformat(record["created_at"].replace("Z", "+00:00")) if record.get("created_at") else None
    job.updated_at = datetime.fromisoformat(record["updated_at"].replace("Z", "+00:00")) if record.get("updated_at") else None
    return job


def _job_to_supabase_dict(data: JobCreateData) -> Dict[str, Any]:
    """Convert JobCreateData to Supabase-compatible dict."""
    result = {
        "company": data.get("company", ""),
        "job_title": data.get("job_title", ""),
        "status": data.get("status", "applied"),
        "source": data.get("source", "manual"),
    }
    
    if data.get("date"):
        result["date"] = data["date"].isoformat()
    if data.get("job_url"):
        result["job_url"] = data["job_url"]
    if data.get("notes"):
        result["notes"] = data["notes"]
    if data.get("resume"):
        result["resume"] = data["resume"].isoformat()
    if data.get("salary_range"):
        result["salary_range"] = data["salary_range"]
    if data.get("lead_status_id"):
        result["lead_status_id"] = data["lead_status_id"]
    
    return result


def find_all_jobs() -> List[Job]:
    """Find all jobs."""
    client = _get_supabase_client()
    if client:
        try:
            response = client.table("Job").select("*").execute()
            return [_supabase_to_job(record) for record in response.data]
        except Exception as e:
            logger.error(f"Failed to fetch jobs from Supabase: {e}")
            raise
    
    session = _get_session()
    try:
        stmt = select(Job)
        return list(session.execute(stmt).scalars().all())
    finally:
        session.close()


def count_jobs() -> int:
    """Count total jobs."""
    client = _get_supabase_client()
    if client:
        try:
            response = client.table("Job").select("*", count="exact", head=True).execute()
            return response.count or 0
        except Exception as e:
            logger.error(f"Failed to count jobs from Supabase: {e}")
            raise
    
    session = _get_session()
    try:
        stmt = select(sql_func.count(Job.id))
        return session.execute(stmt).scalar() or 0
    finally:
        session.close()


def find_job_by_id(job_id: str) -> Optional[Job]:
    """Find job by ID."""
    client = _get_supabase_client()
    if client:
        try:
            response = client.table("Job").select("*").eq("id", job_id).execute()
            if response.data:
                return _supabase_to_job(response.data[0])
            return None
        except Exception as e:
            logger.error(f"Failed to find job from Supabase: {e}")
            raise
    
    session = _get_session()
    try:
        return session.get(Job, job_id)
    finally:
        session.close()


def create_job(data: JobCreateData) -> Job:
    """Create a new job."""
    client = _get_supabase_client()
    if client:
        try:
            supabase_data = _job_to_supabase_dict(data)
            response = client.table("Job").insert(supabase_data).execute()
            if response.data:
                return _supabase_to_job(response.data[0])
            raise Exception("No data returned from Supabase insert")
        except Exception as e:
            logger.error(f"Failed to create job in Supabase: {e}")
            raise
    
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


def update_job(job_id: str, data: JobUpdateData) -> Optional[Job]:
    """Update an existing job."""
    client = _get_supabase_client()
    if client:
        try:
            update_dict = {}
            if "company" in data and data["company"] is not None:
                update_dict["company"] = data["company"]
            if "job_title" in data and data["job_title"] is not None:
                update_dict["job_title"] = data["job_title"]
            if "date" in data:
                update_dict["date"] = data["date"].isoformat() if data["date"] else None
            if "job_url" in data:
                update_dict["job_url"] = data["job_url"]
            if "notes" in data:
                update_dict["notes"] = data["notes"]
            if "resume" in data:
                update_dict["resume"] = data["resume"].isoformat() if data["resume"] else None
            if "salary_range" in data:
                update_dict["salary_range"] = data["salary_range"]
            if "status" in data and data["status"] is not None:
                update_dict["status"] = data["status"]
            if "source" in data and data["source"] is not None:
                update_dict["source"] = data["source"]
            if "lead_status_id" in data:
                update_dict["lead_status_id"] = data["lead_status_id"]
            
            response = client.table("Job").update(update_dict).eq("id", job_id).execute()
            if response.data:
                return _supabase_to_job(response.data[0])
            return None
        except Exception as e:
            logger.error(f"Failed to update job in Supabase: {e}")
            raise
    
    session = _get_session()
    try:
        job = session.get(Job, job_id)
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
    client = _get_supabase_client()
    if client:
        try:
            response = client.table("Job").delete().eq("id", job_id).execute()
            return len(response.data) > 0
        except Exception as e:
            logger.error(f"Failed to delete job from Supabase: {e}")
            raise
    
    session = _get_session()
    try:
        job = session.get(Job, job_id)
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


def find_job_by_company_and_title(company: str, title: str) -> Optional[Job]:
    """Find job by company and title."""
    client = _get_supabase_client()
    if client:
        try:
            response = client.table("Job").select("*").ilike("company", f"%{company.strip()}%").ilike("job_title", f"%{title.strip()}%").execute()
            if response.data:
                return _supabase_to_job(response.data[0])
            return None
        except Exception as e:
            logger.error(f"Failed to find job from Supabase: {e}")
            raise
    
    session = _get_session()
    try:
        stmt = select(Job).where(
            Job.company.ilike(company.strip()),
            Job.job_title.ilike(title.strip())
        )
        return session.execute(stmt).scalar_one_or_none()
    finally:
        session.close()


# Helper function to create job from ORM model (for compatibility)
def create_job_from_orm(job: Job) -> Job:
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


def update_job_orm(job: Job) -> Job:
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


def find_all_lead_statuses() -> List[LeadStatus]:
    """Find all lead statuses."""
    client = _get_supabase_client()
    if client:
        try:
            response = client.table("LeadStatus").select("*").order("name").execute()
            import uuid
            from datetime import datetime
            result = []
            for record in response.data:
                ls = LeadStatus()
                ls.id = uuid.UUID(record["id"]) if record.get("id") else uuid.uuid4()
                ls.name = record.get("name", "")
                ls.description = record.get("description")
                ls.created_at = datetime.fromisoformat(record["created_at"].replace("Z", "+00:00")) if record.get("created_at") else None
                result.append(ls)
            return result
        except Exception as e:
            logger.error(f"Failed to fetch lead statuses from Supabase: {e}")
            raise
    
    session = _get_session()
    try:
        stmt = select(LeadStatus).order_by(LeadStatus.name)
        return list(session.execute(stmt).scalars().all())
    finally:
        session.close()


def find_jobs_with_lead_status(status_names: Optional[List[str]] = None) -> List[Job]:
    """Find jobs with lead_status_id set, optionally filtered by status names."""
    client = _get_supabase_client()
    if client:
        try:
            # First, get lead status IDs if filtering by names
            lead_status_ids = None
            if status_names:
                lead_statuses_response = client.table("LeadStatus").select("id").in_("name", status_names).execute()
                lead_status_ids = [ls["id"] for ls in lead_statuses_response.data]
                if not lead_status_ids:
                    return []  # No matching statuses, return empty
            
            query = client.table("Job").select("*, lead_status:LeadStatus(*)").not_("lead_status_id", "is", "null")
            
            if lead_status_ids:
                query = query.in_("lead_status_id", lead_status_ids)
            
            response = query.order("created_at", desc=True).execute()
            jobs = []
            for record in response.data:
                job = _supabase_to_job(record)
                # Load lead_status relationship if present
                if record.get("lead_status"):
                    from models.LeadStatus import LeadStatus
                    import uuid
                    from datetime import datetime
                    ls = LeadStatus()
                    ls_data = record["lead_status"]
                    ls.id = uuid.UUID(ls_data["id"]) if ls_data.get("id") else uuid.uuid4()
                    ls.name = ls_data.get("name", "")
                    ls.description = ls_data.get("description")
                    ls.created_at = datetime.fromisoformat(ls_data["created_at"].replace("Z", "+00:00")) if ls_data.get("created_at") else None
                    job.lead_status = ls
                jobs.append(job)
            return jobs
        except Exception as e:
            logger.error(f"Failed to fetch leads from Supabase: {e}")
            raise
    
    session = _get_session()
    try:
        from sqlalchemy.orm import joinedload
        stmt = select(Job).options(joinedload(Job.lead_status)).where(Job.lead_status_id.isnot(None))
        
        if status_names:
            stmt = stmt.join(LeadStatus).where(LeadStatus.name.in_(status_names))
        
        stmt = stmt.order_by(Job.created_at.desc())
        return list(session.execute(stmt).unique().scalars().all())
    finally:
        session.close()
