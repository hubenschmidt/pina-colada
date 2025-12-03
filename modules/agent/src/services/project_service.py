"""Service layer for project business logic."""

import logging
from datetime import date
from typing import Optional, Dict, Any, List

from fastapi import HTTPException
from sqlalchemy import select, func
from sqlalchemy.orm import selectinload

from lib.db import async_get_session
from models.Project import Project
from models.Deal import Deal
from models.Lead import Lead
from models.LeadProject import LeadProject

logger = logging.getLogger(__name__)


def _parse_date(date_str: Optional[str]) -> Optional[date]:
    """Parse date string to date object."""
    if not date_str:
        return None
    return date.fromisoformat(date_str[:10])


async def get_projects(tenant_id: Optional[int]) -> List:
    """Get all projects for tenant."""
    async with async_get_session() as session:
        stmt = select(Project).order_by(Project.name)
        if tenant_id:
            stmt = stmt.where(Project.tenant_id == tenant_id)
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def get_project(project_id: int, tenant_id: Optional[int]):
    """Get project by ID with related counts."""
    async with async_get_session() as session:
        stmt = select(Project).where(Project.id == project_id)
        if tenant_id:
            stmt = stmt.where(Project.tenant_id == tenant_id)
        result = await session.execute(stmt)
        project = result.scalar_one_or_none()

        if not project:
            raise HTTPException(status_code=404, detail="Project not found")

        deals_count_result = await session.execute(
            select(func.count(Deal.id)).where(Deal.project_id == project_id)
        )
        deals_count = deals_count_result.scalar() or 0

        leads_count_result = await session.execute(
            select(func.count(LeadProject.lead_id)).where(LeadProject.project_id == project_id)
        )
        leads_count = leads_count_result.scalar() or 0

        return project, deals_count, leads_count


async def create_project(
    project_data: Dict[str, Any],
    tenant_id: Optional[int],
    user_id: Optional[int],
):
    """Create a new project."""
    async with async_get_session() as session:
        project = Project(
            tenant_id=tenant_id,
            name=project_data.get("name"),
            description=project_data.get("description"),
            status=project_data.get("status"),
            current_status_id=project_data.get("current_status_id"),
            start_date=_parse_date(project_data.get("start_date")),
            end_date=_parse_date(project_data.get("end_date")),
            created_by=user_id,
            updated_by=user_id,
        )
        session.add(project)
        await session.commit()
        await session.refresh(project)
        return project


async def update_project(
    project_id: int,
    project_data: Dict[str, Any],
    tenant_id: Optional[int],
    user_id: Optional[int],
):
    """Update a project."""
    async with async_get_session() as session:
        stmt = select(Project).where(Project.id == project_id)
        if tenant_id:
            stmt = stmt.where(Project.tenant_id == tenant_id)
        result = await session.execute(stmt)
        project = result.scalar_one_or_none()

        if not project:
            raise HTTPException(status_code=404, detail="Project not found")

        if project_data.get("name") is not None:
            project.name = project_data["name"]
        if project_data.get("description") is not None:
            project.description = project_data["description"]
        if project_data.get("status") is not None:
            project.status = project_data["status"]
        if project_data.get("current_status_id") is not None:
            project.current_status_id = project_data["current_status_id"]
        if project_data.get("start_date") is not None:
            project.start_date = _parse_date(project_data["start_date"])
        if project_data.get("end_date") is not None:
            project.end_date = _parse_date(project_data["end_date"])
        project.updated_by = user_id

        await session.commit()
        await session.refresh(project)
        return project


async def delete_project(project_id: int, tenant_id: Optional[int]):
    """Delete a project."""
    async with async_get_session() as session:
        stmt = select(Project).where(Project.id == project_id)
        if tenant_id:
            stmt = stmt.where(Project.tenant_id == tenant_id)
        result = await session.execute(stmt)
        project = result.scalar_one_or_none()

        if not project:
            raise HTTPException(status_code=404, detail="Project not found")

        await session.delete(project)
        await session.commit()
        return True


async def get_project_leads(project_id: int, tenant_id: Optional[int]) -> List:
    """Get leads for a project."""
    async with async_get_session() as session:
        stmt = (
            select(Lead)
            .options(selectinload(Lead.current_status), selectinload(Lead.account))
            .join(LeadProject, Lead.id == LeadProject.lead_id)
            .where(LeadProject.project_id == project_id)
            .order_by(Lead.created_at.desc())
        )
        if tenant_id:
            stmt = stmt.where(Lead.tenant_id == tenant_id)
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def get_project_deals(project_id: int, tenant_id: Optional[int]) -> List:
    """Get deals for a project."""
    async with async_get_session() as session:
        stmt = (
            select(Deal)
            .options(selectinload(Deal.current_status))
            .where(Deal.project_id == project_id)
            .order_by(Deal.created_at.desc())
        )
        if tenant_id:
            stmt = stmt.where(Deal.tenant_id == tenant_id)
        result = await session.execute(stmt)
        return list(result.scalars().all())
