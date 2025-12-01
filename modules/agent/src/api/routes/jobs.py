"""Routes for jobs API endpoints."""

from typing import Optional, List, Type
from fastapi import APIRouter, Query, Request
from pydantic import BaseModel, create_model
from lib.auth import require_auth
from lib.error_logging import log_errors
from controllers.job_controller import (
    get_jobs,
    create_job,
    get_job,
    update_job,
    delete_job,
    get_recent_resume_date,
)

router = APIRouter(prefix="/jobs", tags=["jobs"])


def _make_job_create_model() -> Type[BaseModel]:
    """Create JobCreate model functionally."""
    return create_model(
        "JobCreate",
        account_type=(str, "Organization"),
        account=(Optional[str], None),
        contact_name=(Optional[str], None),
        contacts=(Optional[List[dict]], None),
        industry=(Optional[List[str]], None),
        industry_ids=(Optional[List[int]], None),
        job_title=(str, ...),
        date=(Optional[str], None),
        job_url=(Optional[str], None),
        salary_range=(Optional[str], None),  # Legacy, kept for backwards compat
        salary_range_id=(Optional[int], None),
        description=(Optional[str], None),
        resume=(Optional[str], None),
        status=(str, "applied"),
        source=(str, "manual"),
        project_ids=(Optional[List[int]], None),
    )


def _make_job_update_model() -> Type[BaseModel]:
    """Create JobUpdate model functionally."""
    return create_model(
        "JobUpdate",
        account=(Optional[str], None),
        contacts=(Optional[List[dict]], None),
        job_title=(Optional[str], None),
        date=(Optional[str], None),
        job_url=(Optional[str], None),
        salary_range=(Optional[str], None),  # Legacy, kept for backwards compat
        salary_range_id=(Optional[int], None),
        description=(Optional[str], None),
        resume=(Optional[str], None),
        status=(Optional[str], None),
        source=(Optional[str], None),
        lead_status_id=(Optional[str], None),
        project_ids=(Optional[List[int]], None),
    )


JobCreate = _make_job_create_model()
JobUpdate = _make_job_update_model()


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
    tenant_id = getattr(request.state, "tenant_id", None)
    return await get_jobs(page, limit, order_by, order, search, tenant_id, project_id)


@router.post("")
@log_errors
@require_auth
async def create_job_route(request: Request, job_data: JobCreate):
    """Create a new job."""
    data = job_data.dict()
    data["tenant_id"] = getattr(request.state, "tenant_id", None)
    return await create_job(data)


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
    return await update_job(job_id, job_data.dict(exclude_unset=True))


@router.delete("/{job_id}")
@log_errors
@require_auth
async def delete_job_route(request: Request, job_id: str):
    """Delete a job."""
    return await delete_job(job_id)
