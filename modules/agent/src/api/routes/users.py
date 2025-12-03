"""Routes for user API endpoints."""

from fastapi import APIRouter

from controllers.user_controller import get_user_tenant


router = APIRouter(prefix="/users", tags=["user"])


@router.get("/{email}/tenant")
async def get_user_tenant_route(email: str):
    """Get tenant info for a user by email."""
    return await get_user_tenant(email)
