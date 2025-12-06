"""Routes for leads API endpoints."""

from typing import Optional
from fastapi import APIRouter, Query, Request
from controllers.job_controller import (
    get_leads,
    get_statuses,
    mark_lead_as_applied,
    mark_lead_as_do_not_apply
)

router = APIRouter(prefix="/leads", tags=["leads"])


@router.get("")
async def get_leads_route(request: Request, statuses: Optional[str] = Query(None)):
    """Get all job leads, optionally filtered by status names."""
    return await get_leads(statuses)


@router.get("/statuses")
async def get_lead_statuses_route(request: Request):
    """Get all lead statuses."""
    return await get_statuses()


@router.post("/{job_id}/apply")
async def mark_lead_as_applied_route(request: Request, job_id: str):
    """Mark a lead as applied."""
    return await mark_lead_as_applied(job_id)


@router.post("/{job_id}/do-not-apply")
async def mark_lead_as_do_not_apply_route(request: Request, job_id: str):
    """Mark a lead as do not apply."""
    return await mark_lead_as_do_not_apply(job_id)
