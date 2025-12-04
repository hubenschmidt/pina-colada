"""Repository layer for saved report data access."""

from typing import Any, Dict, List, Optional
from sqlalchemy import delete, func, or_, select
from sqlalchemy.orm import selectinload
from lib.db import async_get_session
from models.SavedReport import SavedReport
from models.SavedReportProject import SavedReportProject
from schemas.report import (

    Aggregation,
    ReportFilter,
    ReportQueryRequest,
    SavedReportCreate,
    SavedReportUpdate,
)

__all__ = [
    "ReportFilter",
    "Aggregation",
    "ReportQueryRequest",
    "SavedReportCreate",
    "SavedReportUpdate",
]


async def find_all_saved_reports(
    tenant_id: int,
    project_id: Optional[int] = None,
    include_global: bool = True,
    search: Optional[str] = None,
    page: int = 1,
    limit: int = 50,
    sort_by: str = "updated_at",
    sort_direction: str = "DESC"
) -> Dict[str, Any]:
    """Find all saved reports for a tenant with pagination and search.

    Args:
        tenant_id: Tenant ID for scoping
        project_id: If provided, filter to reports that include this project
        include_global: If True, also include global reports (no projects assigned)
        search: Optional search string for name/description
        page: Page number (1-indexed)
        limit: Items per page
        sort_by: Column to sort by
        sort_direction: ASC or DESC
    """
    async with async_get_session() as session:
        base_stmt = select(SavedReport).where(SavedReport.tenant_id == tenant_id)

        if project_id is not None:
            has_project = (
                select(SavedReportProject.saved_report_id)
                .where(
                    SavedReportProject.saved_report_id == SavedReport.id,
                    SavedReportProject.project_id == project_id
                )
                .exists()
            )
            is_global = ~(
                select(SavedReportProject.saved_report_id)
                .where(SavedReportProject.saved_report_id == SavedReport.id)
                .exists()
            )

            if include_global:
                base_stmt = base_stmt.where(or_(has_project, is_global))
            else:
                base_stmt = base_stmt.where(has_project)

        if search:
            search_filter = or_(
                SavedReport.name.ilike(f"%{search}%"),
                SavedReport.description.ilike(f"%{search}%")
            )
            base_stmt = base_stmt.where(search_filter)

        # Count total
        count_stmt = select(func.count()).select_from(base_stmt.subquery())
        total_result = await session.execute(count_stmt)
        total = total_result.scalar() or 0

        # Apply sorting
        sort_column = getattr(SavedReport, sort_by, SavedReport.updated_at)
        if sort_direction.upper() == "ASC":
            base_stmt = base_stmt.order_by(sort_column.asc())
        else:
            base_stmt = base_stmt.order_by(sort_column.desc())

        # Apply pagination
        offset = (page - 1) * limit
        stmt = (
            base_stmt
            .options(selectinload(SavedReport.creator), selectinload(SavedReport.projects))
            .offset(offset)
            .limit(limit)
        )

        result = await session.execute(stmt)
        items = list(result.scalars().all())

        total_pages = (total + limit - 1) // limit if limit > 0 else 1

        return {
            "items": items,
            "total": total,
            "page": page,
            "limit": limit,
            "total_pages": total_pages
        }


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
