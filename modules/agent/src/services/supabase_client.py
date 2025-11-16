"""Supabase service for tracking applied jobs.

Uses Supabase database for job application tracking with better
data structure, querying capabilities, and real-time updates.
"""

import logging
import os
from typing import Dict, List, Optional, Set, Any

logger = logging.getLogger(__name__)

try:
    from supabase import create_client, Client

    SUPABASE_AVAILABLE = True
except ImportError:
    SUPABASE_AVAILABLE = False
    logger.warning("supabase-py not installed. Run: pip install supabase")


# Module-level cache
_client: Optional[Client] = None
_applied_jobs_cache: Set[str] = set()
_jobs_details_cache: List[Dict[str, str]] = []


def _normalize_job_identifier(company: str, title: str) -> str:
    """Normalize job identifier for comparison."""
    return f"{company.lower().strip()}|{title.lower().strip()}"


def _get_supabase_client() -> Optional[Client]:
    """Create Supabase client from environment variables."""
    global _client

    if _client is not None:
        return _client

    if not SUPABASE_AVAILABLE:
        return None

    supabase_url = os.getenv("SUPABASE_URL")
    supabase_key = os.getenv("SUPABASE_SERVICE_KEY")

    if not supabase_url or not supabase_key:
        logger.warning("SUPABASE_URL or SUPABASE_SERVICE_KEY not set in environment")
        return None

    _client = create_client(supabase_url, supabase_key)
    if _client:
        logger.info("✓ Supabase client initialized")
    return _client


def _map_db_record_to_job(record: Dict[str, Any]) -> Dict[str, str]:
    """Map database record to job dictionary."""
    return {
        "company": record.get("company", "Unknown Company"),
        "title": record.get("job_title", ""),
        "date_applied": record.get("date", "Not specified"),
        "link": record.get("job_url", ""),
        "status": record.get("status", "applied"),
        "salary_range": record.get("salary_range", ""),
        "notes": record.get("notes", ""),
        "source": record.get("source", "manual"),
        "id": record.get("id", ""),
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

    client = _get_supabase_client()
    if not client:
        logger.warning("Supabase client unavailable - returning empty set")
        return set()

    try:
        # Filter by status='applied' only
        response = client.table("Job").select("*").eq("status", "applied").execute()
        records = response.data

        logger.info(f"✓ Found {len(records)} rows in Job table with status='applied'")

        applied_jobs = set()
        jobs_details = []

        def _process_record(record: Dict[str, Any]) -> Optional[Dict[str, str]]:
            job_data = _map_db_record_to_job(record)
            if not job_data["title"]:
                return None
            # Double-check status (should already be filtered by query, but be safe)
            if job_data.get("status", "").lower() != "applied":
                return None
            return job_data

        def _add_job(job_data: Dict[str, str]) -> None:
            identifier = _normalize_job_identifier(
                job_data["company"], job_data["title"]
            )
            applied_jobs.add(identifier)
            jobs_details.append(job_data)

        for record in records:
            job_data = _process_record(record)
            if job_data:
                _add_job(job_data)

        logger.info(
            f"✓ Loaded {len(applied_jobs)} applied jobs (status='applied') from Supabase"
        )
        logger.info(f"✓ Total job details: {len(jobs_details)}")
        _applied_jobs_cache = applied_jobs
        _jobs_details_cache = jobs_details
        return applied_jobs

    except Exception as e:
        logger.error(f"Failed to fetch applied jobs from Supabase: {e}")
        return _applied_jobs_cache


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
        job
        for job in jobs
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
    source: str = "agent",
) -> Optional[Dict[str, Any]]:
    """Add a new job application to the database."""
    client = _get_supabase_client()
    if not client:
        logger.error("Cannot add job - Supabase client unavailable")
        return None

    try:
        data = {
            "company": company,
            "job_title": job_title,
            "job_url": job_url,
            "salary_range": salary_range,
            "notes": notes,
            "status": status,
            "source": source,
        }

        response = client.table("Job").insert(data).execute()

        if response.data:
            logger.info(f"✓ Added job application: {company} - {job_title}")
            # Invalidate cache to force refresh on next fetch
            _clear_cache()
            return response.data[0]

        return None

    except Exception as e:
        logger.error(f"Failed to add job application: {e}")
        return None


def get_applied_jobs_tracker():
    """Get applied jobs tracker (functional dict interface)."""
    return {
        "get_jobs_details": lambda refresh=False: get_jobs_details(refresh=refresh),
        "fetch_applied_jobs": lambda refresh=False: fetch_applied_jobs(refresh=refresh),
        "is_job_applied": lambda company, title: is_job_applied(company, title),
        "filter_jobs": lambda jobs: filter_jobs(jobs),
        "add_applied_job": lambda **kwargs: add_applied_job(**kwargs),
    }
