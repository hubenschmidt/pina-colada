"""Repository layer for funding stage data access."""

from typing import List, Optional
from sqlalchemy import select
from models.FundingStage import FundingStage
from lib.db import async_get_session

async def find_all_funding_stages() -> List[FundingStage]:
    """Find all funding stages, ordered by display_order."""
    async with async_get_session() as session:
        stmt = select(FundingStage).order_by(FundingStage.display_order)
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def find_funding_stage_by_id(stage_id: int) -> Optional[FundingStage]:
    """Find funding stage by ID."""
    async with async_get_session() as session:
        return await session.get(FundingStage, stage_id)
