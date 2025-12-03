"""Routes for jobs API endpoints."""

from typing import Optional

from fastapi import APIRouter, Query, Request

from controllers.job_controller import (
    get_jobs,
    create_job,
    get_job,
    update_job,
    delete_job,
    get_recent_resume_date,
    JobCreate,
    JobUpdate,
)
from lib.auth import require_auth
from lib.error_logging import log_errors


router = APIRouter(prefix="/jobs", tags=["jobs"])


@router.get("")
@log_errors
@require_auth
async def get_jobs_route(
    request: Request,
    page: int = Query(1, ge=1),
    limit: int = Query(50, ge=1, le=100),
    order_by: str = Query("updated_at", alias="orderBy"),
    order: str = Query("DESC", regex="^(ASC|DESC)$"),
    search: Optional[str] = Query(None),
    project_id: Optional[int] = Query(None, alias="projectId"),
):
    """Get all jobs with pagination."""
    return await get_jobs(request, page, limit, order_by, order, search, project_id)


@router.post("")
@log_errors
@require_auth
async def create_job_route(request: Request, job_data: JobCreate):
    """Create a new job."""
    return await create_job(request, job_data)


@router.get("/recent-resume-date")
@log_errors
@require_auth
async def get_recent_resume_date_route(request: Request):
    """Get the most recent resume date."""
    return await get_recent_resume_date()


@router.get("/{job_id}")
@log_errors
@require_auth
async def get_job_route(request: Request, job_id: str):
    """Get a job by ID."""
    return await get_job(job_id)


@router.put("/{job_id}")
@log_errors
@require_auth
async def update_job_route(request: Request, job_id: str, job_data: JobUpdate):
    """Update a job."""
    return await update_job(request, job_id, job_data)


@router.delete("/{job_id}")
@log_errors
@require_auth
async def delete_job_route(request: Request, job_id: str):
    """Delete a job."""
    return await delete_job(job_id)
