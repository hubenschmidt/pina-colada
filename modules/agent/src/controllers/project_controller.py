"""Controller layer for project routing to services."""

import logging
from typing import List

from fastapi import Request

from lib.decorators import handle_http_exceptions
from serializers.project import project_to_response, lead_to_response, deal_to_response
from schemas.project import ProjectCreate, ProjectUpdate
from services.project_service import (
    get_projects as get_projects_service,
    get_project as get_project_service,
    create_project as create_project_service,
    update_project as update_project_service,
    delete_project as delete_project_service,
    get_project_leads as get_leads_service,
    get_project_deals as get_deals_service,
)

# Re-export for routes
__all__ = ["ProjectCreate", "ProjectUpdate"]


@handle_http_exceptions
async def get_projects(request: Request) -> List[dict]:
    """Get all projects for tenant."""
    tenant_id = request.state.tenant_id
    projects = await get_projects_service(tenant_id)
    return [project_to_response(p) for p in projects]


@handle_http_exceptions
async def get_project(request: Request, project_id: int) -> dict:
    """Get project by ID with related counts."""
    tenant_id = request.state.tenant_id
    project, deals_count, leads_count = await get_project_service(project_id, tenant_id)
    return project_to_response(project, deals_count, leads_count)


@handle_http_exceptions
async def create_project(request: Request, data: ProjectCreate) -> dict:
    """Create a new project."""
    tenant_id = request.state.tenant_id
    user_id = request.state.user_id
    project = await create_project_service(data.model_dump(), tenant_id, user_id)
    return project_to_response(project)


@handle_http_exceptions
async def update_project(request: Request, project_id: int, data: ProjectUpdate) -> dict:
    """Update a project."""
    tenant_id = request.state.tenant_id
    user_id = request.state.user_id
    project = await update_project_service(project_id, data.model_dump(exclude_unset=True), tenant_id, user_id)
    return project_to_response(project)


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
    return [lead_to_response(lead) for lead in leads]


@handle_http_exceptions
async def get_project_deals(request: Request, project_id: int) -> List[dict]:
    """Get deals for a project."""
    tenant_id = request.state.tenant_id
    deals = await get_deals_service(project_id, tenant_id)
    return [deal_to_response(deal) for deal in deals]
