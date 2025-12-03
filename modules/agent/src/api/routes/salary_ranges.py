"""Routes for salary ranges API endpoints."""

from fastapi import APIRouter, Request

from controllers.salary_range_controller import get_all_salary_ranges
from lib.auth import require_auth
from lib.error_logging import log_errors


router = APIRouter(prefix="/salary-ranges", tags=["salary-ranges"])


@router.get("")
@log_errors
@require_auth
async def get_salary_ranges(request: Request):
    """Get all salary ranges."""
    return await get_all_salary_ranges()
