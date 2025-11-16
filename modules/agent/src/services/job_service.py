"""Service layer for job business logic."""

import logging
from typing import Dict, List, Optional, Any
from datetime import datetime
from fastapi import HTTPException
from lib.serialization import model_to_dict
from repositories.job_repository import (
    find_all_jobs,
    create_job as create_job_repo,
    find_job_by_id,
    update_job as update_job_repo,
    delete_job as delete_job_repo,
    find_all_statuses,
    find_jobs_with_status,
)
from repositories.organization_repository import get_or_create_organization
from repositories.job_repository import find_all_jobs as find_all_jobs_repo

logger = logging.getLogger(__name__)


def _normalize_identifier(company: str, title: str) -> str:
    """Normalize job identifier for comparison."""
    return f"{company.lower().strip()}|{title.lower().strip()}"


def _map_to_dict(job) -> Dict[str, str]:
    """Map database model to dictionary for API compatibility."""
    # Use model_to_dict which includes relationships
    job_dict = model_to_dict(job, include_relationships=True)

    # Extract organization name for company field (backward compatibility)
    company = job_dict.get("organization", {}).get("name", "Unknown Company")

    # Extract status name from current_status (backward compatibility)
    status = job_dict.get("current_status", {}).get("name", "Applied")

    created_at = job_dict.get("created_at")
    date_applied = str(created_at)[:10] if created_at else "Not specified"

    return {
        "company": company,
        "title": job_dict.get("job_title", ""),
        "date_applied": date_applied,
        "link": job_dict.get("job_url") or "",
        "status": status,
        "salary_range": job_dict.get("salary_range") or "",
        "notes": job_dict.get("notes") or "",
        "source": job_dict.get("source", "manual"),
        "id": str(job_dict.get("id", "")),
    }


async def get_all_jobs(refresh: bool = False) -> List[Dict[str, str]]:
    """Get all jobs as dictionaries."""
    jobs, _ = await find_all_jobs()
    # Map to internal format with company/date_applied keys for consistency
    details = [_map_to_dict(job) for job in jobs]
    return details


async def get_applied_jobs_only(refresh: bool = False) -> List[Dict[str, str]]:
    """Get only jobs with status 'Applied'."""
    all_jobs = await get_all_jobs(refresh=refresh)
    applied_jobs = [
        job for job in all_jobs if job.get("status", "") == "Applied"
    ]
    return applied_jobs


async def get_applied_identifiers(refresh: bool = False) -> set[str]:
    """Get set of normalized job identifiers (only status='applied')."""
    applied_jobs = await get_applied_jobs_only(refresh=refresh)
    return {
        _normalize_identifier(j["company"], j["title"])
        for j in applied_jobs
        if j.get("title")
    }


async def fetch_applied_jobs(refresh: bool = False) -> set[str]:
    """Fetch set of normalized job identifiers (for compatibility)."""
    return await get_applied_identifiers(refresh=refresh)


async def get_jobs_details(refresh: bool = False) -> List[Dict[str, str]]:
    """Get detailed list of applied jobs (status='applied' only, for compatibility)."""
    return await get_applied_jobs_only(refresh=refresh)


async def is_job_applied(company: str, title: str) -> bool:
    """Check if a job has been applied to (status='applied' only)."""
    applied_jobs = await get_applied_jobs_only()
    identifier = _normalize_identifier(company, title)
    applied_identifiers = {
        _normalize_identifier(j["company"], j["title"])
        for j in applied_jobs
        if j.get("title")
    }
    return identifier in applied_identifiers


async def filter_jobs(jobs: List[Dict[str, str]]) -> List[Dict[str, str]]:
    """Filter out jobs that have already been applied to (status='applied' only)."""
    identifiers = await get_applied_identifiers()

    if not identifiers:
        logger.warning("No applied jobs (status='applied') found - returning all jobs")
        return jobs

    filtered = []
    for job in jobs:
        is_applied = await is_job_applied(job.get("company", ""), job.get("title", ""))
        if not is_applied:
            filtered.append(job)

    return filtered


async def add_job(
    organization_name: str,
    job_title: str,
    job_url: str = "",
    salary_range: str = "",
    notes: str = "",
    status: str = "applied",
    source: str = "agent",
) -> Optional[Dict[str, str]]:
    """Add a new job application."""

    # Get or create organization
    org = await get_or_create_organization(organization_name)

    data: Dict[str, Any] = {
        "organization_id": org.id,
        "job_title": job_title,
        "job_url": job_url or None,
        "salary_range": salary_range or None,
        "notes": notes or None,
        "status": status,  # Will be converted to status_id in repository
        "source": source,
    }

    created = await create_job(data)
    return _map_to_dict(created)


