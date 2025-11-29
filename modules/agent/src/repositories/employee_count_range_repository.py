"""Repository layer for employee count range data access."""

from typing import List, Optional
from sqlalchemy import select
from models.EmployeeCountRange import EmployeeCountRange
from lib.db import async_get_session


async def find_all_employee_count_ranges() -> List[EmployeeCountRange]:
    """Find all employee count ranges, ordered by display_order."""
    async with async_get_session() as session:
        stmt = select(EmployeeCountRange).order_by(EmployeeCountRange.display_order)
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def find_employee_count_range_by_id(range_id: int) -> Optional[EmployeeCountRange]:
    """Find employee count range by ID."""
    async with async_get_session() as session:
        return await session.get(EmployeeCountRange, range_id)
