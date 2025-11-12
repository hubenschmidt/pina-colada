"""Routes for lead status API endpoints."""

from fastapi import APIRouter
from api.controllers.jobs_controller import get_lead_statuses

router = APIRouter(prefix="/api/lead-statuses", tags=["lead-statuses"])


@router.get("/")
async def get_lead_statuses_route():
    """Get all lead statuses."""
    return get_lead_statuses()
