"""Service layer for project business logic."""

import logging
from datetime import date
from typing import Optional, Dict, Any, List

from fastapi import HTTPException

from repositories.project_repository import (
    find_all_projects,
    find_project_by_id,
    count_project_deals,
    count_project_leads,
    create_project as create_project_repo,
    update_project as update_project_repo,
    delete_project as delete_project_repo,
    find_project_leads,
    find_project_deals,
    ProjectCreate,
    ProjectUpdate,
)

# Re-export Pydantic models for controllers
__all__ = ["ProjectCreate", "ProjectUpdate"]

logger = logging.getLogger(__name__)


def _parse_date(date_str: Optional[str]) -> Optional[date]:
    """Parse date string to date object."""
    if not date_str:
        return None
    return date.fromisoformat(date_str[:10])


async def get_projects(tenant_id: Optional[int]) -> List:
    """Get all projects for tenant."""
    return await find_all_projects(tenant_id)


async def get_project(project_id: int, tenant_id: Optional[int]):
    """Get project by ID with related counts."""
    project = await find_project_by_id(project_id, tenant_id)

    if not project:
        raise HTTPException(status_code=404, detail="Project not found")

    deals_count = await count_project_deals(project_id)
    leads_count = await count_project_leads(project_id)

    return project, deals_count, leads_count


async def create_project(
    project_data: Dict[str, Any],
    tenant_id: Optional[int],
    user_id: Optional[int],
):
    """Create a new project."""
    return await create_project_repo(
        tenant_id=tenant_id,
        name=project_data.get("name"),
        description=project_data.get("description"),
        status=project_data.get("status"),
        current_status_id=project_data.get("current_status_id"),
        start_date=_parse_date(project_data.get("start_date")),
        end_date=_parse_date(project_data.get("end_date")),
        user_id=user_id,
    )


async def update_project(
    project_id: int,
    project_data: Dict[str, Any],
    tenant_id: Optional[int],
    user_id: Optional[int],
):
    """Update a project."""
    # Parse dates before passing to repository
    data = dict(project_data)
    if data.get("start_date") is not None:
        data["start_date"] = _parse_date(data["start_date"])
    if data.get("end_date") is not None:
        data["end_date"] = _parse_date(data["end_date"])

    project = await update_project_repo(project_id, tenant_id, data, user_id)

    if not project:
        raise HTTPException(status_code=404, detail="Project not found")

    return project


async def delete_project(project_id: int, tenant_id: Optional[int]):
    """Delete a project."""
    deleted = await delete_project_repo(project_id, tenant_id)

    if not deleted:
        raise HTTPException(status_code=404, detail="Project not found")

    return True


async def get_project_leads(project_id: int, tenant_id: Optional[int]) -> List:
    """Get leads for a project."""
    return await find_project_leads(project_id, tenant_id)


async def get_project_deals(project_id: int, tenant_id: Optional[int]) -> List:
    """Get deals for a project."""
    return await find_project_deals(project_id, tenant_id)
