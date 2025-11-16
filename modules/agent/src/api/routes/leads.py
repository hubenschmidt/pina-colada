"""Routes for leads API endpoints."""

from typing import Optional
from fastapi import APIRouter, Query, Request
from lib.auth import require_auth
from lib.error_logging import log_errors
from controllers.job_controller import (
    get_leads,
    mark_lead_as_applied,
    mark_lead_as_do_not_apply
)

router = APIRouter(prefix="/leads", tags=["leads"])


@router.get("")
@log_errors
@require_auth
async def get_leads_route(request: Request, statuses: Optional[str] = Query(None)):
    """Get all job leads, optionally filtered by status names."""
    return get_leads(statuses)


@router.post("/{job_id}/apply")
@log_errors
@require_auth
async def mark_lead_as_applied_route(request: Request, job_id: str):
    """Mark a lead as applied."""
    return mark_lead_as_applied(job_id)


@router.post("/{job_id}/do-not-apply")
@log_errors
@require_auth
async def mark_lead_as_do_not_apply_route(request: Request, job_id: str):
    """Mark a lead as do not apply."""
    return mark_lead_as_do_not_apply(job_id)
