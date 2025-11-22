"""Routes for jobs API endpoints."""

from typing import Optional, Type
from fastapi import APIRouter, Query, HTTPException, Request
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
        lead_status_id=(Optional[str], None),
    )


JobCreate = _make_job_create_model()
JobUpdate = _make_job_update_model()


@router.get("")
@log_errors
@require_auth
async def get_jobs_route(
    request: Request,
    page: int = Query(1, ge=1),
    limit: int = Query(25, ge=1, le=100),
    order_by: str = Query("date", alias="orderBy"),
    order: str = Query("DESC", regex="^(ASC|DESC)$"),
    search: Optional[str] = Query(None),
):
    """Get all jobs with pagination."""
    tenant_id = getattr(request.state, "tenant_id", None)
    return await get_jobs(page, limit, order_by, order, search, tenant_id)


@router.post("")
@log_errors
@require_auth
async def create_job_route(request: Request, job_data: JobCreate):
    """Create a new job."""
    return await create_job(job_data.dict())


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
