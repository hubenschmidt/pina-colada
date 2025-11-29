"""Routes for projects API endpoints."""

from datetime import date
from typing import Optional
from fastapi import APIRouter, Request, Query, HTTPException
from pydantic import BaseModel
from sqlalchemy import select, func
from sqlalchemy.orm import selectinload
from lib.auth import require_auth
from lib.error_logging import log_errors
from lib.db import async_get_session
from models.Project import Project
from models.Deal import Deal
from models.Lead import Lead
from models.LeadProject import LeadProject


def _parse_date(date_str: Optional[str]) -> Optional[date]:
    """Parse date string to date object."""
    if not date_str:
        return None
    return date.fromisoformat(date_str[:10])


router = APIRouter(prefix="/projects", tags=["projects"])


class ProjectCreate(BaseModel):
    name: str
    description: Optional[str] = None
    status: Optional[str] = None
    current_status_id: Optional[int] = None
    start_date: Optional[str] = None
    end_date: Optional[str] = None


class ProjectUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    status: Optional[str] = None
    current_status_id: Optional[int] = None
    start_date: Optional[str] = None
    end_date: Optional[str] = None


def _project_to_dict(project, deals_count: int = 0, leads_count: int = 0):
    """Convert project to dict."""
    return {
        "id": project.id,
        "name": project.name,
        "description": project.description,
        "status": project.status,
        "current_status_id": project.current_status_id,
        "start_date": str(project.start_date) if project.start_date else None,
        "end_date": str(project.end_date) if project.end_date else None,
        "created_at": project.created_at.isoformat() if project.created_at else None,
        "updated_at": project.updated_at.isoformat() if project.updated_at else None,
        "deals_count": deals_count,
        "leads_count": leads_count,
    }


@router.get("")
@log_errors
@require_auth
async def get_projects(request: Request):
    """Get all projects for tenant."""
    tenant_id = getattr(request.state, "tenant_id", None)

    async with async_get_session() as session:
        stmt = select(Project).order_by(Project.name)

        if tenant_id:
            stmt = stmt.where(Project.tenant_id == tenant_id)

        result = await session.execute(stmt)
        projects = list(result.scalars().all())

        return [_project_to_dict(p) for p in projects]


@router.get("/{project_id}")
@log_errors
@require_auth
async def get_project(request: Request, project_id: int):
    """Get project by ID with related counts."""
    tenant_id = getattr(request.state, "tenant_id", None)

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

        return _project_to_dict(project, deals_count, leads_count)


@router.post("")
@log_errors
@require_auth
async def create_project(request: Request, body: ProjectCreate):
    """Create a new project."""
    tenant_id = getattr(request.state, "tenant_id", None)

    async with async_get_session() as session:
        project = Project(
            tenant_id=tenant_id,
            name=body.name,
            description=body.description,
            status=body.status,
            current_status_id=body.current_status_id,
            start_date=_parse_date(body.start_date),
            end_date=_parse_date(body.end_date),
        )

        session.add(project)
        await session.commit()
        await session.refresh(project)

        return _project_to_dict(project)


@router.put("/{project_id}")
@log_errors
@require_auth
async def update_project(request: Request, project_id: int, body: ProjectUpdate):
    """Update a project."""
    tenant_id = getattr(request.state, "tenant_id", None)

    async with async_get_session() as session:
        stmt = select(Project).where(Project.id == project_id)

        if tenant_id:
            stmt = stmt.where(Project.tenant_id == tenant_id)

        result = await session.execute(stmt)
        project = result.scalar_one_or_none()

        if not project:
            raise HTTPException(status_code=404, detail="Project not found")

        if body.name is not None:
            project.name = body.name
        if body.description is not None:
            project.description = body.description
        if body.status is not None:
            project.status = body.status
        if body.current_status_id is not None:
            project.current_status_id = body.current_status_id
        if body.start_date is not None:
            project.start_date = _parse_date(body.start_date)
        if body.end_date is not None:
            project.end_date = _parse_date(body.end_date)

        await session.commit()
        await session.refresh(project)

        return _project_to_dict(project)


@router.delete("/{project_id}")
@log_errors
@require_auth
async def delete_project(request: Request, project_id: int):
    """Delete a project."""
    tenant_id = getattr(request.state, "tenant_id", None)

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

        return {"success": True}


@router.get("/{project_id}/leads")
@log_errors
@require_auth
async def get_project_leads(request: Request, project_id: int):
    """Get leads for a project."""
    tenant_id = getattr(request.state, "tenant_id", None)

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
        leads = list(result.scalars().all())

        return [
            {
                "id": lead.id,
                "title": lead.title,
                "type": lead.type,
                "description": lead.description,
                "source": lead.source,
                "current_status": lead.current_status.name if lead.current_status else None,
                "account_name": lead.account.name if lead.account else None,
                "created_at": lead.created_at.isoformat() if lead.created_at else None,
            }
            for lead in leads
        ]


@router.get("/{project_id}/deals")
@log_errors
@require_auth
async def get_project_deals(request: Request, project_id: int):
    """Get deals for a project."""
    tenant_id = getattr(request.state, "tenant_id", None)

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
        deals = list(result.scalars().all())

        return [
            {
                "id": deal.id,
                "name": deal.name,
                "description": deal.description,
                "current_status": deal.current_status.name if deal.current_status else None,
                "value_amount": float(deal.value_amount) if deal.value_amount else None,
                "value_currency": deal.value_currency,
                "created_at": deal.created_at.isoformat() if deal.created_at else None,
            }
            for deal in deals
        ]
