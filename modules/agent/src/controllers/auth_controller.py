"""Controller layer for auth routing to services."""

from typing import Dict, Any
from lib.decorators import handle_http_exceptions
from lib.serialization import model_to_dict
from services.auth_service import (
    get_user_tenants as get_user_tenants_service,
    create_tenant_for_user as create_tenant_service,
)


@handle_http_exceptions
async def get_current_user_with_tenants(user, tenant_id: int) -> Dict[str, Any]:
    """Get current user profile with tenant associations."""
    tenants = await get_user_tenants_service(user.id)

    return {
        "user": model_to_dict(user, include_relationships=False),
        "tenants": tenants,
        "current_tenant_id": tenant_id
    }


@handle_http_exceptions
async def create_tenant(user_id: int, name: str, slug: str = None, plan: str = "free") -> Dict[str, Any]:
    """Create new tenant and assign current user as owner."""
    tenant = await create_tenant_service(user_id, name, slug, plan)
    return {"tenant": tenant}
