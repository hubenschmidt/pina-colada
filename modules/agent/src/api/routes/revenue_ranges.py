"""Routes for revenue ranges API endpoints."""

from fastapi import APIRouter, Request
from lib.auth import require_auth
from lib.error_logging import log_errors
from lib.serialization import model_to_dict
from repositories.revenue_range_repository import find_all_revenue_ranges

router = APIRouter(prefix="/revenue-ranges", tags=["revenue-ranges"])


@router.get("")
@log_errors
@require_auth
async def get_revenue_ranges(request: Request):
    """Get all revenue ranges."""
    ranges = await find_all_revenue_ranges()
    return [model_to_dict(r, include_relationships=False) for r in ranges]
