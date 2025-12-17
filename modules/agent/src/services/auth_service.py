"""Authentication service for user and tenant management."""

import re
from typing import TYPE_CHECKING, Dict, Any, List

if TYPE_CHECKING:
    from models.User import User

from repositories.user_repository import (
    find_user_by_auth0_sub,
    find_user_by_email,
    create_user,
    update_user,
    get_user_global_roles as get_user_global_roles_repo,
)
from repositories.tenant_repository import (
    find_or_create_tenant_with_user,
    get_user_tenants_with_roles,
    find_role_by_tenant_and_name,
    create_user_role,
    TenantCreate,
)

# Re-export Pydantic models for controllers
__all__ = ["TenantCreate"]


async def get_or_create_user(auth0_sub: str, email: str) -> "User":
    """Get or create user from Auth0 sub."""
    user = await find_user_by_auth0_sub(auth0_sub)
    if user:
        return user

    # Check if user exists by email (e.g., seeded user without auth0_sub)
    user = await find_user_by_email(email)
    if user:
        # Update with auth0_sub
        return await update_user(user.id, {"auth0_sub": auth0_sub})

    # First login - create user without tenant
    data: Dict[str, Any] = {"auth0_sub": auth0_sub, "email": email, "status": "active"}
    return await create_user(data)


async def get_user_tenants(user_id: int) -> List[dict]:
    """Get all tenants user belongs to via UserRole relationships."""
    results = await get_user_tenants_with_roles(user_id)
    return [
        {"id": tenant.id, "name": tenant.name, "role": role.name}
        for tenant, role in results
    ]


async def create_tenant_for_user(
    user_id: int, tenant_name: str, slug: str = None, plan: str = "free"
) -> dict:
    """Find or create tenant and assign user as owner."""
    if slug is None:
        slug = re.sub(r'[^a-z0-9]+', '-', tenant_name.lower()).strip('-')

    tenant, role = await find_or_create_tenant_with_user(user_id, tenant_name, slug, plan)

    return {
        "id": tenant.id,
        "name": tenant.name,
        "slug": tenant.slug,
        "plan": tenant.plan,
        "role": role.name,
    }


async def add_user_to_tenant(user_id: int, tenant_id: int, role_name: str) -> None:
    """Add user to existing tenant with role."""
    # Find role by name in tenant
    role = await find_role_by_tenant_and_name(tenant_id, role_name)
    if not role:
        raise ValueError(f"Role {role_name} not found in tenant")

    # Create UserRole assignment
    await create_user_role(user_id, role.id)


async def get_user_global_roles(user_id: int) -> List[str]:
    """Get global role names for a user (roles with NULL tenant_id)."""
    return await get_user_global_roles_repo(user_id)
