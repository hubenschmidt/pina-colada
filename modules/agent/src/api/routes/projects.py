"""Routes for projects API endpoints."""

from fastapi import APIRouter, Request

from controllers.project_controller import (
    create_project,
    delete_project,
    get_project,
    get_project_deals,
    get_project_leads,
    get_projects,
    update_project,
)
from schemas.project import ProjectCreate, ProjectUpdate
from lib.auth import require_auth
from lib.error_logging import log_errors


router = APIRouter(prefix="/projects", tags=["projects"])


@router.get("")
@log_errors
@require_auth
async def get_projects_route(request: Request):
    """Get all projects for tenant."""
    return await get_projects(request)


@router.get("/{project_id}")
@log_errors
@require_auth
async def get_project_route(request: Request, project_id: int):
    """Get project by ID with related counts."""
    return await get_project(request, project_id)


@router.post("")
@log_errors
@require_auth
async def create_project_route(request: Request, body: ProjectCreate):
    """Create a new project."""
    return await create_project(request, body)


@router.put("/{project_id}")
@log_errors
@require_auth
async def update_project_route(request: Request, project_id: int, body: ProjectUpdate):
    """Update a project."""
    return await update_project(request, project_id, body)


@router.delete("/{project_id}")
@log_errors
@require_auth
async def delete_project_route(request: Request, project_id: int):
    """Delete a project."""
    return await delete_project(request, project_id)


@router.get("/{project_id}/leads")
@log_errors
@require_auth
async def get_project_leads_route(request: Request, project_id: int):
    """Get leads for a project."""
    return await get_project_leads(request, project_id)


@router.get("/{project_id}/deals")
@log_errors
@require_auth
async def get_project_deals_route(request: Request, project_id: int):
    """Get deals for a project."""
    return await get_project_deals(request, project_id)
