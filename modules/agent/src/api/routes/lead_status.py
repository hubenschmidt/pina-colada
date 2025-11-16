"""Routes for lead status API endpoints."""

from fastapi import APIRouter, Request
from lib.auth import require_auth
from lib.error_logging import log_errors
from controllers.job_controller import get_statuses

router = APIRouter(prefix="/lead-statuses", tags=["lead-statuses"])


@router.get("/")
@log_errors
@require_auth
async def get_lead_statuses_route(request: Request):
    """Get all lead statuses."""
    return get_statuses()
