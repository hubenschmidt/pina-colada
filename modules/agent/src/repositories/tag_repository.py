"""Repository layer for tag data access."""

from typing import List

from sqlalchemy import select

from lib.db import async_get_session
from models.Tag import Tag


async def find_all_tags() -> List[Tag]:
    """Find all tags ordered by name."""
    async with async_get_session() as session:
        stmt = select(Tag).order_by(Tag.name)
        result = await session.execute(stmt)
        return list(result.scalars().all())
