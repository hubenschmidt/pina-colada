"""Routes for user API endpoints."""

from fastapi import APIRouter, Request

from controllers.user_controller import (
    get_user_tenant,
    set_selected_project,
    get_tenant_users,
)
from schemas.user import SetSelectedProjectRequest


router = APIRouter(prefix="/users", tags=["user"])


@router.get("/")
async def get_tenant_users_route(request: Request):
    """Get all users for the current tenant."""
    return await get_tenant_users(request)


@router.get("/{email}/tenant")
async def get_user_tenant_route(email: str):
    """Get tenant info for a user by email."""
    return await get_user_tenant(email)


@router.put("/me/selected-project")
async def set_selected_project_route(request: Request, body: SetSelectedProjectRequest):
    """Set the current user's selected project."""
    return await set_selected_project(request, body.project_id)
