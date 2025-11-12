"""Controller for jobs business logic."""

import logging
from typing import List, Optional, Dict, Any
from datetime import datetime
from fastapi import HTTPException
from api.repositories.job_repository import (
    find_all_jobs,
    count_jobs,
    find_job_by_id,
    create_job as create_job_repo,
    update_job as update_job_repo,
    delete_job as delete_job_repo,
    find_all_lead_statuses,
    find_jobs_with_lead_status
)
from models.Job import JobCreateData, JobUpdateData, orm_to_dict
from models.LeadStatus import orm_to_dict as lead_status_to_dict

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
    return {
        "id": str(job.id),
        "company": job.company or "",
        "job_title": job.job_title or "",
        "date": job.date.isoformat() if job.date else "",
        "status": job.status or "applied",
        "job_url": job.job_url,
        "salary_range": job.salary_range,
        "notes": job.notes,
        "resume": job.resume.isoformat() if job.resume else None,
        "source": job.source or "manual",
        "created_at": job.created_at.isoformat() if job.created_at else "",
        "updated_at": job.updated_at.isoformat() if job.updated_at else ""
    }


def _parse_date_string(date_str: Optional[str]) -> Optional[datetime]:
    """Parse ISO date string to datetime."""
    if not date_str:
        return None
    try:
        return datetime.fromisoformat(date_str.replace('Z', '+00:00'))
    except:
        return None


def get_jobs(
    page: int,
    limit: int,
    order_by: str,
    order: str
) -> dict:
    """Get all jobs with pagination."""
    paging = _parse_paging(page, limit, order_by, order)
    
    total_count = count_jobs()
    
    # Get paginated jobs with sorting
    all_jobs = find_all_jobs()
    
    # Simple sorting (in-memory for now - can be optimized with DB queries later)
    reverse = paging["order"] == "DESC"
    order_by_field = paging["order_by"]
    if order_by_field == "application_date" or order_by_field == "date":
        all_jobs.sort(key=lambda j: j.date or "", reverse=reverse)
    elif order_by_field == "company":
        all_jobs.sort(key=lambda j: (j.company or "").lower(), reverse=reverse)
    elif order_by_field == "status":
        all_jobs.sort(key=lambda j: j.status or "", reverse=reverse)
    elif order_by_field == "job_title":
        all_jobs.sort(key=lambda j: (j.job_title or "").lower(), reverse=reverse)
    elif order_by_field == "resume":
        all_jobs.sort(key=lambda j: j.resume or "", reverse=reverse)
    
    # Paginate
    start = paging["offset"]
    end = start + paging["limit"]
    paginated_jobs = all_jobs[start:end]
    
    items = [_job_to_response_dict(job) for job in paginated_jobs]
    
    return _to_paged_response(total_count, page, limit, items)


def create_job(job_data: Dict[str, Any]) -> Dict[str, Any]:
    """Create a new job."""
    date_obj = _parse_date_string(job_data.get("date"))
    resume_obj = _parse_date_string(job_data.get("resume"))
    
    data: JobCreateData = {
        "company": job_data["company"],
        "job_title": job_data["job_title"],
        "date": date_obj,
        "job_url": job_data.get("job_url"),
        "salary_range": job_data.get("salary_range"),
        "notes": job_data.get("notes"),
        "resume": resume_obj,
        "status": job_data.get("status", "applied"),
        "source": job_data.get("source", "manual")
    }
    
    created = create_job_repo(data)
    return _job_to_response_dict(created)


def get_job(job_id: str) -> Dict[str, Any]:
    """Get a job by ID."""
    job = find_job_by_id(job_id)
    
    if not job:
        raise HTTPException(status_code=404, detail="Job not found")
    
    return _job_to_response_dict(job)


def update_job(job_id: str, job_data: Dict[str, Any]) -> Dict[str, Any]:
    """Update a job."""
    date_obj = _parse_date_string(job_data.get("date"))
    resume_obj = _parse_date_string(job_data.get("resume"))
    
    data: JobUpdateData = {}
    
    if "company" in job_data:
        data["company"] = job_data["company"]
    if "job_title" in job_data:
        data["job_title"] = job_data["job_title"]
    if date_obj is not None or "date" in job_data:
        data["date"] = date_obj
    if "job_url" in job_data:
        data["job_url"] = job_data["job_url"]
    if "salary_range" in job_data:
        data["salary_range"] = job_data["salary_range"]
    if "notes" in job_data:
        data["notes"] = job_data["notes"]
    if resume_obj is not None or "resume" in job_data:
        data["resume"] = resume_obj
    if "status" in job_data:
        data["status"] = job_data["status"]
    if "source" in job_data:
        data["source"] = job_data["source"]
    if "lead_status_id" in job_data:
        data["lead_status_id"] = job_data["lead_status_id"]
    
    # Remove None values
    data = {k: v for k, v in data.items() if v is not None}
    
    updated = update_job_repo(job_id, data)
    
    if not updated:
        raise HTTPException(status_code=404, detail="Job not found")
    
    return _job_to_response_dict(updated)


def delete_job(job_id: str) -> dict:
    """Delete a job."""
    deleted = delete_job_repo(job_id)
    
    if not deleted:
        raise HTTPException(status_code=404, detail="Job not found")
    
    return {"success": True}


def get_lead_statuses() -> List[Dict[str, Any]]:
    """Get all lead statuses."""
    lead_statuses = find_all_lead_statuses()
    return [lead_status_to_dict(ls) for ls in lead_statuses]


def get_leads(statuses: Optional[str] = None) -> List[Dict[str, Any]]:
    """Get all job leads, optionally filtered by status names."""
    status_names = None
    if statuses:
        status_names = [s.strip() for s in statuses.split(",")]
    
    jobs = find_jobs_with_lead_status(status_names)
    
    items = []
    for job in jobs:
        job_dict = orm_to_dict(job)
        if job.lead_status:
            job_dict["lead_status"] = lead_status_to_dict(job.lead_status)
        items.append(job_dict)
    
    return items


def mark_lead_as_applied(job_id: str) -> Dict[str, Any]:
    """Mark a lead as applied."""
    job = find_job_by_id(job_id)
    if not job:
        raise HTTPException(status_code=404, detail="Job not found")
    
    update_data: JobUpdateData = {
        "status": "applied",
        "lead_status_id": None,
    }
    
    if not job.date:
        update_data["date"] = datetime.now()
    
    updated = update_job_repo(job_id, update_data)
    if not updated:
        raise HTTPException(status_code=404, detail="Job not found")
    
    return _job_to_response_dict(updated)


def mark_lead_as_do_not_apply(job_id: str) -> Dict[str, Any]:
    """Mark a lead as do not apply."""
    update_data: JobUpdateData = {
        "status": "do_not_apply",
        "lead_status_id": None,
    }
    
    updated = update_job_repo(job_id, update_data)
    if not updated:
        raise HTTPException(status_code=404, detail="Job not found")
    
    return _job_to_response_dict(updated)
