"""Controller for jobs business logic."""

import logging
from typing import List, Optional, Dict, Any
from datetime import datetime
from fastapi import HTTPException
from repositories.job_repository import (
    find_all_jobs,
    count_jobs,
    find_job_by_id,
    create_job as create_job_repo,
    update_job as update_job_repo,
    delete_job as delete_job_repo,
    find_all_statuses,
    find_jobs_with_status
)
from lib.serialization import model_to_dict

logger = logging.getLogger(__name__)


def _parse_paging(
    page: int,
    limit: int,
    order_by: str,
    order: str
) -> dict:
    """Parse pagination parameters."""
    offset = (page - 1) * limit
    return {
        "page": page,
        "limit": limit,
        "offset": offset,
        "order_by": order_by,
        "order": order.upper(),
    }


def _to_paged_response(count: int, page: int, limit: int, items: List) -> dict:
    """Convert to paged response format."""
    return {
        "items": items,
        "currentPage": page,
        "totalPages": max(1, (count + limit - 1) // limit),
        "total": count,
        "pageSize": limit,
    }


def _job_to_response_dict(job) -> Dict[str, Any]:
    """Convert job ORM to response dictionary."""
    job_dict = model_to_dict(job, include_relationships=True)

    # Extract company name from organization
    company = job_dict.get("organization", {}).get("name", "")

    # Extract status name from current_status
    status = job_dict.get("current_status", {}).get("name", "applied").lower()

    # Use Lead created_at for date (application date)
    created_at = job_dict.get("created_at", "")
    date_str = created_at.isoformat() if isinstance(created_at, datetime) else str(created_at)

    resume_date = job_dict.get("resume_date")
    resume_str = resume_date.isoformat() if resume_date else None

    return {
        "id": str(job_dict.get("id", "")),
        "company": company,
        "job_title": job_dict.get("job_title", ""),
        "date": date_str[:10] if date_str else "",  # YYYY-MM-DD format
        "status": status,
        "job_url": job_dict.get("job_url"),
        "salary_range": job_dict.get("salary_range"),
        "notes": job_dict.get("notes"),
        "resume": resume_str,
        "source": job_dict.get("source", "manual"),
        "created_at": date_str,
        "updated_at": job_dict.get("updated_at", "")
    }


def _parse_date_string(date_str: Optional[str]) -> Optional[datetime]:
    """Parse ISO date string to datetime."""
    if not date_str:
        return None
    try:
        return datetime.fromisoformat(date_str.replace('Z', '+00:00'))
    except (ValueError, TypeError):
        return None


def get_jobs(
    page: int,
    limit: int,
    order_by: str,
    order: str,
    search: Optional[str] = None
) -> dict:
    """Get all jobs with pagination."""
    paging = _parse_paging(page, limit, order_by, order)

    # Get all jobs
    all_jobs = find_all_jobs()

    # Apply search filter if provided
    if search and search.strip():
        search_lower = search.strip().lower()
        all_jobs = [
            job for job in all_jobs
            if (job.organization and search_lower in job.organization.name.lower()) or
               (job.job_title and search_lower in job.job_title.lower())
        ]

    total_count = len(all_jobs)

    # Simple sorting (in-memory for now - can be optimized with DB queries later)
    reverse = paging["order"] == "DESC"
    order_by_field = paging["order_by"]

    sort_functions = {
        "application_date": lambda j: j.lead.created_at if j.lead else "",
        "date": lambda j: j.lead.created_at if j.lead else "",
        "company": lambda j: (j.organization.name if j.organization else "").lower(),
        "status": lambda j: j.lead.current_status.name if (j.lead and j.lead.current_status) else "",
        "job_title": lambda j: (j.job_title or "").lower(),
        "resume": lambda j: j.resume_date or "",
    }
    
    sort_fn = sort_functions.get(order_by_field)
    if sort_fn:
        all_jobs.sort(key=sort_fn, reverse=reverse)

    # Paginate
    start = paging["offset"]
    end = start + paging["limit"]
    paginated_jobs = all_jobs[start:end]

    items = [_job_to_response_dict(job) for job in paginated_jobs]

    return _to_paged_response(total_count, page, limit, items)


def create_job(job_data: Dict[str, Any]) -> Dict[str, Any]:
    """Create a new job."""
    from repositories.organization_repository import get_or_create_organization

    resume_obj = _parse_date_string(job_data.get("resume"))

    # Get or create organization from company/organization_name
    organization_name = job_data.get("company") or job_data.get("organization_name")
    if not organization_name:
        raise HTTPException(status_code=400, detail="company or organization_name is required")

    org = get_or_create_organization(organization_name)

    data: Dict[str, Any] = {
        "organization_id": org.id,
        "job_title": job_data.get("job_title", ""),
        "job_url": job_data.get("job_url"),
        "salary_range": job_data.get("salary_range"),
        "notes": job_data.get("notes"),
        "resume_date": resume_obj,
        "status": job_data.get("status", "applied"),  # Will be converted to status_id
        "source": job_data.get("source", "manual")
    }

    created = create_job_repo(data)
    return _job_to_response_dict(created)


def get_job(job_id: str) -> Dict[str, Any]:
    """Get a job by ID."""
    try:
        job_id_int = int(job_id)
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid job ID format")

    job = find_job_by_id(job_id_int)

    if not job:
        raise HTTPException(status_code=404, detail="Job not found")

    return _job_to_response_dict(job)


def update_job(job_id: str, job_data: Dict[str, Any]) -> Dict[str, Any]:
    """Update a job."""
    from repositories.organization_repository import get_or_create_organization

    try:
        job_id_int = int(job_id)
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid job ID format")

    resume_obj = _parse_date_string(job_data.get("resume"))

    data: Dict[str, Any] = {}

    # Handle organization update
    organization_name = job_data.get("company") or job_data.get("organization_name")
    if organization_name:
        org = get_or_create_organization(organization_name)
        data["organization_id"] = org.id

    # Update simple fields
    allowed_fields = ["job_title", "job_url", "salary_range", "notes", "status", "source", "current_status_id"]
    data.update({k: job_data[k] for k in allowed_fields if k in job_data})
    
    # Handle resume_date separately (can be None or explicitly set)
    if resume_obj is not None or "resume" in job_data:
        data["resume_date"] = resume_obj

    updated = update_job_repo(job_id_int, data)

    if not updated:
        raise HTTPException(status_code=404, detail="Job not found")

    return _job_to_response_dict(updated)


def delete_job(job_id: str) -> dict:
    """Delete a job."""
    try:
        job_id_int = int(job_id)
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid job ID format")

    deleted = delete_job_repo(job_id_int)

    if not deleted:
        raise HTTPException(status_code=404, detail="Job not found")

    return {"success": True}


def get_statuses() -> List[Dict[str, Any]]:
    """Get all job statuses."""
    statuses = find_all_statuses()
    return [model_to_dict(s, include_relationships=False) for s in statuses]


def get_leads(statuses: Optional[str] = None) -> List[Dict[str, Any]]:
    """Get all job leads, optionally filtered by status names."""
    status_names = None
    if statuses:
        status_names = [s.strip() for s in statuses.split(",")]

    jobs = find_jobs_with_status(status_names)

    items = []
    for job in jobs:
        job_dict = _job_to_response_dict(job)
        if job.lead and job.lead.current_status:
            job_dict["lead_status"] = model_to_dict(job.lead.current_status, include_relationships=False)
        items.append(job_dict)

    return items


def mark_lead_as_applied(job_id: str) -> Dict[str, Any]:
    """Mark a lead as applied."""
    try:
        job_id_int = int(job_id)
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid job ID format")

    job = find_job_by_id(job_id_int)
    if not job:
        raise HTTPException(status_code=404, detail="Job not found")

    update_data: Dict[str, Any] = {
        "status": "applied",
    }

    updated = update_job_repo(job_id_int, update_data)
    if not updated:
        raise HTTPException(status_code=404, detail="Job not found")

    return _job_to_response_dict(updated)


def mark_lead_as_do_not_apply(job_id: str) -> Dict[str, Any]:
    """Mark a lead as do not apply."""
    try:
        job_id_int = int(job_id)
    except ValueError:
        raise HTTPException(status_code=400, detail="Invalid job ID format")

    update_data: Dict[str, Any] = {
        "status": "do_not_apply",
    }

    updated = update_job_repo(job_id_int, update_data)
    if not updated:
        raise HTTPException(status_code=404, detail="Job not found")

    return _job_to_response_dict(updated)


def get_recent_resume_date() -> Dict[str, Optional[str]]:
    """Get the most recent resume date."""
    from repositories.job_repository import find_all_jobs

    all_jobs = find_all_jobs()

    # Filter jobs with resume dates and sort by Lead created_at DESC
    jobs_with_resume = [job for job in all_jobs if job.resume_date]
    if not jobs_with_resume:
        return {"resume_date": None}

    # Sort by Lead created_at (application date) DESC to get most recent application
    jobs_with_resume.sort(
        key=lambda j: j.lead.created_at if j.lead else datetime.min,
        reverse=True
    )

    # Get the resume date from the most recently applied job
    most_recent_job = jobs_with_resume[0]
    if most_recent_job.resume_date:
        # Convert to YYYY-MM-DD format
        resume_date_str = most_recent_job.resume_date.isoformat().split('T')[0]
        return {"resume_date": resume_date_str}

    return {"resume_date": None}
