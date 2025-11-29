"""Repository layer for saved report data access."""

import logging
from typing import List, Optional, Dict, Any

from sqlalchemy import select, delete, or_, exists
from sqlalchemy.orm import selectinload

from models.SavedReport import SavedReport
from models.SavedReportProject import SavedReportProject
from lib.db import async_get_session

logger = logging.getLogger(__name__)


async def find_all_saved_reports(
    tenant_id: int,
    project_id: Optional[int] = None,
    include_global: bool = True
) -> List[SavedReport]:
    """Find all saved reports for a tenant, optionally filtered by project.

    Args:
        tenant_id: Tenant ID for scoping
        project_id: If provided, filter to reports that include this project
        include_global: If True, also include global reports (no projects assigned)
    """
    async with async_get_session() as session:
        stmt = (
            select(SavedReport)
            .options(selectinload(SavedReport.creator), selectinload(SavedReport.projects))
            .where(SavedReport.tenant_id == tenant_id)
        )

        if project_id is not None:
            # Subquery to check if report has this project
            has_project = (
                select(SavedReportProject.saved_report_id)
                .where(
                    SavedReportProject.saved_report_id == SavedReport.id,
                    SavedReportProject.project_id == project_id
                )
                .exists()
            )
            # Subquery to check if report has no projects (global)
            is_global = ~(
                select(SavedReportProject.saved_report_id)
                .where(SavedReportProject.saved_report_id == SavedReport.id)
                .exists()
            )

            if include_global:
                stmt = stmt.where(or_(has_project, is_global))
            else:
                stmt = stmt.where(has_project)
        # else: No project filter - return all reports (both global and project-specific)

        stmt = stmt.order_by(SavedReport.updated_at.desc())
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def find_saved_report_by_id(report_id: int, tenant_id: int) -> Optional[SavedReport]:
    """Find saved report by ID, scoped to tenant."""
    async with async_get_session() as session:
        stmt = (
            select(SavedReport)
            .options(selectinload(SavedReport.creator), selectinload(SavedReport.projects))
            .where(SavedReport.id == report_id, SavedReport.tenant_id == tenant_id)
        )
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def create_saved_report(data: Dict[str, Any], project_ids: Optional[List[int]] = None) -> SavedReport:
    """Create a new saved report with optional project assignments."""
    async with async_get_session() as session:
        report = SavedReport(**data)
        session.add(report)
        await session.flush()  # Get the report ID

        # Add project associations
        if project_ids:
            for pid in project_ids:
                link = SavedReportProject(saved_report_id=report.id, project_id=pid)
                session.add(link)

        await session.commit()

        # Re-fetch with relationships loaded
        stmt = (
            select(SavedReport)
            .options(selectinload(SavedReport.creator), selectinload(SavedReport.projects))
            .where(SavedReport.id == report.id)
        )
        result = await session.execute(stmt)
        return result.scalar_one()


async def update_saved_report(
    report_id: int,
    tenant_id: int,
    data: Dict[str, Any],
    project_ids: Optional[List[int]] = None
) -> Optional[SavedReport]:
    """Update an existing saved report, optionally updating project assignments."""
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

        # Update project associations if provided
        if project_ids is not None:
            # Delete existing links
            await session.execute(
                delete(SavedReportProject).where(SavedReportProject.saved_report_id == report_id)
            )
            # Add new links
            for pid in project_ids:
                link = SavedReportProject(saved_report_id=report_id, project_id=pid)
                session.add(link)

        await session.commit()

        # Re-fetch with relationships loaded
        stmt = (
            select(SavedReport)
            .options(selectinload(SavedReport.creator), selectinload(SavedReport.projects))
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
