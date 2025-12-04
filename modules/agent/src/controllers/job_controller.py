"""Controller layer for job routing to services"""

from typing import List, Optional, Dict, Any

from fastapi import Request

from lib.decorators import handle_http_exceptions
from serializers.base import model_to_dict
from serializers.common import to_paged_response
from serializers.job import job_to_list_response, job_to_detail_response
from schemas.job import JobCreate, JobUpdate
from services.job_service import (
    create_job as create_job_service,
    delete_job as delete_job_service,
    get_job as get_job_service,
    get_jobs_paginated,
    get_jobs_with_status as get_jobs_with_status_service,
    get_recent_resume_date as get_recent_resume_date_service,
    get_statuses as get_statuses_service,
    update_job as update_job_service,
)


@handle_http_exceptions
async def get_jobs(
    request: Request,
    page: int,
    limit: int,
    order_by: str,
    order: str,
    search: Optional[str] = None,
    project_id: Optional[int] = None,
) -> dict:
    """Get all jobs with pagination."""
    tenant_id = getattr(request.state, "tenant_id", None)
    paginated_jobs, total_count = await get_jobs_paginated(
        page, limit, order_by, order, search, tenant_id, project_id
    )
    items = [job_to_list_response(job) for job in paginated_jobs]
    return to_paged_response(total_count, page, limit, items)


@handle_http_exceptions
async def create_job(request: Request, data: JobCreate) -> Dict[str, Any]:
    """Create a new job."""
    job_data = data.dict()
    job_data["tenant_id"] = getattr(request.state, "tenant_id", None)
    job_data["user_id"] = getattr(request.state, "user_id", None)
    created = await create_job_service(job_data)
    return job_to_detail_response(created)


@handle_http_exceptions
async def get_job(job_id: str) -> Dict[str, Any]:
    """Get a job by ID."""
    job = await get_job_service(job_id)
    return job_to_detail_response(job)


@handle_http_exceptions
async def update_job(request: Request, job_id: str, data: JobUpdate) -> Dict[str, Any]:
    """Update a job."""
    job_data = data.dict(exclude_unset=True)
    job_data["user_id"] = getattr(request.state, "user_id", None)
    updated = await update_job_service(job_id, job_data)
    return job_to_detail_response(updated)


@handle_http_exceptions
async def delete_job(job_id: str) -> dict:
    """Delete a job."""
    await delete_job_service(job_id)
    return {"success": True}


@handle_http_exceptions
async def get_statuses() -> List[Dict[str, Any]]:
    """Get all job statuses."""
    statuses = await get_statuses_service()
    return [model_to_dict(s, include_relationships=False) for s in statuses]


@handle_http_exceptions
async def get_leads(statuses: Optional[str] = None) -> List[Dict[str, Any]]:
    """Get all job leads, optionally filtered by status names."""
    status_names = None
    if statuses:
        status_names = [s.strip() for s in statuses.split(",")]

    jobs = await get_jobs_with_status_service(status_names)

    items = []
    for job in jobs:
        job_dict = job_to_detail_response(job)
        if job.lead and job.lead.current_status:
            job_dict["lead_status"] = model_to_dict(
                job.lead.current_status, include_relationships=False
            )
        items.append(job_dict)

    return items


@handle_http_exceptions
async def mark_lead_as_applied(job_id: str) -> Dict[str, Any]:
    """Mark a lead as applied."""
    update_data: Dict[str, Any] = {"status": "applied"}
    updated = await update_job_service(job_id, update_data)
    return job_to_detail_response(updated)


@handle_http_exceptions
async def mark_lead_as_do_not_apply(job_id: str) -> Dict[str, Any]:
    """Mark a lead as do not apply."""
    update_data: Dict[str, Any] = {"status": "do_not_apply"}
    updated = await update_job_service(job_id, update_data)
    return job_to_detail_response(updated)


@handle_http_exceptions
async def get_recent_resume_date() -> Dict[str, Optional[str]]:
    """Get the most recent resume date."""
    resume_date = await get_recent_resume_date_service()
    return {"resume_date": resume_date}
