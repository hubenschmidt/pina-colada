"""Routes for employee count ranges API endpoints."""

from fastapi import APIRouter, Request

from controllers.employee_count_range_controller import get_all_employee_count_ranges
from lib.auth import require_auth
from lib.error_logging import log_errors


router = APIRouter(prefix="/employee-count-ranges", tags=["employee-count-ranges"])


@router.get("")
@log_errors
@require_auth
async def get_employee_count_ranges(request: Request):
    """Get all employee count ranges."""
    return await get_all_employee_count_ranges()
