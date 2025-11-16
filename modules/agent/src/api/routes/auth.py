"""Routes for authentication API endpoints."""

from typing import Optional
from fastapi import APIRouter, Request
from pydantic import BaseModel
from lib.auth import require_auth
from lib.error_logging import log_errors
from services.auth_service import get_user_tenants, create_tenant_for_user
from lib.serialization import model_to_dict

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
    user = request.state.user
    tenants = get_user_tenants(user.id)

    return {
        "user": model_to_dict(user, include_relationships=False),
        "tenants": tenants,
        "current_tenant_id": request.state.tenant_id
    }


@router.post("/tenant/create")
@log_errors
@require_auth
async def create_tenant(request: Request, tenant_data: TenantCreate):
    """Create new tenant and assign current user as owner."""
    user_id = request.state.user_id
    tenant = create_tenant_for_user(user_id, tenant_data.name, tenant_data.slug, tenant_data.plan)
    return {"tenant": tenant}
