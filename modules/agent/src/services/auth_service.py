"""Authentication service for user and tenant management."""

import re
from typing import Dict, Any, List
from sqlalchemy import select
from lib.db import get_session
from models.User import User
from models.Tenant import Tenant
from models.Role import Role
from models.UserRole import UserRole
from repositories.user_repository import find_user_by_auth0_sub, create_user


async def get_or_create_user(auth0_sub: str, email: str) -> User:
    """Get or create user from Auth0 sub."""
    user = find_user_by_auth0_sub(auth0_sub)
    if user:
        return user

    # First login - create user without tenant
    data: Dict[str, Any] = {"auth0_sub": auth0_sub, "email": email, "status": "active"}
    return create_user(data)


def get_user_tenants(user_id: int) -> List[dict]:
    """Get all tenants user belongs to via UserRole relationships."""
    session = get_session()
    try:
        stmt = (
            select(Tenant, Role)
            .join(Role, Role.tenant_id == Tenant.id)
            .join(UserRole, UserRole.role_id == Role.id)
            .where(UserRole.user_id == user_id)
        )
        results = session.execute(stmt).all()

        tenants = []
        for tenant, role in results:
            tenants.append({"id": tenant.id, "name": tenant.name, "role": role.name})
        return tenants
    finally:
        session.close()


def create_tenant_for_user(
    user_id: int, tenant_name: str, slug: str = None, plan: str = "free"
) -> dict:
    """Create new tenant and assign user as owner."""
    session = get_session()
    try:
        # Generate slug from name if not provided
        if not slug:
            slug = re.sub(r'[^a-z0-9]+', '-', tenant_name.lower()).strip('-')

        # Create Tenant
        tenant = Tenant(name=tenant_name, slug=slug, plan=plan)
        session.add(tenant)
        session.flush()

        # Create owner Role
        owner_role = Role(
            tenant_id=tenant.id,
            name="owner",
            description="Full access to all resources",
        )
        session.add(owner_role)
        session.flush()

        # Create UserRole assignment
        user_role = UserRole(user_id=user_id, role_id=owner_role.id)
        session.add(user_role)

        # Update User.tenant_id to this tenant (primary tenant)
        user = session.get(User, user_id)
        if user:
            user.tenant_id = tenant.id

        session.commit()

        return {
            "id": tenant.id,
            "name": tenant.name,
            "slug": tenant.slug,
            "plan": tenant.plan,
            "role": "owner",
        }
    finally:
        session.close()


def add_user_to_tenant(user_id: int, tenant_id: int, role_name: str) -> None:
    """Add user to existing tenant with role."""
    session = get_session()
    try:
        # Find role by name in tenant
        stmt = select(Role).where(Role.tenant_id == tenant_id, Role.name == role_name)
        role = session.execute(stmt).scalar_one_or_none()
        if not role:
            raise ValueError(f"Role {role_name} not found in tenant")

        # Create UserRole assignment
        user_role = UserRole(user_id=user_id, role_id=role.id)
        session.add(user_role)
        session.commit()
    finally:
        session.close()
