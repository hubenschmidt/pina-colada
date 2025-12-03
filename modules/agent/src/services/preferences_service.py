"""Service layer for user and tenant preferences."""

import logging
from typing import Optional

from fastapi import HTTPException
from sqlalchemy import select
from sqlalchemy.orm import selectinload

from lib.db import async_get_session
from lib.date_utils import is_valid_timezone
from models.User import User
from models.UserPreferences import UserPreferences
from models.TenantPreferences import TenantPreferences
from models.Tenant import Tenant
from models.UserRole import UserRole
from repositories.preferences_repository import (
    UpdateUserPreferencesRequest,
    UpdateTenantPreferencesRequest,
)

# Re-export Pydantic models for controllers
__all__ = ["UpdateUserPreferencesRequest", "UpdateTenantPreferencesRequest"]

logger = logging.getLogger(__name__)


def resolve_theme(user: User) -> str:
    """Resolve effective theme using hierarchy: user > tenant > system default."""
    if user.preferences and user.preferences.theme:
        return user.preferences.theme
    if user.tenant and user.tenant.preferences and user.tenant.preferences.theme:
        return user.tenant.preferences.theme
    return "light"


def user_has_admin_role(user: User) -> bool:
    """Check if user has Admin or SuperAdmin role."""
    if not user.user_roles:
        return False
    role_names = [ur.role.name for ur in user.user_roles]
    return "Admin" in role_names or "SuperAdmin" in role_names


async def get_user_preferences(user_id: int):
    """Get current user's preferences with resolved theme."""
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
        user = result.scalar_one_or_none()

        if not user:
            raise HTTPException(status_code=404, detail="User not found")

        if not user.preferences:
            user_prefs = UserPreferences(user_id=user.id, theme=None)
            session.add(user_prefs)
            await session.commit()
            await session.refresh(user, ["preferences"])

        effective_theme = resolve_theme(user)
        can_edit_tenant = user_has_admin_role(user)
        user_theme = user.preferences.theme
        user_timezone = user.preferences.timezone or "America/New_York"

    return user_theme, user_timezone, effective_theme, can_edit_tenant


async def update_user_preferences(
    user_id: int,
    theme: Optional[str] = None,
    timezone: Optional[str] = None,
):
    """Update current user's preferences."""
    if theme is not None and theme not in ["light", "dark"]:
        raise HTTPException(status_code=400, detail="Invalid theme value")

    if timezone is not None and not is_valid_timezone(timezone):
        raise HTTPException(status_code=400, detail="Invalid timezone value")

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
            raise HTTPException(status_code=404, detail="User not found")

        if not user.preferences:
            user.preferences = UserPreferences(user_id=user.id)
            session.add(user.preferences)

        if theme is not None:
            user.preferences.theme = theme
        if timezone is not None:
            user.preferences.timezone = timezone

        await session.commit()
        await session.refresh(user, ["preferences"])

        user_theme = user.preferences.theme
        user_timezone = user.preferences.timezone or "America/New_York"
        effective_theme = resolve_theme(user)

    return user_theme, user_timezone, effective_theme


async def get_tenant_preferences(user_id: int):
    """Get tenant preferences (Admin/SuperAdmin only)."""
    async with async_get_session() as session:
        result = await session.execute(
            select(User)
            .options(
                selectinload(User.user_roles).selectinload(UserRole.role),
                selectinload(User.tenant).selectinload(Tenant.preferences)
            )
            .filter(User.id == user_id)
        )
        user = result.scalar_one_or_none()

        if not user:
            raise HTTPException(status_code=404, detail="User not found")

        if not user_has_admin_role(user):
            raise HTTPException(status_code=403, detail="Insufficient permissions")

        if not user.tenant.preferences:
            tenant_prefs = TenantPreferences(tenant_id=user.tenant_id, theme="light")
            session.add(tenant_prefs)
            await session.commit()
            await session.refresh(user.tenant, ["preferences"])

        return user.tenant.preferences.theme


async def update_tenant_preferences(user_id: int, theme: str):
    """Update tenant theme preference (Admin/SuperAdmin only)."""
    if theme not in ["light", "dark"]:
        raise HTTPException(status_code=400, detail="Invalid theme value")

    async with async_get_session() as session:
        result = await session.execute(
            select(User)
            .options(
                selectinload(User.user_roles).selectinload(UserRole.role),
                selectinload(User.tenant).selectinload(Tenant.preferences)
            )
            .filter(User.id == user_id)
        )
        user = result.scalar_one_or_none()

        if not user:
            raise HTTPException(status_code=404, detail="User not found")

        if not user_has_admin_role(user):
            raise HTTPException(status_code=403, detail="Insufficient permissions")

        if not user.tenant.preferences:
            user.tenant.preferences = TenantPreferences(tenant_id=user.tenant_id)
            session.add(user.tenant.preferences)

        user.tenant.preferences.theme = theme
        await session.commit()

        return user.tenant.preferences.theme
