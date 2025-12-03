"""Controller layer for project routing to services."""

import logging
from typing import List

from fastapi import Request

from lib.decorators import handle_http_exceptions
from repositories.project_repository import ProjectCreate, ProjectUpdate
from services.project_service import (
    get_projects as get_projects_service,
    get_project as get_project_service,
    create_project as create_project_service,
    update_project as update_project_service,
    delete_project as delete_project_service,
    get_project_leads as get_leads_service,
    get_project_deals as get_deals_service,
)

logger = logging.getLogger(__name__)

# Re-export for routes
__all__ = ["ProjectCreate", "ProjectUpdate"]


def _project_to_response(project, deals_count: int = 0, leads_count: int = 0) -> dict:
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


def _lead_to_response(lead) -> dict:
    """Convert lead to response dict."""
    return {
        "id": lead.id,
        "title": lead.title,
        "type": lead.type,
        "description": lead.description,
        "source": lead.source,
        "current_status": lead.current_status.name if lead.current_status else None,
        "account_name": lead.account.name if lead.account else None,
        "created_at": lead.created_at.isoformat() if lead.created_at else None,
    }


def _deal_to_response(deal) -> dict:
    """Convert deal to response dict."""
    return {
        "id": deal.id,
        "name": deal.name,
        "description": deal.description,
        "current_status": deal.current_status.name if deal.current_status else None,
        "value_amount": float(deal.value_amount) if deal.value_amount else None,
        "value_currency": deal.value_currency,
        "created_at": deal.created_at.isoformat() if deal.created_at else None,
    }


@handle_http_exceptions
async def get_projects(request: Request) -> List[dict]:
    """Get all projects for tenant."""
    tenant_id = request.state.tenant_id
    projects = await get_projects_service(tenant_id)
    return [_project_to_response(p) for p in projects]


@handle_http_exceptions
async def get_project(request: Request, project_id: int) -> dict:
    """Get project by ID with related counts."""
    tenant_id = request.state.tenant_id
    project, deals_count, leads_count = await get_project_service(project_id, tenant_id)
    return _project_to_response(project, deals_count, leads_count)


@handle_http_exceptions
async def create_project(request: Request, data: ProjectCreate) -> dict:
    """Create a new project."""
    tenant_id = request.state.tenant_id
    user_id = request.state.user_id
    project = await create_project_service(data.model_dump(), tenant_id, user_id)
    return _project_to_response(project)


@handle_http_exceptions
async def update_project(request: Request, project_id: int, data: ProjectUpdate) -> dict:
    """Update a project."""
    tenant_id = request.state.tenant_id
    user_id = request.state.user_id
    project = await update_project_service(project_id, data.model_dump(exclude_unset=True), tenant_id, user_id)
    return _project_to_response(project)


@handle_http_exceptions
async def delete_project(request: Request, project_id: int) -> dict:
    """Delete a project."""
    tenant_id = request.state.tenant_id
    await delete_project_service(project_id, tenant_id)
    return {"success": True}


@handle_http_exceptions
async def get_project_leads(request: Request, project_id: int) -> List[dict]:
    """Get leads for a project."""
    tenant_id = request.state.tenant_id
    leads = await get_leads_service(project_id, tenant_id)
    return [_lead_to_response(lead) for lead in leads]


@handle_http_exceptions
async def get_project_deals(request: Request, project_id: int) -> List[dict]:
    """Get deals for a project."""
    tenant_id = request.state.tenant_id
    deals = await get_deals_service(project_id, tenant_id)
    return [_deal_to_response(deal) for deal in deals]
