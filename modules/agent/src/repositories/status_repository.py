"""Repository layer for status data access."""

import logging
from typing import List, Optional, Dict, Any
from sqlalchemy import select
from models.Status import Status
from lib.db import async_get_session

logger = logging.getLogger(__name__)


async def find_all_statuses() -> List[Status]:
    """Find all statuses."""
    async with async_get_session() as session:
        stmt = select(Status).order_by(Status.category, Status.name)
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def find_statuses_by_category(category: str) -> List[Status]:
    """Find statuses by category."""
    async with async_get_session() as session:
        stmt = select(Status).where(Status.category == category).order_by(Status.name)
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def find_status_by_id(status_id: int) -> Optional[Status]:
    """Find status by ID."""
    async with async_get_session() as session:
        return await session.get(Status, status_id)


async def find_status_by_name(name: str, category: Optional[str] = None) -> Optional[Status]:
    """Find status by name and optional category."""
    async with async_get_session() as session:
        stmt = select(Status).where(Status.name == name)
        if category:
            stmt = stmt.where(Status.category == category)
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def create_status(data: Dict[str, Any]) -> Status:
    """Create a new status."""
    async with async_get_session() as session:
        try:
            status = Status(**data)
            session.add(status)
            await session.commit()
            await session.refresh(status)
            return status
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to create status: {e}")
            raise


async def update_status(status_id: int, data: Dict[str, Any]) -> Optional[Status]:
    """Update an existing status."""
    async with async_get_session() as session:
        try:
            status = await session.get(Status, status_id)
            if not status:
                return None

            for key, value in data.items():
                if hasattr(status, key):
                    setattr(status, key, value)
            await session.commit()
            await session.refresh(status)
            return status
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to update status: {e}")
            raise


async def delete_status(status_id: int) -> bool:
    """Delete a status by ID."""
    async with async_get_session() as session:
        try:
            status = await session.get(Status, status_id)
            if not status:
                return False
            await session.delete(status)
            await session.commit()
            return True
        except Exception as e:
            await session.rollback()
            logger.error(f"Failed to delete status: {e}")
            raise
