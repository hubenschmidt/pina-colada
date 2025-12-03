"""Routes for revenue ranges API endpoints."""

from fastapi import APIRouter, Request

from controllers.revenue_range_controller import get_all_revenue_ranges
from lib.auth import require_auth
from lib.error_logging import log_errors


router = APIRouter(prefix="/revenue-ranges", tags=["revenue-ranges"])


@router.get("")
@log_errors
@require_auth
async def get_revenue_ranges(request: Request):
    """Get all revenue ranges."""
    return await get_all_revenue_ranges()
