"""Repository layer for signal data access."""

from datetime import date
from typing import List, Optional
from sqlalchemy import select
from models.Signal import Signal
from lib.db import async_get_session


async def find_signals_by_account(
    account_id: int,
    signal_type: Optional[str] = None,
    limit: int = 20
) -> List[Signal]:
    """Find signals for an account, optionally filtered by type."""
    async with async_get_session() as session:
        stmt = (
            select(Signal)
            .where(Signal.account_id == account_id)
            .order_by(Signal.signal_date.desc().nullslast())
            .limit(limit)
        )
        if signal_type:
            stmt = stmt.where(Signal.signal_type == signal_type)
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def find_signal_by_id(signal_id: int) -> Optional[Signal]:
    """Find signal by ID."""
    async with async_get_session() as session:
        return await session.get(Signal, signal_id)


async def create_signal(
    account_id: int,
    signal_type: str,
    headline: str,
    description: Optional[str] = None,
    signal_date: Optional[str] = None,
    source: Optional[str] = None,
    source_url: Optional[str] = None,
    sentiment: Optional[str] = None,
    relevance_score: Optional[float] = None
) -> Signal:
    """Create a new signal."""
    async with async_get_session() as session:
        signal = Signal(
            account_id=account_id,
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
    """Delete a signal."""
    async with async_get_session() as session:
        signal = await session.get(Signal, signal_id)
        if not signal:
            return False
        await session.delete(signal)
        await session.commit()
        return True
