"""Routes for authentication API endpoints."""

from fastapi import APIRouter, Request

from controllers.auth_controller import (
    create_tenant as create_tenant_controller,
    get_current_user_with_tenants,
)
from schemas.tenant import TenantCreate


router = APIRouter(prefix="/auth", tags=["auth"])


@router.get("/me")
async def get_current_user(request: Request):
    """Get current user profile with tenant associations."""
    return await get_current_user_with_tenants(request)


@router.post("/tenant/create")
async def create_tenant(request: Request, tenant_data: TenantCreate):
    """Create new tenant and assign current user as owner."""
    return await create_tenant_controller(request, tenant_data)
