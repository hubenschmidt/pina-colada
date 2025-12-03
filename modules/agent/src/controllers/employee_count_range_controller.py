"""Controller layer for employee count range routing to services."""

from typing import List

from lib.decorators import handle_http_exceptions
from lib.serialization import model_to_dict
from repositories.employee_count_range_repository import find_all_employee_count_ranges


@handle_http_exceptions
async def get_all_employee_count_ranges() -> List[dict]:
    """Get all employee count ranges."""
    ranges = await find_all_employee_count_ranges()
    return [model_to_dict(r, include_relationships=False) for r in ranges]
