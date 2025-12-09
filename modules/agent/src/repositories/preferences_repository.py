"""Repository layer for preferences data access."""

from typing import Optional

from sqlalchemy import select
from sqlalchemy.orm import selectinload

from lib.db import async_get_session
from models.User import User
from models.UserPreferences import UserPreferences
from models.TenantPreferences import TenantPreferences
from models.Tenant import Tenant
from models.UserRole import UserRole
from schemas.preferences import UpdateTenantPreferencesRequest, UpdateUserPreferencesRequest

__all__ = ["UpdateUserPreferencesRequest", "UpdateTenantPreferencesRequest"]


async def find_user_with_preferences(user_id: int) -> Optional[User]:
    """Find user with preferences, tenant preferences, and roles loaded."""
    async with async_get_session() as session:
        result = await session.execute(
            select(User)
            .options(
                selectinload(User.preferences),
                selectinload(User.tenant).selectinload(Tenant.preferences),
                selectinload(User.user_roles).selectinload(UserRole.role)
            )
            .filter(User.id == user_id)
        )
        return result.scalar_one_or_none()


async def ensure_user_preferences(user_id: int) -> UserPreferences:
    """Ensure user preferences exist, create if needed."""
    async with async_get_session() as session:
        user = await session.get(User, user_id)
        if not user:
            return None

        result = await session.execute(
            select(UserPreferences).where(UserPreferences.user_id == user_id)
        )
        prefs = result.scalar_one_or_none()

        if not prefs:
            prefs = UserPreferences(user_id=user_id, theme=None)
            session.add(prefs)
            await session.commit()
            await session.refresh(prefs)

        return prefs


async def update_user_preferences_data(
    user_id: int,
    theme: Optional[str] = None,
    timezone: Optional[str] = None,
) -> Optional[User]:
    """Update user preferences and return updated user with relationships."""
    async with async_get_session() as session:
        result = await session.execute(
            select(User)
            .options(
                selectinload(User.preferences),
                selectinload(User.tenant).selectinload(Tenant.preferences)
            )
            .filter(User.id == user_id)
        )
        user = result.scalar_one_or_none()

        if not user:
            return None

        if not user.preferences:
            user.preferences = UserPreferences(user_id=user.id)
            session.add(user.preferences)

        if theme is not None:
            user.preferences.theme = theme
        if timezone is not None:
            user.preferences.timezone = timezone

        await session.commit()
        await session.refresh(user, ["preferences"])
        return user


async def find_user_with_tenant_and_roles(user_id: int) -> Optional[User]:
    """Find user with tenant preferences and roles loaded."""
    async with async_get_session() as session:
        result = await session.execute(
            select(User)
            .options(
                selectinload(User.user_roles).selectinload(UserRole.role),
                selectinload(User.tenant).selectinload(Tenant.preferences)
            )
            .filter(User.id == user_id)
        )
        return result.scalar_one_or_none()


async def ensure_tenant_preferences(tenant_id: int) -> TenantPreferences:
    """Ensure tenant preferences exist, create if needed."""
    async with async_get_session() as session:
        result = await session.execute(
            select(TenantPreferences).where(TenantPreferences.tenant_id == tenant_id)
        )
        prefs = result.scalar_one_or_none()

        if not prefs:
            prefs = TenantPreferences(tenant_id=tenant_id, theme="light")
            session.add(prefs)
            await session.commit()
            await session.refresh(prefs)

        return prefs


async def update_tenant_preferences_data(tenant_id: int, theme: str) -> TenantPreferences:
    """Update tenant theme preference."""
    async with async_get_session() as session:
        result = await session.execute(
            select(TenantPreferences).where(TenantPreferences.tenant_id == tenant_id)
        )
        prefs = result.scalar_one_or_none()

        if not prefs:
            prefs = TenantPreferences(tenant_id=tenant_id)
            session.add(prefs)

        prefs.theme = theme
        await session.commit()
        await session.refresh(prefs)
        return prefs
