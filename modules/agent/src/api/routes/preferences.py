"""Routes for user and tenant preferences API endpoints."""

from fastapi import APIRouter, Request

from controllers.preferences_controller import (
    get_tenant_preferences,
    get_timezones,
    get_user_preferences,
    update_tenant_preferences,
    update_user_preferences,
)
from schemas.preferences import UpdateTenantPreferencesRequest, UpdateUserPreferencesRequest


router = APIRouter(prefix="/preferences", tags=["preferences"])


@router.get("/timezones")
async def get_timezones_route():
    """Get list of common timezones for dropdown selection."""
    return get_timezones()


@router.get("/user")
async def get_user_preferences_route(request: Request):
    """Get current user's preferences with resolved theme."""
    return await get_user_preferences(request)


@router.patch("/user")
async def update_user_preferences_route(request: Request, body: UpdateUserPreferencesRequest):
    """Update current user's preferences (theme and/or timezone)."""
    return await update_user_preferences(request, body)


@router.get("/tenant")
async def get_tenant_preferences_route(request: Request):
    """Get tenant preferences (Admin/SuperAdmin only)."""
    return await get_tenant_preferences(request)


@router.patch("/tenant")
async def update_tenant_preferences_route(request: Request, body: UpdateTenantPreferencesRequest):
    """Update tenant theme preference (Admin/SuperAdmin only)."""
    return await update_tenant_preferences(request, body)
