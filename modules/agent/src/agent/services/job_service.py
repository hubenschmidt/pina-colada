"""Service layer for job business logic."""

import logging
from typing import Dict, List, Optional
from agent.models.Job import JobCreateData, orm_to_dict
from agent.repositories.job_repository import (
    find_all_jobs,
    create_job,
    find_job_by_id,
    update_job,
    delete_job
)

logger = logging.getLogger(__name__)

# Module-level cache
_cache: Optional[set[str]] = None
_details_cache: Optional[List[Dict[str, str]]] = None


def _normalize_identifier(company: str, title: str) -> str:
    """Normalize job identifier for comparison."""
    return f"{company.lower().strip()}|{title.lower().strip()}"


def _map_to_dict(job) -> Dict[str, str]:
    """Map database model to dictionary."""
    if hasattr(job, 'id'):
        # It's an ORM object, use orm_to_dict
        job_dict = orm_to_dict(job)
        return {
            "company": job_dict.get("company") or "Unknown Company",
            "title": job_dict.get("job_title") or "",
            "date_applied": str(job_dict.get("date")) if job_dict.get("date") else "Not specified",
            "link": job_dict.get("job_url") or "",
            "status": job_dict.get("status") or "applied",
            "salary_range": job_dict.get("salary_range") or "",
            "notes": job_dict.get("notes") or "",
            "source": job_dict.get("source") or "manual",
            "id": job_dict.get("id") or "",
        }
    else:
        # It's already a dict
        return {
        "company": job.get("company") or "Unknown Company",
        "title": job.get("job_title") or job.get("title") or "",
            "date_applied": str(job.get("date")) if job.get("date") else (str(job.get("application_date")) if job.get("application_date") else "Not specified"),
        "link": job.get("job_url") or job.get("link") or "",
        "status": job.get("status") or "applied",
        "salary_range": job.get("salary_range") or "",
        "notes": job.get("notes") or "",
        "source": job.get("source") or "manual",
        "id": str(job.get("id")) if job.get("id") else "",
    }


def _clear_cache() -> None:
    """Clear module-level cache."""
    global _cache, _details_cache
    _cache = None
    _details_cache = None


def get_all_jobs(refresh: bool = False) -> List[Dict[str, str]]:
    """Get all jobs as dictionaries."""
    global _details_cache, _cache
    
    if _details_cache and not refresh:
        return _details_cache.copy()
    
    jobs = find_all_jobs()
    # Map to internal format with company/date_applied keys for consistency
    details = [_map_to_dict(job) for job in jobs]
    
    _details_cache = details
    _cache = {_normalize_identifier(j["company"], j["title"]) for j in details if j.get("title")}
    
    logger.info(f"Loaded {len(details)} jobs from repository")
    return details


def get_applied_jobs_only(refresh: bool = False) -> List[Dict[str, str]]:
    """Get only jobs with status 'applied'."""
    all_jobs = get_all_jobs(refresh=refresh)
    applied_jobs = [job for job in all_jobs if job.get("status", "").lower() == "applied"]
    logger.info(f"Filtered to {len(applied_jobs)} jobs with status 'applied' (from {len(all_jobs)} total)")
    return applied_jobs


def get_applied_identifiers(refresh: bool = False) -> set[str]:
    """Get set of normalized job identifiers (only status='applied')."""
    applied_jobs = get_applied_jobs_only(refresh=refresh)
    return {_normalize_identifier(j["company"], j["title"]) for j in applied_jobs if j.get("title")}


def fetch_applied_jobs(refresh: bool = False) -> set[str]:
    """Fetch set of normalized job identifiers (for compatibility)."""
    return get_applied_identifiers(refresh=refresh)


def get_jobs_details(refresh: bool = False) -> List[Dict[str, str]]:
    """Get detailed list of applied jobs (status='applied' only, for compatibility)."""
    return get_applied_jobs_only(refresh=refresh)


def is_job_applied(company: str, title: str) -> bool:
    """Check if a job has been applied to (status='applied' only)."""
    applied_jobs = get_applied_jobs_only()
    identifier = _normalize_identifier(company, title)
    applied_identifiers = {_normalize_identifier(j["company"], j["title"]) for j in applied_jobs if j.get("title")}
    return identifier in applied_identifiers


def filter_jobs(jobs: List[Dict[str, str]]) -> List[Dict[str, str]]:
    """Filter out jobs that have already been applied to (status='applied' only)."""
    identifiers = get_applied_identifiers()
    
    if not identifiers:
        logger.warning("No applied jobs (status='applied') found - returning all jobs")
        return jobs
    
    filtered = [
        job for job in jobs
        if not is_job_applied(job.get("company", ""), job.get("title", ""))
    ]
    
    filtered_count = len(jobs) - len(filtered)
    if filtered_count > 0:
        logger.info(f"Filtered {filtered_count} jobs with status 'applied', {len(filtered)} remaining")
    
    return filtered


def add_job(
    company: str,
    job_title: str,
    job_url: str = "",
    location: str = "",  # Deprecated, kept for API compatibility but ignored
    salary_range: str = "",
    notes: str = "",
    status: str = "applied",
    source: str = "agent"
) -> Optional[Dict[str, str]]:
    """Add a new job application."""
    data: JobCreateData = {
        "company": company,
        "job_title": job_title,
        "job_url": job_url or None,
        "salary_range": salary_range or None,
        "notes": notes or None,
        "status": status,
        "source": source
    }
    
    created = create_job(data)
    logger.info(f"Added job application: {company} - {job_title}")
    
    _clear_cache()
    
    return orm_to_dict(created)


def add_applied_job(
    company: str,
    job_title: str,
    job_url: str = "",
    location: str = "",  # Deprecated, kept for API compatibility but ignored
    salary_range: str = "",
    notes: str = "",
    status: str = "applied",
    source: str = "agent"
) -> Optional[Dict[str, str]]:
    """Add a new job application (for compatibility)."""
    return add_job(
        company=company,
        job_title=job_title,
        job_url=job_url,
        salary_range=salary_range,
        notes=notes,
        status=status,
        source=source
    )
