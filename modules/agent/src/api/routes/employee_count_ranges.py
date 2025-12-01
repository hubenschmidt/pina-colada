"""Routes for employee count ranges API endpoints."""

from fastapi import APIRouter, Request
from lib.auth import require_auth
from lib.error_logging import log_errors
from lib.serialization import model_to_dict
from repositories.employee_count_range_repository import find_all_employee_count_ranges

router = APIRouter(prefix="/employee-count-ranges", tags=["employee-count-ranges"])


@router.get("")
@log_errors
@require_auth
async def get_employee_count_ranges(request: Request):
    """Get all employee count ranges."""
    ranges = await find_all_employee_count_ranges()
    return [model_to_dict(r, include_relationships=False) for r in ranges]
