"""Controller layer for user routing to services."""

from typing import Optional

from fastapi import Request

from lib.decorators import handle_http_exceptions
from serializers.user import tenant_to_response
from services.user_service import (
    get_user_tenant as get_user_tenant_service,
    set_selected_project as set_selected_project_service,
    get_tenant_users as get_tenant_users_service,
)


@handle_http_exceptions
async def get_user_tenant(email: str) -> dict:
    """Get tenant info for a user."""
    result = await get_user_tenant_service(email)
    return tenant_to_response(result["tenant"], result["individual"])


@handle_http_exceptions
async def set_selected_project(request: Request, project_id: Optional[int]) -> dict:
    """Set user's selected project."""
    user_id = request.state.user_id
    tenant_id = request.state.tenant_id
    selected_id = await set_selected_project_service(user_id, tenant_id, project_id)
    return {"selected_project_id": selected_id}


@handle_http_exceptions
async def get_tenant_users(request: Request) -> list:
    """Get all users for the current tenant."""
    tenant_id = request.state.tenant_id
    return await get_tenant_users_service(tenant_id)
