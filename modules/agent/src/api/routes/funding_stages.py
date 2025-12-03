"""Routes for funding stages API endpoints."""

from fastapi import APIRouter, Request

from controllers.funding_stage_controller import get_all_funding_stages
from lib.auth import require_auth
from lib.error_logging import log_errors


router = APIRouter(prefix="/funding-stages", tags=["funding-stages"])


@router.get("")
@log_errors
@require_auth
async def get_funding_stages(request: Request):
    """Get all funding stages."""
    return await get_all_funding_stages()
