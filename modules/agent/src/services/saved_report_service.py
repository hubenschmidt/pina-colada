"""Service layer for saved report CRUD operations."""

from typing import Optional, List

from repositories.saved_report_repository import (
    find_all_saved_reports,
    find_saved_report_by_id,
    create_saved_report,
    update_saved_report,
    delete_saved_report,
    ReportQueryRequest,
    SavedReportCreate,
    SavedReportUpdate,
)

# Re-export schemas for controller
__all__ = [
    "ReportQueryRequest",
    "SavedReportCreate",
    "SavedReportUpdate",
    "get_saved_reports",
    "get_saved_report_by_id",
    "create_report",
    "update_report",
    "delete_report",
]


async def get_saved_reports(
    tenant_id: int,
    project_id: Optional[int],
    include_global: bool,
    search: Optional[str],
    page: int,
    limit: int,
    sort_by: str,
    sort_direction: str,
) -> dict:
    """Get all saved reports with pagination."""
    return await find_all_saved_reports(
        tenant_id,
        project_id,
        include_global,
        search=search,
        page=page,
        limit=limit,
        sort_by=sort_by,
        sort_direction=sort_direction,
    )


async def get_saved_report_by_id(report_id: int, tenant_id: int):
    """Get a saved report by ID."""
    return await find_saved_report_by_id(report_id, tenant_id)


async def create_report(data: dict, project_ids: Optional[List[int]] = None):
    """Create a new saved report."""
    return await create_saved_report(data, project_ids=project_ids)


async def update_report(
    report_id: int,
    tenant_id: int,
    update_data: dict,
    project_ids: Optional[List[int]] = None,
):
    """Update a saved report."""
    return await update_saved_report(report_id, tenant_id, update_data, project_ids)


async def delete_report(report_id: int, tenant_id: int) -> bool:
    """Delete a saved report."""
    return await delete_saved_report(report_id, tenant_id)
