"""Routes for authentication API endpoints."""

from typing import Optional
from fastapi import APIRouter, Request
from pydantic import BaseModel
from lib.auth import require_auth
from lib.error_logging import log_errors
from controllers.auth_controller import (
    get_current_user_with_tenants,
    create_tenant as create_tenant_controller,
)

router = APIRouter(prefix="/auth", tags=["auth"])


class TenantCreate(BaseModel):
    """Model for tenant creation."""
    name: str
    slug: Optional[str] = None
    plan: str = "free"


@router.get("/me")
@log_errors
@require_auth
async def get_current_user(request: Request):
    """Get current user profile with tenant associations."""
    return await get_current_user_with_tenants(request.state.user, request.state.tenant_id)


@router.post("/tenant/create")
@log_errors
@require_auth
async def create_tenant(request: Request, tenant_data: TenantCreate):
    """Create new tenant and assign current user as owner."""
    return await create_tenant_controller(request.state.user_id, tenant_data.name, tenant_data.slug, tenant_data.plan)
