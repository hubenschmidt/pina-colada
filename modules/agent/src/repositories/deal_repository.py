"""Repository layer for deal data access."""

import logging
from typing import List, Optional, Dict, Any
from sqlalchemy import select
from sqlalchemy.orm import joinedload
from models.Deal import Deal
from lib.db import async_get_session

logger = logging.getLogger(__name__)


async def find_all_deals(tenant_id: Optional[int] = None) -> List[Deal]:
    """Find all deals, optionally filtered by tenant."""
    async with async_get_session() as session:
        stmt = select(Deal).options(joinedload(Deal.current_status)).order_by(Deal.created_at.desc())
        if tenant_id is not None:
            stmt = stmt.where(Deal.tenant_id == tenant_id)
        result = await session.execute(stmt)
        return list(result.unique().scalars().all())


async def find_deal_by_id(deal_id: int) -> Optional[Deal]:
    """Find deal by ID."""
    async with async_get_session() as session:
        stmt = select(Deal).options(joinedload(Deal.current_status)).where(Deal.id == deal_id)
        result = await session.execute(stmt)
        return result.unique().scalar_one_or_none()


async def find_deal_by_name(name: str, tenant_id: Optional[int] = None) -> Optional[Deal]:
    """Find deal by name, optionally scoped to tenant."""
    async with async_get_session() as session:
        stmt = select(Deal).where(Deal.name == name)
        if tenant_id is not None:
            stmt = stmt.where(Deal.tenant_id == tenant_id)
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def create_deal(data: Dict[str, Any]) -> Deal:
    """Create a new deal."""
    async with async_get_session() as session:
        try:
            deal = Deal(**data)
            session.add(deal)
            await session.commit()
            await session.refresh(deal)
            return deal
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to create deal: {e}")
            raise


async def get_or_create_deal(name: str, tenant_id: Optional[int] = None) -> Deal:
    """Get or create deal by name."""
    existing = await find_deal_by_name(name, tenant_id)
    if existing:
        return existing

    data: Dict[str, Any] = {
        "name": name,
        "tenant_id": tenant_id
    }
    return await create_deal(data)


async def update_deal(deal_id: int, data: Dict[str, Any]) -> Optional[Deal]:
    """Update an existing deal."""
    async with async_get_session() as session:
        try:
            deal = await session.get(Deal, deal_id)
            if not deal:
                return None

            for key, value in data.items():
                if hasattr(deal, key):
                    setattr(deal, key, value)
            await session.commit()
            await session.refresh(deal)
            return deal
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to update deal: {e}")
            raise


async def delete_deal(deal_id: int) -> bool:
    """Delete a deal by ID."""
    async with async_get_session() as session:
        try:
            deal = await session.get(Deal, deal_id)
            if not deal:
                return False
            await session.delete(deal)
            await session.commit()
            return True
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to delete deal: {e}")
            raise
