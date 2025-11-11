"""Controller for jobs API endpoints."""

import logging
from typing import List, Optional, Type
from fastapi import APIRouter, HTTPException, Query
from pydantic import BaseModel, create_model
from agent.repositories.job_repository import (
    find_all_jobs,
    count_jobs,
    find_job_by_id,
    create_job as create_job_repo,
    update_job as update_job_repo,
    delete_job as delete_job_repo
)
from agent.models.Job import JobCreateData, JobUpdateData, orm_to_dict

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/api/jobs", tags=["jobs"])


def _parse_paging(
    page: int = Query(1, ge=1),
    limit: int = Query(25, ge=1, le=100),
    order_by: str = Query("date", alias="orderBy"),
    order: str = Query("DESC", regex="^(ASC|DESC)$"),
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


def _make_job_create_model() -> Type[BaseModel]:
    """Create JobCreate model functionally."""
    return create_model(
        "JobCreate",
        company=(str, ...),
        job_title=(str, ...),
        date=(Optional[str], None),
        job_url=(Optional[str], None),
        salary_range=(Optional[str], None),
        notes=(Optional[str], None),
        resume=(Optional[str], None),
        status=(str, "applied"),
        source=(str, "manual"),
    )


def _make_job_update_model() -> Type[BaseModel]:
    """Create JobUpdate model functionally."""
    return create_model(
        "JobUpdate",
        company=(Optional[str], None),
        job_title=(Optional[str], None),
        date=(Optional[str], None),
        job_url=(Optional[str], None),
        salary_range=(Optional[str], None),
        notes=(Optional[str], None),
        resume=(Optional[str], None),
        status=(Optional[str], None),
        source=(Optional[str], None),
    )


def _make_job_response_model() -> Type[BaseModel]:
    """Create JobResponse model functionally."""
    return create_model(
        "JobResponse",
        id=(str, ...),
        company=(str, ...),
        job_title=(str, ...),
        date=(str, ...),
        status=(str, ...),
        job_url=(Optional[str], None),
        salary_range=(Optional[str], None),
        notes=(Optional[str], None),
        resume=(Optional[str], None),
        source=(str, ...),
        created_at=(str, ...),
        updated_at=(str, ...),
    )


# Create models functionally (required by FastAPI for validation/serialization)
JobCreate = _make_job_create_model()
JobUpdate = _make_job_update_model()
JobResponse = _make_job_response_model()


@router.get("/")
async def get_jobs(
    page: int = Query(1, ge=1),
    limit: int = Query(25, ge=1, le=100),
    order_by: str = Query("date", alias="orderBy"),
    order: str = Query("DESC", regex="^(ASC|DESC)$"),
):
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
    
    items = [
        JobResponse(
            id=str(job.id),
            company=job.company or "",
            job_title=job.job_title or "",
            date=job.date.isoformat() if job.date else "",
            status=job.status or "applied",
            job_url=job.job_url,
            salary_range=job.salary_range,
            notes=job.notes,
            resume=job.resume.isoformat() if job.resume else None,
            source=job.source or "manual",
            created_at=job.created_at.isoformat() if job.created_at else "",
            updated_at=job.updated_at.isoformat() if job.updated_at else ""
        )
        for job in paginated_jobs
    ]
    
    return _to_paged_response(total_count, page, limit, items)


@router.post("/", response_model=JobResponse)
async def create_job(job_data: JobCreate):
    """Create a new job."""
    from datetime import datetime
    
    date_obj = None
    if hasattr(job_data, 'date') and job_data.date:
        try:
            date_obj = datetime.fromisoformat(job_data.date.replace('Z', '+00:00'))
        except:
            pass
    
    resume_obj = None
    if hasattr(job_data, 'resume') and job_data.resume:
        try:
            resume_obj = datetime.fromisoformat(job_data.resume.replace('Z', '+00:00'))
        except:
            pass
    
    data: JobCreateData = {
        "company": job_data.company,
        "job_title": job_data.job_title,
        "date": date_obj,
        "job_url": job_data.job_url,
        "salary_range": job_data.salary_range,
        "notes": job_data.notes,
        "resume": resume_obj,
        "status": job_data.status,
        "source": job_data.source
    }
    
    created = create_job_repo(data)
    
    return JobResponse(
        id=str(created.id),
        company=created.company or "",
        job_title=created.job_title or "",
        date=created.date.isoformat() if created.date else "",
        status=created.status or "applied",
        job_url=created.job_url,
        salary_range=created.salary_range,
        notes=created.notes,
        resume=created.resume.isoformat() if created.resume else None,
        source=created.source or "manual",
        created_at=created.created_at.isoformat() if created.created_at else "",
        updated_at=created.updated_at.isoformat() if created.updated_at else ""
    )


@router.get("/{job_id}", response_model=JobResponse)
async def get_job(job_id: str):
    """Get a job by ID."""
    job = find_job_by_id(job_id)
    
    if not job:
        raise HTTPException(status_code=404, detail="Job not found")
    
    return JobResponse(
        id=str(job.id),
        company=job.company or "",
        job_title=job.job_title or "",
        date=job.date.isoformat() if job.date else "",
        status=job.status or "applied",
        job_url=job.job_url,
        salary_range=job.salary_range,
        notes=job.notes,
        resume=job.resume.isoformat() if job.resume else None,
        source=job.source or "manual",
        created_at=job.created_at.isoformat() if job.created_at else "",
        updated_at=job.updated_at.isoformat() if job.updated_at else ""
    )


@router.put("/{job_id}", response_model=JobResponse)
async def update_job(job_id: str, job_data: JobUpdate):
    """Update a job."""
    from datetime import datetime
    
    date_obj = None
    if hasattr(job_data, 'date') and job_data.date is not None:
        try:
            date_obj = datetime.fromisoformat(job_data.date.replace('Z', '+00:00'))
        except:
            pass
    
    resume_obj = None
    if hasattr(job_data, 'resume') and job_data.resume is not None:
        try:
            resume_obj = datetime.fromisoformat(job_data.resume.replace('Z', '+00:00'))
        except:
            pass
    
    data: JobUpdateData = {
        "company": job_data.company if hasattr(job_data, 'company') else None,
        "job_title": job_data.job_title if hasattr(job_data, 'job_title') else None,
        "date": date_obj,
        "job_url": job_data.job_url if hasattr(job_data, 'job_url') else None,
        "salary_range": job_data.salary_range if hasattr(job_data, 'salary_range') else None,
        "notes": job_data.notes if hasattr(job_data, 'notes') else None,
        "resume": resume_obj,
        "status": job_data.status if hasattr(job_data, 'status') else None,
        "source": job_data.source if hasattr(job_data, 'source') else None
    }
    
    # Remove None values
    data = {k: v for k, v in data.items() if v is not None}
    
    updated = update_job_repo(job_id, data)
    
    if not updated:
        raise HTTPException(status_code=404, detail="Job not found")
    
    return JobResponse(
        id=str(updated.id),
        company=updated.company or "",
        job_title=updated.job_title or "",
        date=updated.date.isoformat() if updated.date else "",
        status=updated.status or "applied",
        job_url=updated.job_url,
        salary_range=updated.salary_range,
        notes=updated.notes,
        resume=updated.resume.isoformat() if updated.resume else None,
        source=updated.source or "manual",
        created_at=updated.created_at.isoformat() if updated.created_at else "",
        updated_at=updated.updated_at.isoformat() if updated.updated_at else ""
    )


@router.delete("/{job_id}")
async def delete_job(job_id: str):
    """Delete a job."""
    deleted = delete_job_repo(job_id)
    
    if not deleted:
        raise HTTPException(status_code=404, detail="Job not found")
    
    return {"success": True}

