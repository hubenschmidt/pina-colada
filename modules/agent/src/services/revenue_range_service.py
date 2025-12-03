"""Service layer for revenue range business logic."""

from typing import List

from repositories.revenue_range_repository import find_all_revenue_ranges


async def get_all_revenue_ranges() -> List:
    """Get all revenue ranges."""
    return await find_all_revenue_ranges()
