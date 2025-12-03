"""Service layer for employee count range business logic."""

from typing import List

from repositories.employee_count_range_repository import find_all_employee_count_ranges


async def get_all_employee_count_ranges() -> List:
    """Get all employee count ranges."""
    return await find_all_employee_count_ranges()
