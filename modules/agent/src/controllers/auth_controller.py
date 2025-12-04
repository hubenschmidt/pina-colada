"""Controller layer for auth routing to services."""

from typing import Dict, Any

from fastapi import Request

from lib.decorators import handle_http_exceptions
from serializers.base import model_to_dict
from services.auth_service import (
    TenantCreate,
    get_user_tenants as get_user_tenants_service,
    create_tenant_for_user as create_tenant_service,
)

# Re-export for routes
__all__ = ["TenantCreate"]


@handle_http_exceptions
async def get_current_user_with_tenants(request: Request) -> Dict[str, Any]:
    """Get current user profile with tenant associations."""
    user = request.state.user
    tenant_id = request.state.tenant_id
    tenants = await get_user_tenants_service(user.id)

    return {
        "user": model_to_dict(user, include_relationships=False),
        "tenants": tenants,
        "current_tenant_id": tenant_id
    }


@handle_http_exceptions
async def create_tenant(request: Request, data: TenantCreate) -> Dict[str, Any]:
    """Create new tenant and assign current user as owner."""
    user_id = request.state.user_id
    tenant = await create_tenant_service(user_id, data.name, data.slug, data.plan)
    return {"tenant": tenant}
