"""Routes for salary ranges API endpoints."""

from fastapi import APIRouter, Request
from lib.auth import require_auth
from lib.error_logging import log_errors
from lib.serialization import model_to_dict
from repositories.salary_range_repository import find_all_salary_ranges

router = APIRouter(prefix="/salary-ranges", tags=["salary-ranges"])


@router.get("")
@log_errors
@require_auth
async def get_salary_ranges(request: Request):
    """Get all salary ranges."""
    ranges = await find_all_salary_ranges()
    return [model_to_dict(r, include_relationships=False) for r in ranges]
