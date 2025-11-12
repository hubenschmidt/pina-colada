"""Routes for leads API endpoints."""

from typing import Optional
from fastapi import APIRouter, Query
from modules.agent.src.controllers.job_controller import (
    get_leads,
    mark_lead_as_applied,
    mark_lead_as_do_not_apply
)

router = APIRouter(prefix="/api/leads", tags=["leads"])


@router.get("/")
async def get_leads_route(statuses: Optional[str] = Query(None)):
    """Get all job leads, optionally filtered by status names."""
    return get_leads(statuses)


@router.post("/{job_id}/apply")
async def mark_lead_as_applied_route(job_id: str):
    """Mark a lead as applied."""
    return mark_lead_as_applied(job_id)


@router.post("/{job_id}/do-not-apply")
async def mark_lead_as_do_not_apply_route(job_id: str):
    """Mark a lead as do not apply."""
    return mark_lead_as_do_not_apply(job_id)