async def add_applied_job(
    company: str,
    job_title: str,
    job_url: str = "",
    location: str = "",  # Deprecated, kept for backward compatibility but ignored
    salary_range: str = "",
    notes: str = "",
    status: str = "applied",
    source: str = "agent",
) -> Optional[Dict[str, str]]:
    """Add a new job application (backward compatibility wrapper)."""
    return await add_job(
        organization_name=company,
        job_title=job_title,
        job_url=job_url,
        salary_range=salary_range,
        notes=notes,
        status=status,
        source=source,
    )


def _fuzzy_match_company(search_company: str, db_company: str) -> bool:
    """Check if company names match using fuzzy matching."""
    if not search_company or not db_company:
        return False

    search_norm = search_company.lower().strip()
    db_norm = db_company.lower().strip()

    # Exact match
    if search_norm == db_norm:
        return True

    # Remove common suffixes for comparison
    search_clean = (
        search_norm.replace(" inc", "")
        .replace(" llc", "")
        .replace(" corp", "")
        .replace(" ltd", "")
        .strip()
    )
    db_clean = (
        db_norm.replace(" inc", "")
        .replace(" llc", "")
        .replace(" corp", "")
        .replace(" ltd", "")
        .strip()
    )

    if search_clean == db_clean:
        return True

    # One is substring of the other (require at least 4 chars to avoid false matches)
    if len(search_clean) < 4 or len(db_clean) < 4:
        return False

    return search_clean in db_clean or db_clean in search_clean


def _matches_job(job, company: str, job_title: str) -> bool:
    """Check if job matches company and title using fuzzy matching."""
    job_dict = model_to_dict(job, include_relationships=True)
    job_company = job_dict.get("organization", {}).get("name", "")

    if not _fuzzy_match_company(company, job_company):
        return False

    job_title_lower = job_title.lower().strip()
    db_title_lower = job_dict.get("job_title", "").lower()
    return job_title_lower in db_title_lower or db_title_lower in job_title_lower


async def get_jobs_paginated(
    page: int, limit: int, order_by: str, order: str, search: Optional[str] = None
) -> tuple[List[Any], int]:
    """Get all jobs with search, sorting, and pagination logic.

    Args:
        page: Page number
        limit: Items per page
        order_by: Field to sort by
        order: Sort direction (ASC/DESC)
        search: Optional search query

    Returns:
        Tuple of (paginated_jobs, total_count)
    """
    return await find_all_jobs_repo(
        page=page,
        page_size=limit,
        search=search,
        order_by=order_by,
        order=order
    )


async def create_job(job_data: Dict[str, Any]) -> Any:
    """Create a new job.

    Handles:
    - Organization lookup/creation
    - Date parsing
    - Field validation
    - Repository creation

    Args:
        job_data: Dictionary of job fields

    Returns:
        Created Job ORM object

    Raises:
        HTTPException: If required fields are missing
    """
    # Get or create organization
    organization_name = job_data.get("company") or job_data.get("organization_name")
    if not organization_name:
        raise HTTPException(
            status_code=400, detail="company or organization_name is required"
        )

    org = await get_or_create_organization(organization_name)

    # Parse resume date
    resume_str = job_data.get("resume")
    resume_obj = None
    if not resume_str:
        pass  # No resume date provided
    if resume_str:
        try:
            resume_obj = datetime.fromisoformat(resume_str.replace("Z", "+00:00"))
        except (ValueError, TypeError):
            pass

    data: Dict[str, Any] = {
        "organization_id": org.id,
        "job_title": job_data.get("job_title", ""),
        "job_url": job_data.get("job_url"),
        "salary_range": job_data.get("salary_range"),
        "notes": job_data.get("notes"),
        "resume_date": resume_obj,
        "status": job_data.get("status", "applied"),
        "source": job_data.get("source", "manual"),
    }

    created = await create_job_repo(data)
    return created


async def get_job(job_id: str) -> Any:
    """Get a job by ID.

    Args:
        job_id: Job ID (string from route)

    Returns:
        Job ORM object

    Raises:
        HTTPException: If job_id is invalid or job not found
    """
    try:
        job_id_int = int(job_id)
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid job ID format")

    job = await find_job_by_id(job_id_int)

    if not job:
        raise HTTPException(status_code=404, detail="Job not found")

    return job


