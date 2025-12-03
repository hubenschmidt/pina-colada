"""Controller layer for preferences routing to services."""

import logging
from typing import Optional

from fastapi import Request

from lib.decorators import handle_http_exceptions
from lib.date_utils import get_common_timezones
from repositories.preferences_repository import (
    UpdateUserPreferencesRequest,
    UpdateTenantPreferencesRequest,
)
from services.preferences_service import (
    get_user_preferences as get_user_prefs_service,
    update_user_preferences as update_user_prefs_service,
    get_tenant_preferences as get_tenant_prefs_service,
    update_tenant_preferences as update_tenant_prefs_service,
)

# Re-export for routes
__all__ = ["UpdateUserPreferencesRequest", "UpdateTenantPreferencesRequest"]

logger = logging.getLogger(__name__)


def get_timezones() -> list:
    """Get list of common timezones for dropdown selection."""
    return get_common_timezones()


@handle_http_exceptions
async def get_user_preferences(request: Request) -> dict:
    """Get current user's preferences with resolved theme."""
    user_id = request.state.user_id
    user_theme, user_timezone, effective_theme, can_edit_tenant = await get_user_prefs_service(user_id)
    return {
        "theme": user_theme,
        "timezone": user_timezone,
        "effective_theme": effective_theme,
        "can_edit_tenant": can_edit_tenant,
    }


@handle_http_exceptions
async def update_user_preferences(request: Request, data: UpdateUserPreferencesRequest) -> dict:
    """Update current user's preferences."""
    user_id = request.state.user_id
    user_theme, user_timezone, effective_theme = await update_user_prefs_service(user_id, data.theme, data.timezone)
    return {
        "theme": user_theme,
        "timezone": user_timezone,
        "effective_theme": effective_theme,
    }


@handle_http_exceptions
async def get_tenant_preferences(request: Request) -> dict:
    """Get tenant preferences (Admin/SuperAdmin only)."""
    user_id = request.state.user_id
    tenant_theme = await get_tenant_prefs_service(user_id)
    return {"theme": tenant_theme}


@handle_http_exceptions
async def update_tenant_preferences(request: Request, data: UpdateTenantPreferencesRequest) -> dict:
    """Update tenant theme preference (Admin/SuperAdmin only)."""
    user_id = request.state.user_id
    tenant_theme = await update_tenant_prefs_service(user_id, data.theme)
    return {"theme": tenant_theme}
