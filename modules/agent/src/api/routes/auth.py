"""Routes for authentication API endpoints."""

from typing import Optional
from fastapi import APIRouter, Request, HTTPException
from pydantic import BaseModel
from lib.auth import require_auth
from services.auth_service import get_user_tenants, create_tenant_for_user, add_user_to_tenant
from models.User import orm_to_dict

router = APIRouter(prefix="/api/auth", tags=["auth"])


class TenantCreate(BaseModel):
    """Model for tenant creation."""
    name: str


class TenantSwitch(BaseModel):
    """Model for tenant switching."""
    tenant_id: int


@router.get("/me")
@require_auth
async def get_current_user(request: Request):
    """Get current user profile with tenant associations."""
    user = request.state.user
    tenants = get_user_tenants(user.id)

    return {
        "user": orm_to_dict(user),
        "tenants": tenants,
        "current_tenant_id": request.state.tenant_id
    }


@router.post("/tenant/create")
@require_auth
async def create_tenant(request: Request, tenant_data: TenantCreate):
    """Create new tenant and assign current user as owner."""
    user_id = request.state.user_id
    tenant = create_tenant_for_user(user_id, tenant_data.name)
    return {"tenant": tenant}


@router.post("/tenant/switch")
@require_auth
async def switch_tenant(request: Request, tenant_data: TenantSwitch):
    """Switch active tenant for current user."""
    user_id = request.state.user_id
    tenant_id = tenant_data.tenant_id

    # Verify user has access to tenant
    tenants = get_user_tenants(user_id)
    tenant_ids = [t["id"] for t in tenants]

    if tenant_id not in tenant_ids:
        raise HTTPException(
            status_code=403,
            detail="User does not have access to this tenant"
        )

    # Return success - frontend will update cookie
    return {
        "success": True,
        "tenant_id": tenant_id
    }
