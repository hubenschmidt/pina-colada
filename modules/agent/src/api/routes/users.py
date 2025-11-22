"""Routes for user API endpoints."""

from fastapi import APIRouter, HTTPException
from repositories.user_repository import find_user_by_email
from lib.db import async_get_session
from models.Tenant import Tenant
from models.Individual import Individual

router = APIRouter(prefix="/users", tags=["user"])


@router.get("/{email}/tenant")
async def get_user_tenant(email: str):
    """
    Check if a user has an associated tenant.
    Returns tenant info if found, 404 if not.
    """
    # Find user by email
    user = await find_user_by_email(email)

    if not user:
        raise HTTPException(status_code=404, detail="User not found")

    # Check if user has a tenant
    if not user.tenant_id:
        raise HTTPException(status_code=404, detail="No tenant associated with user")

    # Get tenant and individual details
    async with async_get_session() as session:
        tenant = await session.get(Tenant, user.tenant_id)
        if not tenant:
            raise HTTPException(status_code=404, detail="Tenant not found")

        individual = None
        if user.individual_id:
            individual = await session.get(Individual, user.individual_id)

        return {
            "id": tenant.id,
            "tenant": {
                "name": tenant.name,
                "slug": tenant.slug,
                "plan": tenant.plan,
                "industry": tenant.industry,
                "website": tenant.website,
                "employee_count": tenant.employee_count,
            },
            "individual": {
                "first_name": individual.first_name if individual else "",
                "last_name": individual.last_name if individual else "",
                "phone": individual.phone if individual else "",
                "linkedin_url": individual.linkedin_url if individual else "",
                "title": individual.title if individual else "",
            } if individual else None,
            "created_at": tenant.created_at.isoformat() if tenant.created_at else None,
        }
