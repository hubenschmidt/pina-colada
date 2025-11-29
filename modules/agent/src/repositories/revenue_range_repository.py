"""Repository layer for revenue range data access."""

from typing import List, Optional
from sqlalchemy import select
from models.RevenueRange import RevenueRange
from lib.db import async_get_session


async def find_all_revenue_ranges() -> List[RevenueRange]:
    """Find all revenue ranges, ordered by display_order."""
    async with async_get_session() as session:
        stmt = select(RevenueRange).order_by(RevenueRange.display_order)
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def find_revenue_range_by_id(range_id: int) -> Optional[RevenueRange]:
    """Find revenue range by ID."""
    async with async_get_session() as session:
        return await session.get(RevenueRange, range_id)
