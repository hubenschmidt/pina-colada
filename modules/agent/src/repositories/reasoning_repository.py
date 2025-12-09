"""Repository layer for reasoning schema data access."""

from typing import List

from sqlalchemy import select

from lib.db import async_get_session
from models.Reasoning import Reasoning


async def find_reasoning_table_names(reasoning_type: str) -> List[str]:
    """Find all table names for a reasoning type."""
    async with async_get_session() as session:
        stmt = select(Reasoning.table_name).where(Reasoning.type == reasoning_type)
        result = await session.execute(stmt)
        return [row[0] for row in result.fetchall()]


async def find_reasoning_by_type(reasoning_type: str) -> List[Reasoning]:
    """Find all reasoning records for a given type."""
    async with async_get_session() as session:
        stmt = select(Reasoning).where(Reasoning.type == reasoning_type)
        result = await session.execute(stmt)
        return list(result.scalars().all())
