"""Repository layer for tenant data access."""

from typing import Optional, List, Tuple
from sqlalchemy import select
from models.Tenant import Tenant
from models.Role import Role
from models.UserRole import UserRole
from models.User import User
from models.Organization import Organization
from models.Account import Account
from models.Individual import Individual
from lib.db import async_get_session


async def find_tenant_by_slug(slug: str) -> Optional[Tenant]:
    """Find tenant by slug."""
    async with async_get_session() as session:
        stmt = select(Tenant).where(Tenant.slug == slug)
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def create_tenant(name: str, slug: str, plan: str = "free") -> Tenant:
    """Create a new tenant."""
    async with async_get_session() as session:
        tenant = Tenant(name=name, slug=slug, plan=plan)
        session.add(tenant)
        await session.commit()
        await session.refresh(tenant)
        return tenant


async def find_role_by_tenant_and_name(tenant_id: int, role_name: str) -> Optional[Role]:
    """Find role by tenant_id and name."""
    async with async_get_session() as session:
        stmt = select(Role).where(Role.tenant_id == tenant_id, Role.name == role_name)
        result = await session.execute(stmt)
        return result.scalar_one_or_none()


async def create_role(tenant_id: int, name: str, description: str = "") -> Role:
    """Create a new role."""
    async with async_get_session() as session:
        role = Role(tenant_id=tenant_id, name=name, description=description)
        session.add(role)
        await session.commit()
        await session.refresh(role)
        return role


async def find_user_role_in_tenant(user_id: int, tenant_id: int) -> Optional[Tuple[Role, UserRole]]:
    """Find user's role in a specific tenant."""
    async with async_get_session() as session:
        stmt = (
            select(Role, UserRole)
            .join(UserRole, UserRole.role_id == Role.id)
            .where(Role.tenant_id == tenant_id, UserRole.user_id == user_id)
        )
        result = await session.execute(stmt)
        return result.first()


async def get_user_tenants_with_roles(user_id: int) -> List[Tuple[Tenant, Role]]:
    """Get all tenants user belongs to with their roles."""
    async with async_get_session() as session:
        stmt = (
            select(Tenant, Role)
            .join(Role, Role.tenant_id == Tenant.id)
            .join(UserRole, UserRole.role_id == Role.id)
            .where(UserRole.user_id == user_id)
        )
        result = await session.execute(stmt)
        return result.all()


async def create_user_role(user_id: int, role_id: int) -> UserRole:
    """Create a user role assignment."""
    async with async_get_session() as session:
        user_role = UserRole(user_id=user_id, role_id=role_id)
        session.add(user_role)
        await session.commit()
        await session.refresh(user_role)
        return user_role


async def find_or_create_tenant_with_user(
    user_id: int, tenant_name: str, slug: str, plan: str = "free"
) -> Tuple[Tenant, Role]:
    """
    Find or create tenant and assign user as owner.
    Returns (tenant, role) tuple.
    This is a complex transaction that handles multiple operations atomically.
    """
    async with async_get_session() as session:
        stmt = select(Tenant).where(Tenant.slug == slug)
        result = await session.execute(stmt)
        tenant = result.scalar_one_or_none()

        # Tenant doesn't exist - create new one
        if not tenant:
            return await _create_new_tenant_with_user(session, user_id, tenant_name, slug, plan)

        # Check if user already has role in tenant
        stmt = (
            select(Role, UserRole)
            .join(UserRole, UserRole.role_id == Role.id)
            .where(Role.tenant_id == tenant.id, UserRole.user_id == user_id)
        )
        result = await session.execute(stmt)
        role_result = result.first()

        # User already has role - return existing
        if role_result:
            role, _ = role_result
            return (tenant, role)

        # User doesn't have role - assign as owner
        return await _assign_user_as_owner_to_tenant(session, user_id, tenant)


async def _create_new_tenant_with_user(
    session, user_id: int, tenant_name: str, slug: str, plan: str
) -> Tuple[Tenant, Role]:
    """Create new tenant with default Organization and assign user as owner."""
    tenant = Tenant(name=tenant_name, slug=slug, plan=plan)
    session.add(tenant)
    await session.flush()

    # Create Account for the default Organization (with tenant_id)
    org_account = Account(name=tenant_name, tenant_id=tenant.id)
    session.add(org_account)
    await session.flush()

    # Create default Organization for the Tenant
    organization = Organization(
        account_id=org_account.id,
        name=tenant_name,
    )
    session.add(organization)
    await session.flush()

    owner_role = Role(
        tenant_id=tenant.id,
        name="owner",
        description="Full access to all resources",
    )
    session.add(owner_role)
    await session.flush()

    user_role = UserRole(user_id=user_id, role_id=owner_role.id)
    session.add(user_role)

    user = await session.get(User, user_id)
    if not user:
        await session.commit()
        await session.refresh(tenant)
        await session.refresh(owner_role)
        return (tenant, owner_role)

    user.tenant_id = tenant.id

    # Also update the user's Individual's Account to belong to this tenant
    if not user.individual_id:
        await session.commit()
        await session.refresh(tenant)
        await session.refresh(owner_role)
        return (tenant, owner_role)

    individual = await session.get(Individual, user.individual_id)
    if individual and individual.account_id:
        ind_account = await session.get(Account, individual.account_id)
        if ind_account:
            ind_account.tenant_id = tenant.id

    await session.commit()
    await session.refresh(tenant)
    await session.refresh(owner_role)

    return (tenant, owner_role)


async def _assign_user_as_owner_to_tenant(
    session, user_id: int, tenant: Tenant
) -> Tuple[Tenant, Role]:
    """Assign user as owner to existing tenant."""
    owner_role = await _get_or_create_owner_role(session, tenant.id)

    user_role = UserRole(user_id=user_id, role_id=owner_role.id)
    session.add(user_role)

    user = await session.get(User, user_id)
    if user and not user.tenant_id:
        user.tenant_id = tenant.id

    await session.commit()
    return (tenant, owner_role)


async def _get_or_create_owner_role(session, tenant_id: int) -> Role:
    """Get existing owner role or create new one."""
    stmt = select(Role).where(Role.tenant_id == tenant_id, Role.name == "owner")
    result = await session.execute(stmt)
    owner_role = result.scalar_one_or_none()

    if owner_role:
        return owner_role

    owner_role = Role(
        tenant_id=tenant_id,
        name="owner",
        description="Full access to all resources",
    )
    session.add(owner_role)
    await session.flush()
    return owner_role