async def delete_job(job_id: str) -> bool:
    """Delete a job by ID.

    Args:
        job_id: Job ID (string from route)

    Returns:
        True if deleted

    Raises:
        HTTPException: If job_id is invalid or job not found
    """
    try:
        job_id_int = int(job_id)
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid job ID format")

    deleted = await delete_job_repo(job_id_int)

    if not deleted:
        raise HTTPException(status_code=404, detail="Job not found")

    return deleted


async def get_statuses() -> List[Any]:
    """Get all job statuses.

    Returns:
        List of Status ORM objects
    """
    return await find_all_statuses()


async def get_jobs_with_status(status_names: Optional[List[str]] = None) -> List[Any]:
    """Get jobs filtered by status names.

    Args:
        status_names: Optional list of status names to filter by

    Returns:
        List of Job ORM objects
    """
    return await find_jobs_with_status(status_names)


async def get_recent_resume_date() -> Optional[str]:
    """Get the most recent resume date from all jobs.

    Returns:
        Resume date string (YYYY-MM-DD) or None
    """
    all_jobs, _ = await find_all_jobs_repo()

    # Filter jobs with resume dates and sort by Lead created_at DESC
    jobs_with_resume = [job for job in all_jobs if job.resume_date]
    if not jobs_with_resume:
        return None

    # Sort by Lead created_at (application date) DESC to get most recent application
    jobs_with_resume.sort(
        key=lambda j: j.lead.created_at if j.lead else datetime.min, reverse=True
    )

    # Get the resume date from the most recently applied job
    most_recent_job = jobs_with_resume[0]
    if most_recent_job.resume_date:
        return most_recent_job.resume_date.isoformat().split("T")[0]

    return None


async def update_job(job_id: str, job_data: Dict[str, Any]) -> Any:
    """Update a job with the provided data.

    Handles:
    - ID validation
    - Organization lookup/creation
    - Date parsing
    - Field validation
    - Repository update
    - Error handling

    Args:
        job_id: Job ID to update (string from route)
        job_data: Dictionary of fields to update

    Returns:
        Updated Job ORM object

    Raises:
        HTTPException: If job_id is invalid or job not found
    """
    # Validate job_id
    try:
        job_id_int = int(job_id)
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid job ID format")

    data: Dict[str, Any] = {}

    # Handle organization update
    organization_name = job_data.get("company") or job_data.get("organization_name")
    if organization_name:
        org = await get_or_create_organization(organization_name)
        data["organization_id"] = org.id

    # Update simple fields
    allowed_fields = [
        "job_title",
        "job_url",
        "salary_range",
        "notes",
        "status",
        "source",
        "current_status_id",
    ]
    data.update({k: job_data[k] for k in allowed_fields if k in job_data})

    # Handle resume_date parsing
    resume_str = job_data.get("resume")
    if resume_str is None:
        pass  # No update to resume_date
    if resume_str == "":
        data["resume_date"] = None
    if resume_str:
        try:
            data["resume_date"] = datetime.fromisoformat(resume_str.replace("Z", "+00:00"))
        except (ValueError, TypeError):
            data["resume_date"] = None

    # Update via repository
    updated = await update_job_repo(job_id_int, data)

    # Handle not found
    if not updated:
        raise HTTPException(status_code=404, detail="Job not found")

    return updated


async def update_job_by_company(
    company: str,
    job_title: str,
    status: Optional[str] = None,
    job_url: Optional[str] = None,
    notes: Optional[str] = None,
) -> Optional[Dict[str, str]]:
    """
    Find and update a job by company name and job title (fuzzy matching).
    Returns updated job or None if not found.
    """
    # Get all jobs and find matching one
    all_jobs, _ = await find_all_jobs()
    matching_job = next(
        (job for job in all_jobs if _matches_job(job, company, job_title)), None
    )

    if not matching_job:
        return None

    # Build update data
    update_data: Dict[str, Any] = {
        k: v
        for k, v in [
            ("status", status),
            ("job_url", job_url),
            ("notes", notes),
        ]
        if v is not None
    }

    # Update the job
    updated_job = await update_job(str(matching_job.id), update_data)

    if not updated_job:
        return None

    return _map_to_dict(updated_job)
