"""Repository layer for saved report data access."""

import logging
from typing import List, Optional, Dict, Any

from sqlalchemy import select
from sqlalchemy.orm import selectinload

from models.SavedReport import SavedReport
from lib.db import async_get_session

logger = logging.getLogger(__name__)


async def find_all_saved_reports(tenant_id: int) -> List[SavedReport]:
    """Find all saved reports for a tenant."""
    async with async_get_session() as session:
        stmt = (
            select(SavedReport)
            .options(selectinload(SavedReport.creator))
            .where(SavedReport.tenant_id == tenant_id)
            .order_by(SavedReport.updated_at.desc())
        )
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def find_saved_report_by_id(report_id: int, tenant_id: int) -> Optional[SavedReport]:
    """Find saved report by ID, scoped to tenant."""
    async with async_get_session() as session:
        stmt = (
            select(SavedReport)
            .options(selectinload(SavedReport.creator))
            .where(SavedReport.id == report_id, SavedReport.tenant_id == tenant_id)
        )
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def create_saved_report(data: Dict[str, Any]) -> SavedReport:
    """Create a new saved report."""
    async with async_get_session() as session:
        report = SavedReport(**data)
        session.add(report)
        await session.commit()
        # Re-fetch with relationships loaded
        stmt = (
            select(SavedReport)
            .options(selectinload(SavedReport.creator))
            .where(SavedReport.id == report.id)
        )
        result = await session.execute(stmt)
        return result.scalar_one()


async def update_saved_report(report_id: int, tenant_id: int, data: Dict[str, Any]) -> Optional[SavedReport]:
    """Update an existing saved report."""
    async with async_get_session() as session:
        stmt = select(SavedReport).where(
            SavedReport.id == report_id,
            SavedReport.tenant_id == tenant_id
        )
        result = await session.execute(stmt)
        report = result.scalar_one_or_none()
        if not report:
            return None
        for key, value in data.items():
            setattr(report, key, value)
        await session.commit()
        # Re-fetch with relationships loaded
        stmt = (
            select(SavedReport)
            .options(selectinload(SavedReport.creator))
            .where(SavedReport.id == report_id)
        )
        result = await session.execute(stmt)
        return result.scalar_one()


async def delete_saved_report(report_id: int, tenant_id: int) -> bool:
    """Delete a saved report. Returns True if deleted, False if not found."""
    async with async_get_session() as session:
        stmt = select(SavedReport).where(
            SavedReport.id == report_id,
            SavedReport.tenant_id == tenant_id
        )
        result = await session.execute(stmt)
        report = result.scalar_one_or_none()
        if not report:
            return False
        await session.delete(report)
        await session.commit()
        return True
