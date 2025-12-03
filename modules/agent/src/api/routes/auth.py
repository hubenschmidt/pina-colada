"""Routes for authentication API endpoints."""

from fastapi import APIRouter, Request

from controllers.auth_controller import (
    get_current_user_with_tenants,
    create_tenant as create_tenant_controller,
    TenantCreate,
)
from lib.auth import require_auth
from lib.error_logging import log_errors


router = APIRouter(prefix="/auth", tags=["auth"])


@router.get("/me")
@log_errors
@require_auth
async def get_current_user(request: Request):
    """Get current user profile with tenant associations."""
    return await get_current_user_with_tenants(request)


@router.post("/tenant/create")
@log_errors
@require_auth
async def create_tenant(request: Request, tenant_data: TenantCreate):
    """Create new tenant and assign current user as owner."""
    return await create_tenant_controller(request, tenant_data)
