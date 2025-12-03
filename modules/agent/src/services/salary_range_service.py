"""Service layer for salary range business logic."""

from typing import List

from repositories.salary_range_repository import find_all_salary_ranges


async def get_all_salary_ranges() -> List:
    """Get all salary ranges."""
    return await find_all_salary_ranges()
