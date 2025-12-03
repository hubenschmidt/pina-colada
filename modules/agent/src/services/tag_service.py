"""Service layer for tag business logic."""

from typing import List

from sqlalchemy import select

from lib.db import async_get_session
from models.Tag import Tag


async def get_all_tags() -> List[Tag]:
    """Get all tags ordered by name."""
    async with async_get_session() as session:
        stmt = select(Tag).order_by(Tag.name)
        result = await session.execute(stmt)
        return result.scalars().all()
