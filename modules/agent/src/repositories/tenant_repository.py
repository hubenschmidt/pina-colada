"""Repository layer for tenant data access."""

from typing import Optional, List, Tuple
from sqlalchemy import select
from models.Tenant import Tenant
from models.Role import Role
from models.UserRole import UserRole
from models.User import User
from lib.db import get_session


def find_tenant_by_slug(slug: str) -> Optional[Tenant]:
    """Find tenant by slug."""
    session = get_session()
    try:
        stmt = select(Tenant).where(Tenant.slug == slug)
        return session.execute(stmt).scalar_one_or_none()
    finally:
        session.close()


def create_tenant(name: str, slug: str, plan: str = "free") -> Tenant:
    """Create a new tenant."""
    session = get_session()
    try:
        tenant = Tenant(name=name, slug=slug, plan=plan)
        session.add(tenant)
        session.commit()
        session.refresh(tenant)
        return tenant
    finally:
        session.close()


def find_role_by_tenant_and_name(tenant_id: int, role_name: str) -> Optional[Role]:
    """Find role by tenant_id and name."""
    session = get_session()
    try:
        stmt = select(Role).where(Role.tenant_id == tenant_id, Role.name == role_name)
        return session.execute(stmt).scalar_one_or_none()
    finally:
        session.close()


def create_role(tenant_id: int, name: str, description: str = "") -> Role:
    """Create a new role."""
    session = get_session()
    try:
        role = Role(tenant_id=tenant_id, name=name, description=description)
        session.add(role)
        session.commit()
        session.refresh(role)
        return role
    finally:
        session.close()


def find_user_role_in_tenant(user_id: int, tenant_id: int) -> Optional[Tuple[Role, UserRole]]:
    """Find user's role in a specific tenant."""
    session = get_session()
    try:
        stmt = (
            select(Role, UserRole)
            .join(UserRole, UserRole.role_id == Role.id)
            .where(Role.tenant_id == tenant_id, UserRole.user_id == user_id)
        )
        return session.execute(stmt).first()
    finally:
        session.close()


def get_user_tenants_with_roles(user_id: int) -> List[Tuple[Tenant, Role]]:
    """Get all tenants user belongs to with their roles."""
    session = get_session()
    try:
        stmt = (
            select(Tenant, Role)
            .join(Role, Role.tenant_id == Tenant.id)
            .join(UserRole, UserRole.role_id == Role.id)
            .where(UserRole.user_id == user_id)
        )
        return session.execute(stmt).all()
    finally:
        session.close()


def create_user_role(user_id: int, role_id: int) -> UserRole:
    """Create a user role assignment."""
    session = get_session()
    try:
        user_role = UserRole(user_id=user_id, role_id=role_id)
        session.add(user_role)
        session.commit()
        session.refresh(user_role)
        return user_role
    finally:
        session.close()


def find_or_create_tenant_with_user(
    user_id: int, tenant_name: str, slug: str, plan: str = "free"
) -> Tuple[Tenant, Role]:
    """
    Find or create tenant and assign user as owner.
    Returns (tenant, role) tuple.
    This is a complex transaction that handles multiple operations atomically.
    """
    session = get_session()
    try:
        stmt = select(Tenant).where(Tenant.slug == slug)
        tenant = session.execute(stmt).scalar_one_or_none()

        # Tenant doesn't exist - create new one
        if not tenant:
            return _create_new_tenant_with_user(session, user_id, tenant_name, slug, plan)

        # Check if user already has role in tenant
        stmt = (
            select(Role, UserRole)
            .join(UserRole, UserRole.role_id == Role.id)
            .where(Role.tenant_id == tenant.id, UserRole.user_id == user_id)
        )
        result = session.execute(stmt).first()

        # User already has role - return existing
        if result:
            role, _ = result
            return (tenant, role)

        # User doesn't have role - assign as owner
        return _assign_user_as_owner_to_tenant(session, user_id, tenant)
    finally:
        session.close()


def _create_new_tenant_with_user(
    session, user_id: int, tenant_name: str, slug: str, plan: str
) -> Tuple[Tenant, Role]:
    """Create new tenant and assign user as owner."""
    tenant = Tenant(name=tenant_name, slug=slug, plan=plan)
    session.add(tenant)
    session.flush()

    owner_role = Role(
        tenant_id=tenant.id,
        name="owner",
        description="Full access to all resources",
    )
    session.add(owner_role)
    session.flush()

    user_role = UserRole(user_id=user_id, role_id=owner_role.id)
    session.add(user_role)

    user = session.get(User, user_id)
    if not user:
        session.commit()
        session.refresh(tenant)
        session.refresh(owner_role)
        return (tenant, owner_role)

    user.tenant_id = tenant.id
    session.commit()
    session.refresh(tenant)
    session.refresh(owner_role)

    return (tenant, owner_role)


def _assign_user_as_owner_to_tenant(
    session, user_id: int, tenant: Tenant
) -> Tuple[Tenant, Role]:
    """Assign user as owner to existing tenant."""
    owner_role = _get_or_create_owner_role(session, tenant.id)

    user_role = UserRole(user_id=user_id, role_id=owner_role.id)
    session.add(user_role)

    user = session.get(User, user_id)
    if user and not user.tenant_id:
        user.tenant_id = tenant.id

    session.commit()
    return (tenant, owner_role)


def _get_or_create_owner_role(session, tenant_id: int) -> Role:
    """Get existing owner role or create new one."""
    stmt = select(Role).where(Role.tenant_id == tenant_id, Role.name == "owner")
    owner_role = session.execute(stmt).scalar_one_or_none()

    if owner_role:
        return owner_role

    owner_role = Role(
        tenant_id=tenant_id,
        name="owner",
        description="Full access to all resources",
    )
    session.add(owner_role)
    session.flush()
    return owner_role
