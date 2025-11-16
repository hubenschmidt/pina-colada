"""Routes for user API endpoints."""

from fastapi import APIRouter, HTTPException
from repositories.user_repository import find_user_by_email
from lib.db import get_session
from models.Tenant import Tenant

router = APIRouter(prefix="/users", tags=["user"])


@router.get("/{email}/tenant")
async def get_user_tenant(email: str):
    """
    Check if a user has an associated tenant.
    Returns tenant info if found, 404 if not.
    """
    # Find user by email
    user = find_user_by_email(email)

    if not user:
        raise HTTPException(status_code=404, detail="User not found")

    # Check if user has a tenant
    if not user.tenant_id:
        raise HTTPException(status_code=404, detail="No tenant associated with user")

    # Get tenant details
    session = get_session()
    try:
        tenant = session.get(Tenant, user.tenant_id)
        if not tenant:
            raise HTTPException(status_code=404, detail="Tenant not found")

        return {
            "id": tenant.id,
            "name": tenant.name,
            "created_at": tenant.created_at.isoformat() if tenant.created_at else None,
        }
    finally:
        session.close()
