"""Controller layer for salary range routing to services."""

from typing import List

from lib.decorators import handle_http_exceptions
from serializers.base import model_to_dict
from services.salary_range_service import get_all_salary_ranges as find_all_salary_ranges


@handle_http_exceptions
async def get_all_salary_ranges() -> List[dict]:
    """Get all salary ranges."""
    ranges = await find_all_salary_ranges()
    return [model_to_dict(r, include_relationships=False) for r in ranges]
