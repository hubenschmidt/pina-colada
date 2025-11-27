"""Routes for revenue ranges API endpoints."""

from fastapi import APIRouter, Request, Query
from lib.auth import require_auth
from lib.error_logging import log_errors
from lib.serialization import model_to_dict
from repositories.revenue_range_repository import find_revenue_ranges_by_category

router = APIRouter(prefix="/revenue-ranges", tags=["revenue-ranges"])


@router.get("")
@log_errors
@require_auth
async def get_revenue_ranges(request: Request, category: str = Query(..., description="Category (e.g., 'salary')")):
    """Get all revenue ranges for a category."""
    ranges = await find_revenue_ranges_by_category(category)
    return [model_to_dict(r, include_relationships=False) for r in ranges]
