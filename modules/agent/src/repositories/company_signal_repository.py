"""Repository layer for company signal data access."""

from datetime import date
from typing import List, Optional
from sqlalchemy import select
from models.CompanySignal import CompanySignal
from lib.db import async_get_session


async def find_signals_by_org(
    org_id: int,
    signal_type: Optional[str] = None,
    limit: int = 20
) -> List[CompanySignal]:
    """Find signals for an organization, optionally filtered by type."""
    async with async_get_session() as session:
        stmt = (
            select(CompanySignal)
            .where(CompanySignal.organization_id == org_id)
            .order_by(CompanySignal.signal_date.desc().nullslast())
            .limit(limit)
        )
        if signal_type:
            stmt = stmt.where(CompanySignal.signal_type == signal_type)
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def find_signal_by_id(signal_id: int) -> Optional[CompanySignal]:
    """Find signal by ID."""
    async with async_get_session() as session:
        return await session.get(CompanySignal, signal_id)


async def create_signal(
    org_id: int,
    signal_type: str,
    headline: str,
    description: Optional[str] = None,
    signal_date: Optional[str] = None,
    source: Optional[str] = None,
    source_url: Optional[str] = None,
    sentiment: Optional[str] = None,
    relevance_score: Optional[float] = None
) -> CompanySignal:
    """Create a new company signal."""
    async with async_get_session() as session:
        signal = CompanySignal(
            organization_id=org_id,
            signal_type=signal_type,
            headline=headline,
            description=description,
            signal_date=date.fromisoformat(signal_date) if signal_date else None,
            source=source,
            source_url=source_url,
            sentiment=sentiment,
            relevance_score=relevance_score
        )
        session.add(signal)
        await session.commit()
        await session.refresh(signal)
        return signal


async def delete_signal(signal_id: int) -> bool:
    """Delete a company signal."""
    async with async_get_session() as session:
        signal = await session.get(CompanySignal, signal_id)
        if not signal:
            return False
        await session.delete(signal)
        await session.commit()
        return True
