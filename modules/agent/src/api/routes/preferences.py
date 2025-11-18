"""Routes for user and tenant preferences API endpoints."""

from fastapi import APIRouter, HTTPException, Request
from pydantic import BaseModel
from sqlalchemy import select
from sqlalchemy.orm import selectinload
from lib.auth import require_auth
from lib.db import async_get_session
from lib.error_logging import log_errors
from models.User import User
from models.UserPreferences import UserPreferences
from models.TenantPreferences import TenantPreferences
from models.Tenant import Tenant

router = APIRouter(prefix="/preferences", tags=["preferences"])


class UpdateUserPreferencesRequest(BaseModel):
    """Model for updating user preferences."""
    theme: str | None = None


class UpdateTenantPreferencesRequest(BaseModel):
    """Model for updating tenant preferences."""
    theme: str


def resolve_theme(user: User) -> str:
    """Resolve effective theme using hierarchy: user > tenant > system default."""
    if user.preferences and user.preferences.theme:
        return user.preferences.theme
    if user.tenant.preferences and user.tenant.preferences.theme:
        return user.tenant.preferences.theme
    return "light"


def user_has_admin_role(user: User) -> bool:
    """Check if user has Admin or SuperAdmin role."""
    if not user.user_roles:
        return False
    role_names = [ur.role.name for ur in user.user_roles]
    return "Admin" in role_names or "SuperAdmin" in role_names


@router.get("/user/{user_id}")
@log_errors
@require_auth
async def get_user_preferences(request: Request, user_id: int):
    """Get user preferences by user ID with resolved theme."""
    async with async_get_session() as session:
        result = await session.execute(
            select(User)
            .options(
                selectinload(User.preferences),
                selectinload(User.tenant).selectinload(Tenant.preferences),
                selectinload(User.user_roles)
            )
            .filter(User.id == user_id)
        )
        user = result.scalar_one_or_none()

        if not user:
            raise HTTPException(status_code=404, detail="User not found")

        # Ensure preferences exist
        if not user.preferences:
            user_prefs = UserPreferences(user_id=user.id, theme=None)
            session.add(user_prefs)
            await session.commit()
            await session.refresh(user, ["preferences"])

        effective_theme = resolve_theme(user)
        can_edit_tenant = user_has_admin_role(user)

        return {
            "theme": user.preferences.theme,
            "effective_theme": effective_theme,
            "can_edit_tenant": can_edit_tenant
        }


@router.get("/user")
@log_errors
@require_auth
async def get_current_user_preferences(request: Request):
    """Get current authenticated user's preferences with resolved theme."""
    user_id = request.state.user_id

    async with async_get_session() as session:
        result = await session.execute(
            select(User)
            .options(
                selectinload(User.preferences),
                selectinload(User.tenant).selectinload(Tenant.preferences),
                selectinload(User.user_roles)
            )
            .filter(User.id == user_id)
        )
        user = result.scalar_one_or_none()

        if not user:
            raise HTTPException(status_code=404, detail="User not found")

        # Ensure preferences exist
        if not user.preferences:
            user_prefs = UserPreferences(user_id=user.id, theme=None)
            session.add(user_prefs)
            await session.commit()
            await session.refresh(user, ["preferences"])

        effective_theme = resolve_theme(user)
        can_edit_tenant = user_has_admin_role(user)

        return {
            "theme": user.preferences.theme,
            "effective_theme": effective_theme,
            "can_edit_tenant": can_edit_tenant
        }


@router.patch("/user")
@log_errors
@require_auth
async def update_user_preferences(request: Request, body: UpdateUserPreferencesRequest):
    """Update current user's theme preference."""
    user_id = request.state.user_id
    theme = body.theme

    if theme not in ["light", "dark", None]:
        raise HTTPException(status_code=400, detail="Invalid theme value")

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

        user.preferences.theme = theme
        await session.commit()
        await session.refresh(user, ["preferences"])

        return {
            "theme": user.preferences.theme,
            "effective_theme": resolve_theme(user)
        }


@router.get("/tenant")
@log_errors
@require_auth
async def get_tenant_preferences(request: Request):
    """Get tenant preferences (Admin/SuperAdmin only)."""
    user_id = request.state.user_id

    async with async_get_session() as session:
        result = await session.execute(
            select(User)
            .options(
                selectinload(User.user_roles),
                selectinload(User.tenant).selectinload(Tenant.preferences)
            )
            .filter(User.id == user_id)
        )
        user = result.scalar_one_or_none()

        if not user:
            raise HTTPException(status_code=404, detail="User not found")

        if not user_has_admin_role(user):
            raise HTTPException(status_code=403, detail="Insufficient permissions")

        # Ensure tenant preferences exist
        if not user.tenant.preferences:
            tenant_prefs = TenantPreferences(tenant_id=user.tenant_id, theme="light")
            session.add(tenant_prefs)
            await session.commit()
            await session.refresh(user.tenant, ["preferences"])

        return {"theme": user.tenant.preferences.theme}


@router.patch("/tenant")
@log_errors
@require_auth
async def update_tenant_preferences(request: Request, body: UpdateTenantPreferencesRequest):
    """Update tenant theme preference (Admin/SuperAdmin only)."""
    user_id = request.state.user_id
    theme = body.theme

    if theme not in ["light", "dark"]:
        raise HTTPException(status_code=400, detail="Invalid theme value")

    async with async_get_session() as session:
        result = await session.execute(
            select(User)
            .options(
                selectinload(User.user_roles),
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

        return {"theme": user.tenant.preferences.theme}
