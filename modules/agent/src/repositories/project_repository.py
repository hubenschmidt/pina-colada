"""Repository layer for project data access."""

from datetime import date
from typing import Optional, Dict, Any, List

from sqlalchemy import select, func
from sqlalchemy.orm import selectinload

from lib.db import async_get_session
from models.Project import Project
from models.Deal import Deal
from models.Lead import Lead
from models.LeadProject import LeadProject
from schemas.project import ProjectCreate, ProjectUpdate

__all__ = ["ProjectCreate", "ProjectUpdate"]


async def find_all_projects(tenant_id: Optional[int]) -> List[Project]:
    """Find all projects for tenant."""
    async with async_get_session() as session:
        stmt = select(Project).order_by(Project.name)
        if tenant_id:
            stmt = stmt.where(Project.tenant_id == tenant_id)
        result = await session.execute(stmt)
        return list(result.scalars().all())


async def find_project_by_id(project_id: int, tenant_id: Optional[int]) -> Optional[Project]:
    """Find project by ID."""
    async with async_get_session() as session:
        stmt = select(Project).where(Project.id == project_id)
        if tenant_id:
            stmt = stmt.where(Project.tenant_id == tenant_id)
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def count_project_deals(project_id: int) -> int:
    """Count deals for a project."""
    async with async_get_session() as session:
        result = await session.execute(
            select(func.count(Deal.id)).where(Deal.project_id == project_id)
        )
        return result.scalar() or 0


async def count_project_leads(project_id: int) -> int:
    """Count leads for a project."""
    async with async_get_session() as session:
        result = await session.execute(
            select(func.count(LeadProject.lead_id)).where(LeadProject.project_id == project_id)
        )
        return result.scalar() or 0


async def create_project(
    tenant_id: Optional[int],
    name: str,
    description: Optional[str],
    status: Optional[str],
    current_status_id: Optional[int],
    start_date: Optional[date],
    end_date: Optional[date],
    user_id: Optional[int],
) -> Project:
    """Create a new project."""
    async with async_get_session() as session:
        project = Project(
            tenant_id=tenant_id,
            name=name,
            description=description,
            status=status,
            current_status_id=current_status_id,
            start_date=start_date,
            end_date=end_date,
            created_by=user_id,
            updated_by=user_id,
        )
        session.add(project)
        await session.commit()
        await session.refresh(project)
        return project


async def update_project(
    project_id: int,
    tenant_id: Optional[int],
    data: Dict[str, Any],
    user_id: Optional[int],
) -> Optional[Project]:
    """Update a project. Returns None if not found."""
    async with async_get_session() as session:
        stmt = select(Project).where(Project.id == project_id)
        if tenant_id:
            stmt = stmt.where(Project.tenant_id == tenant_id)
        result = await session.execute(stmt)
        project = result.scalar_one_or_none()

        if not project:
            return None

        if data.get("name") is not None:
            project.name = data["name"]
        if data.get("description") is not None:
            project.description = data["description"]
        if data.get("status") is not None:
            project.status = data["status"]
        if data.get("current_status_id") is not None:
            project.current_status_id = data["current_status_id"]
        if data.get("start_date") is not None:
            project.start_date = data["start_date"]
        if data.get("end_date") is not None:
            project.end_date = data["end_date"]
        project.updated_by = user_id

        await session.commit()
        await session.refresh(project)
        return project


async def delete_project(project_id: int, tenant_id: Optional[int]) -> bool:
    """Delete a project. Returns False if not found."""
    async with async_get_session() as session:
        stmt = select(Project).where(Project.id == project_id)
        if tenant_id:
            stmt = stmt.where(Project.tenant_id == tenant_id)
        result = await session.execute(stmt)
        project = result.scalar_one_or_none()

        if not project:
            return False

        await session.delete(project)
        await session.commit()
        return True


async def find_project_leads(project_id: int, tenant_id: Optional[int]) -> List[Lead]:
    """Find leads for a project."""
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


async def find_project_deals(project_id: int, tenant_id: Optional[int]) -> List[Deal]:
    """Find deals for a project."""
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
