"""Routes for individuals API endpoints."""

from fastapi import APIRouter, Request
from lib.auth import require_auth
from lib.error_logging import log_errors
from repositories.individual_repository import find_all_individuals

router = APIRouter(prefix="/individuals", tags=["individuals"])


@router.get("")
@log_errors
@require_auth
async def get_individuals_route(request: Request):
    """Get all individuals for the current tenant."""
    tenant_id = getattr(request.state, "tenant_id", None)
    individuals = await find_all_individuals(tenant_id=tenant_id)
    return [
        {
            "id": ind.id,
            "first_name": ind.first_name,
            "last_name": ind.last_name,
            "email": ind.email,
            "phone": ind.phone,
            "linkedin_url": ind.linkedin_url,
            "title": ind.title,
            "notes": ind.notes,
            "created_at": ind.created_at.isoformat() if ind.created_at else None,
            "updated_at": ind.updated_at.isoformat() if ind.updated_at else None,
        }
        for ind in individuals
    ]
