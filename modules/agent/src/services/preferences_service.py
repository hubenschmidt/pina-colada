"""Service layer for user and tenant preferences."""

import logging
from typing import Optional

from fastapi import HTTPException

from lib.date_utils import is_valid_timezone
from repositories.preferences_repository import (
    find_user_with_preferences,
    ensure_user_preferences,
    update_user_preferences_data,
    find_user_with_tenant_and_roles,
    ensure_tenant_preferences,
    update_tenant_preferences_data,
    UpdateUserPreferencesRequest,
    UpdateTenantPreferencesRequest,
)

# Re-export Pydantic models for controllers
__all__ = ["UpdateUserPreferencesRequest", "UpdateTenantPreferencesRequest"]

logger = logging.getLogger(__name__)


def _resolve_theme(user) -> str:
    """Resolve effective theme using hierarchy: user > tenant > system default."""
    if user.preferences and user.preferences.theme:
        return user.preferences.theme
    if user.tenant and user.tenant.preferences and user.tenant.preferences.theme:
        return user.tenant.preferences.theme
    return "light"


def _user_has_admin_role(user) -> bool:
    """Check if user has Admin or SuperAdmin role."""
    if not user.user_roles:
        return False
    role_names = [ur.role.name for ur in user.user_roles]
    return "Admin" in role_names or "SuperAdmin" in role_names


async def get_user_preferences(user_id: int):
    """Get current user's preferences with resolved theme."""
    user = await find_user_with_preferences(user_id)

    if not user:
        raise HTTPException(status_code=404, detail="User not found")

    if not user.preferences:
        await ensure_user_preferences(user_id)
        user = await find_user_with_preferences(user_id)

    effective_theme = _resolve_theme(user)
    can_edit_tenant = _user_has_admin_role(user)
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

    user = await update_user_preferences_data(user_id, theme, timezone)

    if not user:
        raise HTTPException(status_code=404, detail="User not found")

    user_theme = user.preferences.theme
    user_timezone = user.preferences.timezone or "America/New_York"
    effective_theme = _resolve_theme(user)

    return user_theme, user_timezone, effective_theme


async def get_tenant_preferences(user_id: int):
    """Get tenant preferences (Admin/SuperAdmin only)."""
    user = await find_user_with_tenant_and_roles(user_id)

    if not user:
        raise HTTPException(status_code=404, detail="User not found")

    if not _user_has_admin_role(user):
        raise HTTPException(status_code=403, detail="Insufficient permissions")

    if not user.tenant.preferences:
        await ensure_tenant_preferences(user.tenant_id)
        user = await find_user_with_tenant_and_roles(user_id)

    return user.tenant.preferences.theme


async def update_tenant_preferences(user_id: int, theme: str):
    """Update tenant theme preference (Admin/SuperAdmin only)."""
    if theme not in ["light", "dark"]:
        raise HTTPException(status_code=400, detail="Invalid theme value")

    user = await find_user_with_tenant_and_roles(user_id)

    if not user:
        raise HTTPException(status_code=404, detail="User not found")

    if not _user_has_admin_role(user):
        raise HTTPException(status_code=403, detail="Insufficient permissions")

    prefs = await update_tenant_preferences_data(user.tenant_id, theme)
    return prefs.theme
