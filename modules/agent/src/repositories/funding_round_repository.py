"""Repository layer for funding round data access."""

from typing import List, Optional
from sqlalchemy import select
from models.FundingRound import FundingRound
from lib.db import async_get_session


async def find_funding_rounds_by_org(org_id: int) -> List[FundingRound]:
    """Find all funding rounds for an organization, ordered by date descending."""
    async with async_get_session() as session:
        stmt = (
            select(FundingRound)
            .where(FundingRound.organization_id == org_id)
            .order_by(FundingRound.announced_date.desc().nullslast())
        )
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def find_funding_round_by_id(round_id: int) -> Optional[FundingRound]:
    """Find funding round by ID."""
    async with async_get_session() as session:
        return await session.get(FundingRound, round_id)


async def create_funding_round(
    org_id: int,
    round_type: str,
    amount: Optional[int] = None,
    announced_date: Optional[str] = None,
    lead_investor: Optional[str] = None,
    source_url: Optional[str] = None
) -> FundingRound:
    """Create a new funding round."""
    from datetime import date
    async with async_get_session() as session:
        funding_round = FundingRound(
            organization_id=org_id,
            round_type=round_type,
            amount=amount,
            announced_date=date.fromisoformat(announced_date) if announced_date else None,
            lead_investor=lead_investor,
            source_url=source_url
        )
        session.add(funding_round)
        await session.commit()
        await session.refresh(funding_round)
        return funding_round


async def delete_funding_round(round_id: int) -> bool:
    """Delete a funding round."""
    async with async_get_session() as session:
        funding_round = await session.get(FundingRound, round_id)
        if not funding_round:
            return False
        await session.delete(funding_round)
        await session.commit()
        return True
