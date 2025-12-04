"""Repository layer for salary range data access."""

from typing import List, Optional
from sqlalchemy import select
from models.SalaryRange import SalaryRange
from lib.db import async_get_session

async def find_all_salary_ranges() -> List[SalaryRange]:
    """Find all salary ranges, ordered by display_order."""
    async with async_get_session() as session:
        stmt = select(SalaryRange).order_by(SalaryRange.display_order)
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def find_salary_range_by_id(range_id: int) -> Optional[SalaryRange]:
    """Find salary range by ID."""
    async with async_get_session() as session:
        return await session.get(SalaryRange, range_id)
