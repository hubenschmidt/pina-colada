"""Repository layer for industry data access."""

from typing import List, Optional
from sqlalchemy import select
from lib.db import async_get_session
from models.Industry import Industry
from schemas.industry import IndustryCreate

__all__ = ["IndustryCreate"]


async def find_all_industries() -> List[Industry]:
    """Find all industries ordered by name."""
    async with async_get_session() as session:
        stmt = select(Industry).order_by(Industry.name)
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def find_industry_by_id(industry_id: int) -> Optional[Industry]:
    """Find industry by ID."""
    async with async_get_session() as session:
        return await session.get(Industry, industry_id)


async def find_industry_by_name(name: str) -> Optional[Industry]:
    """Find industry by name (case-insensitive)."""
    async with async_get_session() as session:
        stmt = select(Industry).where(Industry.name.ilike(name))
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def create_industry(name: str) -> Industry:
    """Create a new industry."""
    async with async_get_session() as session:
        industry = Industry(name=name)
        session.add(industry)
        await session.commit()
        await session.refresh(industry)
        return industry
